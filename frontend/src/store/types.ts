/**
 * Store Types - Application State Shape Definitions
 * See FRONTEND_SCOPE.md Section 5.1
 *
 * Defines the TypeScript interfaces for all application state.
 * These types are used by the Signals-based store.
 */

import type { UserRole } from '../types/enums';

// ============================================================================
// Auth State
// ============================================================================

/**
 * Authenticated user data stored in state.
 * Matches AuthUser from API but uses camelCase for frontend convention.
 */
export interface User {
    id: string;
    email: string;
    name: string;
    role: UserRole;
    orgId: string;
}

/**
 * Authentication state slice.
 */
export interface AuthState {
    /** The currently authenticated user, or null if not authenticated */
    user: User | null;
    /** JWT auth token, or null if not authenticated */
    token: string | null;
    /** Whether an auth operation is in progress */
    isLoading: boolean;
    /** Last auth error message, if any */
    error: string | null;
}

// ============================================================================
// Project State
// ============================================================================

/**
 * Project data stored in state.
 * Simplified view for list/selection.
 */
export interface ProjectSummary {
    id: string;
    name: string;
    address: string;
    status: string;
    completionPercentage: number;
    createdAt: string;
    updatedAt: string;
}

/**
 * Full project detail.
 */
export interface ProjectDetail extends ProjectSummary {
    orgId: string;
    squareFootage: number;
    bedrooms: number;
    bathrooms: number;
    lotSize: number;
    foundationType: string;
    startDate: string;
    projectedEndDate: string | null;
}

/**
 * Project state slice.
 */
export interface ProjectState {
    /** List of project summaries */
    items: ProjectSummary[];
    /** Currently selected project ID, or null */
    currentId: string | null;
    /** Full detail of current project, or null */
    currentDetail: ProjectDetail | null;
    /** Whether project data is loading */
    isLoading: boolean;
    /** Last project error message, if any */
    error: string | null;
}

// ============================================================================
// Chat State
// ============================================================================

/**
 * Chat message in state.
 */
export interface ChatMessage {
    id: string;
    role: 'user' | 'assistant';
    content: string;
    createdAt: string;
    /** Whether this message is still being streamed */
    isStreaming?: boolean;
    /** Optional action card if agent is requesting approval */
    actionCard?: ActionCard;
    /** Optional artifact reference to display inline */
    artifactRef?: ArtifactRef;
}

/**
 * Action card embedded in assistant message.
 * Represents a recommended action requiring user approval.
 */
export interface ActionCard {
    id: string;
    type: 'invoice_approval' | 'schedule_change' | 'material_order' | 'confirmation' | 'general';
    title: string;
    summary: string;
    status: 'pending' | 'approved' | 'denied' | 'edited';
    /** Optional data payload for the action */
    data?: Record<string, unknown>;
}

/**
 * Reference to an artifact (can be displayed inline or in artifact panel).
 */
export interface ArtifactRef {
    id: string;
    type: 'gantt' | 'budget' | 'invoice' | 'table' | 'chart';
    title: string;
    /** Whether this is a project-level (global) or thread-level artifact */
    scope: 'project' | 'thread';
}

/**
 * Chat state slice.
 */
export interface ChatState {
    /** Messages for the current project */
    messages: ChatMessage[];
    /** Whether a message is being sent/processed */
    isLoading: boolean;
    /** Last chat error, if any */
    error: string | null;
}

// ============================================================================
// Thread State (New for Agent Command Center)
// ============================================================================

/**
 * Conversation thread within a project.
 * Projects can have multiple concurrent threads.
 */
export interface Thread {
    id: string;
    projectId: string;
    title: string;
    createdAt: string;
    updatedAt: string;
    /** Messages in this thread */
    messages: ChatMessage[];
    /** Whether this thread has unread messages */
    hasUnread: boolean;
}

/**
 * Thread state slice.
 */
export interface ThreadState {
    /** All threads for the current project */
    items: Thread[];
    /** Currently active thread ID */
    activeId: string | null;
    /** Whether thread data is loading */
    isLoading: boolean;
}

// ============================================================================
// Agent Activity State
// ============================================================================

/**
 * Autonomous agent activity log entry.
 * Shows what the AI agent did without user prompting.
 */
export interface AgentActivity {
    id: string;
    action: string;
    description: string;
    timestamp: string;
    /** Associated project ID if applicable */
    projectId?: string;
    /** Associated thread ID if applicable */
    threadId?: string;
    /** Activity status */
    status: 'running' | 'completed' | 'failed';
}

// ============================================================================
// Daily Focus State
// ============================================================================

/**
 * Daily Focus task - prioritized agent-recommended action.
 */
export interface FocusTask {
    id: string;
    title: string;
    description: string;
    priority: 'high' | 'medium' | 'low';
    projectId: string;
    projectName: string;
    /** Thread ID to navigate to when clicked */
    threadId?: string;
    /** Action type this task relates to */
    actionType: 'approval' | 'review' | 'confirmation' | 'urgent';
    createdAt: string;
}

// ============================================================================
// UI State
// ============================================================================

/**
 * Theme preference.
 */
export type Theme = 'light' | 'dark' | 'system';

/**
 * UI state slice.
 * Updated for 3-panel Agent Command Center layout.
 */
export interface UIState {
    /** Whether the left panel is visible (projects/threads) */
    leftPanelOpen: boolean;
    /** Whether the right panel is visible (artifacts) */
    rightPanelOpen: boolean;
    /** Current theme preference */
    theme: Theme;
    /** Whether the app is in mobile viewport */
    isMobile: boolean;
    /** Whether the app is in tablet viewport */
    isTablet: boolean;
    /** Currently active project ID */
    activeProjectId: string | null;
    /** Currently active thread ID */
    activeThreadId: string | null;
}

// ============================================================================
// Aggregate State
// ============================================================================

/**
 * Complete application state shape.
 * This is the root type for the entire store.
 */
export interface AppState {
    auth: AuthState;
    project: ProjectState;
    thread: ThreadState;
    chat: ChatState;
    ui: UIState;
    /** Daily Focus tasks */
    focusTasks: FocusTask[];
    /** Agent activity log */
    agentActivity: AgentActivity[];
}

// ============================================================================
// Action Types
// ============================================================================

/**
 * Store actions interface.
 * All state mutations go through these methods.
 */
export interface StoreActions {
    // Auth actions
    login(user: User, token: string): void;
    logout(): void;
    setAuthLoading(loading: boolean): void;
    setAuthError(error: string | null): void;

    // Project actions
    setProjects(projects: ProjectSummary[]): void;
    selectProject(id: string | null): void;
    setCurrentProjectDetail(detail: ProjectDetail | null): void;
    setProjectLoading(loading: boolean): void;
    setProjectError(error: string | null): void;

    // Thread actions
    setThreads(threads: Thread[]): void;
    selectThread(id: string | null): void;
    addThread(thread: Thread): void;
    markThreadRead(id: string): void;

    // Chat actions
    addMessage(message: ChatMessage): void;
    setMessages(messages: ChatMessage[]): void;
    updateMessage(id: string, updates: Partial<ChatMessage>): void;
    setChatLoading(loading: boolean): void;
    setChatError(error: string | null): void;
    updateActionCard(messageId: string, status: ActionCard['status']): void;

    // Focus & Activity actions
    setFocusTasks(tasks: FocusTask[]): void;
    dismissFocusTask(id: string): void;
    addAgentActivity(activity: AgentActivity): void;
    updateAgentActivity(id: string, updates: Partial<AgentActivity>): void;

    // UI actions
    toggleLeftPanel(): void;
    toggleRightPanel(): void;
    setLeftPanelOpen(open: boolean): void;
    setRightPanelOpen(open: boolean): void;
    setTheme(theme: Theme): void;
    setIsMobile(isMobile: boolean): void;
    setIsTablet(isTablet: boolean): void;
    setActiveProject(projectId: string | null): void;
    setActiveThread(threadId: string | null): void;
}
