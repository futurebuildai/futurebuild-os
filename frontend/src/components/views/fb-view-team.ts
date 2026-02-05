/**
 * FBViewTeam - Team Management View
 * See PHASE_12_PRD.md Step 80: Organization Manager
 *
 * Custom team UI that lists org members and pending invitations.
 * Replaces the Clerk OrganizationProfile component.
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { api, UserProfile, Invite } from '../../services/api';

type ViewState = 'loading' | 'ready' | 'error';
type ModalState = 'closed' | 'creating' | 'confirming' | 'submitting';

@customElement('fb-view-team')
export class FBViewTeam extends FBViewElement {
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
                display: flex;
                justify-content: space-between;
                align-items: center;
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

            .btn-danger {
                padding: var(--fb-spacing-xs) var(--fb-spacing-sm);
                background: transparent;
                color: var(--fb-error);
                border: 1px solid var(--fb-error);
                border-radius: var(--fb-radius-sm);
                font-size: var(--fb-text-xs);
                cursor: pointer;
                transition: all 0.2s ease;
            }

            .btn-danger:hover {
                background: var(--fb-error);
                color: white;
            }

            .section-title {
                font-size: var(--fb-text-lg);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin-bottom: var(--fb-spacing-md);
            }

            .section {
                margin-bottom: var(--fb-spacing-xl);
            }

            .table-container {
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                overflow: hidden;
            }

            table {
                width: 100%;
                border-collapse: collapse;
            }

            th, td {
                padding: var(--fb-spacing-md);
                text-align: left;
                border-bottom: 1px solid var(--fb-border-light);
            }

            th {
                background: var(--fb-bg-tertiary);
                font-size: var(--fb-text-sm);
                font-weight: 600;
                color: var(--fb-text-secondary);
                text-transform: uppercase;
                letter-spacing: 0.05em;
            }

            td {
                font-size: var(--fb-text-sm);
                color: var(--fb-text-primary);
            }

            tr:last-child td {
                border-bottom: none;
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

            .empty-state {
                text-align: center;
                padding: var(--fb-spacing-2xl);
                color: var(--fb-text-muted);
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

            .date-cell {
                color: var(--fb-text-secondary);
                font-size: var(--fb-text-xs);
            }

            /* Modal styles */
            .modal-backdrop {
                position: fixed;
                top: 0;
                left: 0;
                right: 0;
                bottom: 0;
                background: rgba(0, 0, 0, 0.7);
                display: flex;
                align-items: center;
                justify-content: center;
                z-index: 1000;
            }

            .modal {
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-xl);
                max-width: 400px;
                width: 90%;
            }

            .modal-title {
                font-size: var(--fb-text-lg);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin-bottom: var(--fb-spacing-lg);
            }

            .form-group {
                margin-bottom: var(--fb-spacing-md);
            }

            label {
                display: block;
                font-size: var(--fb-text-sm);
                color: var(--fb-text-secondary);
                margin-bottom: var(--fb-spacing-xs);
            }

            input, select {
                width: 100%;
                padding: var(--fb-spacing-md);
                background: var(--fb-bg-tertiary);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-md);
                color: var(--fb-text-primary);
                font-size: var(--fb-text-base);
                box-sizing: border-box;
            }

            input:focus, select:focus {
                outline: none;
                border-color: var(--fb-primary);
            }

            .modal-actions {
                display: flex;
                gap: var(--fb-spacing-sm);
                justify-content: flex-end;
                margin-top: var(--fb-spacing-lg);
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

            .form-error {
                color: var(--fb-error);
                font-size: var(--fb-text-sm);
                margin-bottom: var(--fb-spacing-md);
            }
        `,
    ];

    @state() private _viewState: ViewState = 'loading';
    @state() private _members: UserProfile[] = [];
    @state() private _invites: Invite[] = [];
    @state() private _error = '';

    // Invite modal state
    @state() private _modalState: ModalState = 'closed';
    @state() private _formEmail = '';
    @state() private _formRole = 'Builder';
    @state() private _formError = '';

    // Revoke confirmation state
    @state() private _confirmRevokeId = '';
    @state() private _confirmRevokeEmail = '';

    override connectedCallback(): void {
        super.connectedCallback();
        void this._loadData();
    }

    private async _loadData(): Promise<void> {
        this._viewState = 'loading';
        this._error = '';

        try {
            // Fetch members and invites in parallel; invites may 403 for non-admins
            const [members, invites] = await Promise.all([
                api.team.listMembers(),
                api.invites.list().catch(() => [] as Invite[]),
            ]);
            this._members = members;
            this._invites = invites;
            this._viewState = 'ready';
        } catch (err: unknown) {
            this._viewState = 'error';
            this._error = this._extractErrorMessage(err);
        }
    }

    // ========================================================================
    // Invite modal handlers
    // ========================================================================

    private _openCreateModal(): void {
        this._modalState = 'creating';
        this._formEmail = '';
        this._formRole = 'Builder';
        this._formError = '';
    }

    private _closeModal(): void {
        this._modalState = 'closed';
    }

    private _handleCreateStep1(): void {
        if (!this._formEmail || !this._formEmail.includes('@')) {
            this._formError = 'Please enter a valid email address';
            return;
        }
        this._formError = '';
        this._modalState = 'confirming';
    }

    private async _handleCreateConfirm(): Promise<void> {
        this._modalState = 'submitting';
        this._formError = '';

        try {
            await api.invites.create(this._formEmail, this._formRole);
            this._modalState = 'closed';
            void this._loadData();
        } catch (err: unknown) {
            this._modalState = 'confirming';
            this._formError = this._extractErrorMessage(err);
        }
    }

    private _openRevokeConfirm(id: string, email: string): void {
        this._confirmRevokeId = id;
        this._confirmRevokeEmail = email;
    }

    private _closeRevokeConfirm(): void {
        this._confirmRevokeId = '';
        this._confirmRevokeEmail = '';
    }

    private async _handleRevokeConfirm(): Promise<void> {
        const id = this._confirmRevokeId;
        this._confirmRevokeId = '';
        this._confirmRevokeEmail = '';

        try {
            await api.invites.revoke(id);
            void this._loadData();
        } catch (err) {
            console.error('Failed to revoke invitation:', err instanceof Error ? err.message : err);
        }
    }

    // ========================================================================
    // Error helpers
    // ========================================================================

    /** Extract a human-readable message from an unknown error. */
    private _extractErrorMessage(err: unknown): string {
        if (err instanceof Error) {
            return err.message;
        }
        if (typeof err === 'string') {
            return err;
        }
        // Backend returns {"error": {"code": N, "message": "..."}}
        const obj = err as Record<string, unknown>;
        if (obj && typeof obj === 'object') {
            if (typeof obj['message'] === 'string') return obj['message'];
            const inner = obj['error'];
            if (inner && typeof inner === 'object') {
                const detail = inner as Record<string, unknown>;
                if (typeof detail['message'] === 'string') return detail['message'];
            }
        }
        return 'Failed to load team data';
    }

    // ========================================================================
    // Formatting helpers
    // ========================================================================

    private _formatDate(dateStr: string): string {
        const date = new Date(dateStr);
        return date.toLocaleDateString('en-US', {
            month: 'short',
            day: 'numeric',
            year: 'numeric',
        });
    }

    private _formatDateTime(dateStr: string): string {
        const date = new Date(dateStr);
        return date.toLocaleDateString('en-US', {
            month: 'short',
            day: 'numeric',
            year: 'numeric',
            hour: 'numeric',
            minute: '2-digit',
        });
    }

    // ========================================================================
    // Render
    // ========================================================================

    override render(): TemplateResult {
        return html`
            <div class="header">
                <div>
                    <div class="title">Team</div>
                    <div class="subtitle">Manage your organization members and invitations</div>
                </div>
                <button class="btn-primary" @click=${this._openCreateModal.bind(this)}>
                    + Invite User
                </button>
            </div>

            ${this._renderContent()}
            ${this._modalState !== 'closed' ? this._renderModal() : nothing}
            ${this._confirmRevokeId ? this._renderRevokeConfirm() : nothing}
        `;
    }

    private _renderContent(): TemplateResult {
        switch (this._viewState) {
            case 'loading':
                return html`
                    <div class="table-container">
                        <div class="loading-state">
                            <div class="spinner"></div>
                        </div>
                    </div>
                `;

            case 'error':
                return html`
                    <div class="table-container">
                        <div class="error-state">
                            <p>${this._error}</p>
                            <button class="btn-secondary" @click=${this._loadData.bind(this)}>
                                Retry
                            </button>
                        </div>
                    </div>
                `;

            case 'ready':
                return html`
                    ${this._renderMembersSection()}
                    ${this._renderInvitesSection()}
                `;
        }
    }

    // ========================================================================
    // Members Section
    // ========================================================================

    private _renderMembersSection(): TemplateResult {
        return html`
            <div class="section">
                <div class="section-title">Members</div>
                <div class="table-container">
                    ${this._members.length === 0
                        ? html`
                            <div class="empty-state">
                                <p>No members found</p>
                            </div>
                        `
                        : html`
                            <table>
                                <thead>
                                    <tr>
                                        <th>Name</th>
                                        <th>Email</th>
                                        <th>Role</th>
                                        <th>Member Since</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    ${this._members.map(
                                        (member) => html`
                                            <tr>
                                                <td>${member.name}</td>
                                                <td>${member.email}</td>
                                                <td><span class="role-badge">${member.role}</span></td>
                                                <td class="date-cell">${this._formatDate(member.created_at)}</td>
                                            </tr>
                                        `
                                    )}
                                </tbody>
                            </table>
                        `}
                </div>
            </div>
        `;
    }

    // ========================================================================
    // Invites Section
    // ========================================================================

    private _renderInvitesSection(): TemplateResult {
        return html`
            <div class="section">
                <div class="section-title">Pending Invitations</div>
                <div class="table-container">
                    ${this._invites.length === 0
                        ? html`
                            <div class="empty-state">
                                <p>No pending invitations</p>
                            </div>
                        `
                        : html`
                            <table>
                                <thead>
                                    <tr>
                                        <th>Email</th>
                                        <th>Role</th>
                                        <th>Expires</th>
                                        <th></th>
                                    </tr>
                                </thead>
                                <tbody>
                                    ${this._invites.map(
                                        (invite) => html`
                                            <tr>
                                                <td>${invite.email}</td>
                                                <td><span class="role-badge">${invite.role}</span></td>
                                                <td class="date-cell">${this._formatDateTime(invite.expires_at)}</td>
                                                <td>
                                                    <button
                                                        class="btn-danger"
                                                        @click=${() => this._openRevokeConfirm(invite.id, invite.email)}
                                                    >
                                                        Revoke
                                                    </button>
                                                </td>
                                            </tr>
                                        `
                                    )}
                                </tbody>
                            </table>
                        `}
                </div>
            </div>
        `;
    }

    // ========================================================================
    // Invite Modal
    // ========================================================================

    private _renderModal(): TemplateResult {
        if (this._modalState === 'confirming' || this._modalState === 'submitting') {
            return this._renderConfirmStep();
        }
        return this._renderFormStep();
    }

    private _renderFormStep(): TemplateResult {
        return html`
            <div class="modal-backdrop" @click=${this._closeModal.bind(this)}>
                <div class="modal" @click=${(e: Event) => e.stopPropagation()}>
                    <div class="modal-title">Invite Team Member</div>

                    ${this._formError ? html`<p class="form-error">${this._formError}</p>` : nothing}

                    <div class="form-group">
                        <label for="invite-email">Email Address</label>
                        <input
                            id="invite-email"
                            type="email"
                            placeholder="user@example.com"
                            .value=${this._formEmail}
                            @input=${(e: Event) => {
                                this._formEmail = (e.target as HTMLInputElement).value;
                            }}
                        />
                    </div>

                    <div class="form-group">
                        <label for="invite-role">Role</label>
                        <select
                            id="invite-role"
                            .value=${this._formRole}
                            @change=${(e: Event) => {
                                this._formRole = (e.target as HTMLSelectElement).value;
                            }}
                        >
                            <option value="Admin">Admin</option>
                            <option value="Builder">Builder</option>
                            <option value="Viewer">Viewer</option>
                            <option value="Client">Client</option>
                            <option value="Subcontractor">Subcontractor</option>
                        </select>
                    </div>

                    <div class="modal-actions">
                        <button
                            class="btn-secondary"
                            @click=${this._closeModal.bind(this)}
                        >
                            Cancel
                        </button>
                        <button
                            class="btn-primary"
                            @click=${this._handleCreateStep1.bind(this)}
                        >
                            Send Invitation
                        </button>
                    </div>
                </div>
            </div>
        `;
    }

    private _renderConfirmStep(): TemplateResult {
        const isSubmitting = this._modalState === 'submitting';

        return html`
            <div class="modal-backdrop" @click=${this._closeModal.bind(this)}>
                <div class="modal" @click=${(e: Event) => e.stopPropagation()}>
                    <div class="modal-title">Confirm Invitation</div>

                    ${this._formError ? html`<p class="form-error">${this._formError}</p>` : nothing}

                    <p style="color: var(--fb-text-secondary); margin-bottom: var(--fb-spacing-lg); line-height: 1.5;">
                        Send an invitation to the following user?
                    </p>

                    <div style="background: var(--fb-bg-tertiary); border-radius: var(--fb-radius-md); padding: var(--fb-spacing-md); margin-bottom: var(--fb-spacing-lg);">
                        <div style="font-size: var(--fb-text-sm); color: var(--fb-text-muted); margin-bottom: var(--fb-spacing-xs);">Email</div>
                        <div style="color: var(--fb-text-primary); font-weight: 500;">${this._formEmail}</div>
                        <div style="font-size: var(--fb-text-sm); color: var(--fb-text-muted); margin-top: var(--fb-spacing-sm); margin-bottom: var(--fb-spacing-xs);">Role</div>
                        <div><span class="role-badge">${this._formRole}</span></div>
                    </div>

                    <div class="modal-actions">
                        <button
                            class="btn-secondary"
                            ?disabled=${isSubmitting}
                            @click=${() => { this._modalState = 'creating'; }}
                        >
                            Back
                        </button>
                        <button
                            class="btn-primary"
                            ?disabled=${isSubmitting}
                            @click=${this._handleCreateConfirm.bind(this)}
                        >
                            ${isSubmitting ? 'Sending...' : 'Confirm & Send'}
                        </button>
                    </div>
                </div>
            </div>
        `;
    }

    private _renderRevokeConfirm(): TemplateResult {
        return html`
            <div class="modal-backdrop" @click=${this._closeRevokeConfirm.bind(this)}>
                <div class="modal" @click=${(e: Event) => e.stopPropagation()}>
                    <div class="modal-title">Revoke Invitation</div>

                    <p style="color: var(--fb-text-secondary); margin-bottom: var(--fb-spacing-lg); line-height: 1.5;">
                        Are you sure you want to revoke the invitation for
                        <strong style="color: var(--fb-text-primary);">${this._confirmRevokeEmail}</strong>?
                        They will no longer be able to accept it.
                    </p>

                    <div class="modal-actions">
                        <button
                            class="btn-secondary"
                            @click=${this._closeRevokeConfirm.bind(this)}
                        >
                            Cancel
                        </button>
                        <button
                            class="btn-danger"
                            @click=${this._handleRevokeConfirm.bind(this)}
                        >
                            Revoke
                        </button>
                    </div>
                </div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-team': FBViewTeam;
    }
}
