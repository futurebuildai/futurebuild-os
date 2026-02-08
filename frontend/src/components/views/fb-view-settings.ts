/**
 * FBViewSettings - User Settings View
 * See LAUNCH_PLAN.md Section: User Settings View (P1)
 *
 * Allows users to:
 * - View their profile information (pre-filled from store, enriched via API)
 * - Update their name
 * - View organization info (readonly)
 * - Configure Construction Physics (Step 86) — admin/builder only
 *
 * Graceful degradation: renders immediately from store data even if
 * the Go backend is down. API data enriches (created_at, etc.) when available.
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBViewElement } from '../base/FBViewElement';
import { api, UserProfile } from '../../services/api';
import { store } from '../../store/store';
import type { User } from '../../store/types';
import { UserRole } from '../../types/enums';

type ProfileLoadState = 'idle' | 'loading' | 'ready' | 'saving' | 'failed';

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

            /* Tab navigation */
            .tab-bar {
                display: flex;
                gap: var(--fb-spacing-xs);
                margin-bottom: var(--fb-spacing-xl);
                border-bottom: 1px solid var(--fb-border);
                padding-bottom: 0;
            }

            .tab-btn {
                padding: var(--fb-spacing-sm) var(--fb-spacing-lg);
                background: transparent;
                color: var(--fb-text-secondary);
                border: none;
                border-bottom: 2px solid transparent;
                font-size: var(--fb-text-sm);
                font-weight: 500;
                cursor: pointer;
                transition: all 0.15s ease;
                margin-bottom: -1px;
            }

            .tab-btn:hover {
                color: var(--fb-text-primary);
            }

            .tab-btn.active {
                color: var(--fb-primary);
                border-bottom-color: var(--fb-primary);
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

            .warning-message {
                color: var(--fb-warning, #f59e0b);
                font-size: var(--fb-text-sm);
                margin-bottom: var(--fb-spacing-md);
                padding: var(--fb-spacing-sm) var(--fb-spacing-md);
                background: rgba(245, 158, 11, 0.1);
                border: 1px solid rgba(245, 158, 11, 0.2);
                border-radius: var(--fb-radius-sm);
                display: flex;
                align-items: center;
                justify-content: space-between;
                gap: var(--fb-spacing-sm);
            }

            .warning-message .btn-secondary {
                padding: var(--fb-spacing-xs) var(--fb-spacing-sm);
                font-size: var(--fb-text-xs);
                flex-shrink: 0;
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

            .speed-slider:disabled {
                opacity: 0.5;
                cursor: not-allowed;
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

            .speed-slider:disabled::-webkit-slider-thumb {
                cursor: not-allowed;
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

            .day-toggle:hover:not(:disabled) {
                border-color: var(--fb-primary);
            }

            .day-toggle:disabled {
                opacity: 0.5;
                cursor: not-allowed;
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

            .readonly-banner {
                font-size: var(--fb-text-xs);
                color: var(--fb-text-muted);
                padding: var(--fb-spacing-sm) var(--fb-spacing-md);
                background: var(--fb-bg-tertiary);
                border-radius: var(--fb-radius-sm);
                margin-bottom: var(--fb-spacing-md);
            }

            /* Slider wrapper with baseline markers */
            .slider-wrapper {
                position: relative;
                flex: 1;
                padding-top: 22px;
            }

            .slider-wrapper .speed-slider {
                width: 100%;
            }

            .baseline-marker {
                position: absolute;
                top: 18px;
                width: 2px;
                height: 14px;
                transform: translateX(-50%);
                pointer-events: none;
                z-index: 1;
            }

            .baseline-marker.industry {
                background: var(--fb-text-muted, #666);
            }

            .baseline-marker.org {
                background: var(--fb-primary);
            }

            .baseline-label {
                position: absolute;
                top: 0;
                transform: translateX(-50%);
                font-size: 10px;
                font-weight: 600;
                white-space: nowrap;
                pointer-events: none;
            }

            .baseline-label.industry {
                color: var(--fb-text-muted, #666);
            }

            .baseline-label.org {
                color: var(--fb-primary);
            }

            .gap-indicator {
                position: absolute;
                top: 22px;
                height: 6px;
                background: var(--fb-primary);
                opacity: 0.2;
                border-radius: 3px;
                pointer-events: none;
                z-index: 0;
            }

            /* Confirmation overlay */
            .confirm-overlay {
                position: fixed;
                inset: 0;
                background: rgba(0, 0, 0, 0.6);
                display: flex;
                align-items: center;
                justify-content: center;
                z-index: 1000;
            }

            .confirm-modal {
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-xl);
                max-width: 440px;
                width: 90%;
                color: var(--fb-text-primary);
            }

            .confirm-modal h3 {
                margin: 0 0 var(--fb-spacing-md) 0;
                font-size: var(--fb-text-lg);
                font-weight: 600;
            }

            .confirm-modal p {
                margin: 0 0 var(--fb-spacing-md) 0;
                font-size: var(--fb-text-sm);
                color: var(--fb-text-secondary);
                line-height: 1.5;
            }

            .confirm-actions {
                display: flex;
                gap: var(--fb-spacing-sm);
                justify-content: flex-end;
                margin-top: var(--fb-spacing-lg);
            }

            .scope-option {
                display: flex;
                align-items: flex-start;
                gap: var(--fb-spacing-sm);
                padding: var(--fb-spacing-md);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-md);
                cursor: pointer;
                transition: border-color 0.15s ease;
                margin-bottom: var(--fb-spacing-sm);
            }

            .scope-option:hover {
                border-color: var(--fb-primary);
            }

            .scope-option.selected {
                border-color: var(--fb-primary);
                background: rgba(59, 130, 246, 0.08);
            }

            .scope-option input[type="radio"] {
                margin-top: 2px;
                accent-color: var(--fb-primary);
            }

            .scope-option-text {
                display: flex;
                flex-direction: column;
                gap: 2px;
            }

            .scope-option-label {
                font-size: var(--fb-text-sm);
                font-weight: 500;
                color: var(--fb-text-primary);
            }

            .scope-option-desc {
                font-size: var(--fb-text-xs);
                color: var(--fb-text-muted);
            }
        `,
    ];

    // Tab state
    @state() private _activeTab: 'account' | 'organization' = 'account';

    // Profile state — no longer gates rendering
    @state() private _profileLoadState: ProfileLoadState = 'idle';
    @state() private _profile: UserProfile | null = null;
    @state() private _storeUser: User | null = null;
    @state() private _editName = '';
    @state() private _error = '';
    @state() private _success = '';
    @state() private _profileWarning = '';

    // Step 86: Construction Physics state
    @state() private _speedMultiplier = 1.0;
    @state() private _workDays: number[] = [1, 2, 3, 4, 5];
    @state() private _physicsDirty = false;
    @state() private _physicsSuccess = '';
    @state() private _physicsError = '';
    @state() private _physicsUsingDefaults = false;

    // Baseline & confirmation state
    @state() private _savedOrgBaseline = 1.0;
    @state() private _confirmStep: 0 | 1 | 2 = 0;
    @state() private _applyToExisting = false;
    @state() private _confirmProcessing = false;

    private _disposeEffect: (() => void) | null = null;

    override connectedCallback(): void {
        super.connectedCallback();

        // Pre-fill from store so the page renders immediately
        this._disposeEffect = effect(() => {
            const user = store.user$.value;
            if (user) {
                this._storeUser = user;
                // Only pre-fill name if we haven't loaded from API yet and user hasn't started editing
                if (!this._profile && !this._editName) {
                    this._editName = user.name;
                }
            }
        });

        void this._loadProfile();
        void this._loadPhysics();
    }

    override disconnectedCallback(): void {
        super.disconnectedCallback();
        this._disposeEffect?.();
        this._disposeEffect = null;
    }

    private async _loadProfile(): Promise<void> {
        this._profileLoadState = 'loading';
        this._profileWarning = '';

        try {
            this._profile = await api.users.getMe();
            this._editName = this._profile.name;
            this._profileLoadState = 'ready';
        } catch (err) {
            this._profileLoadState = 'failed';
            this._profileWarning = 'Could not load full profile from server';
            console.warn('[FBViewSettings] Profile load failed:', err instanceof Error ? err.message : err);
        }
    }

    private async _handleSave(): Promise<void> {
        if (!this._editName.trim()) {
            this._error = 'Name is required';
            return;
        }

        this._profileLoadState = 'saving';
        this._error = '';
        this._success = '';

        try {
            this._profile = await api.users.updateMe({ name: this._editName.trim() });
            this._profileLoadState = 'ready';
            this._profileWarning = '';
            this._success = 'Profile updated successfully';

            setTimeout(() => {
                this._success = '';
            }, 3000);
        } catch (err) {
            this._profileLoadState = this._profile ? 'ready' : 'failed';
            this._error = err instanceof Error ? err.message : 'Failed to update profile';
        }
    }

    private _handleCancel(): void {
        // Reset to API data if available, otherwise store data
        this._editName = this._profile?.name ?? this._storeUser?.name ?? '';
        this._error = '';
        this._success = '';
    }

    private _hasChanges(): boolean {
        const currentName = this._profile?.name ?? this._storeUser?.name ?? '';
        return this._editName.trim() !== '' && this._editName.trim() !== currentName;
    }

    private _canEditOrgSettings(): boolean {
        const role = this._storeUser?.role;
        return role === UserRole.Admin || role === UserRole.Builder;
    }

    // ========================================================================
    // Render
    // ========================================================================

    override render(): TemplateResult {
        return html`
            <div class="header">
                <div class="title">Settings</div>
                <div class="subtitle">Manage your account and preferences</div>
            </div>

            <div class="tab-bar">
                <button
                    class="tab-btn ${this._activeTab === 'account' ? 'active' : ''}"
                    @click=${() => { this._activeTab = 'account'; }}
                >
                    Account
                </button>
                <button
                    class="tab-btn ${this._activeTab === 'organization' ? 'active' : ''}"
                    @click=${() => { this._activeTab = 'organization'; }}
                >
                    Organization
                </button>
            </div>

            ${this._activeTab === 'account'
                ? this._renderAccountTab()
                : this._renderOrganizationTab()}
        `;
    }

    // ========================================================================
    // Account Tab
    // ========================================================================

    private _renderAccountTab(): TemplateResult {
        const isSaving = this._profileLoadState === 'saving';
        const email = this._profile?.email ?? this._storeUser?.email ?? '';
        const role = this._profile?.role ?? this._storeUser?.role ?? '';
        const userId = this._profile?.id ?? this._storeUser?.id ?? '';
        const orgId = this._profile?.org_id ?? this._storeUser?.orgId ?? '';
        const memberSince = this._formatDate(this._profile?.created_at);

        return html`
            <!-- Profile Card -->
            <div class="settings-card">
                <div class="card-title">Profile</div>

                ${this._profileWarning
                    ? html`
                        <div class="warning-message">
                            <span>${this._profileWarning}</span>
                            <button class="btn-secondary" @click=${this._loadProfile.bind(this)}>Retry</button>
                        </div>`
                    : nothing}
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
                        .value=${email}
                        readonly
                    />
                    <div class="readonly-hint">Email cannot be changed</div>
                </div>

                <div class="form-group">
                    <label>Role</label>
                    <div style="padding-top: var(--fb-spacing-xs);">
                        ${role
                            ? html`<span class="role-badge">${role}</span>`
                            : html`<span class="info-value">--</span>`}
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

            <!-- Account Info Card -->
            <div class="settings-card">
                <div class="card-title">Account Information</div>

                <div class="info-row">
                    <span class="info-label">User ID</span>
                    <span class="info-value">${userId || '--'}</span>
                </div>

                <div class="info-row">
                    <span class="info-label">Organization ID</span>
                    <span class="info-value">${orgId || '--'}</span>
                </div>

                <div class="info-row">
                    <span class="info-label">Member Since</span>
                    <span class="info-value">${memberSince || '--'}</span>
                </div>
            </div>
        `;
    }

    // ========================================================================
    // Organization Tab
    // ========================================================================

    private _renderOrganizationTab(): TemplateResult {
        return this._renderPhysicsCard();
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
            this._savedOrgBaseline = config.speed_multiplier;
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

    // Step 87: Save physics config — gates through confirmation when baseline changes
    private _handlePhysicsSave(): void {
        this._physicsError = '';
        // If speed multiplier changed from saved org baseline, require confirmation
        if (this._speedMultiplier !== this._savedOrgBaseline) {
            this._confirmStep = 1;
            this._applyToExisting = false;
            return;
        }
        // Only work_days changed — save directly
        void this._executePhysicsSave();
    }

    private async _executePhysicsSave(): Promise<void> {
        this._confirmProcessing = true;
        this._physicsError = '';
        try {
            const baselineChanged = this._speedMultiplier !== this._savedOrgBaseline;
            const config = await api.settings.updatePhysics({
                speed_multiplier: this._speedMultiplier,
                work_days: this._workDays,
                ...(baselineChanged ? { apply_to_existing: this._applyToExisting } : {}),
            });
            // Sync with server response (in case of server-side rounding)
            this._speedMultiplier = config.speed_multiplier;
            this._workDays = config.work_days;
            this._savedOrgBaseline = config.speed_multiplier;
            this._physicsDirty = false;
            this._physicsUsingDefaults = false;
            this._confirmStep = 0;
            this._physicsSuccess = 'Physics settings saved';
            setTimeout(() => { this._physicsSuccess = ''; }, 3000);
        } catch (err) {
            // M-3: Use dedicated physics error state, not shared _error
            this._physicsError = err instanceof Error ? err.message : 'Failed to save physics settings';
            // If in modal, keep it open so user sees the error
            if (this._confirmStep === 0) {
                // Direct save path — error already shown in card
            }
        } finally {
            this._confirmProcessing = false;
        }
    }

    private _handlePhysicsReset(): void {
        this._speedMultiplier = 1.0;
        this._workDays = [1, 2, 3, 4, 5];
        this._physicsDirty = true; // Mark dirty so user can save the reset
    }

    /** Convert a value in the 0.5–1.5 range to a percentage position. */
    private _toSliderPercent(value: number): number {
        return ((value - 0.5) / 1.0) * 100;
    }

    private _renderPhysicsCard(): TemplateResult {
        const paceLabel = getPaceLabel(this._speedMultiplier);
        const paceClass = getPaceClass(this._speedMultiplier);
        const canEdit = this._canEditOrgSettings();

        const industryPct = this._toSliderPercent(1.0);
        const orgPct = this._toSliderPercent(this._savedOrgBaseline);
        const currentPct = this._toSliderPercent(this._speedMultiplier);
        const gapLeft = Math.min(orgPct, currentPct);
        const gapWidth = Math.abs(currentPct - orgPct);
        const showGap = this._speedMultiplier !== this._savedOrgBaseline;

        return html`
            <div class="settings-card">
                <div class="card-title">Construction Physics</div>

                ${!canEdit
                    ? html`<div class="readonly-banner">These settings are managed by your organization's Admin or Builder. Contact your administrator to request changes.</div>`
                    : nothing}

                <!-- M1/M3: Physics-scoped success and error messages -->
                ${this._physicsSuccess ? html`<div class="success-message">${this._physicsSuccess}</div>` : nothing}
                ${this._physicsError && this._confirmStep === 0 ? html`<div class="error-message">${this._physicsError}</div>` : nothing}
                ${this._physicsUsingDefaults ? html`<div class="info-message">Using default settings. Save to persist your configuration.</div>` : nothing}

                <div class="form-group">
                    <span class="physics-section-label">My Pace (Schedule Padding)</span>
                    <div class="slider-container">
                        <div class="slider-wrapper">
                            <!-- Industry baseline marker (always at 1.0) -->
                            <span class="baseline-label industry" style="left: ${industryPct}%">Industry</span>
                            <div class="baseline-marker industry" style="left: ${industryPct}%"></div>

                            <!-- Org baseline marker (last-saved value) -->
                            <span class="baseline-label org" style="left: ${orgPct}%">Org</span>
                            <div class="baseline-marker org" style="left: ${orgPct}%"></div>

                            <!-- Gap indicator between org baseline and current thumb -->
                            ${showGap
                                ? html`<div class="gap-indicator" style="left: ${gapLeft}%; width: ${gapWidth}%"></div>`
                                : nothing}

                            <input
                                type="range"
                                class="speed-slider"
                                min="0.5"
                                max="1.5"
                                step="0.1"
                                .value=${String(this._speedMultiplier)}
                                ?disabled=${!canEdit}
                                @input=${this._handleSpeedChange.bind(this)}
                                aria-label="Schedule speed multiplier"
                            />
                        </div>
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
                                    ?disabled=${!canEdit}
                                    @click=${() => { this._toggleWorkDay(dayNum); }}
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

                ${canEdit && this._physicsDirty
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

            <!-- Confirmation modal -->
            ${this._confirmStep > 0 ? this._renderConfirmModal() : nothing}
        `;
    }

    private _renderConfirmModal(): TemplateResult {
        return html`
            <div class="confirm-overlay" @click=${(e: Event) => {
                if (e.target === e.currentTarget) this._confirmStep = 0;
            }}>
                <div class="confirm-modal">
                    ${this._confirmStep === 1
                        ? this._renderConfirmStep1()
                        : this._renderConfirmStep2()}
                </div>
            </div>
        `;
    }

    private _renderConfirmStep1(): TemplateResult {
        const oldVal = this._savedOrgBaseline.toFixed(1);
        const newVal = this._speedMultiplier.toFixed(1);
        const newPace = getPaceLabel(this._speedMultiplier);

        return html`
            <h3>Change Organization Baseline?</h3>
            <p>
                You're changing the org pace from ${oldVal}x to ${newVal}x (${newPace}).
            </p>
            <p>This affects how all project schedules are calculated.</p>
            <div class="confirm-actions">
                <button class="btn-secondary" @click=${() => { this._confirmStep = 0; }}>Cancel</button>
                <button class="btn-primary" @click=${() => { this._confirmStep = 2; }}>Continue</button>
            </div>
        `;
    }

    private _renderConfirmStep2(): TemplateResult {
        return html`
            <h3>Apply to which schedules?</h3>
            ${this._physicsError ? html`<div class="error-message">${this._physicsError}</div>` : nothing}

            <label
                class="scope-option ${this._applyToExisting ? 'selected' : ''}"
                @click=${() => { this._applyToExisting = true; }}
            >
                <input
                    type="radio"
                    name="scope"
                    .checked=${this._applyToExisting}
                />
                <div class="scope-option-text">
                    <span class="scope-option-label">In-progress schedules</span>
                    <span class="scope-option-desc">Recalculate active project timelines</span>
                </div>
            </label>

            <label
                class="scope-option ${!this._applyToExisting ? 'selected' : ''}"
                @click=${() => { this._applyToExisting = false; }}
            >
                <input
                    type="radio"
                    name="scope"
                    .checked=${!this._applyToExisting}
                />
                <div class="scope-option-text">
                    <span class="scope-option-label">Future projects only</span>
                    <span class="scope-option-desc">Existing schedules remain unchanged</span>
                </div>
            </label>

            <div class="confirm-actions">
                <button
                    class="btn-secondary"
                    ?disabled=${this._confirmProcessing}
                    @click=${() => { this._confirmStep = 1; }}
                >Back</button>
                <button
                    class="btn-primary"
                    ?disabled=${this._confirmProcessing}
                    @click=${() => { void this._executePhysicsSave(); }}
                >${this._confirmProcessing ? 'Saving...' : 'Confirm & Save'}</button>
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
