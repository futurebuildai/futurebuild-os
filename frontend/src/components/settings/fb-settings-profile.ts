/**
 * fb-settings-profile — User profile settings page.
 * See FRONTEND_V2_SPEC.md §10.2.A
 *
 * Route: /settings/profile
 * Shows: name (editable), email (read-only), role (read-only), member since
 */
import { html, css, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { api, UserProfile } from '../../services/api';
import { store } from '../../store/store';
import type { User } from '../../store/types';

@customElement('fb-settings-profile')
export class FBSettingsProfile extends FBElement {
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
                color: var(--fb-text-primary, #F0F0F5);
            }

            .subtitle {
                font-size: 14px;
                color: var(--fb-text-secondary, #8B8D98);
                margin-top: 4px;
            }

            .card {
                background: var(--fb-surface-1, #161821);
                border: 1px solid var(--fb-border, rgba(255,255,255,0.05));
                border-radius: 12px;
                padding: 24px;
                margin-bottom: 20px;
            }

            .card-title {
                font-size: 16px;
                font-weight: 600;
                color: var(--fb-text-primary, #F0F0F5);
                margin-bottom: 20px;
                padding-bottom: 12px;
                border-bottom: 1px solid var(--fb-border, rgba(255,255,255,0.05));
            }

            .form-group {
                margin-bottom: 20px;
            }

            .form-group:last-child {
                margin-bottom: 0;
            }

            label {
                display: block;
                font-size: 13px;
                font-weight: 500;
                color: var(--fb-text-secondary, #8B8D98);
                margin-bottom: 6px;
            }

            input {
                width: 100%;
                padding: 10px 12px;
                background: var(--fb-surface-2, #1E2029);
                border: 1px solid var(--fb-border, rgba(255,255,255,0.05));
                border-radius: 8px;
                color: var(--fb-text-primary, #F0F0F5);
                font-size: 14px;
                box-sizing: border-box;
            }

            input:focus {
                outline: none;
                border-color: var(--fb-accent, #00FFA3);
            }

            input:disabled {
                opacity: 0.6;
                cursor: not-allowed;
            }

            input[readonly] {
                background: var(--fb-surface-0, #12121e);
                cursor: default;
            }

            .hint {
                font-size: 12px;
                color: var(--fb-text-tertiary, #5A5B66);
                margin-top: 4px;
            }

            .role-badge {
                display: inline-block;
                padding: 4px 10px;
                background: var(--fb-accent, #00FFA3);
                color: #fff;
                border-radius: 4px;
                font-size: 12px;
                font-weight: 600;
                text-transform: uppercase;
            }

            .info-row {
                display: flex;
                justify-content: space-between;
                align-items: center;
                padding: 12px 0;
                border-bottom: 1px solid var(--fb-border, rgba(255,255,255,0.05));
            }

            .info-row:last-child {
                border-bottom: none;
            }

            .info-label {
                font-size: 13px;
                color: var(--fb-text-secondary, #8B8D98);
            }

            .info-value {
                font-size: 13px;
                color: var(--fb-text-primary, #F0F0F5);
                font-weight: 500;
            }

            .actions {
                display: flex;
                gap: 12px;
                margin-top: 20px;
                padding-top: 20px;
                border-top: 1px solid var(--fb-border, rgba(255,255,255,0.05));
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
                background: var(--fb-accent, #00FFA3);
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
                border: 1px solid var(--fb-border, rgba(255,255,255,0.05));
                color: var(--fb-text-secondary, #8B8D98);
            }

            .btn-secondary:hover {
                background: var(--fb-surface-2, #1E2029);
                color: var(--fb-text-primary, #F0F0F5);
            }

            .message {
                padding: 10px 14px;
                border-radius: 8px;
                font-size: 13px;
                margin-bottom: 16px;
            }

            .message-success {
                background: rgba(34, 197, 94, 0.1);
                color: #00FFA3;
            }

            .message-error {
                background: rgba(239, 68, 68, 0.1);
                color: #F43F5E;
            }

            .back-link {
                display: inline-flex;
                align-items: center;
                gap: 6px;
                font-size: 13px;
                color: var(--fb-text-secondary, #8B8D98);
                cursor: pointer;
                margin-bottom: 16px;
            }

            .back-link:hover {
                color: var(--fb-text-primary, #F0F0F5);
            }
        `,
    ];

    @state() private _loading = true;
    @state() private _saving = false;
    @state() private _profile: UserProfile | null = null;
    @state() private _storeUser: User | null = null;
    @state() private _editName = '';
    @state() private _success = '';
    @state() private _error = '';

    private _disposeEffect: (() => void) | null = null;

    override connectedCallback() {
        super.connectedCallback();
        this._disposeEffect = effect(() => {
            const user = store.user$.value;
            if (user) {
                this._storeUser = user;
                if (!this._profile && !this._editName) {
                    this._editName = user.name;
                }
            }
        });
        this._loadProfile();
    }

    override disconnectedCallback() {
        super.disconnectedCallback();
        this._disposeEffect?.();
    }

    private async _loadProfile() {
        this._loading = true;
        try {
            this._profile = await api.users.getMe();
            this._editName = this._profile.name;
        } catch (err) {
            console.warn('[FBSettingsProfile] Failed to load profile:', err);
        } finally {
            this._loading = false;
        }
    }

    private async _handleSave() {
        if (!this._editName.trim()) {
            this._error = 'Name is required';
            return;
        }

        this._saving = true;
        this._error = '';
        this._success = '';

        try {
            this._profile = await api.users.updateMe({ name: this._editName.trim() });
            this._success = 'Profile updated successfully';
            setTimeout(() => { this._success = ''; }, 3000);
        } catch (err) {
            this._error = err instanceof Error ? err.message : 'Failed to update profile';
        } finally {
            this._saving = false;
        }
    }

    private _handleCancel() {
        this._editName = this._profile?.name ?? this._storeUser?.name ?? '';
        this._error = '';
    }

    private _hasChanges(): boolean {
        const currentName = this._profile?.name ?? this._storeUser?.name ?? '';
        return this._editName.trim() !== currentName;
    }

    private _handleBack() {
        this.emit('fb-navigate', { view: 'home' });
    }

    private _formatDate(dateStr?: string): string {
        if (!dateStr) return '--';
        return new Date(dateStr).toLocaleDateString('en-US', {
            month: 'long',
            day: 'numeric',
            year: 'numeric',
        });
    }

    override render() {
        // Use loading state to show skeleton while data loads
        void this._loading; // Suppress unused warning - used for future loading skeleton
        const email = this._profile?.email ?? this._storeUser?.email ?? '';
        const role = this._profile?.role ?? this._storeUser?.role ?? '';
        const memberSince = this._formatDate(this._profile?.created_at);

        return html`
            <div class="back-link" @click=${this._handleBack}>
                ← Back to Feed
            </div>

            <div class="header">
                <div class="title">My Profile</div>
                <div class="subtitle">Manage your account information</div>
            </div>

            ${this._success ? html`<div class="message message-success">${this._success}</div>` : nothing}
            ${this._error ? html`<div class="message message-error">${this._error}</div>` : nothing}

            <div class="card">
                <div class="card-title">Profile Information</div>

                <div class="form-group">
                    <label for="name">Display Name</label>
                    <input
                        id="name"
                        type="text"
                        .value=${this._editName}
                        ?disabled=${this._saving}
                        @input=${(e: Event) => { this._editName = (e.target as HTMLInputElement).value; }}
                    />
                </div>

                <div class="form-group">
                    <label for="email">Email Address</label>
                    <input
                        id="email"
                        type="email"
                        .value=${email}
                        readonly
                    />
                    <div class="hint">Managed by Clerk</div>
                </div>

                <div class="form-group">
                    <label>Role</label>
                    <div style="padding-top: 4px;">
                        ${role ? html`<span class="role-badge">${role}</span>` : html`<span class="info-value">--</span>`}
                    </div>
                    <div class="hint">Contact your administrator to change your role</div>
                </div>

                ${this._hasChanges() ? html`
                    <div class="actions">
                        <button class="btn btn-secondary" ?disabled=${this._saving} @click=${this._handleCancel}>
                            Cancel
                        </button>
                        <button class="btn btn-primary" ?disabled=${this._saving} @click=${this._handleSave}>
                            ${this._saving ? 'Saving...' : 'Save Changes'}
                        </button>
                    </div>
                ` : nothing}
            </div>

            <div class="card">
                <div class="card-title">Account Details</div>

                <div class="info-row">
                    <span class="info-label">Member Since</span>
                    <span class="info-value">${memberSince}</span>
                </div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-settings-profile': FBSettingsProfile;
    }
}
