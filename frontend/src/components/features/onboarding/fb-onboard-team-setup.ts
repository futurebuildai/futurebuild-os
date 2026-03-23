/**
 * FBOnboardTeamSetup - Post-Creation Team Assignment
 *
 * After project creation, prompts the builder to assign key trade contacts
 * to schedule phases. Shows required trades from the schedule preview's trade_gaps.
 *
 * Each trade gap shows the trade name, phase, start date, and input for contact info.
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBElement } from '../../base/FBElement';
import {
    schedulePreview,
    createdProjectId,
    completeTeamSetup,
    isTeamSetupStage,
} from '../../../store/onboarding-store';

interface TradeAssignment {
    wbsCode: string;
    phaseName: string;
    requiredTrade: string;
    startDate: string;
    contactName: string;
    contactEmail: string;
    contactPhone: string;
    skipped: boolean;
}

@customElement('fb-onboard-team-setup')
export class FBOnboardTeamSetup extends FBElement {
    @state() private assignments: TradeAssignment[] = [];
    @state() private submitting = false;

    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                padding: var(--spacing-lg);
            }
            .header {
                margin-bottom: var(--spacing-lg);
            }
            .header h2 {
                color: var(--text-primary);
                margin: 0 0 var(--spacing-xs) 0;
            }
            .header p {
                color: var(--text-secondary);
                margin: 0;
            }
            .trade-list {
                display: flex;
                flex-direction: column;
                gap: var(--spacing-md);
            }
            .trade-card {
                background: var(--surface-secondary);
                border: 1px solid var(--border-primary);
                border-radius: var(--radius-md);
                padding: var(--spacing-md);
            }
            .trade-card.skipped {
                opacity: 0.5;
            }
            .trade-header {
                display: flex;
                justify-content: space-between;
                align-items: center;
                margin-bottom: var(--spacing-sm);
            }
            .trade-name {
                font-weight: 600;
                color: var(--text-primary);
            }
            .trade-phase {
                font-size: var(--text-sm);
                color: var(--text-secondary);
            }
            .trade-start {
                font-size: var(--text-sm);
                color: var(--accent-primary);
            }
            .inputs {
                display: grid;
                grid-template-columns: 1fr 1fr 1fr;
                gap: var(--spacing-sm);
            }
            input {
                padding: var(--spacing-xs) var(--spacing-sm);
                background: var(--surface-primary);
                border: 1px solid var(--border-primary);
                border-radius: var(--radius-sm);
                color: var(--text-primary);
                font-size: var(--text-sm);
            }
            input::placeholder {
                color: var(--text-tertiary);
            }
            .actions {
                display: flex;
                gap: var(--spacing-sm);
                margin-top: var(--spacing-sm);
            }
            .btn-skip {
                font-size: var(--text-sm);
                color: var(--text-secondary);
                background: none;
                border: none;
                cursor: pointer;
                text-decoration: underline;
            }
            .footer {
                display: flex;
                justify-content: flex-end;
                gap: var(--spacing-md);
                margin-top: var(--spacing-lg);
            }
            .btn-primary {
                padding: var(--spacing-sm) var(--spacing-lg);
                background: var(--accent-primary);
                color: var(--text-on-accent);
                border: none;
                border-radius: var(--radius-sm);
                font-weight: 600;
                cursor: pointer;
            }
            .btn-secondary {
                padding: var(--spacing-sm) var(--spacing-lg);
                background: var(--surface-secondary);
                color: var(--text-primary);
                border: 1px solid var(--border-primary);
                border-radius: var(--radius-sm);
                cursor: pointer;
            }
        `,
    ];

    connectedCallback(): void {
        super.connectedCallback();
        this.loadTradeGaps();
    }

    private loadTradeGaps(): void {
        const preview = schedulePreview.value;
        if (!preview?.trade_gaps) {
            this.assignments = [];
            return;
        }
        this.assignments = preview.trade_gaps
            .filter(g => !g.has_contact)
            .map(g => ({
                wbsCode: g.wbs_code,
                phaseName: g.phase_name,
                requiredTrade: g.required_trade,
                startDate: g.start_date,
                contactName: '',
                contactEmail: '',
                contactPhone: '',
                skipped: false,
            }));
    }

    private updateAssignment(index: number, field: 'contactName' | 'contactEmail' | 'contactPhone', value: string): void {
        const updated = [...this.assignments];
        const assignment = updated[index];
        if (assignment) {
            assignment[field] = value;
            this.assignments = updated;
        }
    }

    private toggleSkip(index: number): void {
        const updated = [...this.assignments];
        const assignment = updated[index];
        if (assignment) {
            assignment.skipped = !assignment.skipped;
            this.assignments = updated;
        }
    }

    private async handleSubmit(): Promise<void> {
        const projectId = createdProjectId.value;
        if (!projectId) return;

        const toSubmit = this.assignments.filter(
            a => !a.skipped && (a.contactName || a.contactEmail)
        );
        if (toSubmit.length === 0) {
            completeTeamSetup();
            return;
        }

        this.submitting = true;
        try {
            // Emit event for parent to handle API call
            this.emit('team-assignments', {
                projectId,
                assignments: toSubmit.map(a => ({
                    wbs_code: a.wbsCode,
                    trade: a.requiredTrade,
                    contact_name: a.contactName,
                    contact_email: a.contactEmail,
                    contact_phone: a.contactPhone,
                })),
            });
            completeTeamSetup();
        } finally {
            this.submitting = false;
        }
    }

    private handleSkipAll(): void {
        completeTeamSetup();
    }

    protected render(): TemplateResult {
        if (!isTeamSetupStage.value) return html`${nothing}`;

        return html`
            <div class="header">
                <h2>Assign Your Team</h2>
                <p>Assign subcontractors to key phases. You can skip any and add them later.</p>
            </div>

            <div class="trade-list">
                ${this.assignments.map((a, i) => html`
                    <div class="trade-card ${a.skipped ? 'skipped' : ''}">
                        <div class="trade-header">
                            <div>
                                <span class="trade-name">${a.requiredTrade}</span>
                                <span class="trade-phase"> — ${a.phaseName} (${a.wbsCode})</span>
                            </div>
                            <span class="trade-start">Starts ${a.startDate}</span>
                        </div>
                        ${a.skipped ? nothing : html`
                            <div class="inputs">
                                <input
                                    type="text"
                                    placeholder="Contact name"
                                    .value=${a.contactName}
                                    @input=${(e: Event) => this.updateAssignment(i, 'contactName', (e.target as HTMLInputElement).value)}
                                />
                                <input
                                    type="email"
                                    placeholder="Email"
                                    .value=${a.contactEmail}
                                    @input=${(e: Event) => this.updateAssignment(i, 'contactEmail', (e.target as HTMLInputElement).value)}
                                />
                                <input
                                    type="tel"
                                    placeholder="Phone"
                                    .value=${a.contactPhone}
                                    @input=${(e: Event) => this.updateAssignment(i, 'contactPhone', (e.target as HTMLInputElement).value)}
                                />
                            </div>
                        `}
                        <div class="actions">
                            <button class="btn-skip" @click=${() => this.toggleSkip(i)}>
                                ${a.skipped ? 'Assign' : 'Skip for now'}
                            </button>
                        </div>
                    </div>
                `)}
            </div>

            <div class="footer">
                <button class="btn-secondary" @click=${this.handleSkipAll}>Skip All</button>
                <button class="btn-primary" @click=${this.handleSubmit} ?disabled=${this.submitting}>
                    ${this.submitting ? 'Saving...' : 'Save & Continue'}
                </button>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-onboard-team-setup': FBOnboardTeamSetup;
    }
}
