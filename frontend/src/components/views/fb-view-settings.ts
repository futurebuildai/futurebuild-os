/**
 * FBViewSettings - User Settings View
 * See LAUNCH_PLAN.md Section: User Settings View (P1)
 *
 * Allows users to:
 * - View their profile information
 * - Update their name
 * - View organization info (readonly)
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { api, UserProfile } from '../../services/api';

type ViewState = 'loading' | 'ready' | 'saving' | 'error';

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

            input {
                width: 100%;
                padding: var(--fb-spacing-md);
                background: var(--fb-bg-tertiary);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-md);
                color: var(--fb-text-primary);
                font-size: var(--fb-text-base);
                box-sizing: border-box;
            }

            input:focus {
                outline: none;
                border-color: var(--fb-primary);
            }

            input:disabled {
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
        `,
    ];

    @state() private _viewState: ViewState = 'loading';
    @state() private _profile: UserProfile | null = null;
    @state() private _editName = '';
    @state() private _error = '';
    @state() private _success = '';

    override connectedCallback(): void {
        super.connectedCallback();
        void this._loadProfile();
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
