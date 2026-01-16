/**
 * Global State Store - Signals-Based Reactive State
 * See FRONTEND_SCOPE.md Section 5.1
 *
 * Implements a centralized state store using @preact/signals-core.
 * Key patterns:
 * - Readonly signal exposure to prevent arbitrary mutation
 * - Action-based mutations for predictable state changes
 * - localStorage persistence for auth token
 * - Computed values for derived state
 *
 * @example
 * ```typescript
 * import { store } from '@/store/store';
 *
 * // Read state (reactive)
 * const isAuth = store.isAuthenticated$.value;
 *
 * // Mutate via actions
 * store.actions.login(user, token);
 * store.actions.logout();
 * ```
 */

import { signal, computed, effect, type ReadonlySignal } from '@preact/signals-core';
import type {
    User,
    ProjectSummary,
    ProjectDetail,
    ChatMessage,
    Theme,
    StoreActions,
} from './types';
import { setTokenGetter, setUnauthorizedHandler } from '../services/http';

// ============================================================================
// Constants
// ============================================================================

const STORAGE_KEY_TOKEN = 'fb_token';
const STORAGE_KEY_THEME = 'fb_theme';

// ============================================================================
// Internal Writable Signals
// ============================================================================

// Auth
const _user$ = signal<User | null>(null);
const _token$ = signal<string | null>(null);
const _authLoading$ = signal<boolean>(false);
const _authError$ = signal<string | null>(null);

// Projects
const _projects$ = signal<ProjectSummary[]>([]);
const _currentProjectId$ = signal<string | null>(null);
const _currentProjectDetail$ = signal<ProjectDetail | null>(null);
const _projectLoading$ = signal<boolean>(false);
const _projectError$ = signal<string | null>(null);

// Chat
const _messages$ = signal<ChatMessage[]>([]);
const _chatLoading$ = signal<boolean>(false);
const _chatError$ = signal<string | null>(null);

// UI
const _sidebarOpen$ = signal<boolean>(true);
const _theme$ = signal<Theme>('system');
const _isMobile$ = signal<boolean>(false);
const _activeView$ = signal<string>('dashboard');

// ============================================================================
// Computed Values
// ============================================================================

/**
 * Computed: Whether the user is authenticated.
 */
const _isAuthenticated$ = computed(() => _token$.value !== null && _user$.value !== null);

/**
 * Computed: The currently selected project from the list.
 */
const _currentProject$ = computed(() => {
    const id = _currentProjectId$.value;
    if (!id) return null;
    return _projects$.value.find((p) => p.id === id) ?? null;
});

/**
 * Computed: Whether any async operation is in progress.
 */
const _isLoading$ = computed(
    () => _authLoading$.value || _projectLoading$.value || _chatLoading$.value
);

// ============================================================================
// Actions
// ============================================================================

const actions: StoreActions = {
    // ---- Auth Actions ----

    login(user: User, token: string): void {
        _user$.value = user;
        _token$.value = token;
        _authError$.value = null;
        _authLoading$.value = false;
    },

    logout(): void {
        _user$.value = null;
        _token$.value = null;
        _currentProjectId$.value = null;
        _currentProjectDetail$.value = null;
        _messages$.value = [];
        _authError$.value = null;
    },

    setAuthLoading(loading: boolean): void {
        _authLoading$.value = loading;
    },

    setAuthError(error: string | null): void {
        _authError$.value = error;
        _authLoading$.value = false;
    },

    // ---- Project Actions ----

    setProjects(projects: ProjectSummary[]): void {
        _projects$.value = projects;
        _projectError$.value = null;
        _projectLoading$.value = false;
    },

    selectProject(id: string | null): void {
        _currentProjectId$.value = id;
        // Clear detail when switching projects
        if (_currentProjectDetail$.value?.id !== id) {
            _currentProjectDetail$.value = null;
        }
        // Clear chat when switching projects
        _messages$.value = [];
    },

    setCurrentProjectDetail(detail: ProjectDetail | null): void {
        _currentProjectDetail$.value = detail;
        _projectLoading$.value = false;
    },

    setProjectLoading(loading: boolean): void {
        _projectLoading$.value = loading;
    },

    setProjectError(error: string | null): void {
        _projectError$.value = error;
        _projectLoading$.value = false;
    },

    // ---- Chat Actions ----

    addMessage(message: ChatMessage): void {
        _messages$.value = [..._messages$.value, message];
    },

    setMessages(messages: ChatMessage[]): void {
        _messages$.value = messages;
        _chatError$.value = null;
        _chatLoading$.value = false;
    },

    updateMessage(id: string, updates: Partial<ChatMessage>): void {
        _messages$.value = _messages$.value.map((m) =>
            m.id === id ? { ...m, ...updates } : m
        );
    },

    setChatLoading(loading: boolean): void {
        _chatLoading$.value = loading;
    },

    setChatError(error: string | null): void {
        _chatError$.value = error;
        _chatLoading$.value = false;
    },

    // ---- UI Actions ----

    toggleSidebar(): void {
        _sidebarOpen$.value = !_sidebarOpen$.value;
    },

    setSidebarOpen(open: boolean): void {
        _sidebarOpen$.value = open;
    },

    setTheme(theme: Theme): void {
        _theme$.value = theme;
    },

    setIsMobile(isMobile: boolean): void {
        _isMobile$.value = isMobile;
        // Auto-close sidebar on mobile
        if (isMobile) {
            _sidebarOpen$.value = false;
        }
    },

    setActiveView(view: string): void {
        _activeView$.value = view;
    },
};

// ============================================================================
// Store Singleton
// ============================================================================

/**
 * Global application store.
 *
 * State is exposed as readonly signals to prevent arbitrary mutation.
 * All changes must go through the actions object.
 */
export const store = {
    // ---- Auth State (readonly) ----
    user$: _user$ as ReadonlySignal<User | null>,
    token$: _token$ as ReadonlySignal<string | null>,
    isAuthenticated$: _isAuthenticated$,
    authLoading$: _authLoading$ as ReadonlySignal<boolean>,
    authError$: _authError$ as ReadonlySignal<string | null>,

    // ---- Project State (readonly) ----
    projects$: _projects$ as ReadonlySignal<ProjectSummary[]>,
    currentProjectId$: _currentProjectId$ as ReadonlySignal<string | null>,
    currentProject$: _currentProject$,
    currentProjectDetail$: _currentProjectDetail$ as ReadonlySignal<ProjectDetail | null>,
    projectLoading$: _projectLoading$ as ReadonlySignal<boolean>,
    projectError$: _projectError$ as ReadonlySignal<string | null>,

    // ---- Chat State (readonly) ----
    messages$: _messages$ as ReadonlySignal<ChatMessage[]>,
    chatLoading$: _chatLoading$ as ReadonlySignal<boolean>,
    chatError$: _chatError$ as ReadonlySignal<string | null>,

    // ---- UI State (readonly) ----
    sidebarOpen$: _sidebarOpen$ as ReadonlySignal<boolean>,
    theme$: _theme$ as ReadonlySignal<Theme>,
    isMobile$: _isMobile$ as ReadonlySignal<boolean>,
    activeView$: _activeView$ as ReadonlySignal<string>,

    // ---- Global Computed ----
    isLoading$: _isLoading$,

    // ---- Actions ----
    actions,
} as const;

// ============================================================================
// Effects (Side Effects)
// ============================================================================

/**
 * Effect: Persist auth token to localStorage.
 */
effect(() => {
    const token = _token$.value;
    if (token) {
        localStorage.setItem(STORAGE_KEY_TOKEN, token);
    } else {
        localStorage.removeItem(STORAGE_KEY_TOKEN);
    }
});

/**
 * Effect: Persist theme preference to localStorage.
 */
effect(() => {
    const theme = _theme$.value;
    localStorage.setItem(STORAGE_KEY_THEME, theme);
});

// ============================================================================
// Initialization
// ============================================================================

/**
 * Initialize the store with persisted state.
 * Call this once at application bootstrap.
 */
export function initializeStore(): void {
    // Restore token from localStorage
    const storedToken = localStorage.getItem(STORAGE_KEY_TOKEN);
    if (storedToken) {
        _token$.value = storedToken;
        // Note: User data should be fetched via api.auth.me() after restore
    }

    // Restore theme from localStorage
    const storedTheme = localStorage.getItem(STORAGE_KEY_THEME) as Theme | null;
    if (storedTheme && ['light', 'dark', 'system'].includes(storedTheme)) {
        _theme$.value = storedTheme;
    }

    // Detect mobile viewport
    if (typeof window !== 'undefined') {
        const mql = window.matchMedia('(max-width: 768px)');
        _isMobile$.value = mql.matches;
        mql.addEventListener('change', (e) => {
            actions.setIsMobile(e.matches);
        });
    }

    // Wire up HTTP service callbacks to prevent circular dependencies
    setTokenGetter(() => _token$.value);
    setUnauthorizedHandler(() => {
        actions.logout();
    });
}

// ============================================================================
// Type Exports
// ============================================================================

export type { User, ProjectSummary, ProjectDetail, ChatMessage, Theme, StoreActions };
