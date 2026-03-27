/**
 * API Service - Domain-Specific API Bindings
 * See FRONTEND_SCOPE.md Section 5.2
 *
 * Provides typed API methods organized by domain:
 * - auth: Authentication (Clerk-managed, backend principal lookup)
 * - projects: Project CRUD operations
 * - chat: Chat messaging
 * - schedule: Gantt/Schedule data
 * - tasks: Task management
 *
 * IMPORTANT: Uses types from src/types/ for all request/response bodies.
 * Does NOT import from store to prevent circular dependencies.
 */

import { get, post, put, del, getAuthToken } from './http';
import type { CorrectionEvent } from '../types/corrections';
import type { GanttData, Contact, CreateContactRequest, CreateContactResponse, BulkCreateContactsResponse, AssignmentRow, InvoiceExtraction, Thread as ApiThread, CompletionReport } from '../types/models';
import type { InvoiceStatus } from '../types/enums';
import type { UserRole } from '../types/enums';
import type { PortfolioFeedResponse, ActionResponse } from '../types/feed';
import type { MaterialEstimate, BudgetEstimate, ProjectMaterial, ProjectBudget, CreateMaterialRequest, MaterialUpdateRequest, BudgetSeedRequest } from '../types/materials';
import type { SchedulePreviewRequest, SchedulePreviewResponse, ScenarioComparisonRequest, ScenarioComparisonResponse } from '../types/schedule';
import type { CorporateBudget, GLSyncLog, ARAgingSnapshot } from '../types/corporate';
import type { Employee, TimeLog, Certification } from '../types/employee';
import type { FleetAsset, EquipmentAllocation, MaintenanceLog } from '../types/fleet';
import type { A2AExecutionLog, ActiveAgentConnection } from '../types/a2a';

// ============================================================================
// Auth Types
// ============================================================================

/**
 * Authenticated principal data (matches Go types.Principal).
 */
export interface AuthPrincipal {
    id: string;
    email: string;
    name: string;
    role: UserRole;
    org_id: string;
    subject_type: string;
    created_at: string;
}

// ============================================================================
// Project Types
// ============================================================================

/**
 * Project summary for list views.
 */
export interface Project {
    id: string;
    name: string;
    address: string;
    status: string;
    completion_percentage: number;
    created_at: string;
    updated_at: string;
}

/**
 * Project detail with full data.
 */
export interface ProjectDetail extends Project {
    org_id: string;
    square_footage: number;
    bedrooms: number;
    bathrooms: number;
    lot_size: number;
    foundation_type: string;
    start_date: string;
    projected_end_date: string | null;
}

/**
 * Create project request.
 */
export interface CreateProjectRequest {
    name: string;
    address: string;
    square_footage: number;
    bedrooms: number;
    bathrooms: number;
    lot_size?: number;
    foundation_type?: string;
    start_date: string;
    stories?: number;
    topography?: string;
    soil_conditions?: string;
}

// ============================================================================
// Onboarding Types
// ============================================================================

/**
 * Request to the Interrogator Agent for onboarding.
 * For file uploads, use FormData with multipart/form-data instead of this interface.
 */
export interface OnboardProcessRequest {
    session_id: string;
    message: string;
    document_url?: string;
    current_state: Partial<CreateProjectRequest>;
}

/**
 * Long-lead item that affects schedule generation.
 * Items with significant lead times need early procurement.
 */
export interface LongLeadItem {
    name: string;
    brand?: string;
    model?: string;
    category: 'windows' | 'doors' | 'hvac' | 'appliances' | 'millwork' | 'finishes';
    estimated_lead_weeks: number;
    wbs_code?: string;
    notes?: string;
}

/**
 * Response from the Interrogator Agent.
 * Must match Go models.OnboardResponse exactly (Rosetta Stone).
 */
export interface OnboardProcessResponse {
    session_id: string;
    reply: string;
    extracted_values?: Record<string, unknown>;
    confidence_scores?: Record<string, number>;
    long_lead_items?: LongLeadItem[];
    clarifying_question?: string;
    ready_to_create: boolean;
    next_priority_field?: string;
    /** Interrogator gate status. Controls when schedule generation is allowed. */
    status?: 'gathering' | 'clarifying' | 'satisfied' | 'error';
    /** Material estimates extracted from blueprint or computed from project attributes. */
    material_estimates?: MaterialEstimate[];
    /** Budget estimate computed from material estimates. */
    budget_estimate?: BudgetEstimate;
}

/**
 * Confidence data for extracted fields.
 */
export interface ConfidenceReport {
    overall_confidence: number;
    field_confidences: Record<string, number>;
    warnings: string[];
    suggested_questions: string[];
}

/**
 * Response from the VisionExtractionResponse endpoint.
 */
export interface VisionExtractionResponse {
    extracted_values: Record<string, unknown>;
    confidence_report: ConfidenceReport;
    long_lead_items?: LongLeadItem[];
    raw_text?: string;
}

// ============================================================================
// Chat Types
// ============================================================================

/**
 * Chat message request.
 */
export interface ChatRequest {
    project_id: string;
    thread_id: string;
    message: string;
}

/**
 * Chat message response.
 */
export interface ChatResponse {
    id: string;
    role: 'user' | 'assistant';
    content: string;
    artifacts?: ChatArtifact[];
    created_at: string;
}

/**
 * Chat artifact (inline data visualization).
 */
export interface ChatArtifact {
    type: 'invoice' | 'budget' | 'gantt' | 'table';
    data: InvoiceExtraction | GanttData | Record<string, unknown>;
}

/**
 * A single chunk from a streaming chat response.
 */
export interface ChatStreamChunk {
    text?: string;
    tool_use?: { id: string; name: string; input: Record<string, unknown> };
    tool_result?: { tool_use_id: string; content: string; is_error: boolean };
    done?: boolean;
    stop_reason?: string;
    thread_id?: string;
    error?: string;
}

/**
 * Streaming chat connection using fetch + ReadableStream for SSE.
 * Provides an async-iterable interface for consuming chunks.
 */
export class ChatStream {
    private controller: AbortController;
    private _onChunk: ((chunk: ChatStreamChunk) => void) | null = null;
    private _onError: ((err: Error) => void) | null = null;
    private _onDone: (() => void) | null = null;
    private started = false;

    constructor(
        private projectId: string,
        private threadId: string,
        private message: string,
    ) {
        this.controller = new AbortController();
    }

    /** Register a callback for each text/tool chunk. */
    onChunk(cb: (chunk: ChatStreamChunk) => void): this {
        this._onChunk = cb;
        return this;
    }

    /** Register a callback for errors. */
    onError(cb: (err: Error) => void): this {
        this._onError = cb;
        return this;
    }

    /** Register a callback for stream completion. */
    onDone(cb: () => void): this {
        this._onDone = cb;
        return this;
    }

    /** Start the streaming connection. */
    async start(): Promise<void> {
        if (this.started) return;
        this.started = true;

        const token = getAuthToken();
        const headers: Record<string, string> = {
            'Content-Type': 'application/json',
        };
        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }

        try {
            const resp = await fetch(
                `/api/v1/projects/${this.projectId}/chat/stream`,
                {
                    method: 'POST',
                    headers,
                    body: JSON.stringify({
                        thread_id: this.threadId,
                        message: this.message,
                    }),
                    signal: this.controller.signal,
                },
            );

            if (!resp.ok) {
                throw new Error(`Stream failed: ${String(resp.status)}`);
            }

            const reader = resp.body?.getReader();
            if (!reader) throw new Error('No response body');

            const decoder = new TextDecoder();
            let buffer = '';

            // eslint-disable-next-line no-constant-condition
            while (true) {
                const { done, value } = await reader.read();
                if (done) {
                    // Flush any residual bytes from the streaming decoder
                    buffer += decoder.decode();
                    break;
                }

                buffer += decoder.decode(value, { stream: true });

                // Parse SSE events from buffer
                const events = buffer.split('\n\n');
                buffer = events.pop() ?? '';

                for (const event of events) {
                    if (!event.trim()) continue;

                    // Concatenate all data: lines (SSE spec allows multi-line data)
                    const dataLines: string[] = [];
                    for (const line of event.split('\n')) {
                        if (line.startsWith('data:')) {
                            dataLines.push(line.slice(5).trim());
                        }
                    }

                    const data = dataLines.join('\n');
                    if (!data) continue;

                    try {
                        const chunk = JSON.parse(data) as ChatStreamChunk;
                        this._onChunk?.(chunk);

                        if (chunk.done) {
                            this._onDone?.();
                            return;
                        }
                    } catch {
                        // Skip unparseable events (comments, keepalives)
                    }
                }
            }

            this._onDone?.();
        } catch (err) {
            if ((err as Error).name !== 'AbortError') {
                this._onError?.(err as Error);
            }
        }
    }

    /** Abort the stream. */
    abort(): void {
        this.controller.abort();
    }
}

// ============================================================================
// Task Types
// ============================================================================

/**
 * Project task.
 */
export interface Task {
    id: string;
    project_id: string;
    wbs_code: string;
    name: string;
    status: string;
    duration_days: number;
    early_start: string;
    early_finish: string;
    is_critical: boolean;
}

/**
 * Task progress update request.
 */
export interface TaskProgressRequest {
    status: string;
    completion_percentage?: number;
    notes?: string;
}

// ============================================================================
// API Namespace
// ============================================================================

/**
 * Centralized API client with domain-specific methods.
 * All methods use strict typing - no 'any' types.
 *
 * @example
 * ```typescript
 * // Authentication (sign-in handled by Clerk SDK)
 * const me = await api.auth.me();
 *
 * // Projects
 * const projects = await api.projects.list();
 * const project = await api.projects.get('project-uuid');
 *
 * // Chat
 * const response = await api.chat.send('project-uuid', 'What tasks are due?');
 * ```
 */
export const api = {
    /**
     * Authentication endpoints.
     * Note: Sign-in/sign-out are handled by Clerk SDK directly.
     * See STEP_78_AUTH_PROVIDER.md Section 1.
     */
    auth: {
        /**
         * Get the current authenticated user's principal from the backend.
         * Validates that the Clerk JWT is accepted by the backend.
         */
        me(): Promise<AuthPrincipal> {
            return get<AuthPrincipal>('/auth/me');
        },
    },

    /**
     * Project endpoints.
     */
    projects: {
        /**
         * List all projects for the current user's organization.
         */
        list(): Promise<Project[]> {
            return get<Project[]>('/projects');
        },

        /**
         * Get a single project by ID.
         * @param id - Project UUID
         */
        get(id: string): Promise<ProjectDetail> {
            return get<ProjectDetail>(`/projects/${id}`);
        },

        /**
         * Create a new project.
         * @param data - Project creation data
         */
        create(data: CreateProjectRequest): Promise<ProjectDetail> {
            return post<ProjectDetail>('/projects', data);
        },

        /**
         * Update an existing project.
         * @param id - Project UUID
         * @param data - Partial project data to update
         */
        update(
            id: string,
            data: Partial<CreateProjectRequest>
        ): Promise<ProjectDetail> {
            return put<ProjectDetail>(`/projects/${id}`, data);
        },

        /**
         * Delete a project.
         * @param id - Project UUID
         */
        async delete(id: string): Promise<void> {
            await del<undefined>(`/projects/${id}`);
        },
    },

    /**
     * Thread endpoints.
     */
    threads: {
        /**
         * List threads for a project.
         * @param projectId - Project UUID
         * @param includeArchived - Whether to include archived threads
         */
        list(projectId: string, includeArchived = false): Promise<ApiThread[]> {
            const query = includeArchived ? '?archived=true' : '';
            return get<ApiThread[]>(`/projects/${projectId}/threads${query}`);
        },

        /**
         * Create a new thread in a project.
         * @param projectId - Project UUID
         * @param title - Thread title
         */
        create(projectId: string, title: string): Promise<ApiThread> {
            return post<ApiThread>(`/projects/${projectId}/threads`, { title });
        },

        /**
         * Get a single thread.
         * @param projectId - Project UUID
         * @param threadId - Thread UUID
         */
        get(projectId: string, threadId: string): Promise<ApiThread> {
            return get<ApiThread>(`/projects/${projectId}/threads/${threadId}`);
        },

        /**
         * Archive a thread (soft delete).
         * @param projectId - Project UUID
         * @param threadId - Thread UUID
         */
        async archive(projectId: string, threadId: string): Promise<void> {
            await post<undefined>(`/projects/${projectId}/threads/${threadId}/archive`);
        },

        /**
         * Unarchive a thread.
         * @param projectId - Project UUID
         * @param threadId - Thread UUID
         */
        async unarchive(projectId: string, threadId: string): Promise<void> {
            await post<undefined>(`/projects/${projectId}/threads/${threadId}/unarchive`);
        },
    },

    /**
     * Chat endpoints.
     */
    chat: {
        /**
         * Send a chat message and get a response.
         * @param projectId - Project UUID context
         * @param threadId - Thread UUID context
         * @param message - User message text
         */
        send(projectId: string, threadId: string, message: string): Promise<ChatResponse> {
            return post<ChatResponse>('/chat', {
                project_id: projectId,
                thread_id: threadId,
                message,
            } satisfies ChatRequest);
        },

        /**
         * Get chat history for a thread.
         * @param projectId - Project UUID
         * @param threadId - Thread UUID
         * @param limit - Maximum messages to return (default 50)
         */
        history(projectId: string, threadId: string, limit = 50): Promise<ChatResponse[]> {
            return get<ChatResponse[]>(
                `/projects/${projectId}/threads/${threadId}/messages?limit=${String(limit)}`
            );
        },

        /**
         * Stream a chat response via SSE.
         * Returns an object with an EventSource-like interface for consuming
         * text deltas, tool events, and completion signals.
         */
        stream(projectId: string, threadId: string, message: string): ChatStream {
            return new ChatStream(projectId, threadId, message);
        },
    },

    /**
     * Schedule/Gantt endpoints.
     */
    schedule: {
        /**
         * Get the computed schedule for a project.
         * @param projectId - Project UUID
         */
        get(projectId: string): Promise<GanttData> {
            return get<GanttData>(`/projects/${projectId}/schedule`);
        },

        /**
         * Recalculate the schedule.
         * @param projectId - Project UUID
         */
        recalculate(projectId: string): Promise<GanttData> {
            return post<GanttData>(`/projects/${projectId}/schedule/recalculate`);
        },

        /**
         * Generate a new schedule from Interrogator extraction data.
         * Triggers CPM calculation on the backend.
         * @param projectId - Project UUID
         */
        generate(projectId: string): Promise<GanttData> {
            return post<GanttData>(`/projects/${projectId}/schedule/generate`);
        },

        /**
         * Generate an instant schedule preview from onboarding data.
         * No project required — runs physics pipeline in-memory.
         */
        preview(req: SchedulePreviewRequest): Promise<SchedulePreviewResponse> {
            return post<SchedulePreviewResponse>('/schedule/preview', req);
        },

        /**
         * Compare multiple schedule scenarios (what-if analysis).
         */
        compare(req: ScenarioComparisonRequest): Promise<ScenarioComparisonResponse> {
            return post<ScenarioComparisonResponse>('/schedule/compare', req);
        },
    },

    /**
     * Task endpoints.
     */
    tasks: {
        /**
         * List tasks for a project.
         * @param projectId - Project UUID
         */
        list(projectId: string): Promise<Task[]> {
            return get<Task[]>(`/projects/${projectId}/tasks`);
        },

        /**
         * Get a single task.
         * @param projectId - Project UUID
         * @param taskId - Task UUID
         */
        get(projectId: string, taskId: string): Promise<Task> {
            return get<Task>(`/projects/${projectId}/tasks/${taskId}`);
        },

        /**
         * Update task progress.
         * @param projectId - Project UUID
         * @param taskId - Task UUID
         * @param progress - Progress update data
         */
        updateProgress(
            projectId: string,
            taskId: string,
            progress: TaskProgressRequest
        ): Promise<Task> {
            return post<Task>(
                `/projects/${projectId}/tasks/${taskId}/progress`,
                progress
            );
        },
    },

    /**
     * Contact/Directory endpoints.
     * See FRONTEND_V2_SPEC.md §10.3
     */
    contacts: {
        /** List all org contacts. Optional search filter. */
        list(search?: string): Promise<Contact[]> {
            const params = search ? `?search=${encodeURIComponent(search)}` : '';
            return get<Contact[]>(`/contacts${params}`);
        },

        /** Get a single contact by ID. */
        get(contactId: string): Promise<Contact> {
            return get<Contact>(`/contacts/${contactId}`);
        },

        /** Create a single contact with dedup check. */
        create(contact: CreateContactRequest): Promise<CreateContactResponse> {
            return post<CreateContactResponse>('/contacts', contact);
        },

        /** Bulk create contacts. */
        bulkCreate(contacts: CreateContactRequest[]): Promise<BulkCreateContactsResponse> {
            return post<BulkCreateContactsResponse>('/contacts/bulk', { contacts });
        },

        /** List phase assignments for a project (with assigned contacts). */
        listAssignments(projectId: string): Promise<AssignmentRow[]> {
            return get<AssignmentRow[]>(`/projects/${projectId}/assignments`);
        },

        /** Assign a contact to a project phase. */
        createAssignment(projectId: string, contactId: string, wbsPhaseId: string): Promise<{ success: boolean }> {
            return post<{ success: boolean }>(`/projects/${projectId}/assignments`, {
                contact_id: contactId,
                wbs_phase_id: wbsPhaseId,
            });
        },

        /** Bulk assign contacts to project phases. */
        bulkCreateAssignments(projectId: string, assignments: { contact_id: string; wbs_phase_id: string }[]): Promise<{ created: number }> {
            return post<{ created: number }>(`/projects/${projectId}/assignments/bulk`, { assignments });
        },
    },

    /**
     * User profile endpoints.
     * See LAUNCH_PLAN.md User Profile Endpoint (P0).
     */
    users: {
        /**
         * Get the current user's profile.
         */
        getMe(): Promise<UserProfile> {
            return get<UserProfile>('/users/me');
        },

        /**
         * Update the current user's profile.
         * @param data - Profile data to update
         */
        updateMe(data: UpdateProfileRequest): Promise<UserProfile> {
            return put<UserProfile>('/users/me', data);
        },
    },

    /**
     * Invite endpoints.
     * See LAUNCH_STRATEGY.md Task B2.
     */
    invites: {
        /**
         * Get invite info by token (public).
         * @param token - Invite token from email
         */
        getInfo(token: string): Promise<InviteInfo> {
            return get<InviteInfo>(`/invites/info?token=${encodeURIComponent(token)}`, { skipAuth: true });
        },

        /**
         * Accept an invitation and create an account (public).
         * @param token - Invite token from email
         * @param name - User's display name
         * @param password - Account password (min 8 chars)
         */
        accept(token: string, name: string, password: string): Promise<InviteAcceptResponse> {
            return post<InviteAcceptResponse>(
                '/invites/accept',
                { token, name, password },
                { skipAuth: true }
            );
        },

        /**
         * List pending invitations for the organization (admin only).
         */
        list(): Promise<Invite[]> {
            return get<Invite[]>('/admin/invites');
        },

        /**
         * Create a new invitation (admin only).
         * @param email - Email to invite
         * @param role - Role to assign
         */
        create(email: string, role: string): Promise<Invite> {
            return post<Invite>('/admin/invites', { email, role });
        },

        /**
         * Revoke a pending invitation (admin only).
         * @param id - Invite UUID
         */
        async revoke(id: string): Promise<void> {
            await del<undefined>(`/admin/invites/${id}`);
        },
    },
    /**
     * Team management endpoints.
     * Lists org members for the Team page.
     */
    team: {
        /**
         * List all members of the current organization (admin only).
         */
        listMembers(): Promise<UserProfile[]> {
            return get<UserProfile[]>('/org/members');
        },
    },

    /**
     * Portal endpoints.
     * See LAUNCH_PLAN.md P2: Field Portal (Mobile).
     */
    portal: {
        /**
         * Verify an action token and get context (public).
         * @param token - Action token from SMS link
         */
        verifyActionToken(token: string): Promise<PortalActionContext> {
            return get<PortalActionContext>(`/portal/action/${token}`, { skipAuth: true });
        },

        /**
         * Submit an action for a token (public).
         * @param token - Action token from SMS link
         * @param data - Action data (status or photo_id)
         */
        submitAction(token: string, data: PortalSubmitRequest): Promise<PortalSubmitResponse> {
            return post<PortalSubmitResponse>(`/portal/action/${token}`, data, { skipAuth: true });
        },

        /**
         * Create and send an action link (admin only).
         * @param contactId - Contact UUID
         * @param projectId - Project UUID
         * @param taskId - Task UUID
         * @param actionType - Type of action (status_update, photo_upload, view)
         */
        createActionLink(
            contactId: string,
            projectId: string,
            taskId: string,
            actionType: 'status_update' | 'photo_upload' | 'view'
        ): Promise<PortalLinkResponse> {
            return post<PortalLinkResponse>('/admin/portal/link', {
                contact_id: contactId,
                project_id: projectId,
                task_id: taskId,
                action_type: actionType,
            });
        },
    },

    /**
     * Onboarding endpoints (Interrogator Agent).
     * See STEP_75/STEP_76 specs.
     */
    onboard: {
        /**
         * Send a message to the Interrogator Agent for field extraction.
         * @param data - Session ID, message, and current form state
         */
        process(data: OnboardProcessRequest): Promise<OnboardProcessResponse> {
            return post<OnboardProcessResponse>('/agent/onboard', data);
        },
    },

    /**
     * Vision analysis endpoints.
     * See STEP_84_FIELD_FEEDBACK.md Section 2
     */
    vision: {
        /**
         * Get the analysis status of a project asset.
         * @param assetId - Project asset UUID
         */
        status(assetId: string): Promise<VisionStatusResponse> {
            return get<VisionStatusResponse>(`/vision/status/${assetId}`);
        },

        /**
         * Extract data from a document (e.g. plan set or image).
         * @param file - File to extract data from
         */
        async extract(file: File): Promise<VisionExtractionResponse> {
            const formData = new FormData();
            formData.append('file', file);

            // Manual fetch since http wrapper doesn't support FormData directly
            const headers: Record<string, string> = {};
            const token = getAuthToken();
            if (token) {
                headers['Authorization'] = `Bearer ${token}`;
            }

            const response = await fetch('/api/v1/vision/extract', {
                method: 'POST',
                headers,
                body: formData,
            });

            if (!response.ok) {
                const errorBody = await response.json().catch(() => ({}));
                throw new Error(errorBody.error || `HTTP ${response.status}`);
            }

            return response.json() as Promise<VisionExtractionResponse>;
        },
    },

    /**
     * Project asset endpoints.
     * See STEP_85_VISION_BADGES.md Section 2
     */
    assets: {
        /**
         * List all assets for a project with analysis status.
         * @param projectId - Project UUID
         */
        list(projectId: string): Promise<ProjectAssetResponse[]> {
            return get<ProjectAssetResponse[]>(`/projects/${projectId}/assets`);
        },
    },

    /**
     * Organization settings endpoints.
     * See STEP_87_CONFIG_PERSISTENCE.md Section 2
     */
    settings: {
        /**
         * Get the current org's physics configuration.
         */
        getPhysics(): Promise<PhysicsConfigResponse> {
            return get<PhysicsConfigResponse>('/org/settings/physics');
        },

        /**
         * Update the current org's physics configuration.
         * @param data - Physics config to save
         */
        updatePhysics(data: UpdatePhysicsRequest): Promise<PhysicsConfigResponse> {
            return put<PhysicsConfigResponse>('/org/settings/physics', data);
        },

        /** Get agent configuration for the org. */
        getAgents(): Promise<AgentSettingsResponse> {
            return get<AgentSettingsResponse>('/org/settings/agents');
        },

        /** Update agent configuration for the org. */
        updateAgents(data: AgentSettingsResponse): Promise<AgentSettingsResponse> {
            return put<AgentSettingsResponse>('/org/settings/agents', data);
        },

        /** Get FB-Brain connection settings. */
        getBrain(): Promise<BrainConnectionResponse> {
            return get<BrainConnectionResponse>('/org/settings/brain');
        },

        /** Update FB-Brain connection URL. */
        updateBrain(data: { brain_url: string }): Promise<BrainConnectionResponse> {
            return put<BrainConnectionResponse>('/org/settings/brain', data);
        },

        /** Regenerate FB-Brain integration key. */
        regenerateBrainKey(): Promise<{ integration_key: string }> {
            return post<{ integration_key: string }>('/org/settings/brain/regenerate-key');
        },

        /** Get active A2A agent connections. Phase 18: FRONTEND_SCOPE.md §15.1 */
        getBrainAgents(): Promise<ActiveAgentConnection[]> {
            return get<ActiveAgentConnection[]>('/org/settings/brain/agents');
        },

        /** Pause an A2A agent. */
        pauseBrainAgent(agentId: string): Promise<{ status: string }> {
            return post<{ status: string }>(`/org/settings/brain/agents/${agentId}/pause`);
        },

        /** Resume an A2A agent. */
        resumeBrainAgent(agentId: string): Promise<{ status: string }> {
            return post<{ status: string }>(`/org/settings/brain/agents/${agentId}/resume`);
        },

        /** Get A2A execution logs. */
        getBrainLogs(limit = 50): Promise<A2AExecutionLog[]> {
            return get<A2AExecutionLog[]>(`/org/settings/brain/logs?limit=${limit}`);
        },
    },

    /**
     * Corporate financials endpoints.
     * Phase 18: See BACKEND_SCOPE.md Section 20.1
     */
    corporate: {
        /** Get corporate budget for a fiscal year/quarter. */
        getBudget(year: number, quarter: number): Promise<CorporateBudget> {
            return get<CorporateBudget>(`/corporate/budgets?year=${year}&quarter=${quarter}`);
        },

        /** Trigger a budget rollup for a fiscal year/quarter. */
        rollupBudget(year: number, quarter: number): Promise<CorporateBudget> {
            return post<CorporateBudget>('/corporate/budgets/rollup', { year, quarter });
        },

        /** Get AR aging snapshot. */
        getARAging(): Promise<ARAgingSnapshot> {
            return get<ARAgingSnapshot>('/corporate/ar-aging');
        },

        /** Get GL sync logs. */
        getGLSyncLogs(): Promise<GLSyncLog[]> {
            return get<GLSyncLog[]>('/corporate/gl-sync');
        },

        /** Create a GL sync log entry. */
        createGLSync(syncType: string): Promise<GLSyncLog> {
            return post<GLSyncLog>('/corporate/gl-sync', { sync_type: syncType });
        },
    },

    /**
     * Employee management endpoints.
     * Phase 18: See BACKEND_SCOPE.md Section 20.2
     */
    employees: {
        /** List employees, optionally filtered by status. */
        list(status?: string): Promise<Employee[]> {
            const query = status ? `?status=${encodeURIComponent(status)}` : '';
            return get<Employee[]>(`/employees${query}`);
        },

        /** Create a new employee. */
        create(data: Partial<Employee>): Promise<Employee> {
            return post<Employee>('/employees', data);
        },

        /** Get a single employee. */
        get(id: string): Promise<Employee> {
            return get<Employee>(`/employees/${id}`);
        },

        /** Update an employee. */
        update(id: string, data: Partial<Employee>): Promise<Employee> {
            return put<Employee>(`/employees/${id}`, data);
        },

        /** Log time for an employee. */
        logTime(employeeId: string, data: Partial<TimeLog>): Promise<TimeLog> {
            return post<TimeLog>(`/employees/${employeeId}/time-logs`, data);
        },

        /** Get time logs for an employee. */
        getTimeLogs(employeeId: string): Promise<TimeLog[]> {
            return get<TimeLog[]>(`/employees/${employeeId}/time-logs`);
        },

        /** Add a certification for an employee. */
        addCertification(employeeId: string, data: Partial<Certification>): Promise<Certification> {
            return post<Certification>(`/employees/${employeeId}/certifications`, data);
        },

        /** List certifications for an employee. */
        getCertifications(employeeId: string): Promise<Certification[]> {
            return get<Certification[]>(`/employees/${employeeId}/certifications`);
        },

        /** Get expiring certifications across the org. */
        getExpiringCertifications(withinDays = 30): Promise<Certification[]> {
            return get<Certification[]>(`/certifications/expiring?within_days=${withinDays}`);
        },

        /** Approve a time log entry. */
        approveTimeLog(logId: string): Promise<{ status: string }> {
            return post<{ status: string }>(`/time-logs/${logId}/approve`);
        },

        /** Get deterministic labor burden for a project. */
        getLaborBurden(projectId: string): Promise<{ total_labor_cost_cents: number }> {
            return get<{ total_labor_cost_cents: number }>(`/projects/${projectId}/labor-burden`);
        },
    },

    /**
     * Fleet / equipment management endpoints.
     * Phase 18: See BACKEND_SCOPE.md Section 20.3
     */
    fleet: {
        /** List fleet assets, optionally filtered by status and type. */
        list(status?: string, assetType?: string): Promise<FleetAsset[]> {
            const params = new URLSearchParams();
            if (status) params.set('status', status);
            if (assetType) params.set('asset_type', assetType);
            const query = params.toString() ? `?${params.toString()}` : '';
            return get<FleetAsset[]>(`/fleet${query}`);
        },

        /** Create a new fleet asset. */
        create(data: Partial<FleetAsset>): Promise<FleetAsset> {
            return post<FleetAsset>('/fleet', data);
        },

        /** Get a single fleet asset. */
        get(id: string): Promise<FleetAsset> {
            return get<FleetAsset>(`/fleet/${id}`);
        },

        /** Update a fleet asset. */
        update(id: string, data: Partial<FleetAsset>): Promise<FleetAsset> {
            return put<FleetAsset>(`/fleet/${id}`, data);
        },

        /** Allocate equipment to a project. */
        allocate(assetId: string, data: Partial<EquipmentAllocation>): Promise<EquipmentAllocation> {
            return post<EquipmentAllocation>(`/fleet/${assetId}/allocate`, data);
        },

        /** Check equipment availability for a date range. */
        checkAvailability(assetId: string, from: string, to: string): Promise<{ available: boolean }> {
            return get<{ available: boolean }>(`/fleet/${assetId}/availability?from=${from}&to=${to}`);
        },

        /** Get equipment allocations for a project. */
        getProjectEquipment(projectId: string): Promise<EquipmentAllocation[]> {
            return get<EquipmentAllocation[]>(`/projects/${projectId}/equipment`);
        },

        /** Log maintenance for an asset. */
        logMaintenance(assetId: string, data: Partial<MaintenanceLog>): Promise<MaintenanceLog> {
            return post<MaintenanceLog>(`/fleet/${assetId}/maintenance`, data);
        },

        /** Get maintenance history for an asset. */
        getMaintenanceHistory(assetId: string): Promise<MaintenanceLog[]> {
            return get<MaintenanceLog[]>(`/fleet/${assetId}/maintenance`);
        },

        /** Get upcoming maintenance across the org. */
        getUpcomingMaintenance(withinDays = 14): Promise<MaintenanceLog[]> {
            return get<MaintenanceLog[]>(`/maintenance/upcoming?within_days=${withinDays}`);
        },
    },

    /**
     * Invoice endpoints.
     * See PHASE_13_PRD.md Step 82: Interactive Invoice
     */
    invoices: {
        /**
         * Get a single invoice by ID.
         * @param id - Invoice UUID
         */
        get(id: string): Promise<InvoiceResponse> {
            return get<InvoiceResponse>(`/invoices/${id}`);
        },

        /**
         * Update invoice line items (Draft invoices only).
         * @param id - Invoice UUID
         * @param data - Updated line items
         */
        update(id: string, data: UpdateInvoiceRequest): Promise<InvoiceResponse> {
            return put<InvoiceResponse>(`/invoices/${id}`, data);
        },

        /**
         * Approve a Draft invoice. Irreversible.
         * @param id - Invoice UUID
         */
        approve(id: string): Promise<InvoiceResponse> {
            return post<InvoiceResponse>(`/invoices/${id}/approve`);
        },

        /**
         * Reject a Draft invoice with an optional reason.
         * @param id - Invoice UUID
         * @param reason - Rejection reason
         */
        reject(id: string, reason?: string): Promise<InvoiceResponse> {
            return post<InvoiceResponse>(`/invoices/${id}/reject`, { reason: reason ?? '' });
        },
    },

    /**
     * Portfolio feed endpoints.
     * See FRONTEND_V2_SPEC.md §5.1
     */
    portfolio: {
        /**
         * Get the portfolio feed with cards, summary, and project pills.
         * @param projectId - Optional project filter
         */
        getFeed(projectId?: string): Promise<PortfolioFeedResponse> {
            const query = projectId ? `?project_id=${encodeURIComponent(projectId)}` : '';
            return get<PortfolioFeedResponse>(`/portfolio/feed${query}`);
        },

        /**
         * Execute an action on a feed card.
         * @param cardId - Feed card UUID
         * @param actionId - Action identifier
         * @param payload - Optional action data
         */
        executeAction(
            cardId: string,
            actionId: string,
            payload?: Record<string, unknown>
        ): Promise<ActionResponse> {
            return post<ActionResponse>('/portfolio/feed/action', {
                card_id: cardId,
                action_id: actionId,
                payload,
            });
        },

        /**
         * Dismiss a feed card.
         * @param cardId - Feed card UUID
         */
        dismissCard(cardId: string): Promise<{ success: boolean }> {
            return post<{ success: boolean }>('/portfolio/feed/dismiss', {
                card_id: cardId,
            });
        },

        /**
         * Snooze a feed card.
         * @param cardId - Feed card UUID
         * @param hours - Snooze duration in hours
         */
        snoozeCard(cardId: string, hours: number): Promise<{ success: boolean }> {
            return post<{ success: boolean }>('/portfolio/feed/snooze', {
                card_id: cardId,
                hours,
            });
        },
    },

    /**
     * Project completion endpoints.
     * Marks a project as completed and generates a completion report.
     */
    completion: {
        /**
         * Mark a project as completed and generate a completion report.
         * @param projectId - Project UUID
         * @param notes - Optional completion notes
         */
        complete(projectId: string, notes?: string): Promise<CompletionReport> {
            return post<CompletionReport>(`/projects/${projectId}/complete`, { notes: notes ?? '' });
        },

        /**
         * Get the completion report for a project.
         * @param projectId - Project UUID
         */
        getReport(projectId: string): Promise<CompletionReport> {
            return get<CompletionReport>(`/projects/${projectId}/completion-report`);
        },
    },

    /**
     * Correction event endpoints (Sprint 3.2: Interactive Learning).
     * Captures user corrections to AI-extracted artifact fields for model improvement.
     */
    corrections: {
        /**
         * Batch submit correction events. Fire-and-forget — callers should
         * not block on this and should catch errors silently.
         * @param events - Array of correction events to log
         */
        submit(events: CorrectionEvent[]): Promise<void> {
            return post<void>('/corrections', { events });
        },
    },

    /**
     * Material CRUD endpoints.
     * See internal/api/handlers/material_handler.go
     */
    materials: {
        /** List all materials for a project. */
        list(projectId: string): Promise<ProjectMaterial[]> {
            return get<ProjectMaterial[]>(`/projects/${projectId}/materials`);
        },

        /** Create a material manually. */
        create(projectId: string, data: CreateMaterialRequest): Promise<ProjectMaterial[]> {
            return post<ProjectMaterial[]>(`/projects/${projectId}/materials`, data);
        },

        /** Update a material. */
        update(projectId: string, materialId: string, data: MaterialUpdateRequest): Promise<ProjectMaterial> {
            return put<ProjectMaterial>(`/projects/${projectId}/materials/${materialId}`, data);
        },

        /** Delete a material. */
        async remove(projectId: string, materialId: string): Promise<void> {
            await del<undefined>(`/projects/${projectId}/materials/${materialId}`);
        },
    },

    /**
     * Budget endpoints.
     * See internal/api/handlers/budget_handler.go
     */
    budget: {
        /** Get per-phase budget breakdown. */
        getBreakdown(projectId: string): Promise<ProjectBudget[]> {
            return get<ProjectBudget[]>(`/projects/${projectId}/budget`);
        },

        /** Update a single phase budget estimate. */
        updatePhase(projectId: string, budgetId: string, estimatedAmountCents: number): Promise<ProjectBudget> {
            return put<ProjectBudget>(`/projects/${projectId}/budget/${budgetId}`, {
                estimated_amount_cents: estimatedAmountCents,
            });
        },

        /** Seed budget from material estimates. */
        seed(projectId: string, data: BudgetSeedRequest): Promise<BudgetEstimate> {
            return post<BudgetEstimate>(`/projects/${projectId}/budget/seed`, data);
        },
    },

    /**
     * Financial summary endpoints (Sprint 4.1: Service Connection).
     * Replaces mock-financial-service.ts with live backend API.
     * Spend is derived from SUM(total_cents) of approved invoices.
     */
    financials: {
        /**
         * Get financial summary for a specific project.
         * @param projectId - Project UUID
         */
        getSummary(projectId: string): Promise<FinancialSummary> {
            return get<FinancialSummary>(`/projects/${projectId}/financials/summary`);
        },

        /**
         * Get global financial summary aggregated across all projects.
         */
        getGlobalSummary(): Promise<FinancialSummary> {
            return get<FinancialSummary>('/financials/summary');
        },
    },
} as const;

// ============================================================================
// Portal Types
// ============================================================================

/**
 * Action context returned when verifying a portal token.
 * See LAUNCH_PLAN.md P2: Field Portal (Mobile).
 */
export interface PortalActionContext {
    action_type: 'status_update' | 'photo_upload' | 'view';
    contact: { id: string; name: string };
    project: { id: string; name: string; address: string };
    task: {
        id: string;
        wbs_code: string;
        name: string;
        status: string;
        start_date?: string;
        end_date?: string;
    };
}

/**
 * Request body for submitting a portal action.
 */
export interface PortalSubmitRequest {
    status?: string;
    photo_id?: string;
}

/**
 * Response after submitting a portal action.
 */
export interface PortalSubmitResponse {
    success: string;
    message: string;
}

/**
 * Response after creating an action link.
 */
export interface PortalLinkResponse {
    success: boolean;
    message: string;
}

// ============================================================================
// Invite Types
// ============================================================================

/**
 * Invitation info returned for public endpoints.
 */
export interface InviteInfo {
    email: string;
    role: string;
    expires_at: string;
}

/**
 * Response when accepting an invitation.
 */
export interface InviteAcceptResponse {
    message: string;
    email: string;
}

/**
 * Full invitation record for admin endpoints.
 */
export interface Invite {
    id: string;
    email: string;
    role: string;
    expires_at: string;
    created_at: string;
}

// ============================================================================
// User Profile Types
// ============================================================================

/**
 * User profile response.
 * See LAUNCH_PLAN.md User Profile Endpoint (P0).
 */
export interface UserProfile {
    id: string;
    email: string;
    name: string;
    role: string;
    org_id: string;
    created_at: string;
}

/**
 * Request body for updating user profile.
 */
export interface UpdateProfileRequest {
    name: string;
}

// ============================================================================
// Physics Config Types (Step 87)
// ============================================================================

/**
 * Response from GET /api/v1/org/settings/physics.
 * See STEP_87_CONFIG_PERSISTENCE.md Section 2
 */
export interface PhysicsConfigResponse {
    speed_multiplier: number;
    work_days: number[];
}

/**
 * Request body for PUT /api/v1/org/settings/physics.
 * See STEP_87_CONFIG_PERSISTENCE.md Section 2
 */
export interface UpdatePhysicsRequest {
    speed_multiplier: number;
    work_days: number[];
    apply_to_existing?: boolean;
}

// ============================================================================
// Agent Settings Types
// ============================================================================

export interface AgentSettingsResponse {
    daily_focus: {
        enabled: boolean;
        run_time: string;
        ai_provider: string;
        max_focus_cards: number;
    };
    procurement: {
        lead_time_warning_threshold: number;
        staging_buffer_days: number;
        default_weather_buffer_days: number;
    };
    sub_liaison: {
        enabled: boolean;
        confirmation_window: string;
        auto_resend_after: string;
    };
    chat: {
        ai_provider: string;
        max_tool_calls: number;
    };
}

export interface BrainConnectionResponse {
    id: string;
    org_id: string;
    brain_url: string;
    integration_key: string;
    status: string;
    last_sync_at: string | null;
    platforms: Array<{ name: string; type: string; status: string }>;
    updated_at: string;
}

// ============================================================================
// Financial Types (Sprint 4.1: Service Connection)
// ============================================================================

/**
 * Category-level financial summary.
 * Matches Go FinancialCategorySummary.
 */
export interface FinancialCategorySummary {
    name: string;
    budget: number;
    spend: number;
    status: 'on_track' | 'at_risk' | 'over_budget';
}

/**
 * Financial summary response.
 * Matches Go FinancialSummary. Spend is derived from SUM of approved invoices.
 * See Sprint 4.1 Task 4.1.3.
 */
export interface FinancialSummary {
    project_id?: string;
    budget_total: number;
    spend_total: number;
    variance: number;
    last_updated: string;
    categories: FinancialCategorySummary[];
}

// ============================================================================
// Invoice Types (Step 82)
// ============================================================================

/**
 * Line item in an invoice update request.
 */
export interface UpdateInvoiceLineItem {
    description: string;
    quantity: number;
    unit_price_cents: number;
}

/**
 * Request body for updating invoice line items.
 * See STEP_82_INTERACTIVE_INVOICE.md Section 2.1
 */
export interface UpdateInvoiceRequest {
    items: UpdateInvoiceLineItem[];
}

/**
 * Full invoice response from backend.
 * Mirrors Go models.Invoice.
 */
export interface InvoiceResponse {
    id: string;
    project_id: string;
    vendor_name: string;
    amount_cents: number;
    line_items: InvoiceLineItemResponse[];
    detected_wbs_code: string;
    status: InvoiceStatus;
    invoice_date: string | null;
    invoice_number: string | null;
    confidence: number;
    is_human_review_required: boolean;
    source_document_id: string | null;
    // Step 83: Approval metadata (Rosetta Stone parity with Go models.Invoice)
    approved_by_id: string | null;
    approved_at: string | null;
    rejected_by_id: string | null;
    rejected_at: string | null;
    rejection_reason: string | null;
}

/**
 * Line item in an invoice response.
 */
export interface InvoiceLineItemResponse {
    description: string;
    quantity: number;
    unit_price_cents: number;
    total_cents: number;
}

// ============================================================================
// Vision Status Types (Step 84)
// ============================================================================

/**
 * Vision analysis status for a project asset.
 * See STEP_84_FIELD_FEEDBACK.md Section 2.2
 */
export type AnalysisStatus = 'processing' | 'completed' | 'failed';

/**
 * Response from GET /api/v1/vision/status/:id.
 * See STEP_84_FIELD_FEEDBACK.md Section 2.2
 */
export interface VisionStatusResponse {
    id: string;
    status: AnalysisStatus;
    analysis: Record<string, unknown> | null;
}

/**
 * Project asset with analysis status for gallery display.
 * See STEP_85_VISION_BADGES.md Section 2.1
 */
export interface ProjectAssetResponse {
    id: string;
    file_name: string;
    file_url: string;
    analysis_status: AnalysisStatus;
    analysis: Record<string, unknown> | null;
    created_at: string;
}
