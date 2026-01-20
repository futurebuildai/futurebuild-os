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
    PendingUpload,
} from './types';
import { setTokenGetter, setUnauthorizedHandler } from '../services/http';
import { realtimeService, type ConnectionStatus, type ArtifactPayload } from '../services/realtime';
import { normalizeArtifactType } from '../utils/artifact-helpers';

// ============================================================================
// Constants
// ============================================================================

const STORAGE_KEY_TOKEN = 'fb_token';
const STORAGE_KEY_THEME = 'fb_theme';

// Step 56: Allowed MIME types for drag-and-drop file ingestion
// See FRONTEND_SCOPE.md Section 8.2
const ALLOWED_UPLOAD_TYPES = [
    'application/pdf',
    'image/jpeg',
    'image/png',
    'image/gif',
    'image/webp',
] as const;

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

// Upload state (Step 56: Drag-and-Drop Ingestion)
const _isDragging$ = signal<boolean>(false);
const _pendingFiles$ = signal<PendingUpload[]>([]);

// Realtime state (Step 57: Real-Time Messaging)
const _isTyping$ = signal<boolean>(false);
const _connectionStatus$ = signal<ConnectionStatus>('disconnected');
const _activeArtifact$ = signal<ArtifactPayload | null>(null);

// Panel resize state (Step 59.5: UX Enhancements)
const _rightPanelWidth$ = signal<number>(320);
const _popoutArtifact$ = signal<ArtifactPayload | null>(null);

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

    // ---- Upload Actions (Step 56: Drag-and-Drop Ingestion) ----

    setDragging(isDragging: boolean): void {
        _isDragging$.value = isDragging;
    },

    handleFileDrop(files: FileList): void {
        const newUploads: PendingUpload[] = [];

        Array.from(files).forEach((file) => {
            if (ALLOWED_UPLOAD_TYPES.includes(file.type as typeof ALLOWED_UPLOAD_TYPES[number])) {
                newUploads.push({
                    id: crypto.randomUUID(),
                    file,
                    status: 'pending',
                    progress: 0,
                });
            }
        });

        if (newUploads.length === 0) {
            _isDragging$.value = false;
            return;
        }

        _pendingFiles$.value = [..._pendingFiles$.value, ...newUploads];
        _isDragging$.value = false;

        // Create user message with attachment info
        const fileNames = newUploads.map((u) => u.file.name).join(', ');
        actions.addMessage({
            id: crypto.randomUUID(),
            role: 'user',
            content: `📎 Uploaded: ${fileNames}`,
            createdAt: new Date().toISOString(),
        });

        // Step 57: Send file message via RealtimeService (replaces setTimeout)
        // See FRONTEND_SCOPE.md Section 8.4
        const firstUpload = newUploads[0];
        realtimeService.send({
            type: 'file',
            payload: {
                fileName: fileNames,
                fileType: firstUpload?.file.type ?? 'application/octet-stream',
            },
        });
    },

    clearPendingUploads(): void {
        _pendingFiles$.value = [];
    },

    // ---- Realtime Actions (Step 57: Real-Time Messaging) ----

    setTyping(isTyping: boolean): void {
        _isTyping$.value = isTyping;
    },

    setConnectionStatus(status: ConnectionStatus): void {
        _connectionStatus$.value = status;
    },

    setActiveArtifact(artifact: ArtifactPayload | null): void {
        _activeArtifact$.value = artifact;
        if (!artifact) return;

        // Auto-open right panel when artifact is set
        _rightPanelOpen$.value = true;
    },

    // ---- Panel Resize Actions (Step 59.5: UX Enhancements) ----

    setRightPanelWidth(width: number): void {
        // Constrain: min 280px, max 600px
        _rightPanelWidth$.value = Math.max(280, Math.min(600, width));
    },

    setPopoutArtifact(artifact: ArtifactPayload | null): void {
        _popoutArtifact$.value = artifact;
    },

    // ---- Session Reset (Step 58.5: State Hygiene) ----

    resetSession(): void {
        // Auth
        _user$.value = null;
        _token$.value = null;
        _authLoading$.value = false;
        _authError$.value = null;

        // Projects
        _projects$.value = [];
        _currentProjectId$.value = null;
        _currentProjectDetail$.value = null;
        _projectLoading$.value = false;
        _projectError$.value = null;
        _activeProjectId$.value = null;

        // Threads
        _threads$.value = [];
        _activeThreadId$.value = null;
        _threadLoading$.value = false;

        // Chat
        _messages$.value = [];
        _chatLoading$.value = false;
        _chatError$.value = null;

        // Focus & Activity
        _focusTasks$.value = [];
        _agentActivity$.value = [];

        // Uploads
        _pendingFiles$.value = [];
        _isDragging$.value = false;

        // Realtime
        _isTyping$.value = false;
        _connectionStatus$.value = 'disconnected';
        _activeArtifact$.value = null;

        // UI panels (close them)
        _rightPanelOpen$.value = false;
        _leftPanelOpen$.value = false;
        _rightPanelWidth$.value = 320; // Step 59.5: Reset width
        _popoutArtifact$.value = null; // Step 59.5: Close modal

        // Note: localStorage cleared automatically by token effect
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

    // ---- Upload State (readonly, Step 56) ----
    isDragging$: _isDragging$ as ReadonlySignal<boolean>,
    pendingFiles$: _pendingFiles$ as ReadonlySignal<PendingUpload[]>,

    // ---- Realtime State (readonly, Step 57) ----
    isTyping$: _isTyping$ as ReadonlySignal<boolean>,
    connectionStatus$: _connectionStatus$ as ReadonlySignal<ConnectionStatus>,
    activeArtifact$: _activeArtifact$ as ReadonlySignal<ArtifactPayload | null>,

    // ---- Panel Resize State (readonly, Step 59.5) ----
    rightPanelWidth$: _rightPanelWidth$ as ReadonlySignal<number>,
    popoutArtifact$: _popoutArtifact$ as ReadonlySignal<ArtifactPayload | null>,

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

    // Step 57: Initialize RealtimeService and wire event handlers
    // See FRONTEND_SCOPE.md Section 8.4 (Streaming Response Handler)
    realtimeService.on('message', (payload) => {
        const firstArtifact = payload.artifacts?.[0];
        const message: ChatMessage = {
            id: payload.id,
            role: payload.role,
            content: payload.content,
            createdAt: payload.createdAt,
        };
        // Map artifact if present
        if (firstArtifact) {
            message.artifactRef = {
                id: firstArtifact.id,
                type: normalizeArtifactType(firstArtifact.type),
                title: firstArtifact.title,
                scope: 'thread',
            };
        }
        actions.addMessage(message);

        // L7 Update: Set active artifact so it renders in the right panel
        if (firstArtifact) {
            actions.setActiveArtifact(firstArtifact);
        }
    });

    realtimeService.on('typing', (payload) => {
        actions.setTyping(payload.isTyping);
    });

    realtimeService.on('connection_change', (payload) => {
        actions.setConnectionStatus(payload.status);
    });

    // Step 57 L7 Fix: Handle error events for user feedback
    realtimeService.on('error', (payload) => {
        console.error('[Realtime Error]', payload.code, payload.message);
        // Future: Could dispatch to a global error notification system
    });

    // Connect to realtime service
    realtimeService.connect();
}

// mapArtifactType removed - use normalizeArtifactType from artifact-helpers.ts

// ============================================================================
// Type Exports
// ============================================================================

export type { User, ProjectSummary, ProjectDetail, ChatMessage, Theme, StoreActions, Thread, AgentActivity, FocusTask, ConnectionStatus, ArtifactPayload };
