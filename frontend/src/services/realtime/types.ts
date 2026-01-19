/**
 * Realtime Service Types - Event Payloads & Discriminated Unions
 * See FRONTEND_SCOPE.md Section 8.3 (Artifact Mapping) and 8.4 (Streaming)
 *
 * These types define the strict contract between the RealtimeService and the Store.
 * All communication flows through typed events - no magic strings.
 */

import { ArtifactType } from '../../types/enums';

// ============================================================================
// Connection State
// ============================================================================

/**
 * Connection lifecycle states.
 * UI can display connection status indicator based on this.
 */
export type ConnectionStatus = 'connected' | 'disconnected' | 'connecting';

// ============================================================================
// Artifact Payloads (Server → Client)
// ============================================================================

/**
 * Artifact payload matching FRONTEND_SCOPE.md Section 8.3.
 * Uses shared ArtifactType enum to prevent drift with Go backend.
 */
export interface ArtifactPayload {
    /** Artifact type - must match ArtifactType enum exactly */
    type: ArtifactType;
    /** Unique identifier for this artifact instance */
    id: string;
    /** Display title for the artifact */
    title: string;
    /** The structured data for the artifact renderer */
    data: Record<string, unknown>;
}

// ============================================================================
// Server → Client Events (Discriminated Union)
// ============================================================================

/**
 * Incoming message from the AI agent.
 * May include text content and/or artifacts.
 */
export interface IncomingMessageEvent {
    type: 'message';
    payload: {
        id: string;
        role: 'assistant';
        /** Markdown-formatted text content */
        content: string;
        /** Optional artifacts (Invoice, Gantt, Budget) */
        artifacts?: ArtifactPayload[];
        createdAt: string;
    };
}

/**
 * Typing indicator state change.
 */
export interface TypingEvent {
    type: 'typing';
    payload: {
        isTyping: boolean;
    };
}

/**
 * Acknowledgment that server received our message.
 */
export interface AckEvent {
    type: 'ack';
    payload: {
        messageId: string;
        timestamp: string;
    };
}

/**
 * Error event from the server.
 */
export interface ErrorEvent {
    type: 'error';
    payload: {
        code: string;
        message: string;
    };
}

/**
 * Connection state change notification.
 */
export interface ConnectionChangeEvent {
    type: 'connection_change';
    payload: {
        status: ConnectionStatus;
    };
}

/**
 * Union of all possible server events.
 * Use discriminated union pattern for type-safe handling.
 */
export type ServerEvent =
    | IncomingMessageEvent
    | TypingEvent
    | AckEvent
    | ErrorEvent
    | ConnectionChangeEvent;

// ============================================================================
// Client → Server Messages
// ============================================================================

/**
 * Chat message sent by the user.
 */
export interface SendChatMessage {
    type: 'chat';
    payload: {
        content: string;
        threadId?: string;
        projectId?: string;
    };
}

/**
 * File upload notification.
 */
export interface SendFileMessage {
    type: 'file';
    payload: {
        fileName: string;
        fileType: string;
        /** Base64 encoded file content (for small files) or upload reference */
        data?: string;
    };
}

/**
 * Union of all outgoing message types.
 */
export type OutgoingMessage = SendChatMessage | SendFileMessage;

// ============================================================================
// Type Guards
// ============================================================================

/**
 * Type guard for ServerEvent discrimination.
 */
export function isServerEventType<T extends ServerEvent['type']>(
    event: ServerEvent,
    type: T
): event is Extract<ServerEvent, { type: T }> {
    return event.type === type;
}
