/**
 * FBViewChat - AI Chat Interface View
 * See STEP_72_CHAT_VIEW.md
 *
 * The primary conversational interface for FutureBuild.
 * Integrates fb-message-list and fb-input-bar, wires store signals,
 * loads chat history, and handles optimistic message sending.
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBViewElement } from '../base/FBViewElement';
import { store } from '../../store/store';
import { api } from '../../services/api';
import type { ChatMessage } from '../../store/types';
import type { ConnectionStatus } from '../../services/realtime';

import '../chat/fb-message-list';
import '../chat/fb-input-bar';

/**
 * Formats an ISO timestamp to a display time string (e.g., "2:30 PM").
 */
function formatDisplayTime(isoString: string): string {
    try {
        return new Date(isoString).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch {
        return '';
    }
}

@customElement('fb-view-chat')
export class FBViewChat extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                display: flex;
                flex-direction: column;
                overflow: hidden;
            }

            .chat-header {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm);
                padding: var(--fb-spacing-sm) var(--fb-spacing-lg);
                border-bottom: 1px solid var(--fb-border-light);
                flex-shrink: 0;
            }

            .header-title {
                font-size: var(--fb-text-sm);
                font-weight: 500;
                color: var(--fb-text-primary);
            }

            .connection-dot {
                width: 8px;
                height: 8px;
                border-radius: 50%;
                flex-shrink: 0;
            }

            .connection-dot.connected {
                background: var(--fb-success, #22c55e);
            }

            .connection-dot.connecting {
                background: var(--fb-warning, #f59e0b);
                animation: pulse 1.5s ease-in-out infinite;
            }

            .connection-dot.disconnected {
                background: var(--fb-text-muted);
            }

            @keyframes pulse {
                0%, 100% { opacity: 1; }
                50% { opacity: 0.4; }
            }

            fb-message-list {
                flex: 1;
                min-height: 0;
            }

            .error-banner {
                padding: var(--fb-spacing-sm) var(--fb-spacing-lg);
                background: var(--fb-error-bg, #fef2f2);
                color: var(--fb-error, #ef4444);
                font-size: var(--fb-text-sm);
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm);
                flex-shrink: 0;
            }

            .error-dismiss {
                margin-left: auto;
                background: none;
                border: none;
                color: inherit;
                cursor: pointer;
                padding: var(--fb-spacing-xs);
                font-size: var(--fb-text-sm);
            }
        `,
    ];

    @state() private _isLoading = false;
    @state() private _error: string | null = null;
    @state() private _connectionStatus: ConnectionStatus = 'disconnected';
    @state() private _projectId: string | null = null;

    private _disposeEffects: (() => void)[] = [];

    override connectedCallback(): void {
        super.connectedCallback();

        this._disposeEffects.push(
            effect(() => {
                this._isLoading = store.chatLoading$.value;
            }),
            effect(() => {
                this._error = store.chatError$.value;
            }),
            effect(() => {
                this._connectionStatus = store.connectionStatus$.value;
            }),
            effect(() => {
                const projectId = store.activeProjectId$.value;
                if (projectId && projectId !== this._projectId) {
                    this._projectId = projectId;
                    void this._loadHistory(projectId);
                } else if (!projectId) {
                    this._projectId = null;
                }
            })
        );
    }

    override disconnectedCallback(): void {
        this._disposeEffects.forEach((d) => { d(); });
        this._disposeEffects = [];
        super.disconnectedCallback();
    }

    override onViewActive(): void {
        if (this._projectId) {
            void this._loadHistory(this._projectId);
        }
    }

    private async _loadHistory(projectId: string): Promise<void> {
        store.actions.setChatLoading(true);
        try {
            const history = await api.chat.history(projectId);
            const messages: ChatMessage[] = history.map((msg) => ({
                id: msg.id,
                role: msg.role,
                content: msg.content,
                createdAt: msg.created_at,
                displayTime: formatDisplayTime(msg.created_at),
            }));
            store.actions.setMessages(messages);
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to load chat history';
            store.actions.setChatError(message);
        }
    }

    private async _handleSend(e: CustomEvent<{ content: string }>): Promise<void> {
        const content = e.detail.content.trim();
        if (!content) return;

        if (!this._projectId) {
            store.actions.setChatError('No active project selected');
            return;
        }

        const createdAt = new Date().toISOString();
        const optimisticId = `msg-${String(Date.now())}-${crypto.randomUUID()}`;

        // Optimistic user message
        const message: ChatMessage = {
            id: optimisticId,
            role: 'user',
            content,
            createdAt,
            displayTime: formatDisplayTime(createdAt),
        };

        store.actions.addMessage(message);
        store.actions.setChatLoading(true);

        try {
            // REST API call (spec-compliant)
            const response = await api.chat.send(this._projectId, content);

            // Replace optimistic message with server-confirmed message
            store.actions.removeMessage(optimisticId);
            store.actions.addMessage({
                id: response.id,
                role: response.role,
                content: response.content,
                createdAt: response.created_at,
                displayTime: formatDisplayTime(response.created_at),
            });

        } catch (err) {
            // Rollback optimistic message
            store.actions.removeMessage(optimisticId);

            // User-friendly error
            const errorMsg = err instanceof Error
                ? err.message
                : 'Failed to send message. Please try again.';
            store.actions.setChatError(errorMsg);

            console.error('[FBViewChat] Send failed:', err);
        } finally {
            store.actions.setChatLoading(false);
        }
    }

    private _dismissError(): void {
        store.actions.setChatError(null);
    }

    override render(): TemplateResult {
        return html`
            <div class="chat-header">
                <span
                    class="connection-dot ${this._connectionStatus}"
                    title="Connection: ${this._connectionStatus}"
                    aria-label="Connection status: ${this._connectionStatus}"
                ></span>
                <span class="header-title">AI Assistant</span>
            </div>

            ${this._error ? html`
                <div class="error-banner" role="alert">
                    <span>${this._error}</span>
                    <button
                        class="error-dismiss"
                        @click=${this._dismissError.bind(this)}
                        aria-label="Dismiss error"
                    >&times;</button>
                </div>
            ` : nothing}

            <fb-message-list></fb-message-list>
            <fb-input-bar
                ?disabled=${this._isLoading}
                @send=${this._handleSend.bind(this)}
            ></fb-input-bar>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-chat': FBViewChat;
    }
}
