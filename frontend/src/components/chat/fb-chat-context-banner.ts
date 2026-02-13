/**
 * fb-chat-context-banner — Shows feed card context at the top of a chat thread.
 * See FRONTEND_V2_SPEC.md §7 Step 34
 *
 * Displayed when the user navigates to chat via "Tell me more" on a feed card.
 * Shows the card headline and consequence, with a dismiss button.
 */
import { html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

@customElement('fb-chat-context-banner')
export class FBChatContextBanner extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
            }

            .banner {
                display: flex;
                align-items: flex-start;
                gap: 10px;
                padding: 10px 16px;
                background: var(--fb-surface-2, #252540);
                border-left: 3px solid var(--fb-accent, #6366f1);
                border-radius: 0 8px 8px 0;
                margin: 8px 16px;
            }

            .badge {
                font-size: 11px;
                font-weight: 600;
                color: var(--fb-accent, #6366f1);
                background: rgba(99, 102, 241, 0.15);
                padding: 2px 8px;
                border-radius: 4px;
                white-space: nowrap;
                flex-shrink: 0;
                margin-top: 1px;
            }

            .content {
                flex: 1;
                min-width: 0;
            }

            .headline {
                font-size: 13px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
                line-height: 1.3;
            }

            .consequence {
                font-size: 12px;
                color: var(--fb-warning, #f59e0b);
                margin-top: 4px;
                font-style: italic;
            }

            .dismiss {
                background: none;
                border: none;
                color: var(--fb-text-muted, #606070);
                cursor: pointer;
                padding: 2px;
                border-radius: 4px;
                font-size: 16px;
                line-height: 1;
                flex-shrink: 0;
            }

            .dismiss:hover {
                color: var(--fb-text-primary, #e0e0e0);
                background: var(--fb-bg-tertiary, #2a2a3e);
            }
        `,
    ];

    @property({ type: String }) headline = '';
    @property({ type: String }) consequence = '';
    @property({ type: String }) cardType = '';

    private _formatBadge(type: string): string {
        return type.replace(/_/g, ' ');
    }

    override render() {
        if (!this.headline) return nothing;

        return html`
            <div class="banner">
                <span class="badge">${this._formatBadge(this.cardType)}</span>
                <div class="content">
                    <div class="headline">${this.headline}</div>
                    ${this.consequence
                        ? html`<div class="consequence">${this.consequence}</div>`
                        : nothing}
                </div>
                <button
                    class="dismiss"
                    @click=${() => this.emit('fb-banner-dismiss', {})}
                    aria-label="Dismiss context"
                >&times;</button>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-chat-context-banner': FBChatContextBanner;
    }
}
