/**
 * Realtime Service Interface - Transport Abstraction
 * See FRONTEND_SCOPE.md Section 8.4 (Streaming Response Handler)
 *
 * This interface defines the contract that both MockRealtimeService (Phase 7)
 * and WebSocketRealtimeService (Phase 8) must implement.
 *
 * The Store depends ONLY on this interface, never on concrete implementations.
 */

import type { ServerEvent, OutgoingMessage, ConnectionStatus } from './types';

// ============================================================================
// Event Handler Types
// ============================================================================

/**
 * Extract the payload type for a specific event type.
 */
export type EventPayload<T extends ServerEvent['type']> = Extract<
    ServerEvent,
    { type: T }
>['payload'];

/**
 * Handler function for a specific event type.
 */
export type EventHandler<T extends ServerEvent['type']> = (
    payload: EventPayload<T>
) => void;

// ============================================================================
// Core Interface
// ============================================================================

/**
 * RealtimeService interface for WebSocket/SSE abstraction.
 *
 * All transport implementations (Mock, WebSocket, SSE) must implement this.
 * The Store binds to this interface, not concrete classes.
 */
export interface IRealtimeService {
    // ---- Lifecycle ----

    /**
     * Establish connection to the realtime backend.
     * Emits 'connection_change' events during the process.
     */
    connect(): void;

    /**
     * Gracefully disconnect from the realtime backend.
     */
    disconnect(): void;

    /**
     * Current connection status.
     */
    readonly status: ConnectionStatus;

    // ---- Messaging ----

    /**
     * Send a message to the backend.
     * @param message - The typed outgoing message
     */
    send(message: OutgoingMessage): void;

    // ---- Event Subscription ----

    /**
     * Subscribe to a specific event type.
     * @param type - The event type to listen for
     * @param handler - Callback invoked when event occurs
     * @returns Unsubscribe function
     */
    on<T extends ServerEvent['type']>(
        type: T,
        handler: EventHandler<T>
    ): () => void;

    /**
     * Unsubscribe all handlers for a specific event type.
     */
    off(type: ServerEvent['type']): void;
}

// ============================================================================
// DevTools Extension (Mock-Only)
// ============================================================================

/**
 * Extended interface for mock service with DevTools hooks.
 * Only MockRealtimeService implements this.
 */
export interface IRealtimeServiceDevTools extends IRealtimeService {
    /**
     * Manually trigger a text message response.
     * @param text - The message content
     */
    triggerMessage(text: string): void;

    /**
     * Trigger a predefined scenario for testing.
     * @param scenarioId - The scenario identifier
     */
    triggerScenario(scenarioId: string): void;

    /**
     * Manually set typing indicator state.
     * @param isTyping - Whether the agent is "typing"
     */
    setTyping(isTyping: boolean): void;

    /**
     * Get list of available scenario IDs.
     */
    getAvailableScenarios(): string[];
}
