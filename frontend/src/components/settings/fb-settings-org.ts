/**
 * fb-settings-org — Organization settings page.
 * See FRONTEND_V2_SPEC.md §10.2.B
 *
 * Route: /settings/org
 * Shows: Construction physics config (speed slider, work days), org info
 * Access: Admin and Builder roles only
 */
import { html, css, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { api } from '../../services/api';
import { store } from '../../store/store';
import { UserRole } from '../../types/enums';

const DAY_LABELS = ['M', 'T', 'W', 'T', 'F', 'S', 'S'] as const;

function getPaceLabel(value: number): string {
    if (value <= 0.8) return 'Aggressive (Fast Track)';
    if (value <= 1.1) return 'Standard (Industry Avg)';
    return 'Relaxed (Padding Added)';
}

function getPaceClass(value: number): string {
    if (value <= 0.8) return 'aggressive';
    if (value <= 1.1) return 'standard';
    return 'relaxed';
}

@customElement('fb-settings-org')
export class FBSettingsOrg extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                max-width: 600px;
                margin: 0 auto;
                padding: 32px 16px;
            }

            .header {
                margin-bottom: 32px;
            }

            .title {
                font-size: 24px;
                font-weight: 700;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .subtitle {
                font-size: 14px;
                color: var(--fb-text-secondary, #a0a0b0);
                margin-top: 4px;
            }

            .card {
                background: var(--fb-surface-1, #1a1a2e);
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 12px;
                padding: 24px;
                margin-bottom: 20px;
            }

            .card-title {
                font-size: 16px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
                margin-bottom: 20px;
                padding-bottom: 12px;
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
            }

            .form-group {
                margin-bottom: 24px;
            }

            .form-group:last-child {
                margin-bottom: 0;
            }

            label {
                display: block;
                font-size: 13px;
                font-weight: 500;
                color: var(--fb-text-secondary, #a0a0b0);
                margin-bottom: 8px;
            }

            .slider-container {
                display: flex;
                align-items: center;
                gap: 16px;
            }

            .slider-wrapper {
                flex: 1;
                position: relative;
            }

            .speed-slider {
                width: 100%;
                -webkit-appearance: none;
                appearance: none;
                height: 6px;
                background: var(--fb-surface-2, #252540);
                border-radius: 3px;
                outline: none;
                border: none;
            }

            .speed-slider::-webkit-slider-thumb {
                -webkit-appearance: none;
                width: 18px;
                height: 18px;
                border-radius: 50%;
                background: var(--fb-accent, #6366f1);
                cursor: pointer;
                border: 2px solid var(--fb-surface-1, #1a1a2e);
                box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
            }

            .speed-slider:disabled {
                opacity: 0.5;
                cursor: not-allowed;
            }

            .pace-badge {
                font-size: 12px;
                font-weight: 600;
                padding: 6px 12px;
                border-radius: 6px;
                white-space: nowrap;
                min-width: 130px;
                text-align: center;
            }

            .pace-badge.aggressive {
                background: rgba(239, 68, 68, 0.15);
                color: #ef4444;
            }

            .pace-badge.standard {
                background: rgba(34, 197, 94, 0.15);
                color: #22c55e;
            }

            .pace-badge.relaxed {
                background: rgba(59, 130, 246, 0.15);
                color: #3b82f6;
            }

            .speed-value {
                font-size: 12px;
                color: var(--fb-text-tertiary, #707080);
                text-align: center;
                margin-top: 8px;
            }

            .workdays-container {
                display: flex;
                gap: 8px;
            }

            .day-toggle {
                width: 42px;
                height: 42px;
                border-radius: 8px;
                border: 1px solid var(--fb-border, #2a2a3e);
                background: var(--fb-surface-2, #252540);
                color: var(--fb-text-tertiary, #707080);
                font-size: 14px;
                font-weight: 600;
                cursor: pointer;
                transition: all 0.15s ease;
            }

            .day-toggle:hover:not(:disabled) {
                border-color: var(--fb-accent, #6366f1);
            }

            .day-toggle:disabled {
                opacity: 0.5;
                cursor: not-allowed;
            }

            .day-toggle.active {
                background: var(--fb-accent, #6366f1);
                color: #fff;
                border-color: var(--fb-accent, #6366f1);
            }

            .actions {
                display: flex;
                gap: 12px;
                margin-top: 24px;
                padding-top: 20px;
                border-top: 1px solid var(--fb-border, #2a2a3e);
            }

            .btn {
                padding: 10px 20px;
                border-radius: 8px;
                font-size: 14px;
                font-weight: 600;
                cursor: pointer;
                border: none;
                transition: all 0.15s ease;
            }

            .btn-primary {
                background: var(--fb-accent, #6366f1);
                color: #fff;
            }

            .btn-primary:hover:not(:disabled) {
                opacity: 0.9;
            }

            .btn-primary:disabled {
                opacity: 0.5;
                cursor: not-allowed;
            }

            .btn-secondary {
                background: transparent;
                border: 1px solid var(--fb-border, #2a2a3e);
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .btn-secondary:hover {
                background: var(--fb-surface-2, #252540);
            }

            .message {
                padding: 10px 14px;
                border-radius: 8px;
                font-size: 13px;
                margin-bottom: 16px;
            }

            .message-success {
                background: rgba(34, 197, 94, 0.1);
                color: #22c55e;
            }

            .message-error {
                background: rgba(239, 68, 68, 0.1);
                color: #ef4444;
            }

            .message-warning {
                background: rgba(245, 158, 11, 0.1);
                color: #f59e0b;
            }

            .readonly-banner {
                padding: 12px 16px;
                background: var(--fb-surface-2, #252540);
                border-radius: 8px;
                font-size: 13px;
                color: var(--fb-text-secondary, #a0a0b0);
                margin-bottom: 20px;
            }

            .back-link {
                display: inline-flex;
                align-items: center;
                gap: 6px;
                font-size: 13px;
                color: var(--fb-text-secondary, #a0a0b0);
                cursor: pointer;
                margin-bottom: 16px;
            }

            .back-link:hover {
                color: var(--fb-text-primary, #e0e0e0);
            }

            .no-access {
                text-align: center;
                padding: 60px 24px;
            }

            .no-access-title {
                font-size: 20px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
                margin-bottom: 8px;
            }

            .no-access-body {
                font-size: 14px;
                color: var(--fb-text-secondary, #a0a0b0);
            }
        `,
    ];

    @state() private _loading = true;
    @state() private _saving = false;
    @state() private _canEdit = false;
    @state() private _speedMultiplier = 1.0;
    @state() private _workDays: number[] = [1, 2, 3, 4, 5];
    @state() private _dirty = false;
    @state() private _success = '';
    @state() private _error = '';

    private _disposeEffect: (() => void) | null = null;

    override connectedCallback() {
        super.connectedCallback();
        this._disposeEffect = effect(() => {
            const user = store.user$.value;
            if (user) {
                this._canEdit = user.role === UserRole.Admin || user.role === UserRole.Builder;
            }
        });
        this._loadPhysics();
    }

    override disconnectedCallback() {
        super.disconnectedCallback();
        this._disposeEffect?.();
    }

    private async _loadPhysics() {
        this._loading = true;
        try {
            const config = await api.settings.getPhysics();
            this._speedMultiplier = config.speed_multiplier;
            this._workDays = config.work_days;
            this._dirty = false;
        } catch (err) {
            console.warn('[FBSettingsOrg] Failed to load physics:', err);
        } finally {
            this._loading = false;
        }
    }

    private _handleSpeedChange(e: Event) {
        this._speedMultiplier = Math.round(parseFloat((e.target as HTMLInputElement).value) * 10) / 10;
        this._dirty = true;
    }

    private _toggleWorkDay(dayNum: number) {
        const idx = this._workDays.indexOf(dayNum);
        if (idx >= 0) {
            if (this._workDays.length <= 1) return;
            this._workDays = this._workDays.filter(d => d !== dayNum);
        } else {
            this._workDays = [...this._workDays, dayNum].sort((a, b) => a - b);
        }
        this._dirty = true;
    }

    private async _handleSave() {
        this._saving = true;
        this._error = '';
        this._success = '';

        try {
            const config = await api.settings.updatePhysics({
                speed_multiplier: this._speedMultiplier,
                work_days: this._workDays,
            });
            this._speedMultiplier = config.speed_multiplier;
            this._workDays = config.work_days;
            this._dirty = false;
            this._success = 'Settings saved successfully';
            setTimeout(() => { this._success = ''; }, 3000);
        } catch (err) {
            this._error = err instanceof Error ? err.message : 'Failed to save settings';
        } finally {
            this._saving = false;
        }
    }

    private _handleReset() {
        this._speedMultiplier = 1.0;
        this._workDays = [1, 2, 3, 4, 5];
        this._dirty = true;
    }

    private _handleBack() {
        this.emit('fb-navigate', { view: 'home' });
    }

    override render() {
        void this._loading; // Suppress unused warning - used for future loading skeleton
        if (!this._canEdit) {
            return html`
                <div class="back-link" @click=${this._handleBack}>← Back to Feed</div>
                <div class="no-access">
                    <div class="no-access-title">Access Restricted</div>
                    <div class="no-access-body">
                        Organization settings are only available to Admin and Builder roles.
                    </div>
                </div>
            `;
        }

        const paceLabel = getPaceLabel(this._speedMultiplier);
        const paceClass = getPaceClass(this._speedMultiplier);

        return html`
            <div class="back-link" @click=${this._handleBack}>← Back to Feed</div>

            <div class="header">
                <div class="title">Organization</div>
                <div class="subtitle">Configure construction physics and scheduling</div>
            </div>

            ${this._success ? html`<div class="message message-success">${this._success}</div>` : nothing}
            ${this._error ? html`<div class="message message-error">${this._error}</div>` : nothing}

            <div class="card">
                <div class="card-title">Construction Physics</div>

                <div class="form-group">
                    <label>Schedule Pace</label>
                    <div class="slider-container">
                        <div class="slider-wrapper">
                            <input
                                type="range"
                                class="speed-slider"
                                min="0.5"
                                max="1.5"
                                step="0.1"
                                .value=${String(this._speedMultiplier)}
                                ?disabled=${this._saving}
                                @input=${this._handleSpeedChange}
                            />
                        </div>
                        <span class="pace-badge ${paceClass}">${paceLabel}</span>
                    </div>
                    <div class="speed-value">${this._speedMultiplier.toFixed(1)}x multiplier</div>
                </div>

                <div class="form-group">
                    <label>Standard Work Week</label>
                    <div class="workdays-container">
                        ${DAY_LABELS.map((label, i) => {
                            const dayNum = i + 1;
                            const isActive = this._workDays.includes(dayNum);
                            const fullName = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'][i];
                            return html`
                                <button
                                    class="day-toggle ${isActive ? 'active' : ''}"
                                    ?disabled=${this._saving}
                                    @click=${() => this._toggleWorkDay(dayNum)}
                                    title=${fullName}
                                    aria-label=${fullName}
                                    aria-pressed=${isActive}
                                >
                                    ${label}
                                </button>
                            `;
                        })}
                    </div>
                </div>

                ${this._dirty ? html`
                    <div class="actions">
                        <button class="btn btn-secondary" ?disabled=${this._saving} @click=${this._handleReset}>
                            Reset to Default
                        </button>
                        <button class="btn btn-primary" ?disabled=${this._saving} @click=${this._handleSave}>
                            ${this._saving ? 'Saving...' : 'Save Changes'}
                        </button>
                    </div>
                ` : nothing}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-settings-org': FBSettingsOrg;
    }
}
