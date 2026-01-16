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
// UI State
// ============================================================================

/**
 * Theme preference.
 */
export type Theme = 'light' | 'dark' | 'system';

/**
 * UI state slice.
 */
export interface UIState {
    /** Whether the sidebar is expanded */
    sidebarOpen: boolean;
    /** Current theme preference */
    theme: Theme;
    /** Whether the app is in mobile viewport */
    isMobile: boolean;
    /** Currently active view/route */
    activeView: string;
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
    chat: ChatState;
    ui: UIState;
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

    // Chat actions
    addMessage(message: ChatMessage): void;
    setMessages(messages: ChatMessage[]): void;
    updateMessage(id: string, updates: Partial<ChatMessage>): void;
    setChatLoading(loading: boolean): void;
    setChatError(error: string | null): void;

    // UI actions
    toggleSidebar(): void;
    setSidebarOpen(open: boolean): void;
    setTheme(theme: Theme): void;
    setIsMobile(isMobile: boolean): void;
    setActiveView(view: string): void;
}
