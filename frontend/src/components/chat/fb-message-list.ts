import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';
import type { ChatMessage } from '../../store/types';

import './fb-action-card';

/**
 * FBMessageList - Renders the chat message stream.
 * Subscribes to store.messages$ and auto-scrolls when new messages arrive.
 * @element fb-message-list
 */
@customElement('fb-message-list')
export class FBMessageList extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                flex: 1;
                overflow-y: auto;
                padding: var(--fb-spacing-lg);
                scroll-behavior: smooth;
            }

            .list {
                display: flex;
                flex-direction: column;
                gap: var(--fb-spacing-md);
                max-width: 800px;
                margin: 0 auto;
                padding-bottom: var(--fb-spacing-xl);
            }

            .message {
                display: flex;
                flex-direction: column;
                max-width: 85%;
                animation: fadeIn 0.3s ease;
            }

            @keyframes fadeIn {
                from { opacity: 0; transform: translateY(10px); }
                to { opacity: 1; transform: translateY(0); }
            }

            .message-content {
                padding: var(--fb-spacing-md);
                border-radius: var(--fb-radius-lg);
                font-size: var(--fb-text-sm);
                line-height: 1.6;
                position: relative;
                word-wrap: break-word;
            }

            /* User Styles */
            .message.user {
                align-self: flex-end;
                align-items: flex-end;
            }

            .message.user .message-content {
                background: var(--fb-primary);
                color: white;
                border-bottom-right-radius: 2px;
            }

            /* Assistant Styles */
            .message.assistant {
                align-self: flex-start;
                align-items: flex-start;
            }

            .message.assistant .message-content {
                background: var(--fb-bg-tertiary);
                color: var(--fb-text-primary);
                border-bottom-left-radius: 2px;
            }

            /* Metadata */
            .meta {
                font-size: 11px;
                color: var(--fb-text-muted);
                margin-top: 4px;
                opacity: 0.7;
            }

            /* Empty State */
            .empty-state {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                color: var(--fb-text-muted);
                text-align: center;
                padding: var(--fb-spacing-2xl) 0;
                height: 100%;
            }

            .empty-icon {
                font-size: 48px;
                margin-bottom: var(--fb-spacing-md);
                opacity: 0.5;
            }

            /* Screen reader only */
            .sr-only {
                position: absolute;
                width: 1px;
                height: 1px;
                padding: 0;
                margin: -1px;
                overflow: hidden;
                clip: rect(0, 0, 0, 0);
                white-space: nowrap;
                border: 0;
            }
        `
    ];

    /** Threshold in pixels to consider "near bottom" for auto-scroll */
    private static readonly SCROLL_THRESHOLD = 100;

    @state() private _messages: ChatMessage[] = [];
    private _disposeEffects: (() => void)[] = [];

    override connectedCallback(): void {
        super.connectedCallback();

        this._disposeEffects.push(
            effect(() => {
                this._messages = store.messages$.value;
                this._scrollToBottomIfNearEnd();
            })
        );
    }

    override disconnectedCallback(): void {
        this._disposeEffects.forEach((d) => { d(); });
        this._disposeEffects = [];
        super.disconnectedCallback();
    }

    /**
     * Scrolls to the bottom only if the user is already near the bottom.
     * This prevents disrupting users who have scrolled up to read history.
     */
    private _scrollToBottomIfNearEnd(): void {
        requestAnimationFrame(() => {
            const host = this as HTMLElement;
            const isNearBottom =
                (host.scrollHeight - host.scrollTop - host.clientHeight) < FBMessageList.SCROLL_THRESHOLD;

            if (isNearBottom) {
                const last = this.shadowRoot?.querySelector('.message:last-child');
                if (last) {
                    last.scrollIntoView({ behavior: 'smooth', block: 'end' });
                }
            }
        });
    }

    private _formatTime(isoString: string): string {
        try {
            return new Date(isoString).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
        } catch {
            return '';
        }
    }

    override render(): TemplateResult {
        if (this._messages.length === 0) {
            return html`
                <div class="empty-state" role="status" aria-live="polite">
                    <div class="empty-icon" aria-hidden="true">💬</div>
                    <h3>Start a Conversation</h3>
                    <p>Ask the agent to help you manage your project.</p>
                </div>
            `;
        }

        return html`
            <div class="list" role="log" aria-label="Conversation messages" aria-live="polite">
                <span class="sr-only">${this._messages.length} messages in conversation</span>
                ${this._messages.map(msg => html`
                    <article 
                        class="message ${msg.role}" 
                        id=${msg.id}
                        aria-label="${msg.role === 'user' ? 'Your message' : 'Agent response'}"
                    >
                        <div class="message-content">
                            ${msg.content}
                        </div>
                        
                        ${msg.actionCard ? html`
                            <fb-action-card .card=${msg.actionCard} message-id=${msg.id}></fb-action-card>
                        ` : nothing}

                        <span class="meta" aria-label="Sent at ${this._formatTime(msg.createdAt)}">
                            ${msg.role === 'user' ? 'You' : 'Agent'} • ${this._formatTime(msg.createdAt)}
                        </span>
                    </article>
                `)}
            </div>
        `;
    }
}
