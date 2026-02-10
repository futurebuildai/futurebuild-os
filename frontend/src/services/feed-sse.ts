/**
 * FeedSSEService — Server-Sent Events client for live feed card updates.
 * See FRONTEND_V2_SPEC.md §6.5 (Real-Time Feed Updates)
 *
 * Connects to GET /api/v1/portfolio/feed/stream and dispatches
 * card_added, card_updated, card_removed events to registered handlers.
 *
 * Auto-reconnects with exponential backoff on connection loss.
 */

import type { FeedCard, FeedSSEEvent } from '../types/feed';

// ============================================================================
// Types
// ============================================================================

type FeedSSEHandler = (event: FeedSSEEvent) => void;

// ============================================================================
// FeedSSEService
// ============================================================================

const SSE_ENDPOINT = '/api/v1/portfolio/feed/stream';
const MAX_RECONNECT_DELAY = 30_000; // 30s max
const BASE_RECONNECT_DELAY = 2_000; // 2s initial

class FeedSSEService {
    private _eventSource: EventSource | null = null;
    private _handlers: Set<FeedSSEHandler> = new Set();
    private _reconnectAttempt = 0;
    private _reconnectTimer: ReturnType<typeof setTimeout> | null = null;
    private _connected = false;

    /**
     * Subscribe to feed SSE events.
     * @returns Unsubscribe function
     */
    subscribe(handler: FeedSSEHandler): () => void {
        this._handlers.add(handler);
        return () => {
            this._handlers.delete(handler);
        };
    }

    /**
     * Connect to the SSE stream.
     * Safe to call multiple times — will no-op if already connected.
     */
    connect(): void {
        if (this._eventSource) return;

        try {
            this._eventSource = new EventSource(SSE_ENDPOINT);

            this._eventSource.onopen = () => {
                this._connected = true;
                this._reconnectAttempt = 0;
            };

            // Listen to named event types from the server
            this._eventSource.addEventListener('card_added', (e: MessageEvent) => {
                this._dispatch(this._parseCardEvent('card_added', e.data));
            });

            this._eventSource.addEventListener('card_updated', (e: MessageEvent) => {
                this._dispatch(this._parseCardEvent('card_updated', e.data));
            });

            this._eventSource.addEventListener('card_removed', (e: MessageEvent) => {
                this._dispatch(this._parseRemovedEvent(e.data));
            });

            this._eventSource.onerror = () => {
                this._connected = false;
                this._cleanup();
                this._scheduleReconnect();
            };
        } catch {
            this._scheduleReconnect();
        }
    }

    /**
     * Disconnect from the SSE stream.
     */
    disconnect(): void {
        if (this._reconnectTimer) {
            clearTimeout(this._reconnectTimer);
            this._reconnectTimer = null;
        }
        this._cleanup();
        this._reconnectAttempt = 0;
    }

    /** Whether the SSE connection is currently open */
    get connected(): boolean {
        return this._connected;
    }

    // ---- Private ----

    private _cleanup(): void {
        if (this._eventSource) {
            this._eventSource.close();
            this._eventSource = null;
        }
    }

    private _scheduleReconnect(): void {
        const delay = Math.min(
            BASE_RECONNECT_DELAY * Math.pow(2, this._reconnectAttempt),
            MAX_RECONNECT_DELAY,
        );
        this._reconnectAttempt++;
        this._reconnectTimer = setTimeout(() => {
            this._reconnectTimer = null;
            this.connect();
        }, delay);
    }

    private _parseCardEvent(type: 'card_added' | 'card_updated', data: string): FeedSSEEvent | null {
        try {
            const card = JSON.parse(data) as FeedCard;
            return { type, card };
        } catch {
            console.warn(`[FeedSSE] Failed to parse ${type} event`, data);
            return null;
        }
    }

    private _parseRemovedEvent(data: string): FeedSSEEvent | null {
        try {
            const parsed = JSON.parse(data) as { card_id: string };
            return { type: 'card_removed', card_id: parsed.card_id };
        } catch {
            console.warn('[FeedSSE] Failed to parse card_removed event', data);
            return null;
        }
    }

    private _dispatch(event: FeedSSEEvent | null): void {
        if (!event) return;
        for (const handler of this._handlers) {
            try {
                handler(event);
            } catch (err) {
                console.error('[FeedSSE] Handler error', err);
            }
        }
    }
}

/** Singleton feed SSE service */
export const feedSSE = new FeedSSEService();
