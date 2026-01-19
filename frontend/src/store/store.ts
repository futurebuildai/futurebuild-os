/**
 * Global State Store - Signals-Based Reactive State
 * See FRONTEND_SCOPE.md Section 5.1 (Updated for Agent Command Center)
 *
 * Implements a centralized state store using @preact/signals-core.
 * Key patterns:
 * - Readonly signal exposure to prevent arbitrary mutation
 * - Action-based mutations for predictable state changes
 * - localStorage persistence for auth token
 * - Computed values for derived state
 * - 3-panel architecture state (left/center/right)
 */

import { signal, computed, effect, type ReadonlySignal } from '@preact/signals-core';
import type {
    User,
    ProjectSummary,
    ProjectDetail,
    ChatMessage,
    Theme,
    StoreActions,
    Thread,
    AgentActivity,
    FocusTask,
    ActionCard,
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

// Threads (New for Agent Command Center)
const _threads$ = signal<Thread[]>([]);
const _activeThreadId$ = signal<string | null>(null);
const _threadLoading$ = signal<boolean>(false);

// Chat
const _messages$ = signal<ChatMessage[]>([]);
const _chatLoading$ = signal<boolean>(false);
const _chatError$ = signal<string | null>(null);

// Daily Focus & Agent Activity (New for Agent Command Center)
const _focusTasks$ = signal<FocusTask[]>([]);
const _agentActivity$ = signal<AgentActivity[]>([]);

// UI (Updated for 3-panel layout)
const _leftPanelOpen$ = signal<boolean>(true);
const _rightPanelOpen$ = signal<boolean>(true);
const _theme$ = signal<Theme>('system');
const _isMobile$ = signal<boolean>(false);
const _isTablet$ = signal<boolean>(false);
const _activeProjectId$ = signal<string | null>(null);

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
    const id = _activeProjectId$.value;
    if (!id) return null;
    return _projects$.value.find((p) => p.id === id) ?? null;
});

/**
 * Computed: The currently active thread.
 */
const _activeThread$ = computed(() => {
    const id = _activeThreadId$.value;
    if (!id) return null;
    return _threads$.value.find((t) => t.id === id) ?? null;
});

/**
 * Computed: Threads for the currently active project.
 */
const _projectThreads$ = computed(() => {
    const projectId = _activeProjectId$.value;
    if (!projectId) return [];
    return _threads$.value.filter((t) => t.projectId === projectId);
});

/**
 * Computed: Whether any async operation is in progress.
 */
const _isLoading$ = computed(
    () => _authLoading$.value || _projectLoading$.value || _chatLoading$.value || _threadLoading$.value
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
        _activeProjectId$.value = null;
        _activeThreadId$.value = null;
        _currentProjectDetail$.value = null;
        _messages$.value = [];
        _threads$.value = [];
        _focusTasks$.value = [];
        _agentActivity$.value = [];
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
        _activeProjectId$.value = id;
        // Clear detail when switching projects
        if (_currentProjectDetail$.value?.id !== id) {
            _currentProjectDetail$.value = null;
        }
        // Clear thread selection and messages when switching projects
        _activeThreadId$.value = null;
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

    // ---- Thread Actions ----

    setThreads(threads: Thread[]): void {
        _threads$.value = threads;
        _threadLoading$.value = false;
    },

    selectThread(id: string | null): void {
        _activeThreadId$.value = id;
        // Load messages from thread
        const thread = _threads$.value.find((t) => t.id === id);
        _messages$.value = thread?.messages ?? [];
    },

    addThread(thread: Thread): void {
        _threads$.value = [..._threads$.value, thread];
    },

    markThreadRead(id: string): void {
        _threads$.value = _threads$.value.map((t) =>
            t.id === id ? { ...t, hasUnread: false } : t
        );
    },

    // ---- Chat Actions ----

    addMessage(message: ChatMessage): void {
        _messages$.value = [..._messages$.value, message];
        // Also update the thread's messages
        if (_activeThreadId$.value) {
            _threads$.value = _threads$.value.map((t) =>
                t.id === _activeThreadId$.value
                    ? { ...t, messages: [...t.messages, message], updatedAt: new Date().toISOString() }
                    : t
            );
        }
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

    updateActionCard(messageId: string, status: ActionCard['status']): void {
        _messages$.value = _messages$.value.map((m) =>
            m.id === messageId && m.actionCard
                ? { ...m, actionCard: { ...m.actionCard, status } }
                : m
        );
    },

    // ---- Focus & Activity Actions ----

    setFocusTasks(tasks: FocusTask[]): void {
        _focusTasks$.value = tasks;
    },

    dismissFocusTask(id: string): void {
        _focusTasks$.value = _focusTasks$.value.filter((t) => t.id !== id);
    },

    addAgentActivity(activity: AgentActivity): void {
        _agentActivity$.value = [activity, ..._agentActivity$.value].slice(0, 50); // Keep last 50
    },

    updateAgentActivity(id: string, updates: Partial<AgentActivity>): void {
        _agentActivity$.value = _agentActivity$.value.map((a) =>
            a.id === id ? { ...a, ...updates } : a
        );
    },

    // ---- UI Actions ----

    toggleLeftPanel(): void {
        _leftPanelOpen$.value = !_leftPanelOpen$.value;
    },

    toggleRightPanel(): void {
        _rightPanelOpen$.value = !_rightPanelOpen$.value;
    },

    setLeftPanelOpen(open: boolean): void {
        _leftPanelOpen$.value = open;
    },

    setRightPanelOpen(open: boolean): void {
        _rightPanelOpen$.value = open;
    },

    setTheme(theme: Theme): void {
        _theme$.value = theme;
    },

    setIsMobile(isMobile: boolean): void {
        _isMobile$.value = isMobile;
        // Auto-close panels on mobile
        if (isMobile) {
            _leftPanelOpen$.value = false;
            _rightPanelOpen$.value = false;
        }
    },

    setIsTablet(isTablet: boolean): void {
        _isTablet$.value = isTablet;
        // Auto-close right panel on tablet
        if (isTablet) {
            _rightPanelOpen$.value = false;
        }
    },

    setActiveProject(projectId: string | null): void {
        _activeProjectId$.value = projectId;
        // Clear thread selection when switching projects
        _activeThreadId$.value = null;
        _messages$.value = [];
    },

    setActiveThread(threadId: string | null): void {
        _activeThreadId$.value = threadId;
        // Load messages from thread
        const thread = _threads$.value.find((t) => t.id === threadId);
        _messages$.value = thread?.messages ?? [];
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

    // ---- Thread State (readonly) ----
    threads$: _threads$ as ReadonlySignal<Thread[]>,
    activeThreadId$: _activeThreadId$ as ReadonlySignal<string | null>,
    activeThread$: _activeThread$,
    projectThreads$: _projectThreads$,
    threadLoading$: _threadLoading$ as ReadonlySignal<boolean>,

    // ---- Chat State (readonly) ----
    messages$: _messages$ as ReadonlySignal<ChatMessage[]>,
    chatLoading$: _chatLoading$ as ReadonlySignal<boolean>,
    chatError$: _chatError$ as ReadonlySignal<string | null>,

    // ---- Focus & Activity State (readonly) ----
    focusTasks$: _focusTasks$ as ReadonlySignal<FocusTask[]>,
    agentActivity$: _agentActivity$ as ReadonlySignal<AgentActivity[]>,

    // ---- UI State (readonly) ----
    leftPanelOpen$: _leftPanelOpen$ as ReadonlySignal<boolean>,
    rightPanelOpen$: _rightPanelOpen$ as ReadonlySignal<boolean>,
    theme$: _theme$ as ReadonlySignal<Theme>,
    isMobile$: _isMobile$ as ReadonlySignal<boolean>,
    isTablet$: _isTablet$ as ReadonlySignal<boolean>,
    activeProjectId$: _activeProjectId$ as ReadonlySignal<string | null>,

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

    // Detect mobile/tablet viewport
    if (typeof window !== 'undefined') {
        const mobileQuery = window.matchMedia('(max-width: 768px)');
        const tabletQuery = window.matchMedia('(max-width: 1024px)');

        _isMobile$.value = mobileQuery.matches;
        _isTablet$.value = tabletQuery.matches && !mobileQuery.matches;

        // Auto-collapse panels on smaller screens
        if (mobileQuery.matches) {
            _leftPanelOpen$.value = false;
            _rightPanelOpen$.value = false;
        } else if (tabletQuery.matches) {
            _rightPanelOpen$.value = false;
        }

        mobileQuery.addEventListener('change', (e) => {
            actions.setIsMobile(e.matches);
        });
        tabletQuery.addEventListener('change', (e) => {
            actions.setIsTablet(e.matches && !mobileQuery.matches);
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

export type { User, ProjectSummary, ProjectDetail, ChatMessage, Theme, StoreActions, Thread, AgentActivity, FocusTask };
