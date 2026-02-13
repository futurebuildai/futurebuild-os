/**
 * fb-contact-detail — Contact detail slide-over panel.
 * See FRONTEND_V2_SPEC.md §13.7.B
 *
 * Shows full contact info, performance stats, and project history.
 * Slide-over panel accessible from phase grid or directory.
 */
import { html, css, nothing, type TemplateResult } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

export interface ContactDetail {
    id: string;
    name: string;
    phone?: string;
    email?: string;
    company?: string;
    role: 'Subcontractor' | 'Client';
    contact_preference?: 'SMS' | 'Email' | 'Both';
    portal_enabled?: boolean;
    trades?: string[];
    license_number?: string;
    address_city?: string;
    address_state?: string;
    website?: string;
    notes?: string;
    // Agent-computed fields
    last_contacted_at?: string;
    total_projects?: number;
    avg_response_time_hours?: number;
    on_time_rate?: number;
    // Project history
    project_history?: Array<{
        project_id: string;
        project_name: string;
        phase_name: string;
        status: 'active' | 'completed' | 'pending';
    }>;
}

@customElement('fb-contact-detail')
export class FBContactDetail extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
            }

            .overlay {
                position: fixed;
                inset: 0;
                background: rgba(0, 0, 0, 0.5);
                z-index: 1000;
            }

            .panel {
                position: fixed;
                top: 0;
                right: 0;
                bottom: 0;
                width: 420px;
                max-width: 100%;
                background: var(--fb-surface-1, #1a1a2e);
                border-left: 1px solid var(--fb-border, #2a2a3e);
                display: flex;
                flex-direction: column;
                z-index: 1001;
                animation: slideIn 0.2s ease;
            }

            @keyframes slideIn {
                from {
                    transform: translateX(100%);
                }
                to {
                    transform: translateX(0);
                }
            }

            .panel-header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: 20px 24px;
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
            }

            .panel-title {
                font-size: 18px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .close-btn {
                width: 32px;
                height: 32px;
                border-radius: 6px;
                border: none;
                background: transparent;
                color: var(--fb-text-secondary, #a0a0b0);
                cursor: pointer;
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 20px;
                transition: all 0.15s;
            }

            .close-btn:hover {
                background: var(--fb-surface-2, #252540);
                color: var(--fb-text-primary, #e0e0e0);
            }

            .panel-content {
                flex: 1;
                overflow-y: auto;
                padding: 24px;
            }

            .section {
                margin-bottom: 24px;
            }

            .section:last-child {
                margin-bottom: 0;
            }

            .section-title {
                font-size: 11px;
                font-weight: 600;
                color: var(--fb-text-tertiary, #707080);
                text-transform: uppercase;
                letter-spacing: 0.05em;
                margin-bottom: 12px;
                padding-bottom: 8px;
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
            }

            .contact-header {
                display: flex;
                align-items: center;
                gap: 16px;
                margin-bottom: 20px;
            }

            .avatar {
                width: 64px;
                height: 64px;
                border-radius: 50%;
                background: var(--fb-accent, #6366f1);
                color: #fff;
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 24px;
                font-weight: 600;
            }

            .contact-main {
                flex: 1;
            }

            .contact-name {
                font-size: 20px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
                margin-bottom: 4px;
            }

            .contact-company {
                font-size: 14px;
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .info-row {
                display: flex;
                align-items: center;
                gap: 12px;
                padding: 10px 0;
                border-bottom: 1px solid var(--fb-border-light, #1e1e32);
            }

            .info-row:last-child {
                border-bottom: none;
            }

            .info-icon {
                width: 20px;
                height: 20px;
                color: var(--fb-text-tertiary, #707080);
                display: flex;
                align-items: center;
                justify-content: center;
            }

            .info-label {
                width: 80px;
                font-size: 12px;
                color: var(--fb-text-tertiary, #707080);
            }

            .info-value {
                flex: 1;
                font-size: 14px;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .info-value a {
                color: var(--fb-accent, #6366f1);
                text-decoration: none;
            }

            .info-value a:hover {
                text-decoration: underline;
            }

            .trades-list {
                display: flex;
                flex-wrap: wrap;
                gap: 8px;
            }

            .trade-tag {
                padding: 4px 10px;
                background: rgba(99, 102, 241, 0.15);
                color: var(--fb-accent, #6366f1);
                border-radius: 4px;
                font-size: 12px;
                font-weight: 500;
            }

            .portal-status {
                display: flex;
                align-items: center;
                gap: 8px;
            }

            .portal-dot {
                width: 8px;
                height: 8px;
                border-radius: 50%;
            }

            .portal-dot.enabled {
                background: #22c55e;
            }

            .portal-dot.disabled {
                background: var(--fb-text-tertiary, #707080);
            }

            .notes-text {
                font-size: 14px;
                color: var(--fb-text-secondary, #a0a0b0);
                line-height: 1.6;
                font-style: italic;
            }

            /* Stats */
            .stats-grid {
                display: grid;
                grid-template-columns: 1fr 1fr;
                gap: 12px;
            }

            .stat-card {
                padding: 16px;
                background: var(--fb-surface-2, #252540);
                border-radius: 8px;
                text-align: center;
            }

            .stat-value {
                font-size: 24px;
                font-weight: 700;
                color: var(--fb-text-primary, #e0e0e0);
                margin-bottom: 4px;
            }

            .stat-value.good {
                color: #22c55e;
            }

            .stat-value.warning {
                color: #f59e0b;
            }

            .stat-label {
                font-size: 11px;
                color: var(--fb-text-tertiary, #707080);
                text-transform: uppercase;
            }

            /* Project history */
            .project-list {
                display: flex;
                flex-direction: column;
                gap: 8px;
            }

            .project-item {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: 12px 16px;
                background: var(--fb-surface-2, #252540);
                border-radius: 8px;
                cursor: pointer;
                transition: background 0.15s;
            }

            .project-item:hover {
                background: var(--fb-surface-3, #2d2d4a);
            }

            .project-info {
                display: flex;
                flex-direction: column;
                gap: 2px;
            }

            .project-name {
                font-size: 14px;
                font-weight: 500;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .project-phase {
                font-size: 12px;
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .project-status {
                padding: 4px 10px;
                border-radius: 4px;
                font-size: 11px;
                font-weight: 600;
                text-transform: uppercase;
            }

            .project-status.active {
                background: rgba(99, 102, 241, 0.15);
                color: var(--fb-accent, #6366f1);
            }

            .project-status.completed {
                background: rgba(34, 197, 94, 0.15);
                color: #22c55e;
            }

            .project-status.pending {
                background: rgba(156, 163, 175, 0.15);
                color: #9ca3af;
            }

            .empty-state {
                text-align: center;
                padding: 32px;
                color: var(--fb-text-secondary, #a0a0b0);
                font-size: 14px;
            }

            /* Actions */
            .panel-actions {
                display: flex;
                gap: 12px;
                padding: 16px 24px;
                border-top: 1px solid var(--fb-border, #2a2a3e);
            }

            .btn {
                flex: 1;
                padding: 10px 20px;
                border-radius: 6px;
                font-size: 14px;
                font-weight: 600;
                cursor: pointer;
                border: none;
                transition: all 0.15s ease;
                text-align: center;
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

            .btn-danger {
                background: transparent;
                border: 1px solid #ef4444;
                color: #ef4444;
            }

            .btn-danger:hover {
                background: rgba(239, 68, 68, 0.1);
            }

            @media (max-width: 480px) {
                .panel {
                    width: 100%;
                }
            }
        `,
    ];

    /** Contact data to display */
    @property({ attribute: false })
    contact: ContactDetail | null = null;

    /** Whether the panel is open */
    @property({ type: Boolean, reflect: true }) open = false;

    // Reserved for future use when loading contact history
    // private _loading = false;

    private _getInitials(name: string): string {
        const parts = name.trim().split(/\s+/);
        if (parts.length >= 2) {
            return (parts[0]![0]! + parts[parts.length - 1]![0]!).toUpperCase();
        }
        return name.substring(0, 2).toUpperCase();
    }

    private _formatDate(iso: string): string {
        const d = new Date(iso);
        const now = new Date();
        const diffMs = now.getTime() - d.getTime();
        const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

        if (diffDays === 0) return 'Today';
        if (diffDays === 1) return 'Yesterday';
        if (diffDays < 7) return `${diffDays} days ago`;
        if (diffDays < 30) return `${Math.floor(diffDays / 7)} weeks ago`;
        return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
    }

    private _handleClose() {
        this.open = false;
        this.emit('fb-contact-detail-close');
    }

    private _handleEdit() {
        this.emit('fb-contact-edit', { contact: this.contact });
    }

    private _handleProjectClick(projectId: string) {
        this.emit('fb-navigate', { view: 'project', projectId });
        this._handleClose();
    }

    private _handleOverlayClick(e: MouseEvent) {
        if (e.target === e.currentTarget) {
            this._handleClose();
        }
    }

    private _renderContactInfo(): TemplateResult {
        const c = this.contact!;

        return html`
            <div class="section">
                <div class="contact-header">
                    <div class="avatar">${this._getInitials(c.name)}</div>
                    <div class="contact-main">
                        <div class="contact-name">${c.name}</div>
                        ${c.company ? html`<div class="contact-company">${c.company}</div>` : nothing}
                    </div>
                </div>

                ${c.phone
                    ? html`
                          <div class="info-row">
                              <span class="info-label">Phone</span>
                              <span class="info-value">
                                  <a href="tel:${c.phone}">${c.phone}</a>
                              </span>
                          </div>
                      `
                    : nothing}

                ${c.email
                    ? html`
                          <div class="info-row">
                              <span class="info-label">Email</span>
                              <span class="info-value">
                                  <a href="mailto:${c.email}">${c.email}</a>
                              </span>
                          </div>
                      `
                    : nothing}

                ${c.license_number
                    ? html`
                          <div class="info-row">
                              <span class="info-label">License</span>
                              <span class="info-value">${c.license_number}</span>
                          </div>
                      `
                    : nothing}

                ${c.trades && c.trades.length > 0
                    ? html`
                          <div class="info-row">
                              <span class="info-label">Trades</span>
                              <div class="trades-list">
                                  ${c.trades.map((t) => html`<span class="trade-tag">${t}</span>`)}
                              </div>
                          </div>
                      `
                    : nothing}

                <div class="info-row">
                    <span class="info-label">Contact via</span>
                    <span class="info-value">${c.contact_preference ?? 'SMS'}</span>
                </div>

                <div class="info-row">
                    <span class="info-label">Portal</span>
                    <div class="portal-status">
                        <span class="portal-dot ${c.portal_enabled ? 'enabled' : 'disabled'}"></span>
                        <span class="info-value">${c.portal_enabled ? 'Enabled' : 'Disabled'}</span>
                    </div>
                </div>
            </div>
        `;
    }

    private _renderNotes(): TemplateResult | typeof nothing {
        if (!this.contact?.notes) return nothing;

        return html`
            <div class="section">
                <div class="section-title">Notes</div>
                <p class="notes-text">"${this.contact.notes}"</p>
            </div>
        `;
    }

    private _renderStats(): TemplateResult | typeof nothing {
        const c = this.contact!;

        // Only show stats if they have 2+ projects
        if (!c.total_projects || c.total_projects < 2) return nothing;

        const onTimeClass = c.on_time_rate !== undefined
            ? c.on_time_rate >= 0.9 ? 'good' : c.on_time_rate >= 0.7 ? '' : 'warning'
            : '';

        return html`
            <div class="section">
                <div class="section-title">Performance</div>
                <div class="stats-grid">
                    <div class="stat-card">
                        <div class="stat-value">${c.total_projects}</div>
                        <div class="stat-label">Projects</div>
                    </div>
                    ${c.on_time_rate !== undefined
                        ? html`
                              <div class="stat-card">
                                  <div class="stat-value ${onTimeClass}">${Math.round(c.on_time_rate * 100)}%</div>
                                  <div class="stat-label">On-time</div>
                              </div>
                          `
                        : nothing}
                    ${c.avg_response_time_hours !== undefined
                        ? html`
                              <div class="stat-card">
                                  <div class="stat-value">${c.avg_response_time_hours.toFixed(1)}h</div>
                                  <div class="stat-label">Avg Response</div>
                              </div>
                          `
                        : nothing}
                    ${c.last_contacted_at
                        ? html`
                              <div class="stat-card">
                                  <div class="stat-value" style="font-size: 16px;">${this._formatDate(c.last_contacted_at)}</div>
                                  <div class="stat-label">Last Contacted</div>
                              </div>
                          `
                        : nothing}
                </div>
            </div>
        `;
    }

    private _renderProjectHistory(): TemplateResult | typeof nothing {
        const history = this.contact?.project_history;
        if (!history || history.length === 0) return nothing;

        return html`
            <div class="section">
                <div class="section-title">Project History</div>
                <div class="project-list">
                    ${history.map(
                        (p) => html`
                            <div class="project-item" @click=${() => this._handleProjectClick(p.project_id)}>
                                <div class="project-info">
                                    <span class="project-name">${p.project_name}</span>
                                    <span class="project-phase">${p.phase_name}</span>
                                </div>
                                <span class="project-status ${p.status}">${p.status}</span>
                            </div>
                        `
                    )}
                </div>
            </div>
        `;
    }

    override render(): TemplateResult {
        if (!this.open) return html``;

        if (!this.contact) {
            return html`
                <div class="overlay" @click=${this._handleOverlayClick}>
                    <div class="panel">
                        <div class="panel-header">
                            <span class="panel-title">Contact Details</span>
                            <button class="close-btn" @click=${this._handleClose}>&times;</button>
                        </div>
                        <div class="panel-content">
                            <div class="empty-state">No contact selected</div>
                        </div>
                    </div>
                </div>
            `;
        }

        return html`
            <div class="overlay" @click=${this._handleOverlayClick}>
                <div class="panel">
                    <div class="panel-header">
                        <span class="panel-title">Contact Details</span>
                        <button class="close-btn" @click=${this._handleClose}>&times;</button>
                    </div>

                    <div class="panel-content">
                        ${this._renderContactInfo()}
                        ${this._renderNotes()}
                        ${this._renderStats()}
                        ${this._renderProjectHistory()}
                    </div>

                    <div class="panel-actions">
                        <button class="btn btn-secondary" @click=${this._handleEdit}>Edit</button>
                        <button class="btn btn-primary" @click=${this._handleClose}>Close</button>
                    </div>
                </div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-contact-detail': FBContactDetail;
    }
}
