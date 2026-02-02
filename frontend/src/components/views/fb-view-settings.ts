/**
 * FBViewSettings - User Settings View
 * See LAUNCH_PLAN.md Section: User Settings View (P1)
 *
 * Allows users to:
 * - View their profile information
 * - Update their name
 * - View organization info (readonly)
 * - Configure Construction Physics (Step 86)
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { api, UserProfile } from '../../services/api';

type ViewState = 'loading' | 'ready' | 'saving' | 'error';

/** Day labels for the work week toggle buttons. */
const DAY_LABELS = ['M', 'T', 'W', 'T', 'F', 'S', 'S'] as const;

/** Maps slider value ranges to human-readable pace descriptions. */
function getPaceLabel(value: number): string {
    if (value <= 0.8) return 'Aggressive (Fast Track)';
    if (value <= 1.1) return 'Standard (Industry Avg)';
    return 'Relaxed (Padding Added)';
}

/** Maps slider value ranges to CSS modifier classes. */
function getPaceClass(value: number): string {
    if (value <= 0.8) return 'aggressive';
    if (value <= 1.1) return 'standard';
    return 'relaxed';
}

@customElement('fb-view-settings')
export class FBViewSettings extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                display: flex;
                flex-direction: column;
                background: var(--fb-bg-primary);
                padding: var(--fb-spacing-xl);
            }

            .header {
                margin-bottom: var(--fb-spacing-xl);
            }

            .title {
                font-size: var(--fb-text-2xl);
                font-weight: 600;
                color: var(--fb-text-primary);
            }

            .subtitle {
                font-size: var(--fb-text-sm);
                color: var(--fb-text-secondary);
                margin-top: var(--fb-spacing-xs);
            }

            .settings-card {
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-xl);
                margin-bottom: var(--fb-spacing-lg);
            }

            .card-title {
                font-size: var(--fb-text-lg);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin-bottom: var(--fb-spacing-lg);
                padding-bottom: var(--fb-spacing-sm);
                border-bottom: 1px solid var(--fb-border-light);
            }

            .form-group {
                margin-bottom: var(--fb-spacing-lg);
            }

            .form-group:last-child {
                margin-bottom: 0;
            }

            label {
                display: block;
                font-size: var(--fb-text-sm);
                font-weight: 500;
                color: var(--fb-text-secondary);
                margin-bottom: var(--fb-spacing-xs);
            }

            /* H2: Scope text input styles to exclude range sliders */
            input:not([type="range"]) {
                width: 100%;
                padding: var(--fb-spacing-md);
                background: var(--fb-bg-tertiary);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-md);
                color: var(--fb-text-primary);
                font-size: var(--fb-text-base);
                box-sizing: border-box;
            }

            input:not([type="range"]):focus {
                outline: none;
                border-color: var(--fb-primary);
            }

            input:not([type="range"]):disabled {
                opacity: 0.6;
                cursor: not-allowed;
            }

            input[readonly] {
                background: var(--fb-bg-secondary);
                cursor: default;
            }

            .readonly-hint {
                font-size: var(--fb-text-xs);
                color: var(--fb-text-muted);
                margin-top: var(--fb-spacing-xs);
            }

            .btn-primary {
                padding: var(--fb-spacing-sm) var(--fb-spacing-lg);
                background: var(--fb-primary);
                color: white;
                border: none;
                border-radius: var(--fb-radius-md);
                font-size: var(--fb-text-sm);
                font-weight: 600;
                cursor: pointer;
                transition: background 0.2s ease;
            }

            .btn-primary:hover:not(:disabled) {
                background: var(--fb-primary-hover);
            }

            .btn-primary:disabled {
                opacity: 0.6;
                cursor: not-allowed;
            }

            .form-actions {
                display: flex;
                gap: var(--fb-spacing-sm);
                margin-top: var(--fb-spacing-lg);
                padding-top: var(--fb-spacing-lg);
                border-top: 1px solid var(--fb-border-light);
            }

            .loading-state {
                display: flex;
                justify-content: center;
                align-items: center;
                padding: var(--fb-spacing-2xl);
            }

            .spinner {
                width: 32px;
                height: 32px;
                border: 3px solid var(--fb-border);
                border-top-color: var(--fb-primary);
                border-radius: 50%;
                animation: spin 1s linear infinite;
            }

            @keyframes spin {
                to { transform: rotate(360deg); }
            }

            .error-state {
                text-align: center;
                padding: var(--fb-spacing-2xl);
                color: var(--fb-error);
            }

            .btn-secondary {
                padding: var(--fb-spacing-sm) var(--fb-spacing-lg);
                background: transparent;
                color: var(--fb-text-secondary);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-md);
                font-size: var(--fb-text-sm);
                cursor: pointer;
                transition: all 0.2s ease;
            }

            .btn-secondary:hover {
                background: var(--fb-bg-tertiary);
                color: var(--fb-text-primary);
            }

            .success-message {
                color: var(--fb-success);
                font-size: var(--fb-text-sm);
                margin-bottom: var(--fb-spacing-md);
                padding: var(--fb-spacing-sm) var(--fb-spacing-md);
                background: rgba(46, 125, 50, 0.1);
                border-radius: var(--fb-radius-sm);
            }

            .error-message {
                color: var(--fb-error);
                font-size: var(--fb-text-sm);
                margin-bottom: var(--fb-spacing-md);
                padding: var(--fb-spacing-sm) var(--fb-spacing-md);
                background: rgba(198, 40, 40, 0.1);
                border-radius: var(--fb-radius-sm);
            }

            .info-message {
                color: var(--fb-text-secondary);
                font-size: var(--fb-text-xs);
                margin-bottom: var(--fb-spacing-md);
                padding: var(--fb-spacing-sm) var(--fb-spacing-md);
                background: var(--fb-bg-tertiary);
                border-radius: var(--fb-radius-sm);
            }

            .role-badge {
                display: inline-block;
                padding: var(--fb-spacing-xs) var(--fb-spacing-sm);
                background: var(--fb-primary);
                color: white;
                border-radius: var(--fb-radius-sm);
                font-size: var(--fb-text-xs);
                font-weight: 500;
                text-transform: uppercase;
            }

            .info-row {
                display: flex;
                justify-content: space-between;
                align-items: center;
                padding: var(--fb-spacing-md) 0;
                border-bottom: 1px solid var(--fb-border-light);
            }

            .info-row:last-child {
                border-bottom: none;
            }

            .info-label {
                font-size: var(--fb-text-sm);
                color: var(--fb-text-secondary);
            }

            .info-value {
                font-size: var(--fb-text-sm);
                color: var(--fb-text-primary);
                font-weight: 500;
            }

            /* Step 86: Construction Physics Card */
            .physics-section-label {
                display: block;
                font-size: var(--fb-text-sm);
                font-weight: 500;
                color: var(--fb-text-secondary);
                margin-bottom: var(--fb-spacing-sm);
            }

            .slider-container {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-md);
            }

            .speed-slider {
                flex: 1;
                -webkit-appearance: none;
                appearance: none;
                height: 6px;
                background: var(--fb-bg-tertiary);
                border-radius: 3px;
                outline: none;
                border: none;
                padding: 0;
            }

            .speed-slider::-webkit-slider-thumb {
                -webkit-appearance: none;
                appearance: none;
                width: 20px;
                height: 20px;
                border-radius: 50%;
                background: var(--fb-primary);
                cursor: pointer;
                border: 2px solid var(--fb-bg-card);
                box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
                transition: transform 0.15s ease;
            }

            .speed-slider::-webkit-slider-thumb:hover {
                transform: scale(1.15);
            }

            .speed-slider::-moz-range-thumb {
                width: 20px;
                height: 20px;
                border-radius: 50%;
                background: var(--fb-primary);
                cursor: pointer;
                border: 2px solid var(--fb-bg-card);
                box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
            }

            .pace-badge {
                font-size: var(--fb-text-xs);
                font-weight: 600;
                padding: var(--fb-spacing-xs) var(--fb-spacing-sm);
                border-radius: var(--fb-radius-sm);
                white-space: nowrap;
                min-width: 140px;
                text-align: center;
            }

            .pace-badge.aggressive {
                background: rgba(198, 40, 40, 0.15);
                color: #ef5350;
            }

            .pace-badge.standard {
                background: rgba(46, 125, 50, 0.15);
                color: #66bb6a;
            }

            .pace-badge.relaxed {
                background: rgba(33, 150, 243, 0.15);
                color: #42a5f5;
            }

            .speed-value {
                font-size: var(--fb-text-xs);
                color: var(--fb-text-muted);
                text-align: center;
                margin-top: var(--fb-spacing-xs);
            }

            .workdays-container {
                display: flex;
                gap: var(--fb-spacing-xs);
                flex-wrap: wrap;
            }

            .day-toggle {
                width: 40px;
                height: 40px;
                border-radius: var(--fb-radius-md);
                border: 1px solid var(--fb-border);
                background: var(--fb-bg-tertiary);
                color: var(--fb-text-muted);
                font-size: var(--fb-text-sm);
                font-weight: 600;
                cursor: pointer;
                transition: all 0.15s ease;
                display: flex;
                align-items: center;
                justify-content: center;
            }

            .day-toggle:hover {
                border-color: var(--fb-primary);
            }

            .day-toggle.active {
                background: var(--fb-primary);
                color: white;
                border-color: var(--fb-primary);
            }

            .physics-actions {
                display: flex;
                gap: var(--fb-spacing-sm);
                margin-top: var(--fb-spacing-lg);
                padding-top: var(--fb-spacing-lg);
                border-top: 1px solid var(--fb-border-light);
            }
        `,
    ];

    @state() private _viewState: ViewState = 'loading';
    @state() private _profile: UserProfile | null = null;
    @state() private _editName = '';
    @state() private _error = '';
    @state() private _success = '';

    // Step 86: Construction Physics state
    @state() private _speedMultiplier = 1.0;
    @state() private _workDays: number[] = [1, 2, 3, 4, 5];
    @state() private _physicsDirty = false;
    @state() private _physicsSuccess = ''; // M1: Separate from profile success
    @state() private _physicsError = '';   // M3: Separate from profile error
    @state() private _physicsUsingDefaults = false; // Deep dive: track default fallback

    override connectedCallback(): void {
        super.connectedCallback();
        void this._loadProfile();
        void this._loadPhysics(); // Step 87: Fetch physics config on mount
    }

    private async _loadProfile(): Promise<void> {
        this._viewState = 'loading';
        this._error = '';
        this._success = '';

        try {
            this._profile = await api.users.getMe();
            this._editName = this._profile.name;
            this._viewState = 'ready';
        } catch (err) {
            this._viewState = 'error';
            this._error = err instanceof Error ? err.message : 'Failed to load profile';
        }
    }

    private async _handleSave(): Promise<void> {
        if (!this._editName.trim()) {
            this._error = 'Name is required';
            return;
        }

        this._viewState = 'saving';
        this._error = '';
        this._success = '';

        try {
            this._profile = await api.users.updateMe({ name: this._editName.trim() });
            this._viewState = 'ready';
            this._success = 'Profile updated successfully';

            // Clear success message after 3 seconds
            setTimeout(() => {
                this._success = '';
            }, 3000);
        } catch (err) {
            this._viewState = 'ready';
            this._error = err instanceof Error ? err.message : 'Failed to update profile';
        }
    }

    private _handleCancel(): void {
        if (this._profile) {
            this._editName = this._profile.name;
        }
        this._error = '';
        this._success = '';
    }

    private _hasChanges(): boolean {
        return this._profile !== null && this._editName.trim() !== this._profile.name;
    }

    override render(): TemplateResult {
        return html`
            <div class="header">
                <div class="title">Settings</div>
                <div class="subtitle">Manage your account and preferences</div>
            </div>

            ${this._renderContent()}
        `;
    }

    private _renderContent(): TemplateResult {
        if (this._viewState === 'loading') {
            return html`
                <div class="settings-card">
                    <div class="loading-state">
                        <div class="spinner"></div>
                    </div>
                </div>
            `;
        }

        if (this._viewState === 'error' && !this._profile) {
            return html`
                <div class="settings-card">
                    <div class="error-state">
                        <p>${this._error}</p>
                        <button class="btn-secondary" @click=${this._loadProfile.bind(this)}>
                            Retry
                        </button>
                    </div>
                </div>
            `;
        }

        const isSaving = this._viewState === 'saving';

        return html`
            <!-- Profile Card -->
            <div class="settings-card">
                <div class="card-title">Profile</div>

                ${this._success ? html`<div class="success-message">${this._success}</div>` : nothing}
                ${this._error ? html`<div class="error-message">${this._error}</div>` : nothing}

                <div class="form-group">
                    <label for="profile-name">Display Name</label>
                    <input
                        id="profile-name"
                        type="text"
                        .value=${this._editName}
                        ?disabled=${isSaving}
                        @input=${(e: Event) => {
                            this._editName = (e.target as HTMLInputElement).value;
                        }}
                    />
                </div>

                <div class="form-group">
                    <label for="profile-email">Email Address</label>
                    <input
                        id="profile-email"
                        type="email"
                        .value=${this._profile?.email ?? ''}
                        readonly
                    />
                    <div class="readonly-hint">Email cannot be changed</div>
                </div>

                <div class="form-group">
                    <label>Role</label>
                    <div style="padding-top: var(--fb-spacing-xs);">
                        <span class="role-badge">${this._profile?.role ?? ''}</span>
                    </div>
                    <div class="readonly-hint">Contact your administrator to change your role</div>
                </div>

                ${this._hasChanges()
                    ? html`
                          <div class="form-actions">
                              <button
                                  class="btn-secondary"
                                  ?disabled=${isSaving}
                                  @click=${this._handleCancel.bind(this)}
                              >
                                  Cancel
                              </button>
                              <button
                                  class="btn-primary"
                                  ?disabled=${isSaving}
                                  @click=${this._handleSave.bind(this)}
                              >
                                  ${isSaving ? 'Saving...' : 'Save Changes'}
                              </button>
                          </div>
                      `
                    : nothing}
            </div>

            <!-- Construction Physics Card (Step 86) -->
            ${this._renderPhysicsCard()}

            <!-- Account Info Card -->
            <div class="settings-card">
                <div class="card-title">Account Information</div>

                <div class="info-row">
                    <span class="info-label">User ID</span>
                    <span class="info-value">${this._profile?.id ?? ''}</span>
                </div>

                <div class="info-row">
                    <span class="info-label">Organization ID</span>
                    <span class="info-value">${this._profile?.org_id ?? ''}</span>
                </div>

                <div class="info-row">
                    <span class="info-label">Member Since</span>
                    <span class="info-value">${this._formatDate(this._profile?.created_at)}</span>
                </div>
            </div>
        `;
    }

    // ========================================================================
    // Step 86: Construction Physics
    // ========================================================================

    private _handleSpeedChange(e: Event): void {
        // H1: Round to 1dp to prevent floating point artifacts (e.g. 0.7000000000000001)
        this._speedMultiplier = Math.round(parseFloat((e.target as HTMLInputElement).value) * 10) / 10;
        this._physicsDirty = true;
    }

    private _toggleWorkDay(dayIndex: number): void {
        const idx = this._workDays.indexOf(dayIndex);
        if (idx >= 0) {
            // Don't allow deselecting all days
            if (this._workDays.length <= 1) return;
            this._workDays = this._workDays.filter(d => d !== dayIndex);
        } else {
            this._workDays = [...this._workDays, dayIndex].sort((a, b) => a - b);
        }
        this._physicsDirty = true;
    }

    // Step 87: Load physics config from backend
    private async _loadPhysics(): Promise<void> {
        try {
            const config = await api.settings.getPhysics();
            this._speedMultiplier = config.speed_multiplier;
            this._workDays = config.work_days;
            this._physicsDirty = false;
            this._physicsError = '';
            this._physicsUsingDefaults = false;
        } catch (err) {
            // M-5: Non-blocking but differentiated — show error for auth failures
            const msg = err instanceof Error ? err.message : 'Failed to load physics settings';
            if (msg.includes('401') || msg.includes('403') || msg.includes('Unauthorized')) {
                this._physicsError = 'Unable to load physics settings (permission denied)';
            } else {
                // Deep dive fix: show non-intrusive indicator that defaults are being used
                this._physicsUsingDefaults = true;
                console.warn('[FBViewSettings] Physics config load failed, using defaults');
            }
        }
    }

    // Step 87: Save physics config to backend
    private async _handlePhysicsSave(): Promise<void> {
        this._physicsError = '';
        try {
            const config = await api.settings.updatePhysics({
                speed_multiplier: this._speedMultiplier,
                work_days: this._workDays,
            });
            // Sync with server response (in case of server-side rounding)
            this._speedMultiplier = config.speed_multiplier;
            this._workDays = config.work_days;
            this._physicsDirty = false;
            this._physicsUsingDefaults = false;
            this._physicsSuccess = 'Physics settings saved';
            setTimeout(() => { this._physicsSuccess = ''; }, 3000);
        } catch (err) {
            // M-3: Use dedicated physics error state, not shared _error
            this._physicsError = err instanceof Error ? err.message : 'Failed to save physics settings';
        }
    }

    private _handlePhysicsReset(): void {
        this._speedMultiplier = 1.0;
        this._workDays = [1, 2, 3, 4, 5];
        this._physicsDirty = true; // Mark dirty so user can save the reset
    }

    private _renderPhysicsCard(): TemplateResult {
        const paceLabel = getPaceLabel(this._speedMultiplier);
        const paceClass = getPaceClass(this._speedMultiplier);

        return html`
            <div class="settings-card">
                <div class="card-title">Construction Physics</div>

                <!-- M1/M3: Physics-scoped success and error messages -->
                ${this._physicsSuccess ? html`<div class="success-message">${this._physicsSuccess}</div>` : nothing}
                ${this._physicsError ? html`<div class="error-message">${this._physicsError}</div>` : nothing}
                ${this._physicsUsingDefaults ? html`<div class="info-message">Using default settings. Save to persist your configuration.</div>` : nothing}

                <div class="form-group">
                    <span class="physics-section-label">My Pace (Schedule Padding)</span>
                    <div class="slider-container">
                        <input
                            type="range"
                            class="speed-slider"
                            min="0.5"
                            max="1.5"
                            step="0.1"
                            .value=${String(this._speedMultiplier)}
                            @input=${this._handleSpeedChange.bind(this)}
                            aria-label="Schedule speed multiplier"
                        />
                        <span class="pace-badge ${paceClass}">${paceLabel}</span>
                    </div>
                    <div class="speed-value">${this._speedMultiplier.toFixed(1)}x multiplier</div>
                </div>

                <div class="form-group">
                    <span class="physics-section-label">Standard Work Week</span>
                    <div class="workdays-container" role="group" aria-label="Work days selection">
                        ${DAY_LABELS.map((label, i) => {
                            const dayNum = i + 1;
                            const isActive = this._workDays.includes(dayNum);
                            const fullName = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'][i];
                            return html`
                                <button
                                    class="day-toggle ${isActive ? 'active' : ''}"
                                    @click=${() => this._toggleWorkDay(dayNum)}
                                    title="${fullName}"
                                    aria-label="${fullName}"
                                    aria-pressed=${isActive ? 'true' : 'false'}
                                >
                                    ${label}
                                </button>
                            `;
                        })}
                    </div>
                </div>

                ${this._physicsDirty
                    ? html`
                          <div class="physics-actions">
                              <button
                                  class="btn-secondary"
                                  @click=${this._handlePhysicsReset.bind(this)}
                              >
                                  Reset
                              </button>
                              <button
                                  class="btn-primary"
                                  @click=${this._handlePhysicsSave.bind(this)}
                              >
                                  Save Changes
                              </button>
                          </div>
                      `
                    : nothing}
            </div>
        `;
    }

    private _formatDate(dateStr?: string): string {
        if (!dateStr) return '';
        const date = new Date(dateStr);
        return date.toLocaleDateString('en-US', {
            month: 'long',
            day: 'numeric',
            year: 'numeric',
        });
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-settings': FBViewSettings;
    }
}
