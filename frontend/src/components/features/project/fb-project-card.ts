/**
 * FBProjectCard - Presentational Project Card Component
 * See PROJECT_ONBOARDING_SPEC.md Step 62.5, STEP_92_RISK_INDICATORS.md
 *
 * Displays a project summary card with name, address, status badge,
 * and risk indicators (Step 92) for at-risk projects.
 *
 * @element fb-project-card
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../../base/FBElement';
import type { ProjectSummary, RiskLevel } from '../../../store/types';

/**
 * Compute the risk level for a project based on its data.
 * Rules (STEP_92_RISK_INDICATORS.md Section 1.1):
 *  - High: schedule slip > 2 days, OR active blockers, OR budget overrun > 10%
 *  - Medium: schedule slip 1-2 days, OR budget overrun 5-10%
 *  - Low: everything else
 */
function computeRiskLevel(project: ProjectSummary): { level: RiskLevel; reason: string } {
    // If risk data is pre-computed from backend, use it directly
    if (project.riskLevel) {
        return { level: project.riskLevel, reason: project.riskReason ?? '' };
    }

    const reasons: string[] = [];
    let isHigh = false;
    let isMedium = false;

    // Check schedule slip
    const slipDays = project.scheduleSlipDays ?? 0;
    if (slipDays > 2) {
        isHigh = true;
        reasons.push(`Critical Path Delay: +${String(slipDays)} days`);
    } else if (slipDays >= 1) {
        isMedium = true;
        reasons.push(`Schedule Slip: +${String(slipDays)} day${slipDays > 1 ? 's' : ''}`);
    }

    // Check blockers
    if (project.hasBlockers) {
        isHigh = true;
        reasons.push('Active Blocking Issues');
    }

    // Check budget overrun
    const overrun = project.budgetOverrunPct ?? 0;
    if (overrun > 10) {
        isHigh = true;
        reasons.push(`Budget Overrun: ${String(overrun)}%`);
    } else if (overrun > 5) {
        isMedium = true;
        reasons.push(`Budget Warning: ${String(overrun)}%`);
    }

    const level: RiskLevel = isHigh ? 'high' : isMedium ? 'medium' : 'low';
    return { level, reason: reasons.join(' · ') };
}

@customElement('fb-project-card')
export class FBProjectCard extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
            }

            .card {
                position: relative;
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-lg);
                cursor: pointer;
                transition: transform 0.15s ease, box-shadow 0.15s ease, border-color 0.15s ease;
            }

            .card:hover {
                transform: translateY(-2px);
                box-shadow: var(--fb-shadow-md, 0 4px 12px rgba(0, 0, 0, 0.15));
            }

            .card:focus-visible {
                outline: 2px solid var(--fb-primary);
                outline-offset: 2px;
            }

            /* Step 92: Risk border styles */
            .card.risk-high {
                border-left: 4px solid var(--fb-error, #c62828);
            }

            .card.risk-medium {
                border-left: 4px solid var(--fb-warning, #e65100);
            }

            .name {
                font-size: var(--fb-text-lg);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin: 0 0 var(--fb-spacing-xs) 0;
            }

            .address {
                font-size: var(--fb-text-sm);
                color: var(--fb-text-secondary);
                margin: 0 0 var(--fb-spacing-md) 0;
            }

            .footer {
                display: flex;
                align-items: center;
                justify-content: space-between;
            }

            .status {
                display: inline-flex;
                align-items: center;
                padding: var(--fb-spacing-xs) var(--fb-spacing-sm);
                border-radius: var(--fb-radius-full, 9999px);
                font-size: var(--fb-text-xs);
                font-weight: 500;
                text-transform: capitalize;
            }

            .status--planning {
                background: rgba(102, 126, 234, 0.15);
                color: var(--fb-primary, #667eea);
            }

            .status--in_progress,
            .status--active {
                background: rgba(72, 187, 120, 0.15);
                color: #48bb78;
            }

            .status--completed {
                background: rgba(160, 174, 192, 0.15);
                color: #a0aec0;
            }

            .status--on_hold {
                background: rgba(237, 137, 54, 0.15);
                color: #ed8936;
            }

            .progress {
                font-size: var(--fb-text-xs);
                color: var(--fb-text-muted);
            }

            /* Step 92: Risk indicator dot */
            .risk-dot {
                position: absolute;
                top: 12px;
                right: 12px;
                width: 10px;
                height: 10px;
                border-radius: 50%;
                cursor: help;
            }

            .risk-dot.high {
                background: var(--fb-error, #c62828);
                animation: pulse-risk 2s ease-in-out infinite;
            }

            .risk-dot.medium {
                background: var(--fb-warning, #e65100);
            }

            @keyframes pulse-risk {
                0%, 100% { opacity: 1; transform: scale(1); }
                50% { opacity: 0.6; transform: scale(1.2); }
            }

            /* Step 92: Risk tooltip */
            .risk-dot[title]:hover::after {
                content: attr(title);
                position: absolute;
                right: 0;
                top: 100%;
                margin-top: 6px;
                padding: var(--fb-spacing-xs) var(--fb-spacing-sm);
                background: var(--fb-bg-tertiary, #1a1a1a);
                border: 1px solid var(--fb-border, #333);
                border-radius: var(--fb-radius-sm, 4px);
                font-size: var(--fb-text-xs, 0.75rem);
                color: var(--fb-text-primary, #fff);
                white-space: nowrap;
                z-index: 10;
                pointer-events: none;
            }
        `,
    ];

    @property({ type: Object })
    project!: ProjectSummary;

    private _handleClick(): void {
        this.emit('project-selected', { id: this.project.id });
    }

    private _handleKeyDown(e: KeyboardEvent): void {
        if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            this._handleClick();
        }
    }

    private _getStatusClass(): string {
        const status = this.project.status.toLowerCase().replace(/\s+/g, '_');
        return `status--${status}`;
    }

    override render(): TemplateResult {
        const risk = computeRiskLevel(this.project);
        const hasRisk = risk.level !== 'low';
        const cardClass = hasRisk ? `card risk-${risk.level}` : 'card';

        return html`
            <article
                class="${cardClass}"
                role="button"
                tabindex="0"
                @click=${this._handleClick.bind(this)}
                @keydown=${this._handleKeyDown.bind(this)}
                aria-label="Project: ${this.project.name}${hasRisk ? `. Risk: ${risk.reason}` : ''}"
            >
                ${hasRisk ? html`
                    <div
                        class="risk-dot ${risk.level}"
                        title="${risk.reason}"
                        role="img"
                        aria-label="Risk indicator: ${risk.reason}"
                    ></div>
                ` : nothing}
                <h3 class="name">${this.project.name}</h3>
                <p class="address">${this.project.address}</p>
                <div class="footer">
                    <span class="status ${this._getStatusClass()}">
                        ${this.project.status}
                    </span>
                    <span class="progress">
                        ${this.project.completionPercentage}% complete
                    </span>
                </div>
            </article>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-project-card': FBProjectCard;
    }
}
