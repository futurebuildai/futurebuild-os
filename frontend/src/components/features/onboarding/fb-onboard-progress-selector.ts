/**
 * FBOnboardProgressSelector - In-Progress Project Phase Selector
 *
 * When a builder indicates their project is already under construction,
 * this component presents a phase-by-phase checklist to mark completed work.
 *
 * Each checked phase asks for an approximate completion date.
 * Updates the completedPhases signal for schedule preview calculation.
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBElement } from '../../base/FBElement';
import {
    isInProgressProject,
    completedPhases,
    setInProgressProject,
    setCompletedPhases,
} from '../../../store/onboarding-store';
import type { CompletedPhaseInput } from '../../../types/schedule';

interface PhaseOption {
    code: string;
    name: string;
    checked: boolean;
    actualEnd: string;
}

const CONSTRUCTION_PHASES: Array<{ code: string; name: string }> = [
    { code: '7.x', name: 'Site Prep' },
    { code: '8.x', name: 'Foundation' },
    { code: '9.x', name: 'Framing & Dry-In' },
    { code: '10.x', name: 'Rough-Ins (Plumbing, Electrical, HVAC)' },
    { code: '11.x', name: 'Insulation & Drywall' },
    { code: '12.x', name: 'Interior Finishes' },
    { code: '13.x', name: 'Exterior Finishes' },
    { code: '14.x', name: 'Final Inspections & Closeout' },
];

@customElement('fb-onboard-progress-selector')
export class FBOnboardProgressSelector extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                padding: var(--fb-spacing-md);
            }

            .selector-container {
                background: var(--fb-surface-1);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-md);
            }

            .selector-header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                margin-bottom: var(--fb-spacing-md);
            }

            .selector-title {
                font-size: var(--fb-text-sm);
                font-weight: 600;
                color: var(--fb-text-primary);
            }

            .toggle-row {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm);
                margin-bottom: var(--fb-spacing-md);
                padding-bottom: var(--fb-spacing-md);
                border-bottom: 1px solid var(--fb-border);
            }

            .toggle-label {
                font-size: var(--fb-text-sm);
                color: var(--fb-text-secondary);
            }

            .toggle-switch {
                position: relative;
                width: 40px;
                height: 22px;
                cursor: pointer;
            }

            .toggle-switch input {
                opacity: 0;
                width: 0;
                height: 0;
            }

            .toggle-track {
                position: absolute;
                inset: 0;
                background: var(--fb-bg-tertiary);
                border-radius: 11px;
                transition: background 0.2s;
            }

            .toggle-switch input:checked + .toggle-track {
                background: var(--fb-accent);
            }

            .toggle-thumb {
                position: absolute;
                top: 2px;
                left: 2px;
                width: 18px;
                height: 18px;
                background: white;
                border-radius: 50%;
                transition: transform 0.2s;
            }

            .toggle-switch input:checked ~ .toggle-thumb {
                transform: translateX(18px);
            }

            .phase-list {
                display: flex;
                flex-direction: column;
                gap: 6px;
            }

            .phase-row {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm);
                padding: 8px;
                border-radius: var(--fb-radius-sm);
                transition: background 0.15s;
            }

            .phase-row:hover {
                background: var(--fb-bg-tertiary);
            }

            .phase-checkbox {
                width: 18px;
                height: 18px;
                accent-color: var(--fb-accent);
                cursor: pointer;
                flex-shrink: 0;
            }

            .phase-name {
                flex: 1;
                font-size: var(--fb-text-sm);
                color: var(--fb-text-primary);
            }

            .phase-name.checked {
                color: var(--fb-accent);
            }

            .phase-date {
                width: 130px;
                flex-shrink: 0;
            }

            .phase-date input {
                width: 100%;
                padding: 4px 8px;
                background: var(--fb-bg-primary);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-sm);
                color: var(--fb-text-primary);
                font-size: 12px;
                font-family: inherit;
            }

            .phase-date input:focus {
                outline: none;
                border-color: var(--fb-accent);
            }

            .phase-code {
                width: 36px;
                font-size: 10px;
                color: var(--fb-text-muted);
                flex-shrink: 0;
            }

            .done-btn {
                margin-top: var(--fb-spacing-md);
                padding: 8px 16px;
                background: var(--fb-accent);
                color: var(--fb-bg-primary);
                border: none;
                border-radius: var(--fb-radius-sm);
                font-weight: 600;
                font-size: var(--fb-text-sm);
                cursor: pointer;
                transition: opacity 0.15s;
            }

            .done-btn:hover {
                opacity: 0.9;
            }

            .done-btn:disabled {
                opacity: 0.4;
                cursor: not-allowed;
            }

            .hidden {
                display: none;
            }
        `
    ];

    @state() private _phases: PhaseOption[] = CONSTRUCTION_PHASES.map(p => ({
        ...p,
        checked: false,
        actualEnd: '',
    }));

    @state() private _isInProgress = false;

    override connectedCallback(): void {
        super.connectedCallback();
        this._isInProgress = isInProgressProject.value;

        // Restore from existing signal state
        const existing = completedPhases.value;
        if (existing.length > 0) {
            const existingMap = new Map(existing.map(e => [e.wbs_code, e]));
            this._phases = this._phases.map(p => {
                const match = existingMap.get(p.code);
                if (match) {
                    return { ...p, checked: true, actualEnd: match.actual_end };
                }
                return p;
            });
        }
    }

    private _handleToggle(checked: boolean): void {
        this._isInProgress = checked;
        setInProgressProject(checked);
        if (!checked) {
            // Clear phases when turning off in-progress mode
            this._phases = this._phases.map(p => ({ ...p, checked: false, actualEnd: '' }));
            setCompletedPhases([]);
        }
    }

    private _handlePhaseCheck(index: number, checked: boolean): void {
        const updated = [...this._phases];
        const phase = updated[index];
        if (phase) {
            updated[index] = { ...phase, checked };
            if (checked) {
                // Auto-check all preceding phases (construction is sequential)
                for (let i = 0; i < index; i++) {
                    const p = updated[i];
                    if (p && !p.checked) {
                        updated[i] = { ...p, checked: true };
                    }
                }
            } else {
                // Uncheck all phases after this one
                for (let i = index + 1; i < updated.length; i++) {
                    const p = updated[i];
                    if (p) {
                        updated[i] = { ...p, checked: false, actualEnd: '' };
                    }
                }
            }
            this._phases = updated;
            this._syncToStore();
        }
    }

    private _handleDateChange(index: number, date: string): void {
        const updated = [...this._phases];
        const phase = updated[index];
        if (phase) {
            updated[index] = { ...phase, actualEnd: date };
            this._phases = updated;
            this._syncToStore();
        }
    }

    private _syncToStore(): void {
        const completed: CompletedPhaseInput[] = this._phases
            .filter(p => p.checked && p.actualEnd)
            .map(p => ({
                wbs_code: p.code,
                actual_end: p.actualEnd,
                status: 'completed' as const,
            }));
        setCompletedPhases(completed);
    }

    private _handleDone(): void {
        this._syncToStore();
        this.emit('fb-progress-set', {
            phases: completedPhases.value,
            isInProgress: this._isInProgress,
        });
    }

    override render(): TemplateResult {
        return html`
            <div class="selector-container">
                <div class="toggle-row">
                    <label class="toggle-switch">
                        <input
                            type="checkbox"
                            .checked=${this._isInProgress}
                            @change=${(e: Event): void => {
                                this._handleToggle((e.target as HTMLInputElement).checked);
                            }}
                        />
                        <span class="toggle-track"></span>
                        <span class="toggle-thumb"></span>
                    </label>
                    <span class="toggle-label">This project is already under construction</span>
                </div>

                <div class="${this._isInProgress ? '' : 'hidden'}">
                    <div class="selector-header">
                        <span class="selector-title">Which phases are completed?</span>
                    </div>

                    <div class="phase-list">
                        ${this._phases.map((phase, i) => this._renderPhaseRow(phase, i))}
                    </div>

                    <button
                        class="done-btn"
                        ?disabled=${!this._phases.some(p => p.checked && p.actualEnd)}
                        @click=${(): void => { this._handleDone(); }}
                    >
                        Update Schedule
                    </button>
                </div>
            </div>
        `;
    }

    private _renderPhaseRow(phase: PhaseOption, index: number): TemplateResult {
        return html`
            <div class="phase-row">
                <span class="phase-code">${phase.code}</span>
                <input
                    type="checkbox"
                    class="phase-checkbox"
                    .checked=${phase.checked}
                    @change=${(e: Event): void => {
                        this._handlePhaseCheck(index, (e.target as HTMLInputElement).checked);
                    }}
                />
                <span class="phase-name ${phase.checked ? 'checked' : ''}">${phase.name}</span>
                ${phase.checked ? html`
                    <div class="phase-date">
                        <input
                            type="date"
                            .value=${phase.actualEnd}
                            max=${new Date().toISOString().split('T')[0]}
                            @change=${(e: Event): void => {
                                this._handleDateChange(index, (e.target as HTMLInputElement).value);
                            }}
                            placeholder="Completion date"
                        />
                    </div>
                ` : nothing}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-onboard-progress-selector': FBOnboardProgressSelector;
    }
}
