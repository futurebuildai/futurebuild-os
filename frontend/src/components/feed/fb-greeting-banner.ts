/**
 * fb-greeting-banner — Portfolio greeting with summary metrics.
 * See FRONTEND_V2_SPEC.md §4
 *
 * Displays personalized greeting and portfolio summary:
 * "Good morning, Marcus. 5 active projects."
 */
import { html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import type { PortfolioSummary } from '../../types/feed';

@customElement('fb-greeting-banner')
export class FBGreetingBanner extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                margin-bottom: 28px;
            }

            .greeting {
                font-size: 28px;
                font-weight: 700;
                color: var(--fb-text-primary, #e0e0e0);
                margin-bottom: 6px;
                line-height: 1.2;
            }

            .summary {
                font-size: 14px;
                color: var(--fb-text-secondary, #a0a0b0);
                line-height: 1.5;
                display: flex;
                align-items: center;
                flex-wrap: wrap;
                gap: 4px;
            }

            .stat {
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .alert {
                color: var(--fb-warning, #f59e0b);
                font-weight: 600;
            }

            .divider {
                color: var(--fb-text-tertiary, #707080);
                margin: 0 4px;
            }

            .skeleton-greeting {
                width: 220px;
                height: 32px;
                border-radius: 4px;
                margin-bottom: 8px;
            }

            .skeleton-summary {
                width: 160px;
                height: 18px;
                border-radius: 4px;
            }

            @media (max-width: 768px) {
                :host {
                    margin-bottom: 24px;
                }

                .greeting {
                    font-size: 22px;
                }

                .summary {
                    font-size: 13px;
                }
            }
        `,
    ];

    /** Greeting message, e.g., "Good morning, Marcus" */
    @property({ type: String }) greeting = '';

    /** Portfolio summary data */
    @property({ type: Object }) summary: PortfolioSummary | null = null;

    /** Show loading skeleton */
    @property({ type: Boolean }) loading = false;

    override render() {
        if (this.loading) {
            return html`
                <div class="skeleton-greeting skeleton"></div>
                <div class="skeleton-summary skeleton"></div>
            `;
        }

        return html`
            <div class="greeting" role="heading" aria-level="1">${this.greeting}</div>
            ${this._renderSummary()}
        `;
    }

    private _renderSummary() {
        if (!this.summary) return nothing;
        const s = this.summary;

        const parts: { type: 'stat' | 'alert'; text: string }[] = [];

        if (s.active_project_count > 0) {
            const projectText = s.active_project_count === 1
                ? '1 active project'
                : `${s.active_project_count} active projects`;
            parts.push({ type: 'stat', text: projectText });
        }

        if (s.total_tasks > 0) {
            parts.push({ type: 'stat', text: `${s.total_tasks} tasks` });
        }

        if (s.critical_alerts > 0) {
            const alertText = s.critical_alerts === 1
                ? '1 needs attention'
                : `${s.critical_alerts} need attention`;
            parts.push({ type: 'alert', text: alertText });
        }

        if (parts.length === 0) return nothing;

        return html`
            <div class="summary" role="status" aria-label="Portfolio summary">
                ${parts.map((part, i) => html`
                    ${i > 0 ? html`<span class="divider">·</span>` : nothing}
                    <span class="${part.type}">${part.text}</span>
                `)}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-greeting-banner': FBGreetingBanner;
    }
}
