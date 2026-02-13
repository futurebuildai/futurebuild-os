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

import '../chat/fb-message-list';
import '../chat/fb-input-bar';
import '../chat/fb-chat-context-banner';
import { type ChatCardContext } from '../../store/store';

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

            .header-spacer {
                flex: 1;
            }

            .panel-toggle {
                padding: var(--fb-spacing-xs);
                border: none;
                background: transparent;
                color: var(--fb-text-muted);
                cursor: pointer;
                border-radius: var(--fb-radius-sm, 4px);
                flex-shrink: 0;
            }

            .panel-toggle:hover {
                background: var(--fb-bg-tertiary);
                color: var(--fb-text-primary);
            }

            .panel-toggle svg {
                width: 18px;
                height: 18px;
                stroke: currentColor;
                fill: none;
                stroke-width: 2;
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
    @state() private _projectId: string | null = null;
    @state() private _threadId: string | null = null;
    @state() private _isMobile = false;
    @state() private _bannerContext: ChatCardContext | null = null;

    private _disposeEffects: (() => void)[] = [];
    private _loadAbortController: AbortController | null = null;

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
                this._isMobile = store.isMobile$.value;
            }),
            effect(() => {
                this._projectId = store.activeProjectId$.value;
            }),
            effect(() => {
                const threadId = store.activeThreadId$.value;
                if (threadId && this._projectId && threadId !== this._threadId) {
                    this._threadId = threadId;
                    void this._loadHistory(this._projectId, threadId);
                } else if (!threadId) {
                    this._threadId = null;
                }
            }),
            // V2 Phase 5: Check for pending card context ("Tell me more")
            effect(() => {
                const ctx = store.chatCardContext$.value;
                if (ctx && this._threadId && this._projectId) {
                    this._bannerContext = ctx;
                    // Clear pending context from store (banner persists locally)
                    store.actions.setChatCardContext(null);
                    // Auto-send the "Tell me more" prompt
                    void this._autoSendContext(ctx);
                }
            })
        );
    }

    override disconnectedCallback(): void {
        this._disposeEffects.forEach((d) => { d(); });
        this._disposeEffects = [];
        this._loadAbortController?.abort();
        this._loadAbortController = null;
        super.disconnectedCallback();
    }

    override onViewActive(): void {
        if (this._projectId && this._threadId) {
            void this._loadHistory(this._projectId, this._threadId);
        }
    }

    private async _loadHistory(projectId: string, threadId: string): Promise<void> {
        // Abort any in-flight load to prevent race conditions when switching threads
        this._loadAbortController?.abort();
        const controller = new AbortController();
        this._loadAbortController = controller;

        store.actions.setChatLoading(true);
        try {
            const history = await api.chat.history(projectId, threadId);

            // If this load was aborted while awaiting, discard the stale result
            if (controller.signal.aborted) return;

            const messages: ChatMessage[] = history.map((msg) => ({
                id: msg.id,
                role: msg.role,
                content: msg.content,
                createdAt: msg.created_at,
                displayTime: formatDisplayTime(msg.created_at),
            }));
            store.actions.setMessages(messages);
        } catch (err) {
            // Ignore errors from aborted requests
            if (controller.signal.aborted) return;

            const message = err instanceof Error ? err.message : 'Failed to load chat history';
            store.actions.setChatError(message);
        }
    }

    private async _handleSend(e: CustomEvent<{ content: string }>): Promise<void> {
        const content = e.detail.content.trim();
        if (!content) return;

        if (!this._projectId || !this._threadId) {
            store.actions.setChatError('No active project or thread selected');
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
            const response = await api.chat.send(this._projectId, this._threadId, content);

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

    private async _autoSendContext(ctx: ChatCardContext): Promise<void> {
        if (!this._projectId || !this._threadId) return;

        const prompt = `Tell me more about: "${ctx.headline}"`;
        const createdAt = new Date().toISOString();
        const optimisticId = `msg-${String(Date.now())}-${crypto.randomUUID()}`;

        store.actions.addMessage({
            id: optimisticId,
            role: 'user',
            content: prompt,
            createdAt,
            displayTime: formatDisplayTime(createdAt),
        });

        store.actions.setChatLoading(true);

        try {
            const response = await api.chat.send(this._projectId, this._threadId, prompt);
            store.actions.removeMessage(optimisticId);
            store.actions.addMessage({
                id: response.id,
                role: response.role,
                content: response.content,
                createdAt: response.created_at,
                displayTime: formatDisplayTime(response.created_at),
            });
        } catch (err) {
            store.actions.removeMessage(optimisticId);
            const errorMsg = err instanceof Error ? err.message : 'Failed to send message.';
            store.actions.setChatError(errorMsg);
        } finally {
            store.actions.setChatLoading(false);
        }
    }

    private _dismissBanner(): void {
        this._bannerContext = null;
    }

    private _dismissError(): void {
        store.actions.setChatError(null);
    }

    override render(): TemplateResult {
        return html`
            <div class="chat-header">
                ${this._isMobile ? html`
                    <button
                        class="panel-toggle"
                        @click=${(): void => { store.actions.toggleLeftPanel(); }}
                        aria-label="Open navigation panel"
                    >
                        <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M3 12h18M3 6h18M3 18h18"/></svg>
                    </button>
                ` : nothing}
                <span class="header-spacer"></span>
                <button
                    class="panel-toggle"
                    @click=${(): void => { store.actions.toggleRightPanel(); }}
                    aria-label="Toggle artifacts panel"
                >
                    <svg viewBox="0 0 24 24" aria-hidden="true">
                        <rect x="3" y="3" width="18" height="18" rx="2"/>
                        <path d="M9 3v18"/>
                    </svg>
                </button>
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

            ${this._bannerContext ? html`
                <fb-chat-context-banner
                    .headline=${this._bannerContext.headline}
                    .consequence=${this._bannerContext.consequence ?? ''}
                    .cardType=${this._bannerContext.cardType}
                    @fb-banner-dismiss=${this._dismissBanner.bind(this)}
                ></fb-chat-context-banner>
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
