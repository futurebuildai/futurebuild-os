/**
 * fb-engine-calibration — First-project-only work days + inspection latency.
 * See FRONTEND_V2_SPEC.md §2.3.B
 *
 * Shown after extraction, before activation. Subsequent projects inherit
 * org settings silently. Two quick inputs: crew work days and inspection latency.
 */
import { html, css, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBElement } from '../../base/FBElement';
import { api } from '../../../services/api';

const DAY_LABELS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'] as const;
const LATENCY_OPTIONS = [1, 2, 3, 4, 5, 7, 10, 14];

@customElement('fb-engine-calibration')
export class FBEngineCalibration extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                max-width: 520px;
                margin: 0 auto;
                padding: 40px 24px;
            }

            .intro {
                font-size: 16px;
                color: var(--fb-text-primary, #e0e0e0);
                margin-bottom: 32px;
                line-height: 1.5;
            }

            .section {
                margin-bottom: 28px;
            }

            .section-label {
                font-size: 14px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
                margin-bottom: 12px;
            }

            .day-row {
                display: flex;
                gap: 8px;
                flex-wrap: wrap;
            }

            .day-btn {
                width: 48px;
                height: 48px;
                border-radius: 50%;
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 13px;
                font-weight: 600;
                cursor: pointer;
                border: 2px solid var(--fb-border, #2a2a3e);
                background: transparent;
                color: var(--fb-text-secondary, #a0a0b0);
                transition: all 0.15s ease;
            }

            .day-btn:hover {
                border-color: var(--fb-accent, #6366f1);
            }

            .day-btn[data-active] {
                background: var(--fb-accent, #6366f1);
                border-color: var(--fb-accent, #6366f1);
                color: #fff;
            }

            .latency-row {
                display: flex;
                gap: 24px;
                align-items: center;
            }

            .latency-group {
                display: flex;
                flex-direction: column;
                gap: 6px;
            }

            .latency-label {
                font-size: 13px;
                color: var(--fb-text-secondary, #a0a0b0);
            }

            select {
                padding: 8px 12px;
                border-radius: 6px;
                border: 1px solid var(--fb-border, #2a2a3e);
                background: var(--fb-surface-2, #252540);
                color: var(--fb-text-primary, #e0e0e0);
                font-size: 14px;
                outline: none;
                cursor: pointer;
                min-width: 120px;
            }

            select:focus {
                border-color: var(--fb-accent, #6366f1);
            }

            .btn-row {
                display: flex;
                gap: 12px;
                margin-top: 32px;
            }

            button.btn-primary {
                padding: 12px 24px;
                border-radius: 8px;
                background: var(--fb-accent, #6366f1);
                color: #fff;
                font-size: 14px;
                font-weight: 600;
                border: none;
                cursor: pointer;
            }

            button.btn-primary:hover {
                opacity: 0.9;
            }

            button.btn-primary:disabled {
                opacity: 0.5;
                cursor: not-allowed;
            }

            button.btn-skip {
                padding: 12px 24px;
                border-radius: 8px;
                background: transparent;
                color: var(--fb-text-secondary, #a0a0b0);
                font-size: 14px;
                font-weight: 500;
                border: 1px solid var(--fb-border, #2a2a3e);
                cursor: pointer;
            }

            button.btn-skip:hover {
                border-color: var(--fb-text-secondary, #a0a0b0);
                color: var(--fb-text-primary, #e0e0e0);
            }

            .error {
                font-size: 13px;
                color: #ef4444;
                margin-top: 8px;
            }
        `,
    ];

    // Default: Mon-Fri (1-5 in JS Sunday=0 convention, but API uses 1=Mon...7=Sun)
    @state() private _workDays: number[] = [1, 2, 3, 4, 5];
    @state() private _roughLatency = 3;
    @state() private _finalLatency = 5;
    @state() private _saving = false;
    @state() private _error = '';

    private _toggleDay(dayIndex: number) {
        // dayIndex: 0=Sun, 1=Mon, ..., 6=Sat
        // API work_days uses 1=Mon, 2=Tue, ..., 5=Fri, 6=Sat, 7=Sun
        const apiDay = dayIndex === 0 ? 7 : dayIndex;
        if (this._workDays.includes(apiDay)) {
            this._workDays = this._workDays.filter((d) => d !== apiDay);
        } else {
            this._workDays = [...this._workDays, apiDay].sort((a, b) => a - b);
        }
    }

    private _isDayActive(dayIndex: number): boolean {
        const apiDay = dayIndex === 0 ? 7 : dayIndex;
        return this._workDays.includes(apiDay);
    }

    private async _handleApply() {
        if (this._workDays.length === 0) {
            this._error = 'Select at least one work day';
            return;
        }

        this._saving = true;
        this._error = '';

        try {
            await api.settings.updatePhysics({
                speed_multiplier: 1.0,
                work_days: this._workDays,
            });

            this.emit('fb-calibration-applied', {
                workDays: this._workDays,
                roughLatency: this._roughLatency,
                finalLatency: this._finalLatency,
            });
        } catch (err) {
            this._error = err instanceof Error ? err.message : 'Failed to save settings';
        } finally {
            this._saving = false;
        }
    }

    private _handleSkip() {
        this.emit('fb-calibration-skipped', {});
    }

    override render() {
        return html`
            <div class="intro">
                Before I activate your schedule, two quick ones:
            </div>

            <div class="section">
                <div class="section-label">Which days does your crew work?</div>
                <div class="day-row">
                    ${DAY_LABELS.map(
                        (label, i) => html`
                            <button
                                class="day-btn"
                                ?data-active=${this._isDayActive(i)}
                                @click=${() => this._toggleDay(i)}
                            >
                                ${label}
                            </button>
                        `
                    )}
                </div>
            </div>

            <div class="section">
                <div class="section-label">How long do inspections typically take in your area?</div>
                <div class="latency-row">
                    <div class="latency-group">
                        <span class="latency-label">Rough</span>
                        <select
                            .value=${String(this._roughLatency)}
                            @change=${(e: Event) => { this._roughLatency = Number((e.target as HTMLSelectElement).value); }}
                        >
                            ${LATENCY_OPTIONS.map(
                                (d) => html`<option value=${d} ?selected=${d === this._roughLatency}>${d} day${d > 1 ? 's' : ''}</option>`
                            )}
                        </select>
                    </div>
                    <div class="latency-group">
                        <span class="latency-label">Final</span>
                        <select
                            .value=${String(this._finalLatency)}
                            @change=${(e: Event) => { this._finalLatency = Number((e.target as HTMLSelectElement).value); }}
                        >
                            ${LATENCY_OPTIONS.map(
                                (d) => html`<option value=${d} ?selected=${d === this._finalLatency}>${d} day${d > 1 ? 's' : ''}</option>`
                            )}
                        </select>
                    </div>
                </div>
            </div>

            ${this._error ? html`<div class="error">${this._error}</div>` : nothing}

            <div class="btn-row">
                <button class="btn-primary" ?disabled=${this._saving} @click=${this._handleApply}>
                    ${this._saving ? 'Applying...' : 'Apply \u2192 Activate project'}
                </button>
                <button class="btn-skip" @click=${this._handleSkip}>
                    Use defaults and skip
                </button>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-engine-calibration': FBEngineCalibration;
    }
}
