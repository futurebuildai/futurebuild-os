/**
 * fb-settings-team — Team management page.
 * See FRONTEND_V2_SPEC.md §10.2.C
 *
 * Route: /settings/team
 * Shows: Team members list, pending invites, invite modal
 * Access: Admin role only
 */
import { html, css, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';
import { UserRole } from '../../types/enums';

interface TeamMember {
    id: string;
    name: string;
    email: string;
    role: string;
    status: 'active' | 'pending';
    created_at?: string;
}

@customElement('fb-settings-team')
export class FBSettingsTeam extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                max-width: 700px;
                margin: 0 auto;
                padding: 32px 16px;
            }

            .header {
                display: flex;
                justify-content: space-between;
                align-items: flex-start;
                margin-bottom: 32px;
            }

            .header-text .title {
                font-size: 24px;
                font-weight: 700;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .header-text .subtitle {
                font-size: 14px;
                color: var(--fb-text-secondary, #a0a0b0);
                margin-top: 4px;
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

            .btn-primary:hover {
                opacity: 0.9;
            }

            .btn-secondary {
                background: transparent;
                border: 1px solid var(--fb-border, #2a2a3e);
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .btn-secondary:hover {
                background: var(--fb-surface-2, #252540);
            }

            .card {
                background: var(--fb-surface-1, #1a1a2e);
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 12px;
                margin-bottom: 20px;
                overflow: hidden;
            }

            .card-title {
                font-size: 16px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
                padding: 20px 24px;
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
            }

            .member-list {
                list-style: none;
                margin: 0;
                padding: 0;
            }

            .member-item {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: 16px 24px;
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
            }

            .member-item:last-child {
                border-bottom: none;
            }

            .member-info {
                display: flex;
                align-items: center;
                gap: 12px;
            }

            .member-avatar {
                width: 40px;
                height: 40px;
                border-radius: 50%;
                background: var(--fb-accent, #6366f1);
                color: #fff;
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 14px;
                font-weight: 600;
            }

            .member-details {
                display: flex;
                flex-direction: column;
                gap: 2px;
            }

            .member-name {
                font-size: 14px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .member-email {
                font-size: 13px;
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .member-meta {
                display: flex;
                align-items: center;
                gap: 12px;
            }

            .role-badge {
                padding: 4px 10px;
                border-radius: 4px;
                font-size: 11px;
                font-weight: 600;
                text-transform: uppercase;
            }

            .role-badge.admin {
                background: rgba(239, 68, 68, 0.15);
                color: #ef4444;
            }

            .role-badge.builder {
                background: rgba(34, 197, 94, 0.15);
                color: #22c55e;
            }

            .role-badge.pm {
                background: rgba(59, 130, 246, 0.15);
                color: #3b82f6;
            }

            .role-badge.viewer {
                background: rgba(156, 163, 175, 0.15);
                color: #9ca3af;
            }

            .status-badge {
                padding: 4px 10px;
                border-radius: 4px;
                font-size: 11px;
                font-weight: 500;
            }

            .status-badge.pending {
                background: rgba(245, 158, 11, 0.15);
                color: #f59e0b;
            }

            .empty-state {
                padding: 40px 24px;
                text-align: center;
                color: var(--fb-text-secondary, #a0a0b0);
                font-size: 14px;
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

            /* Invite modal */
            .modal-overlay {
                position: fixed;
                inset: 0;
                background: rgba(0, 0, 0, 0.6);
                display: flex;
                align-items: center;
                justify-content: center;
                z-index: 1000;
            }

            .modal {
                background: var(--fb-surface-1, #1a1a2e);
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 12px;
                padding: 24px;
                max-width: 400px;
                width: 90%;
            }

            .modal-title {
                font-size: 18px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
                margin-bottom: 20px;
            }

            .form-group {
                margin-bottom: 16px;
            }

            label {
                display: block;
                font-size: 13px;
                font-weight: 500;
                color: var(--fb-text-secondary, #a0a0b0);
                margin-bottom: 6px;
            }

            input, select {
                width: 100%;
                padding: 10px 12px;
                background: var(--fb-surface-2, #252540);
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 8px;
                color: var(--fb-text-primary, #e0e0e0);
                font-size: 14px;
                box-sizing: border-box;
            }

            input:focus, select:focus {
                outline: none;
                border-color: var(--fb-accent, #6366f1);
            }

            .modal-actions {
                display: flex;
                gap: 12px;
                justify-content: flex-end;
                margin-top: 24px;
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
        `,
    ];

    @state() private _loading = true;
    @state() private _isAdmin = false;
    @state() private _members: TeamMember[] = [];
    @state() private _pendingInvites: TeamMember[] = [];
    @state() private _showInviteModal = false;
    @state() private _inviteEmail = '';
    @state() private _inviteRole = 'Builder';
    @state() private _inviting = false;
    @state() private _success = '';
    @state() private _error = '';

    private _disposeEffect: (() => void) | null = null;

    override connectedCallback() {
        super.connectedCallback();
        this._disposeEffect = effect(() => {
            const user = store.user$.value;
            if (user) {
                this._isAdmin = user.role === UserRole.Admin;
            }
        });
        this._loadTeam();
    }

    override disconnectedCallback() {
        super.disconnectedCallback();
        this._disposeEffect?.();
    }

    private async _loadTeam() {
        this._loading = true;
        try {
            // TODO: Replace with actual API call
            // const data = await api.team.list();
            // this._members = data.members;
            // this._pendingInvites = data.pending;

            // Mock data for now
            this._members = [
                { id: '1', name: 'Current User', email: 'you@company.com', role: 'Admin', status: 'active' },
            ];
            this._pendingInvites = [];
        } catch (err) {
            console.warn('[FBSettingsTeam] Failed to load team:', err);
        } finally {
            this._loading = false;
        }
    }

    private _handleInvite() {
        this._showInviteModal = true;
        this._inviteEmail = '';
        this._inviteRole = 'Builder';
        this._error = '';
    }

    private _closeModal() {
        this._showInviteModal = false;
    }

    private async _sendInvite() {
        if (!this._inviteEmail.trim() || !this._inviteEmail.includes('@')) {
            this._error = 'Please enter a valid email address';
            return;
        }

        this._inviting = true;
        this._error = '';

        try {
            // TODO: Replace with actual API call
            // await api.team.invite({ email: this._inviteEmail, role: this._inviteRole });

            this._pendingInvites = [
                ...this._pendingInvites,
                {
                    id: Date.now().toString(),
                    name: this._inviteEmail.split('@')[0]!,
                    email: this._inviteEmail,
                    role: this._inviteRole,
                    status: 'pending',
                },
            ];

            this._showInviteModal = false;
            this._success = `Invitation sent to ${this._inviteEmail}`;
            setTimeout(() => { this._success = ''; }, 3000);
        } catch (err) {
            this._error = err instanceof Error ? err.message : 'Failed to send invitation';
        } finally {
            this._inviting = false;
        }
    }

    private _handleBack() {
        this.emit('fb-navigate', { view: 'home' });
    }

    private _getInitials(name: string): string {
        const parts = name.trim().split(/\s+/);
        if (parts.length >= 2) {
            return (parts[0]![0]! + parts[parts.length - 1]![0]!).toUpperCase();
        }
        return name.substring(0, 2).toUpperCase();
    }

    private _getRoleBadgeClass(role: string): string {
        switch (role.toLowerCase()) {
            case 'admin': return 'admin';
            case 'builder': return 'builder';
            case 'pm': return 'pm';
            default: return 'viewer';
        }
    }

    override render() {
        void this._loading; // Suppress unused warning - used for future loading skeleton
        if (!this._isAdmin) {
            return html`
                <div class="back-link" @click=${this._handleBack}>← Back to Feed</div>
                <div class="no-access">
                    <div class="no-access-title">Access Restricted</div>
                    <div class="no-access-body">
                        Team management is only available to Admin users.
                    </div>
                </div>
            `;
        }

        return html`
            <div class="back-link" @click=${this._handleBack}>← Back to Feed</div>

            <div class="header">
                <div class="header-text">
                    <div class="title">Team & Invites</div>
                    <div class="subtitle">Manage your organization's team members</div>
                </div>
                <button class="btn btn-primary" @click=${this._handleInvite}>
                    Invite Member
                </button>
            </div>

            ${this._success ? html`<div class="message message-success">${this._success}</div>` : nothing}

            <div class="card">
                <div class="card-title">Team Members</div>
                ${this._members.length > 0 ? html`
                    <ul class="member-list">
                        ${this._members.map(m => html`
                            <li class="member-item">
                                <div class="member-info">
                                    <div class="member-avatar">${this._getInitials(m.name)}</div>
                                    <div class="member-details">
                                        <span class="member-name">${m.name}</span>
                                        <span class="member-email">${m.email}</span>
                                    </div>
                                </div>
                                <div class="member-meta">
                                    <span class="role-badge ${this._getRoleBadgeClass(m.role)}">${m.role}</span>
                                </div>
                            </li>
                        `)}
                    </ul>
                ` : html`
                    <div class="empty-state">No team members found</div>
                `}
            </div>

            ${this._pendingInvites.length > 0 ? html`
                <div class="card">
                    <div class="card-title">Pending Invitations</div>
                    <ul class="member-list">
                        ${this._pendingInvites.map(inv => html`
                            <li class="member-item">
                                <div class="member-info">
                                    <div class="member-avatar">${this._getInitials(inv.name)}</div>
                                    <div class="member-details">
                                        <span class="member-name">${inv.email}</span>
                                        <span class="member-email">Invited as ${inv.role}</span>
                                    </div>
                                </div>
                                <div class="member-meta">
                                    <span class="status-badge pending">Pending</span>
                                </div>
                            </li>
                        `)}
                    </ul>
                </div>
            ` : nothing}

            ${this._showInviteModal ? this._renderInviteModal() : nothing}
        `;
    }

    private _renderInviteModal() {
        return html`
            <div class="modal-overlay" @click=${(e: Event) => {
                if (e.target === e.currentTarget) this._closeModal();
            }}>
                <div class="modal">
                    <div class="modal-title">Invite Team Member</div>

                    ${this._error ? html`<div class="message message-error">${this._error}</div>` : nothing}

                    <div class="form-group">
                        <label for="invite-email">Email Address</label>
                        <input
                            id="invite-email"
                            type="email"
                            placeholder="teammate@company.com"
                            .value=${this._inviteEmail}
                            ?disabled=${this._inviting}
                            @input=${(e: Event) => { this._inviteEmail = (e.target as HTMLInputElement).value; }}
                        />
                    </div>

                    <div class="form-group">
                        <label for="invite-role">Role</label>
                        <select
                            id="invite-role"
                            .value=${this._inviteRole}
                            ?disabled=${this._inviting}
                            @change=${(e: Event) => { this._inviteRole = (e.target as HTMLSelectElement).value; }}
                        >
                            <option value="Admin">Admin</option>
                            <option value="Builder">Builder</option>
                            <option value="PM">PM</option>
                            <option value="Viewer">Viewer</option>
                        </select>
                    </div>

                    <div class="modal-actions">
                        <button class="btn btn-secondary" ?disabled=${this._inviting} @click=${this._closeModal}>
                            Cancel
                        </button>
                        <button class="btn btn-primary" ?disabled=${this._inviting} @click=${this._sendInvite}>
                            ${this._inviting ? 'Sending...' : 'Send Invitation'}
                        </button>
                    </div>
                </div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-settings-team': FBSettingsTeam;
    }
}
