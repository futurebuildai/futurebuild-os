/**
 * API Service - Domain-Specific API Bindings
 * See FRONTEND_SCOPE.md Section 5.2
 *
 * Provides typed API methods organized by domain:
 * - auth: Magic link authentication
 * - projects: Project CRUD operations
 * - chat: Chat messaging
 * - schedule: Gantt/Schedule data
 * - tasks: Task management
 *
 * IMPORTANT: Uses types from src/types/ for all request/response bodies.
 * Does NOT import from store to prevent circular dependencies.
 */

import { get, post, put, del } from './http';
import type { GanttData, Contact, InvoiceExtraction } from '../types/models';
import type { UserRole } from '../types/enums';

// ============================================================================
// Auth Types
// ============================================================================

/**
 * Magic link request payload.
 */
export interface MagicLinkRequest {
    email: string;
}

/**
 * Magic link response.
 */
export interface MagicLinkResponse {
    message: string;
}

/**
 * Token verification response.
 */
export interface AuthResponse {
    token: string;
    user: AuthUser;
}

/**
 * Authenticated user data.
 */
export interface AuthUser {
    id: string;
    email: string;
    name: string;
    role: UserRole;
    org_id: string;
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
}

// ============================================================================
// Chat Types
// ============================================================================

/**
 * Chat message request.
 */
export interface ChatRequest {
    project_id: string;
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
 * // Authentication
 * await api.auth.requestMagicLink('user@example.com');
 * const { token, user } = await api.auth.verifyToken('abc123');
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
     */
    auth: {
        /**
         * Request a magic link email.
         * @param email - User's email address
         */
        requestMagicLink(email: string): Promise<MagicLinkResponse> {
            return post<MagicLinkResponse>(
                '/auth/login',
                { email } satisfies MagicLinkRequest,
                { skipAuth: true }
            );
        },

        /**
         * Verify a magic link token and get auth credentials.
         * @param token - The magic link token from email
         */
        verifyToken(token: string): Promise<AuthResponse> {
            return get<AuthResponse>(
                `/auth/verify?token=${encodeURIComponent(token)}`,
                { skipAuth: true }
            );
        },

        /**
         * Logout and invalidate the current session.
         */
        async logout(): Promise<void> {
            await post<undefined>('/auth/logout');
        },

        /**
         * Get the current authenticated user.
         */
        me(): Promise<AuthUser> {
            return get<AuthUser>('/auth/me');
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
     * Chat endpoints.
     */
    chat: {
        /**
         * Send a chat message and get a response.
         * @param projectId - Project UUID context
         * @param message - User message text
         */
        send(projectId: string, message: string): Promise<ChatResponse> {
            return post<ChatResponse>('/chat', {
                project_id: projectId,
                message,
            } satisfies ChatRequest);
        },

        /**
         * Get chat history for a project.
         * @param projectId - Project UUID
         * @param limit - Maximum messages to return (default 50)
         */
        history(projectId: string, limit = 50): Promise<ChatResponse[]> {
            return get<ChatResponse[]>(
                `/projects/${projectId}/chat?limit=${String(limit)}`
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
         */
        accept(token: string, name: string): Promise<InviteAcceptResponse> {
            return post<InviteAcceptResponse>(
                '/invites/accept',
                { token, name },
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
