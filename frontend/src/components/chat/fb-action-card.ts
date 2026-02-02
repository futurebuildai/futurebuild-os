import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';
import type { ActionCard } from '../../store/types';

/**
 * FBActionCard - Displays agent-generated action cards requiring user approval.
 * Emits 'action' event when user approves/denies.
 * @element fb-action-card
 */
@customElement('fb-action-card')
export class FBActionCard extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                width: 100%;
                margin-top: var(--fb-spacing-sm);
            }

            .card {
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-md);
                border-left: 4px solid var(--fb-border);
                transition: transform 0.2s ease, box-shadow 0.2s ease;
            }

            .card:hover {
                transform: translateY(-2px);
                box-shadow: var(--fb-shadow-md);
            }

            /* Status Colors */
            .status-pending { border-left-color: var(--fb-warning); }
            .status-approved { border-left-color: var(--fb-success); }
            .status-denied { border-left-color: var(--fb-error); }
            .status-edited { border-left-color: var(--fb-primary); }

            .header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                margin-bottom: var(--fb-spacing-sm);
            }

            .type-label {
                font-size: var(--fb-text-xs);
                font-weight: 600;
                text-transform: uppercase;
                letter-spacing: 0.05em;
                color: var(--fb-text-secondary);
            }

            .status-badge {
                font-size: 10px;
                font-weight: 700;
                padding: 2px 6px;
                border-radius: 4px;
                background: var(--fb-bg-tertiary);
                color: var(--fb-text-secondary);
            }

            .title {
                font-size: var(--fb-text-base);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin-bottom: 4px;
            }

            .summary {
                font-size: var(--fb-text-sm);
                color: var(--fb-text-secondary);
                line-height: 1.5;
                margin-bottom: var(--fb-spacing-md);
            }

            .actions {
                display: flex;
                gap: var(--fb-spacing-sm);
            }

            .btn {
                flex: 1;
                padding: var(--fb-spacing-sm);
                border-radius: var(--fb-radius-md);
                border: 1px solid var(--fb-border);
                background: var(--fb-bg-secondary);
                color: var(--fb-text-primary);
                font-size: var(--fb-text-sm);
                font-weight: 500;
                cursor: pointer;
                transition: all 0.2s ease;
                display: flex;
                align-items: center;
                justify-content: center;
                gap: 6px;
            }

            .btn:hover:not(:disabled) {
                background: var(--fb-bg-tertiary);
                border-color: var(--fb-text-muted);
            }

            .btn:focus-visible {
                outline: 2px solid var(--fb-primary);
                outline-offset: 2px;
            }

            .btn-approve { background: var(--fb-success); color: white; border: none; }
            .btn-approve:hover { opacity: 0.9; background: var(--fb-success); }

            .btn-edit { 
                background: var(--fb-bg-tertiary); 
                color: var(--fb-text-muted); 
                border: 1px dashed var(--fb-border);
                cursor: not-allowed;
            }
            
            .btn-deny { background: var(--fb-error); color: white; border: none; }
            .btn-deny:hover { opacity: 0.9; background: var(--fb-error); }
        `
    ];

    @property({ type: Object }) card: ActionCard | undefined;
    @property({ type: String, attribute: 'message-id' }) messageId: string | undefined;

    /**
     * Type guard to check if the card is in pending status.
     */
    private _isPending(): boolean {
        return this.card?.status === 'pending';
    }

    private _handleAction(status: 'approved' | 'denied'): void {
        if (!this.card || !this.messageId) return;
        store.actions.updateActionCard(this.messageId, status);
        this.emit('action', { id: this.card.id, status });
    }

    override render(): TemplateResult {
        if (!this.card || !this.messageId) return html``;
        const isPending = this._isPending();
        const typeLabel = this.card.type.replace(/_/g, ' ');

        return html`
            <div 
                class="card status-${this.card.status}"
                role="region"
                aria-label="Action required: ${this.card.title}"
            >
                <div class="header">
                    <span class="type-label">${typeLabel}</span>
                    ${this.card.status !== 'pending' ? html`
                        <span class="status-badge" role="status">${this.card.status.toUpperCase()}</span>
                    ` : nothing}
                </div>
                
                <div class="title">${this.card.title}</div>
                <div class="summary">${this.card.summary}</div>

                ${isPending ? html`
                    <div class="actions" role="group" aria-label="Action buttons">
                        <button 
                            class="btn btn-approve" 
                            @click=${(): void => { this._handleAction('approved'); }}
                            aria-label="Approve: ${this.card.title}"
                        >
                            ✓ Approve
                        </button>
                        <button 
                            class="btn btn-edit" 
                            disabled
                            aria-label="Edit action (coming soon)"
                            title="Edit functionality coming soon"
                        >
                            ✎ Edit
                        </button>
                        <button 
                            class="btn btn-deny" 
                            @click=${(): void => { this._handleAction('denied'); }}
                            aria-label="Deny: ${this.card.title}"
                        >
                            ✗ Deny
                        </button>
                    </div>
                ` : nothing}
            </div>
        `;
    }
}
