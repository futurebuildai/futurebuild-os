/**
 * fb-feed-card — Single feed card component.
 * See FRONTEND_V2_SPEC.md §3
 *
 * Renders a card with: priority indicator, headline, body, consequence,
 * project label, time-relative deadline, and action buttons.
 */
import { html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { scorePriority } from '../../utils/feed-priority';
import type { FeedCard, FeedCardType } from '../../types/feed';
import './fb-contact-inline-add';

const CARD_ICONS: Partial<Record<FeedCardType, string>> = {
    daily_briefing: '\u2600',      // ☀
    procurement_warning: '\u26A0',  // ⚠
    procurement_critical: '\uD83D\uDEA8', // 🚨 (escaped)
    task_starting: '\u25B6',        // ▶
    task_completed: '\u2714',       // ✔
    inspection_upcoming: '\uD83D\uDD0D', // 🔍
    inspection_result: '\uD83D\uDCCB', // 📋
    schedule_recalc: '\uD83D\uDCC5', // 📅
    weather_risk: '\uD83C\uDF27',   // 🌧
    weather_window: '\u2600',        // ☀
    sub_confirmation: '\uD83D\uDCDE', // 📞
    sub_unconfirmed: '\u2753',      // ❓
    invoice_ready: '\uD83D\uDCB0', // 💰
    setup_team: '\uD83D\uDC65',    // 👥
    setup_contacts: '\uD83D\uDCDE', // 📞
    calibration_drift: '\uD83D\uDCC9', // 📉
    budget_alert: '\uD83D\uDCB8',    // 💸
    milestone: '\uD83C\uDFC1',     // 🏁
    welcome: '\uD83D\uDC4B',       // 👋
};

@customElement('fb-feed-card')
export class FBFeedCard extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
            }

            .card {
                background: var(--fb-surface-1, #1a1a2e);
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 12px;
                padding: 16px 20px;
                transition: border-color 0.15s ease, box-shadow 0.15s ease;
            }

            :host([priority="critical"]) .card {
                border-left: 4px solid #ef4444;
                border-color: #ef4444;
                background: linear-gradient(90deg, rgba(239,68,68,0.05) 0%, var(--fb-surface-1, #1a1a2e) 15%);
            }
            :host([priority="urgent"]) .card {
                border-left: 4px solid #f59e0b;
                border-color: #f59e0b;
                background: linear-gradient(90deg, rgba(245,158,11,0.05) 0%, var(--fb-surface-1, #1a1a2e) 15%);
            }
            :host([priority="routine"]) .card {
                border-left: 4px solid #10b981;
            }

            .card:hover {
                border-color: var(--fb-border-hover, #3a3a5e);
                box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
            }

            .header {
                display: flex;
                align-items: flex-start;
                gap: 12px;
                margin-bottom: 8px;
            }

            .priority-badge {
                display: flex;
                align-items: center;
                gap: 6px;
                margin-top: 4px; /* align with headline */
                flex-shrink: 0;
            }
            .priority-badge .dot, .priority-badge .pulse-dot {
                width: 8px;
                height: 8px;
                border-radius: 50%;
            }
            .priority-badge.critical .pulse-dot {
                background: #ef4444;
                box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.7);
                animation: pulse 2s infinite;
            }
            .priority-badge.urgent .dot {
                background: #f59e0b;
            }
            .priority-badge.routine .dot {
                background: #10b981;
            }
            .badge-text {
                font-size: 10px;
                font-weight: 700;
                letter-spacing: 0.5px;
            }
            .priority-badge.critical .badge-text {
                color: #ef4444;
            }
            .priority-badge.urgent .badge-text {
                color: #f59e0b;
            }
            @keyframes pulse {
                0% { box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.7); }
                70% { box-shadow: 0 0 0 6px rgba(239, 68, 68, 0); }
                100% { box-shadow: 0 0 0 0 rgba(239, 68, 68, 0); }
            }

            .icon {
                font-size: 18px;
                flex-shrink: 0;
                margin-top: 1px;
            }

            .title-area {
                flex: 1;
                min-width: 0;
            }

            .headline {
                font-size: 15px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
                line-height: 1.3;
            }

            .project-label {
                font-size: 12px;
                color: var(--fb-text-tertiary, #707080);
                margin-top: 2px;
            }

            .body {
                font-size: 14px;
                color: var(--fb-text-secondary, #a0a0b0);
                line-height: 1.5;
                margin: 8px 0;
                padding-left: 20px;
            }

            .consequence {
                font-size: 13px;
                color: var(--fb-warning, #f59e0b);
                font-style: italic;
                padding-left: 20px;
                margin-bottom: 12px;
            }

            .actions-row {
                display: flex;
                align-items: center;
                gap: 8px;
                padding-left: 20px;
                margin-top: 12px;
            }

            .action-btn {
                padding: 6px 16px;
                border-radius: 6px;
                font-size: 13px;
                font-weight: 500;
                cursor: pointer;
                border: 1px solid transparent;
                transition: all 0.15s ease;
            }

            .action-btn[data-style='primary'] {
                background: var(--fb-accent, #6366f1);
                color: #fff;
            }

            .action-btn[data-style='primary']:hover {
                opacity: 0.9;
            }

            .action-btn[data-style='secondary'] {
                background: transparent;
                border-color: var(--fb-border, #2a2a3e);
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .action-btn[data-style='secondary']:hover {
                border-color: var(--fb-text-secondary, #a0a0b0);
                color: var(--fb-text-primary, #e0e0e0);
            }

            .action-btn[data-style='danger'] {
                background: transparent;
                border-color: #ef4444;
                color: #ef4444;
            }

            .action-btn[data-style='danger']:hover {
                background: #ef4444;
                color: #fff;
            }

            .action-btn:focus-visible {
                outline: 2px solid var(--fb-accent, #6366f1);
                outline-offset: 2px;
            }

            .deadline {
                font-size: 12px;
                color: var(--fb-text-tertiary, #707080);
                margin-left: auto;
            }

            .inline-form {
                padding: 12px 0 0 20px;
            }

            .tell-me-more {
                display: block;
                padding: 8px 0 0 20px;
                font-size: 13px;
                color: var(--fb-accent, #6366f1);
                cursor: pointer;
                background: none;
                border: none;
                text-align: left;
                transition: opacity 0.15s;
            }

            .tell-me-more:hover {
                opacity: 0.8;
            }

            .tell-me-more:focus-visible {
                outline: 2px solid var(--fb-accent, #6366f1);
                outline-offset: 2px;
                border-radius: 4px;
            }

            @media (max-width: 768px) {
                .card {
                    padding: 12px 14px;
                }

                .body, .consequence, .actions-row, .inline-form {
                    padding-left: 0;
                }
            }
        `,
    ];

    @property({ type: Object }) card!: FeedCard;

    override willUpdate(changedProperties: Map<string, unknown>) {
        super.willUpdate(changedProperties);
        if (changedProperties.has('card') && this.card) {
            this.setAttribute('priority', scorePriority(this.card).priority);
        }
    }

    private _handleAction(actionId: string) {
        this.emit('fb-card-action', {
            cardId: this.card.id,
            actionId,
            projectId: this.card.project_id,
        });
    }

    private _handleContactSaved() {
        // Dismiss the setup_contacts card after a contact is saved + assigned
        this._handleAction('dismiss');
    }

    private _handleContactCancelled() {
        this._handleAction('dismiss');
    }

    private _formatDeadline(deadline: string): string {
        const d = new Date(deadline);
        const now = new Date();
        const diffMs = d.getTime() - now.getTime();
        const diffDays = Math.ceil(diffMs / (1000 * 60 * 60 * 24));

        if (diffDays < 0) return `${Math.abs(diffDays)}d overdue`;
        if (diffDays === 0) return 'Today';
        if (diffDays === 1) return 'Tomorrow';
        if (diffDays <= 7) return `${diffDays}d`;
        return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
    }

    private _renderPriorityBadge() {
        const priority = scorePriority(this.card).priority;
        if (priority === 'critical') {
            return html`
                <div class="priority-badge critical">
                    <span class="pulse-dot"></span>
                    <span class="badge-text">CRITICAL</span>
                </div>
            `;
        } else if (priority === 'urgent') {
            return html`
                <div class="priority-badge urgent">
                    <span class="dot"></span>
                    <span class="badge-text">NEEDS ATTENTION</span>
                </div>
            `;
        } else {
            return html`
                <div class="priority-badge routine">
                    <span class="dot"></span>
                </div>
            `;
        }
    }

    override render() {
        if (!this.card) return nothing;

        const icon = CARD_ICONS[this.card.card_type] ?? '\u2022'; // bullet fallback

        return html`
            <div class="card">
                <div class="header">
                    ${this._renderPriorityBadge()}
                    <span class="icon">${icon}</span>
                    <div class="title-area">
                        <div class="headline">${this.card.headline}</div>
                        ${this.card.project_name
                ? html`<div class="project-label">${this.card.project_name}</div>`
                : nothing}
                    </div>
                    ${this.card.deadline
                ? html`<span class="deadline">${this._formatDeadline(this.card.deadline)}</span>`
                : nothing}
                </div>

                <div class="body">${this.card.body}</div>

                ${this.card.consequence
                ? html`<div class="consequence">${this.card.consequence}</div>`
                : nothing}

                ${this.card.card_type === 'setup_contacts'
                ? html`
                          <div class="inline-form">
                              <fb-contact-inline-add
                                  .projectId=${this.card.project_id}
                                  @fb-contact-saved=${this._handleContactSaved}
                                  @fb-contact-cancelled=${this._handleContactCancelled}
                              ></fb-contact-inline-add>
                          </div>
                      `
                : (this.card.actions?.length ?? 0) > 0
                    ? html`
                              <div class="actions-row">
                                  ${this.card.actions.map(
                        (a) => html`
                                          <button
                                              class="action-btn"
                                              data-style=${a.style}
                                              @click=${() => this._handleAction(a.id)}
                                          >
                                              ${a.label}
                                          </button>
                                      `
                    )}
                              </div>
                          `
                    : nothing}

                ${this._showTellMeMore()
                ? html`
                          <button
                              class="tell-me-more"
                              @click=${(e: Event) => {
                        e.stopPropagation();
                        this._handleAction('tell_me_more');
                    }}
                          >Tell me more \u2192</button>
                      `
                : nothing}
            </div>
        `;
    }

    private _showTellMeMore(): boolean {
        const skip: FeedCardType[] = ['setup_team', 'setup_contacts', 'welcome'];
        return !skip.includes(this.card.card_type);
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-feed-card': FBFeedCard;
    }
}
