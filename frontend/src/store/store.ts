/**
 * Global State Store - Signals-Based Reactive State
 * See FRONTEND_SCOPE.md Section 5.1 (Updated for Agent Command Center)
 *
 * Implements a centralized state store using @preact/signals-core.
 * Key patterns:
 * - Readonly signal exposure to prevent arbitrary mutation
 * - Action-based mutations for predictable state changes
 * - Clerk-managed auth (Phase 12: no localStorage for auth)
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
    ContextScope,
    ContextState,
} from './types';
import type { CompletionReport } from '../types/models';
import { setTokenGetter, setUnauthorizedHandler } from '../services/http';
import { realtimeService, type ConnectionStatus, type ArtifactPayload } from '../services/realtime';
import { normalizeArtifactType } from '../utils/artifact-helpers';
import { clerkService } from '../services/clerk';

/**
 * Formats an ISO timestamp to a display time string (e.g., "2:30 PM").
 * Pre-computing this avoids Date instantiation inside render loops.
 * @see Step 60.1 - Performance Hygiene
 */
function formatDisplayTime(isoString: string): string {
    try {
        return new Date(isoString).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch {
        return '';
    }
}

// ============================================================================
// Constants
// ============================================================================

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

// UI (V2: No left panel, right panel only shows when artifact active)
const _leftPanelOpen$ = signal<boolean>(false);
const _rightPanelOpen$ = signal<boolean>(false);
const _theme$ = signal<Theme>('dark');
const _isMobile$ = signal<boolean>(false);
const _isTablet$ = signal<boolean>(false);
const _activeProjectId$ = signal<string | null>(null);

// Context state (Sprint 1.1: Context Spine)
const _contextScope$ = signal<ContextScope>('global');
const _contextProjectId$ = signal<string | null>(null);

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

// Completion report state
const _completionReport$ = signal<CompletionReport | null>(null);

// Shadow Mode state (SHADOW_VIEWER_specs.md)
const _shadowModeEnabled$ = signal<boolean>(false);
const _shadowActiveView$ = signal<'log' | 'docs'>('log');
const _selectedDecisionId$ = signal<string | null>(null);
const _selectedDocPath$ = signal<string | null>(null);

// V2 Phase 5: Chat context from feed card "Tell me more" flow
// See FRONTEND_V2_SPEC.md §7 Step 33
export interface ChatCardContext {
    cardId: string;
    cardType: string;
    headline: string;
    body: string;
    consequence: string;
    projectId: string;
    projectName: string;
    taskId: string;
}
const _chatCardContext$ = signal<ChatCardContext | null>(null);

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
 * Computed: The full context state (Sprint 1.1).
 */
const _contextState$ = computed<ContextState>(() => ({
    scope: _contextScope$.value,
    projectId: _contextProjectId$.value,
}));

/**
 * Computed: The currently active thread.
 */
const _activeThread$ = computed(() => {
    const id = _activeThreadId$.value;
    if (!id) return null;
    return _threads$.value.find((t) => t.id === id) ?? null;
});

/**
 * Computed: Active (non-archived) threads for the currently active project.
 */
const _projectThreads$ = computed(() => {
    const projectId = _activeProjectId$.value;
    if (!projectId) return [];
    return _threads$.value.filter((t) => t.projectId === projectId && !t.archivedAt);
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
        // Phase 12: Token persistence is managed by Clerk SDK (cookies/session storage).
        // No localStorage writes for auth data.
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
        // Phase 12: No localStorage auth cleanup needed — Clerk manages session.
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

    // ---- Context Actions (Sprint 1.1: Context Spine) ----

    setContext(scope: ContextScope, projectId: string | null): void {
        _contextScope$.value = scope;
        _contextProjectId$.value = projectId;
        // Sync the legacy activeProjectId for backward compat
        _activeProjectId$.value = projectId;
    },

    clearContext(): void {
        _contextScope$.value = 'global';
        _contextProjectId$.value = null;
        _activeProjectId$.value = null;
    },

    selectProject(id: string | null): void {
        _currentProjectId$.value = id;
        _activeProjectId$.value = id;
        // Sprint 1.1: Sync context state
        _contextScope$.value = id ? 'project' : 'global';
        _contextProjectId$.value = id;
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

    archiveThread(id: string): void {
        _threads$.value = _threads$.value.map((t) =>
            t.id === id ? { ...t, archivedAt: new Date().toISOString() } : t
        );
        // If the archived thread was active, deselect it
        if (_activeThreadId$.value === id) {
            _activeThreadId$.value = null;
            _messages$.value = [];
        }
    },

    unarchiveThread(id: string, thread: Thread): void {
        // If thread is already in list, clear archivedAt. Otherwise add it.
        const exists = _threads$.value.some((t) => t.id === id);
        if (exists) {
            _threads$.value = _threads$.value.map((t) => {
                if (t.id !== id) return t;
                // Remove archivedAt by destructuring it out
                const { archivedAt: _removed, ...rest } = t;
                void _removed; // Intentionally unused - destructuring to remove property
                return rest as Thread;
            });
        } else {
            _threads$.value = [..._threads$.value, thread];
        }
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

    /**
     * Remove a message by ID (for optimistic rollback).
     */
    removeMessage(id: string): void {
        _messages$.value = _messages$.value.filter((m) => m.id !== id);
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
        // Sprint 1.1: Sync context state
        _contextScope$.value = projectId ? 'project' : 'global';
        _contextProjectId$.value = projectId;
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
        const createdAt = new Date().toISOString();
        const optimisticId = crypto.randomUUID();
        actions.addMessage({
            id: optimisticId,
            role: 'user',
            content: `📎 Uploaded: ${fileNames}`,
            createdAt,
            displayTime: formatDisplayTime(createdAt),
        });

        // TODO (Phase 10 Remediation): File upload currently uses WebSocket (fire-and-forget).
        // This creates silent failure risk - if upload fails, the optimistic message persists.
        // Required fix:
        // 1. Add REST API endpoint: POST /api/chat/upload or POST /api/files/upload (multipart/form-data)
        // 2. Replace WebSocket call with await api.files.upload(projectId, file)
        // 3. Add error handling with rollback (remove optimistic message on failure)
        // 4. Add progress tracking via _pendingFiles$ status updates
        // 5. Add 5-second timeout to mark upload as failed if no acknowledgment
        // See specs/phase10/STEP_73_DRAG_DROP.md Section 2.1
        //
        // Step 57: Send file message via RealtimeService (replaces setTimeout)
        // See FRONTEND_SCOPE.md Section 8.4
        const firstUpload = newUploads[0];
        try {
            realtimeService.send({
                type: 'file',
                payload: {
                    fileName: fileNames,
                    fileType: firstUpload?.file.type ?? 'application/octet-stream',
                },
            });
        } catch (err) {
            // Rollback the optimistic message on send failure
            actions.removeMessage(optimisticId);
            console.error('[Store] File upload send failed:', err);
        }
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

    // ---- Shadow Mode Actions (SHADOW_VIEWER_specs.md) ----

    toggleShadowMode(): void {
        _shadowModeEnabled$.value = !_shadowModeEnabled$.value;
        // Reset shadow state when disabling
        if (!_shadowModeEnabled$.value) {
            _shadowActiveView$.value = 'log';
            _selectedDecisionId$.value = null;
            _selectedDocPath$.value = null;
        }
    },

    setShadowModeEnabled(enabled: boolean): void {
        _shadowModeEnabled$.value = enabled;
        if (!enabled) {
            _shadowActiveView$.value = 'log';
            _selectedDecisionId$.value = null;
            _selectedDocPath$.value = null;
        }
    },

    setShadowActiveView(view: 'log' | 'docs'): void {
        _shadowActiveView$.value = view;
    },

    selectDecision(id: string | null): void {
        _selectedDecisionId$.value = id;
    },

    selectDoc(path: string | null): void {
        _selectedDocPath$.value = path;
    },

    exitShadowMode(): void {
        _shadowModeEnabled$.value = false;
        _shadowActiveView$.value = 'log';
        _selectedDecisionId$.value = null;
        _selectedDocPath$.value = null;
    },

    // ---- Completion Actions ----

    setCompletionReport(report: CompletionReport | null): void {
        _completionReport$.value = report;
    },

    // V2 Phase 5: Chat card context
    setChatCardContext(ctx: ChatCardContext | null): void {
        _chatCardContext$.value = ctx;
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

        // Context (Sprint 1.1)
        _contextScope$.value = 'global';
        _contextProjectId$.value = null;

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

        // Completion
        _completionReport$.value = null;

        // Chat card context
        _chatCardContext$.value = null;

        // Shadow Mode (SHADOW_VIEWER_specs.md)
        _shadowModeEnabled$.value = false;
        _shadowActiveView$.value = 'log';
        _selectedDecisionId$.value = null;
        _selectedDocPath$.value = null;

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

    // ---- Context State (readonly, Sprint 1.1) ----
    contextScope$: _contextScope$ as ReadonlySignal<ContextScope>,
    contextProjectId$: _contextProjectId$ as ReadonlySignal<string | null>,
    contextState$: _contextState$,

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

    // ---- Completion Report State (readonly) ----
    completionReport$: _completionReport$ as ReadonlySignal<CompletionReport | null>,

    // ---- Chat Card Context (V2 Phase 5: "Tell me more") ----
    chatCardContext$: _chatCardContext$ as ReadonlySignal<ChatCardContext | null>,

    // ---- Shadow Mode State (readonly, SHADOW_VIEWER_specs.md) ----
    shadowModeEnabled$: _shadowModeEnabled$ as ReadonlySignal<boolean>,
    shadowActiveView$: _shadowActiveView$ as ReadonlySignal<'log' | 'docs'>,
    selectedDecisionId$: _selectedDecisionId$ as ReadonlySignal<string | null>,
    selectedDocPath$: _selectedDocPath$ as ReadonlySignal<string | null>,

    // ---- Actions ----
    actions,
} as const;

// ============================================================================
// Initialization
// ============================================================================

// Module-level references for media query listeners so they can be cleaned up on re-init
let _mobileQuery: MediaQueryList | null = null;
let _tabletQuery: MediaQueryList | null = null;
let _mobileHandler: ((e: MediaQueryListEvent) => void) | null = null;
let _tabletHandler: ((e: MediaQueryListEvent) => void) | null = null;

/**
 * Initialize the store with persisted state and Clerk auth observer.
 * Call this once at application bootstrap, after Clerk has loaded.
 *
 * L7 Fix: Side-effect effects are registered here instead of at module level.
 * This prevents localStorage access during import (e.g., JSDOM tests).
 *
 * Phase 12: Auth state is driven by Clerk session observer, not localStorage.
 * Token persistence is managed by Clerk SDK (cookies/session storage).
 */
export function initializeStore(): void {
    // Effect: Persist theme preference to localStorage.
    effect(() => {
        const theme = _theme$.value;
        localStorage.setItem(STORAGE_KEY_THEME, theme);
    });

    // Restore theme from localStorage
    const storedTheme = localStorage.getItem(STORAGE_KEY_THEME) as Theme | null;
    if (storedTheme && ['light', 'dark', 'system'].includes(storedTheme)) {
        _theme$.value = storedTheme;
    }

    // Clean up legacy localStorage keys from magic-link auth
    localStorage.removeItem('fb_token');
    localStorage.removeItem('fb_user');

    // Detect mobile/tablet viewport
    if (typeof window !== 'undefined') {
        // Remove previous listeners if they exist (prevents leak on re-init)
        if (_mobileQuery && _mobileHandler) {
            _mobileQuery.removeEventListener('change', _mobileHandler);
        }
        if (_tabletQuery && _tabletHandler) {
            _tabletQuery.removeEventListener('change', _tabletHandler);
        }

        _mobileQuery = window.matchMedia('(max-width: 768px)');
        _tabletQuery = window.matchMedia('(max-width: 1024px)');

        _isMobile$.value = _mobileQuery.matches;
        _isTablet$.value = _tabletQuery.matches && !_mobileQuery.matches;

        // Auto-collapse panels on smaller screens
        if (_mobileQuery.matches) {
            _leftPanelOpen$.value = false;
            _rightPanelOpen$.value = false;
        } else if (_tabletQuery.matches) {
            _rightPanelOpen$.value = false;
        }

        _mobileHandler = (e: MediaQueryListEvent) => {
            actions.setIsMobile(e.matches);
        };
        _tabletHandler = (e: MediaQueryListEvent) => {
            actions.setIsTablet(e.matches && !(_mobileQuery?.matches ?? false));
        };

        _mobileQuery.addEventListener('change', _mobileHandler);
        _tabletQuery.addEventListener('change', _tabletHandler);
    }

    // Wire up HTTP service callbacks to use Clerk token cache
    // See STEP_78_AUTH_PROVIDER.md Section 1.4
    setTokenGetter(() => clerkService.getToken());
    setUnauthorizedHandler(() => {
        void clerkService.signOut();
    });

    // Wire Clerk auth state changes into the store
    // On sign-in: Clerk fires callback with user info and cached token
    // On sign-out: Clerk fires callback with (null, null)
    // Step 80: Detect org switch (same user, different org) and clear project data
    clerkService.onAuthChange((clerkUser, token) => {
        if (clerkUser && token) {
            const currentUser = _user$.value;
            const isOrgSwitch = currentUser !== null
                && currentUser.id === clerkUser.id
                && currentUser.orgId !== clerkUser.orgId;

            const user: User = {
                id: clerkUser.id,
                email: clerkUser.email,
                name: clerkUser.name,
                role: clerkUser.role as User['role'],
                orgId: clerkUser.orgId,
            };
            actions.login(user, token);

            // On org switch: clear project/thread/chat data but keep auth and UI state
            if (isOrgSwitch) {
                _projects$.value = [];
                _currentProjectId$.value = null;
                _currentProjectDetail$.value = null;
                _activeProjectId$.value = null;
                _threads$.value = [];
                _activeThreadId$.value = null;
                _messages$.value = [];
                _focusTasks$.value = [];
                _agentActivity$.value = [];
                _activeArtifact$.value = null;
            }
        } else {
            actions.logout();
        }
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
            displayTime: formatDisplayTime(payload.createdAt),
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

export type { User, ProjectSummary, ProjectDetail, ChatMessage, Theme, StoreActions, Thread, AgentActivity, FocusTask, ConnectionStatus, ArtifactPayload, CompletionReport };
