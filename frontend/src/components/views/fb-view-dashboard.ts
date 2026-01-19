/**
 * FBViewDashboard - Default Landing View
 * See PRODUCTION_PLAN.md Step 51.4
 *
 * The home base for authenticated users.
 * Displays project overview, metrics, and quick actions.
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBViewElement } from '../base/FBViewElement';
import { store } from '../../store/store';

@customElement('fb-view-dashboard')
export class FBViewDashboard extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                padding: var(--fb-spacing-xl);
            }

            .header {
                margin-bottom: var(--fb-spacing-xl);
            }

            h1 {
                font-size: var(--fb-text-2xl);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin: 0 0 var(--fb-spacing-sm) 0;
            }

            .subtitle {
                color: var(--fb-text-secondary);
                font-size: var(--fb-text-base);
            }

            .metrics-grid {
                display: grid;
                grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
                gap: var(--fb-spacing-lg);
                margin-bottom: var(--fb-spacing-xl);
            }

            .metric-card {
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-lg);
            }

            .metric-label {
                font-size: var(--fb-text-sm);
                color: var(--fb-text-secondary);
                margin-bottom: var(--fb-spacing-xs);
            }

            .metric-value {
                font-size: var(--fb-text-2xl);
                font-weight: 700;
                color: var(--fb-text-primary);
            }

            .metric-trend {
                font-size: var(--fb-text-sm);
                margin-top: var(--fb-spacing-xs);
            }

            .trend-up {
                color: var(--fb-success);
            }

            .trend-down {
                color: var(--fb-error);
            }

            .section-title {
                font-size: var(--fb-text-lg);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin-bottom: var(--fb-spacing-md);
            }

            .placeholder-content {
                background: var(--fb-bg-card);
                border: 1px dashed var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-2xl);
                text-align: center;
                color: var(--fb-text-muted);
            }
        `,
    ];

    @state() private _userName = '';

    private _disposeEffect: (() => void) | null = null;

    override connectedCallback(): void {
        super.connectedCallback();
        this._disposeEffect = effect(() => {
            this._userName = store.user$.value?.name ?? 'Builder';
        });
    }

    override disconnectedCallback(): void {
        this._disposeEffect?.();
        this._disposeEffect = null;
        super.disconnectedCallback();
    }

    override onViewActive(): void {
        // Future: Fetch dashboard metrics from API
        console.log('[FBViewDashboard] View activated - would fetch metrics');
    }

    override render(): TemplateResult {
        return html`
            <div class="header">
                <h1>Welcome back, ${this._userName}</h1>
                <p class="subtitle">Here's what's happening with your projects today.</p>
            </div>

            <div class="metrics-grid">
                <div class="metric-card">
                    <div class="metric-label">Active Projects</div>
                    <div class="metric-value">3</div>
                    <div class="metric-trend trend-up">↑ 1 new this month</div>
                </div>
                <div class="metric-card">
                    <div class="metric-label">Tasks Due This Week</div>
                    <div class="metric-value">12</div>
                    <div class="metric-trend">5 completed</div>
                </div>
                <div class="metric-card">
                    <div class="metric-label">Pending Invoices</div>
                    <div class="metric-value">$24,500</div>
                    <div class="metric-trend trend-down">3 overdue</div>
                </div>
                <div class="metric-card">
                    <div class="metric-label">Avg. Completion</div>
                    <div class="metric-value">67%</div>
                    <div class="metric-trend trend-up">↑ 5% this week</div>
                </div>
            </div>

            <h2 class="section-title">Recent Activity</h2>
            <div class="placeholder-content">
                Activity feed coming soon. Will show task updates, chat messages, and alerts.
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-dashboard': FBViewDashboard;
    }
}
