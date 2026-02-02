/**
 * Mock Realtime Service - Development Simulation Layer
 * See FRONTEND_SCOPE.md Section 8.4 (Streaming Response Handler)
 *
 * Provides a fully functional mock of the WebSocket backend for local development.
 * Features:
 * - Configurable latency simulation
 * - Scenario registry for complex flow testing
 * - DevTools hook (window.fb) for manual triggering
 * - Console logging for "network" traffic debugging
 */

import { ArtifactType } from '../../types/enums';
import type {
    ServerEvent,
    OutgoingMessage,
    ConnectionStatus,
    IncomingMessageEvent,
} from './types';
import type { IRealtimeServiceDevTools, EventHandler } from './interfaces';
import { MOCK_INVOICE_DATA } from '../../fixtures/invoice';
import { MOCK_BUDGET_DATA } from '../../fixtures/budget';
import { MOCK_GANTT_DATA } from '../../fixtures/gantt';

// ============================================================================
// Scenario Registry
// ============================================================================

/**
 * Predefined scenarios for testing complex flows.
 * Each scenario returns a sequence of events to emit.
 */
interface Scenario {
    /** Human-readable description */
    description: string;
    /** Delay before emitting events (ms) */
    delay: number;
    /** Factory function that generates events with fresh IDs */
    createEvents: () => ServerEvent[];
}

/**
 * Generate a unique ID for messages.
 */
function generateId(): string {
    return `msg-${String(Date.now())}-${Math.random().toString(36).slice(2, 9)}`;
}

/**
 * Registry of canned scenarios for testing.
 * Uses factory functions to generate fresh IDs on each invocation.
 */
const SCENARIOS: Record<string, Scenario> = {
    text_reply: {
        description: 'Simple text response',
        delay: 800,
        createEvents: () => [
            {
                type: 'message',
                payload: {
                    id: generateId(),
                    role: 'assistant',
                    content: 'I received your message. How can I help you further?',
                    createdAt: new Date().toISOString(),
                },
            },
        ],
    },
    invoice_success: {
        description: 'Invoice extraction with artifact',
        delay: 2000,
        createEvents: () => [
            {
                type: 'message',
                payload: {
                    id: generateId(),
                    role: 'assistant',
                    content:
                        "I've analyzed your invoice and extracted the following details. Please review and approve.",
                    artifacts: [
                        {
                            type: ArtifactType.Invoice,
                            id: `artifact-${String(Date.now())}`,
                            title: 'Extracted Invoice',
                            data: MOCK_INVOICE_DATA,
                        },
                    ],
                    createdAt: new Date().toISOString(),
                },
            },
        ],
    },
    budget_overview: {
        description: 'Budget artifact response',
        delay: 1500,
        createEvents: () => [
            {
                type: 'message',
                payload: {
                    id: generateId(),
                    role: 'assistant',
                    content: "Here's the current budget overview for your project:",
                    artifacts: [
                        {
                            type: ArtifactType.BudgetView,
                            id: `artifact-${String(Date.now())}`,
                            title: 'Project Budget',
                            data: MOCK_BUDGET_DATA,
                        },
                    ],
                    createdAt: new Date().toISOString(),
                },
            },
        ],
    },
    schedule_change: {
        description: 'Gantt chart schedule response',
        delay: 1800,
        createEvents: () => [
            {
                type: 'message',
                payload: {
                    id: generateId(),
                    role: 'assistant',
                    content: "I've updated the project schedule based on the latest delays. Here is the revised Gantt view:",
                    artifacts: [
                        {
                            type: ArtifactType.GanttView,
                            id: `artifact-${String(Date.now())}`,
                            title: 'Updated Schedule',
                            data: MOCK_GANTT_DATA,
                        },
                    ],
                    createdAt: new Date().toISOString(),
                },
            },
        ],
    },
    typing_long: {
        description: 'Extended typing simulation (5s)',
        delay: 5000,
        createEvents: () => [
            {
                type: 'message',
                payload: {
                    id: generateId(),
                    role: 'assistant',
                    content: 'This response took a while to generate. Complex analysis complete!',
                    createdAt: new Date().toISOString(),
                },
            },
        ],
    },
    error_network: {
        description: 'Simulated network error',
        delay: 1000,
        createEvents: () => [
            {
                type: 'error',
                payload: {
                    code: 'NETWORK_ERROR',
                    message: 'Connection lost. Attempting to reconnect...',
                },
            },
            {
                type: 'connection_change',
                payload: { status: 'disconnected' },
            },
        ],
    },
};

// ============================================================================
// Mock Realtime Service Implementation
// ============================================================================

/**
 * MockRealtimeService - Simulates WebSocket behavior for development.
 *
 * Exposes DevTools hooks via window.fb for manual testing.
 */
export class MockRealtimeService implements IRealtimeServiceDevTools {
    private _status: ConnectionStatus = 'disconnected';
    private _handlers: Map<ServerEvent['type'], Set<EventHandler<ServerEvent['type']>>> = new Map();
    private _pendingTimeouts: Set<ReturnType<typeof setTimeout>> = new Set();

    constructor() {
        // Expose to window.fb for DevTools access (dev only)
        if ((import.meta as unknown as { env?: { DEV?: boolean } }).env?.DEV === true) {
            this._exposeDevTools();
        }
    }

    // ---- IRealtimeService Implementation ----

    get status(): ConnectionStatus {
        return this._status;
    }

    connect(): void {
        this._log('connect', 'Initiating connection...');
        this._setStatus('connecting');

        // Simulate connection handshake
        const timeout = setTimeout(() => {
            this._setStatus('connected');
            this._log('connect', 'Connected successfully');
        }, 300);
        this._pendingTimeouts.add(timeout);
    }

    disconnect(): void {
        this._log('disconnect', 'Disconnecting...');
        this._clearPendingTimeouts();
        this._setStatus('disconnected');
    }

    send(message: OutgoingMessage): void {
        this._log('send', message);

        // Emit acknowledgment
        const ackTimeout = setTimeout(() => {
            this._emit('ack', {
                messageId: generateId(),
                timestamp: new Date().toISOString(),
            });
        }, 50);
        this._pendingTimeouts.add(ackTimeout);

        // Auto-respond based on message type
        this._handleOutgoingMessage(message);
    }

    on<T extends ServerEvent['type']>(type: T, handler: EventHandler<T>): () => void {
        if (!this._handlers.has(type)) {
            this._handlers.set(type, new Set());
        }
        const handlers = this._handlers.get(type);
        if (!handlers) {
            return () => { /* no-op */ };
        }
        // Cast through unknown to handle discriminated union type variance
        handlers.add(handler as unknown as EventHandler<ServerEvent['type']>);

        // Return unsubscribe function
        return () => {
            handlers.delete(handler as unknown as EventHandler<ServerEvent['type']>);
        };
    }

    off(type: ServerEvent['type']): void {
        this._handlers.delete(type);
    }

    // ---- IRealtimeServiceDevTools Implementation ----

    triggerMessage(text: string): void {
        this._log('triggerMessage', text);
        this._emitMessage({
            id: generateId(),
            role: 'assistant',
            content: text,
            createdAt: new Date().toISOString(),
        });
    }

    triggerScenario(scenarioId: string): void {
        const scenario = SCENARIOS[scenarioId];
        if (!scenario) {
            this._log('triggerScenario', `Unknown scenario: ${scenarioId}. Available: ${Object.keys(SCENARIOS).join(', ')}`);
            return;
        }

        this._log('triggerScenario', { scenarioId, description: scenario.description });

        // Emit typing indicator
        this._emit('typing', { isTyping: true });

        // Emit events after delay (generate fresh IDs)
        const timeout = setTimeout(() => {
            this._emit('typing', { isTyping: false });
            scenario.createEvents().forEach((event) => {
                this._emitEvent(event);
            });
        }, scenario.delay);
        this._pendingTimeouts.add(timeout);
    }

    setTyping(isTyping: boolean): void {
        this._log('setTyping', isTyping);
        this._emit('typing', { isTyping });
    }

    getAvailableScenarios(): string[] {
        return Object.keys(SCENARIOS);
    }

    // ---- Private Methods ----

    private _setStatus(status: ConnectionStatus): void {
        this._status = status;
        this._emit('connection_change', { status });
    }

    private _emit<T extends ServerEvent['type']>(
        type: T,
        payload: Extract<ServerEvent, { type: T }>['payload']
    ): void {
        const handlers = this._handlers.get(type);
        if (handlers) {
            handlers.forEach((handler) => {
                (handler as (p: typeof payload) => void)(payload);
            });
        }
    }

    private _emitEvent(event: ServerEvent): void {
        this._emit(event.type, event.payload as never);
    }

    private _emitMessage(payload: IncomingMessageEvent['payload']): void {
        this._emit('message', payload);
    }

    private _handleOutgoingMessage(message: OutgoingMessage): void {
        // Emit typing indicator
        this._emit('typing', { isTyping: true });

        // Determine scenario based on message content
        let scenarioId = 'text_reply';
        if (message.type === 'file') {
            scenarioId = 'invoice_success';
        } else {
            const content = message.payload.content.toLowerCase();
            if (content.includes('budget') || content.includes('cost')) {
                scenarioId = 'budget_overview';
            } else if (content.includes('invoice')) {
                scenarioId = 'invoice_success';
            }
        }

        const scenario = SCENARIOS[scenarioId];
        if (!scenario) {
            this.triggerScenario('text_reply');
            return;
        }
        const timeout = setTimeout(() => {
            this._emit('typing', { isTyping: false });
            scenario.createEvents().forEach((event) => {
                this._emitEvent(event);
            });
        }, scenario.delay);
        this._pendingTimeouts.add(timeout);
    }

    private _clearPendingTimeouts(): void {
        this._pendingTimeouts.forEach((t) => { clearTimeout(t); });
        this._pendingTimeouts.clear();
    }

    private _log(method: string, data: unknown): void {
        if ((import.meta as unknown as { env?: { DEV?: boolean } }).env?.DEV !== true) return;
        console.groupCollapsed(`[MockRealtime] ${method}`);
        console.log(data);
        console.groupEnd();
    }

    private _exposeDevTools(): void {
        // Import store for resetSession access (Step 58.5)
        // Lazy import to avoid circular dependencies
        import('../../store/store').then(({ store }) => {
            // Expose to window.fb for browser console access
            interface FBDevTools {
                triggerMessage: (text: string) => void;
                triggerScenario: (id: string) => void;
                setTyping: (isTyping: boolean) => void;
                getScenarios: () => string[];
                connect: () => void;
                disconnect: () => void;
                store: {
                    resetSession: () => void;
                };
            }
            const devTools: FBDevTools = {
                triggerMessage: this.triggerMessage.bind(this),
                triggerScenario: this.triggerScenario.bind(this),
                setTyping: this.setTyping.bind(this),
                getScenarios: this.getAvailableScenarios.bind(this),
                connect: this.connect.bind(this),
                disconnect: this.disconnect.bind(this),
                store: {
                    resetSession: store.actions.resetSession.bind(store.actions),
                },
            };
            (window as unknown as { fb: FBDevTools }).fb = devTools;
            console.log('[MockRealtime] DevTools exposed at window.fb');
            console.log('[MockRealtime] Available commands: triggerMessage, triggerScenario, setTyping, getScenarios, store.resetSession');
        }).catch(() => {
            console.warn('[MockRealtime] Could not load store for DevTools');
        });
    }
}

// ============================================================================
// Singleton Export
// ============================================================================

/**
 * Singleton instance of the realtime service.
 * Use this throughout the application.
 */
export const realtimeService = new MockRealtimeService();
