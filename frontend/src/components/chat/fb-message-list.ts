import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state, query } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';
import type { ChatMessage } from '../../store/types';

import '@lit-labs/virtualizer';
import { flow } from '@lit-labs/virtualizer/layouts/flow.js';
import type { LitVirtualizer, RangeChangedEvent } from '@lit-labs/virtualizer';

import './fb-action-card';
import './fb-typing-indicator';

/**
 * Virtual item type for the virtualizer.
 * Can be a real message or a typing indicator placeholder.
 * See PRODUCTION_PLAN.md Step 60.2.1
 */
interface VirtualItem {
    id: string;
    role?: 'user' | 'assistant' | 'system';
    content?: string;
    createdAt?: string;
    displayTime?: string;
    isStreaming?: boolean;
    actionCard?: ChatMessage['actionCard'];
    artifactRef?: ChatMessage['artifactRef'];
    /** True if this is the typing indicator placeholder */
    isTypingPlaceholder?: boolean;
}

/**
 * FBMessageList - Virtualized chat message stream.
 * Uses @lit-labs/virtualizer for O(1) DOM at 1,000+ messages.
 * @element fb-message-list
 */
@customElement('fb-message-list')
export class FBMessageList extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            /* Step 60.2.1: Host delegates scrolling to virtualizer */
            :host {
                display: flex;
                flex-direction: column;
                height: 100%;
                overflow: hidden;
            }

            lit-virtualizer {
                flex: 1;
                /* Padding applied inside virtualizer for scroll edge spacing */
                padding: var(--fb-spacing-lg);
            }

            .message {
                display: flex;
                flex-direction: column;
                max-width: var(--fb-msg-max-width, 85%);
                animation: fadeIn var(--fb-anim-duration-fade, 0.3s) ease;
                /* Replaces flex gap - virtualizer doesn't support gap */
                margin-bottom: var(--fb-spacing-md);
            }

            @keyframes fadeIn {
                from { opacity: 0; transform: var(--fb-anim-slide-up, translateY(10px)); }
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

            /* Typing indicator wrapper */
            .typing-wrapper {
                align-self: flex-start;
                margin-bottom: var(--fb-spacing-md);
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

    @query('lit-virtualizer')
    private _virtualizer?: LitVirtualizer;

    @state() private _messages: ChatMessage[] = [];
    @state() private _isTyping = false;

    /** Tracks if user is viewing the bottom of the list */
    private _isAtBottom = true;
    private _disposeEffects: (() => void)[] = [];
    /** L7 Fix: ResizeObserver for virtualizer layout recalculation */
    private _resizeObserver: ResizeObserver | null = null;

    /**
     * Computed items for the virtualizer.
     * Injects typing indicator as a transient virtual item.
     */
    private get _virtualItems(): VirtualItem[] {
        if (this._isTyping) {
            return [
                ...this._messages,
                { id: 'typing-indicator', isTypingPlaceholder: true }
            ];
        }
        return this._messages;
    }

    override connectedCallback(): void {
        super.connectedCallback();

        this._disposeEffects.push(
            effect(() => {
                const prevLength = this._messages.length;
                this._messages = store.messages$.value;

                // Auto-scroll only if user was at bottom before new message
                if (this._messages.length > prevLength && this._isAtBottom) {
                    this._scrollToBottom();
                }
            }),
            effect(() => {
                this._isTyping = store.isTyping$.value;
                // Scroll to show typing indicator if at bottom
                if (this._isTyping && this._isAtBottom) {
                    this._scrollToBottom();
                }
            })
        );

        // L7 Fix: ResizeObserver to trigger virtualizer layout on container resize
        this._resizeObserver = new ResizeObserver(() => {
            // Force the virtualizer to remeasure item heights
            if (this._virtualizer) {
                this._virtualizer.requestUpdate();
            }
        });
        this._resizeObserver.observe(this);
    }

    override disconnectedCallback(): void {
        this._disposeEffects.forEach((d) => { d(); });
        this._disposeEffects = [];
        if (this._resizeObserver) {
            this._resizeObserver.disconnect();
            this._resizeObserver = null;
        }
        super.disconnectedCallback();
    }

    /**
     * Handles virtualizer range changes to track scroll position.
     * Updates _isAtBottom based on whether the last item is visible.
     */
    private _handleRangeChanged = (e: RangeChangedEvent): void => {
        const items = this._virtualItems;
        // User is at bottom if the last item is visible
        this._isAtBottom = e.last >= items.length - 1;
    };

    /**
     * Scrolls to the last item using virtualizer API.
     */
    private _scrollToBottom(): void {
        requestAnimationFrame(() => {
            const items = this._virtualItems;
            if (this._virtualizer && items.length > 0) {
                this._virtualizer.scrollToIndex(items.length - 1, 'end');
            }
        });
    }

    /**
     * Renders individual virtual items (messages or typing indicator).
     */
    private _renderItem = (item: VirtualItem): TemplateResult => {
        // Handle typing indicator placeholder
        if (item.isTypingPlaceholder) {
            return html`
                <div class="typing-wrapper">
                    <fb-typing-indicator></fb-typing-indicator>
                </div>
            `;
        }

        // Render standard message
        return html`
            <article 
                class="message ${item.role}" 
                id=${item.id}
                aria-label="${item.role === 'user' ? 'Your message' : 'Agent response'}"
            >
                <div class="message-content">
                    ${item.content}
                </div>
                
                ${item.actionCard ? html`
                    <fb-action-card .card=${item.actionCard} message-id=${item.id}></fb-action-card>
                ` : nothing}

                <span class="meta" aria-label="Sent at ${item.displayTime ?? ''}"
                    >${item.role === 'user' ? 'You' : 'Agent'} • ${item.displayTime ?? ''}</span
                >
            </article>
        `;
    };

    /**
     * Renders the empty state when no messages exist.
     */
    private _renderEmptyState(): TemplateResult {
        return html`
            <div class="empty-state" role="status" aria-live="polite">
                <div class="empty-icon" aria-hidden="true">💬</div>
                <h3>Start a Conversation</h3>
                <p>Ask the agent to help you manage your project.</p>
            </div>
        `;
    }

    override render(): TemplateResult {
        // Early return for empty state (avoids 0-item sizing glitches)
        if (this._messages.length === 0) {
            return this._renderEmptyState();
        }

        return html`
            <span class="sr-only">${this._messages.length} messages in conversation</span>
            <lit-virtualizer
                scroller
                role="log"
                aria-label="Conversation messages"
                aria-live="polite"
                .layout=${flow({ direction: 'vertical' })}
                .items=${this._virtualItems}
                .renderItem=${this._renderItem}
                @rangeChanged=${this._handleRangeChanged}
            ></lit-virtualizer>
        `;
    }
}
