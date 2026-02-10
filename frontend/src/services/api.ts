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

import { get, post, put, del } from './http';
import type { GanttData, Contact, InvoiceExtraction, Thread as ApiThread, CompletionReport } from '../types/models';
import type { InvoiceStatus } from '../types/enums';
import type { UserRole } from '../types/enums';
import type { PortfolioFeedResponse } from '../types/feed';

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
     */
    contacts: {
        /**
         * List contacts for a project.
         * @param projectId - Project UUID
         */
        list(projectId: string): Promise<Contact[]> {
            return get<Contact[]>(`/projects/${projectId}/contacts`);
        },

        /**
         * Get a single contact.
         * @param contactId - Contact UUID
         */
        get(contactId: string): Promise<Contact> {
            return get<Contact>(`/contacts/${contactId}`);
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
        ): Promise<{ success: boolean; message: string }> {
            return post<{ success: boolean; message: string }>('/portfolio/feed/action', {
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
