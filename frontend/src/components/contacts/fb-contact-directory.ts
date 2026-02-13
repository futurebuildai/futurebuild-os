/**
 * fb-contact-directory — Org-wide contact directory page.
 * See FRONTEND_V2_SPEC.md §13.7.C
 *
 * Searchable, filterable list of all org contacts.
 * Route: /settings/contacts or /contacts
 */
import { html, css, nothing, type TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import './fb-contact-detail';
import type { ContactDetail } from './fb-contact-detail';

interface ContactListItem {
    id: string;
    name: string;
    phone?: string;
    email?: string;
    company?: string;
    role: 'Subcontractor' | 'Client';
    trades?: string[];
    portal_enabled?: boolean;
    total_projects?: number;
    on_time_rate?: number;
}

@customElement('fb-contact-directory')
export class FBContactDirectory extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                max-width: 900px;
                margin: 0 auto;
                padding: 32px 16px;
            }

            .header {
                display: flex;
                justify-content: space-between;
                align-items: flex-start;
                margin-bottom: 24px;
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

            .header-actions {
                display: flex;
                gap: 12px;
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

            /* Search and filters */
            .filters {
                display: flex;
                gap: 12px;
                margin-bottom: 24px;
            }

            .search-box {
                flex: 1;
                position: relative;
            }

            .search-input {
                width: 100%;
                padding: 12px 16px 12px 44px;
                background: var(--fb-surface-1, #1a1a2e);
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 8px;
                color: var(--fb-text-primary, #e0e0e0);
                font-size: 14px;
                box-sizing: border-box;
            }

            .search-input:focus {
                outline: none;
                border-color: var(--fb-accent, #6366f1);
            }

            .search-input::placeholder {
                color: var(--fb-text-tertiary, #707080);
            }

            .search-icon {
                position: absolute;
                left: 16px;
                top: 50%;
                transform: translateY(-50%);
                color: var(--fb-text-tertiary, #707080);
            }

            .filter-select {
                padding: 12px 16px;
                background: var(--fb-surface-1, #1a1a2e);
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 8px;
                color: var(--fb-text-primary, #e0e0e0);
                font-size: 14px;
                min-width: 140px;
            }

            .filter-select:focus {
                outline: none;
                border-color: var(--fb-accent, #6366f1);
            }

            /* Contact list */
            .contact-list {
                background: var(--fb-surface-1, #1a1a2e);
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 12px;
                overflow: hidden;
            }

            .list-header {
                display: grid;
                grid-template-columns: 2fr 1fr 1fr 1fr 60px;
                gap: 16px;
                padding: 12px 20px;
                background: var(--fb-surface-2, #252540);
                font-size: 11px;
                font-weight: 600;
                color: var(--fb-text-tertiary, #707080);
                text-transform: uppercase;
                letter-spacing: 0.05em;
            }

            .contact-row {
                display: grid;
                grid-template-columns: 2fr 1fr 1fr 1fr 60px;
                gap: 16px;
                padding: 16px 20px;
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
                cursor: pointer;
                transition: background 0.15s;
                align-items: center;
            }

            .contact-row:last-child {
                border-bottom: none;
            }

            .contact-row:hover {
                background: var(--fb-surface-2, #252540);
            }

            .contact-info {
                display: flex;
                align-items: center;
                gap: 12px;
            }

            .avatar {
                width: 36px;
                height: 36px;
                border-radius: 50%;
                background: var(--fb-accent, #6366f1);
                color: #fff;
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 13px;
                font-weight: 600;
                flex-shrink: 0;
            }

            .contact-details {
                min-width: 0;
            }

            .contact-name {
                font-size: 14px;
                font-weight: 500;
                color: var(--fb-text-primary, #e0e0e0);
                white-space: nowrap;
                overflow: hidden;
                text-overflow: ellipsis;
            }

            .contact-company {
                font-size: 12px;
                color: var(--fb-text-tertiary, #707080);
                white-space: nowrap;
                overflow: hidden;
                text-overflow: ellipsis;
            }

            .contact-phone {
                font-size: 13px;
                color: var(--fb-text-secondary, #a0a0b0);
                font-family: monospace;
            }

            .trades-list {
                display: flex;
                flex-wrap: wrap;
                gap: 4px;
            }

            .trade-tag {
                padding: 2px 8px;
                background: rgba(99, 102, 241, 0.15);
                color: var(--fb-accent, #6366f1);
                border-radius: 4px;
                font-size: 11px;
                font-weight: 500;
            }

            .badges {
                display: flex;
                gap: 6px;
            }

            .badge {
                padding: 4px 8px;
                border-radius: 4px;
                font-size: 10px;
                font-weight: 600;
            }

            .badge.portal {
                background: rgba(34, 197, 94, 0.15);
                color: #22c55e;
            }

            .badge.sms {
                background: rgba(156, 163, 175, 0.15);
                color: #9ca3af;
            }

            .badge.on-time {
                background: rgba(34, 197, 94, 0.15);
                color: #22c55e;
            }

            .stat-value {
                font-size: 14px;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .stat-value.good {
                color: #22c55e;
            }

            /* Empty state */
            .empty-state {
                padding: 60px 40px;
                text-align: center;
            }

            .empty-icon {
                width: 64px;
                height: 64px;
                margin: 0 auto 16px;
                border-radius: 50%;
                background: var(--fb-surface-2, #252540);
                display: flex;
                align-items: center;
                justify-content: center;
                color: var(--fb-text-tertiary, #707080);
                font-size: 28px;
            }

            .empty-title {
                font-size: 18px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
                margin-bottom: 8px;
            }

            .empty-text {
                font-size: 14px;
                color: var(--fb-text-secondary, #a0a0b0);
                margin-bottom: 20px;
            }

            /* Loading */
            .loading {
                padding: 60px;
                text-align: center;
                color: var(--fb-text-secondary, #a0a0b0);
            }

            /* Pagination */
            .pagination {
                display: flex;
                justify-content: center;
                gap: 8px;
                padding: 16px;
                border-top: 1px solid var(--fb-border, #2a2a3e);
            }

            .page-btn {
                padding: 8px 12px;
                border-radius: 6px;
                border: 1px solid var(--fb-border, #2a2a3e);
                background: transparent;
                color: var(--fb-text-secondary, #a0a0b0);
                font-size: 13px;
                cursor: pointer;
                transition: all 0.15s;
            }

            .page-btn:hover:not(:disabled) {
                background: var(--fb-surface-2, #252540);
                color: var(--fb-text-primary, #e0e0e0);
            }

            .page-btn.active {
                background: var(--fb-accent, #6366f1);
                border-color: var(--fb-accent, #6366f1);
                color: #fff;
            }

            .page-btn:disabled {
                opacity: 0.5;
                cursor: not-allowed;
            }

            @media (max-width: 768px) {
                .header {
                    flex-direction: column;
                    gap: 16px;
                }

                .filters {
                    flex-direction: column;
                }

                .list-header {
                    display: none;
                }

                .contact-row {
                    grid-template-columns: 1fr auto;
                }

                .contact-row > *:not(:first-child):not(:last-child) {
                    display: none;
                }
            }
        `,
    ];

    @state() private _loading = true;
    @state() private _contacts: ContactListItem[] = [];
    @state() private _filteredContacts: ContactListItem[] = [];
    @state() private _searchQuery = '';
    @state() private _tradeFilter = '';
    @state() private _portalFilter = '';
    @state() private _selectedContact: ContactDetail | null = null;
    @state() private _showDetail = false;

    override connectedCallback() {
        super.connectedCallback();
        void this._loadContacts();
    }

    private async _loadContacts() {
        this._loading = true;
        try {
            // TODO: Replace with actual API call
            // this._contacts = await api.contacts.list();

            // Mock data
            this._contacts = [
                { id: '1', name: 'Rodriguez Framing', phone: '555-0101', company: 'Rodriguez Construction', role: 'Subcontractor', trades: ['Framing'], portal_enabled: false, total_projects: 5, on_time_rate: 0.92 },
                { id: '2', name: "Jake's Electric", phone: '555-0199', email: 'jake@jakeselectric.com', company: "Jake's Electric LLC", role: 'Subcontractor', trades: ['Electrical', 'Fire Alarm'], portal_enabled: true, total_projects: 8, on_time_rate: 0.88 },
                { id: '3', name: "Mike's Plumbing", phone: '555-0201', company: "Mike's Plumbing Co", role: 'Subcontractor', trades: ['Plumbing'], portal_enabled: false, total_projects: 3, on_time_rate: 0.95 },
                { id: '4', name: 'Sarah Chen', phone: '555-0300', email: 'sarah@homeowner.com', role: 'Client', portal_enabled: true, total_projects: 1 },
                { id: '5', name: 'Premium Roofing', phone: '555-0401', company: 'Premium Roofing Inc', role: 'Subcontractor', trades: ['Roofing'], portal_enabled: false, total_projects: 2 },
                { id: '6', name: 'Cool Air HVAC', phone: '555-0501', email: 'service@coolairhvac.com', company: 'Cool Air HVAC', role: 'Subcontractor', trades: ['HVAC'], portal_enabled: true, total_projects: 4, on_time_rate: 0.78 },
            ];
            this._applyFilters();
        } catch (err) {
            console.warn('[FBContactDirectory] Failed to load contacts:', err);
        } finally {
            this._loading = false;
        }
    }

    private _applyFilters() {
        let filtered = [...this._contacts];

        // Search filter
        if (this._searchQuery) {
            const query = this._searchQuery.toLowerCase();
            filtered = filtered.filter((c) =>
                c.name.toLowerCase().includes(query) ||
                c.phone?.includes(query) ||
                c.email?.toLowerCase().includes(query) ||
                c.company?.toLowerCase().includes(query)
            );
        }

        // Trade filter
        if (this._tradeFilter) {
            filtered = filtered.filter((c) =>
                c.trades?.some((t) => t.toLowerCase() === this._tradeFilter.toLowerCase())
            );
        }

        // Portal filter
        if (this._portalFilter === 'portal') {
            filtered = filtered.filter((c) => c.portal_enabled);
        } else if (this._portalFilter === 'sms') {
            filtered = filtered.filter((c) => !c.portal_enabled);
        }

        this._filteredContacts = filtered;
    }

    private _handleSearchInput(e: Event) {
        this._searchQuery = (e.target as HTMLInputElement).value;
        this._applyFilters();
    }

    private _handleTradeFilter(e: Event) {
        this._tradeFilter = (e.target as HTMLSelectElement).value;
        this._applyFilters();
    }

    private _handlePortalFilter(e: Event) {
        this._portalFilter = (e.target as HTMLSelectElement).value;
        this._applyFilters();
    }

    private _handleContactClick(contact: ContactListItem) {
        // Load full contact detail
        this._selectedContact = {
            ...contact,
            contact_preference: contact.portal_enabled ? 'Email' : 'SMS',
            project_history: [
                { project_id: 'p1', project_name: '123 Main St', phase_name: contact.trades?.[0] ?? 'General', status: 'active' as const },
            ],
        };
        this._showDetail = true;
    }

    private _handleDetailClose() {
        this._showDetail = false;
        this._selectedContact = null;
    }

    private _handleAddContact() {
        this.emit('fb-navigate', { view: 'contacts-add' });
    }

    private _handleBulkImport() {
        this.emit('fb-navigate', { view: 'contacts-bulk' });
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

    private _getAllTrades(): string[] {
        const trades = new Set<string>();
        for (const contact of this._contacts) {
            contact.trades?.forEach((t) => trades.add(t));
        }
        return Array.from(trades).sort();
    }

    private _renderContactRow(contact: ContactListItem): TemplateResult {
        return html`
            <div class="contact-row" @click=${() => this._handleContactClick(contact)}>
                <div class="contact-info">
                    <div class="avatar">${this._getInitials(contact.name)}</div>
                    <div class="contact-details">
                        <div class="contact-name">${contact.name}</div>
                        ${contact.company ? html`<div class="contact-company">${contact.company}</div>` : nothing}
                    </div>
                </div>
                <div class="contact-phone">${contact.phone ?? '—'}</div>
                <div class="trades-list">
                    ${contact.trades?.slice(0, 2).map((t) => html`<span class="trade-tag">${t}</span>`) ?? html`<span style="color: var(--fb-text-tertiary)">—</span>`}
                </div>
                <div class="badges">
                    ${contact.portal_enabled
                        ? html`<span class="badge portal">Portal</span>`
                        : html`<span class="badge sms">SMS</span>`}
                    ${contact.on_time_rate !== undefined && contact.total_projects && contact.total_projects >= 2
                        ? html`<span class="badge on-time">${Math.round(contact.on_time_rate * 100)}% OT</span>`
                        : nothing}
                </div>
                <div class="stat-value">${contact.total_projects ?? 0}</div>
            </div>
        `;
    }

    override render(): TemplateResult {
        const allTrades = this._getAllTrades();

        return html`
            <div class="back-link" @click=${this._handleBack}>← Back to Feed</div>

            <div class="header">
                <div class="header-text">
                    <div class="title">Contact Directory</div>
                    <div class="subtitle">
                        ${this._contacts.length} contact${this._contacts.length !== 1 ? 's' : ''} in your organization
                    </div>
                </div>
                <div class="header-actions">
                    <button class="btn btn-secondary" @click=${this._handleBulkImport}>Bulk Import</button>
                    <button class="btn btn-primary" @click=${this._handleAddContact}>+ Add Contact</button>
                </div>
            </div>

            <div class="filters">
                <div class="search-box">
                    <span class="search-icon">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <circle cx="11" cy="11" r="8"/>
                            <line x1="21" y1="21" x2="16.65" y2="16.65"/>
                        </svg>
                    </span>
                    <input
                        type="text"
                        class="search-input"
                        placeholder="Search by name, phone, or email..."
                        .value=${this._searchQuery}
                        @input=${this._handleSearchInput}
                    />
                </div>
                <select class="filter-select" @change=${this._handleTradeFilter}>
                    <option value="">All Trades</option>
                    ${allTrades.map((t) => html`<option value=${t}>${t}</option>`)}
                </select>
                <select class="filter-select" @change=${this._handlePortalFilter}>
                    <option value="">All Access</option>
                    <option value="portal">Portal Enabled</option>
                    <option value="sms">SMS Only</option>
                </select>
            </div>

            <div class="contact-list">
                ${this._loading
                    ? html`<div class="loading">Loading contacts...</div>`
                    : this._filteredContacts.length === 0
                    ? html`
                          <div class="empty-state">
                              <div class="empty-icon">
                                  <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                      <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                                      <circle cx="9" cy="7" r="4"/>
                                      <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
                                      <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
                                  </svg>
                              </div>
                              <div class="empty-title">
                                  ${this._searchQuery || this._tradeFilter || this._portalFilter
                                      ? 'No contacts found'
                                      : 'No contacts yet'}
                              </div>
                              <div class="empty-text">
                                  ${this._searchQuery || this._tradeFilter || this._portalFilter
                                      ? 'Try adjusting your filters'
                                      : 'Add your subcontractors and clients to get started'}
                              </div>
                              ${!this._searchQuery && !this._tradeFilter && !this._portalFilter
                                  ? html`<button class="btn btn-primary" @click=${this._handleAddContact}>+ Add Contact</button>`
                                  : nothing}
                          </div>
                      `
                    : html`
                          <div class="list-header">
                              <span>Contact</span>
                              <span>Phone</span>
                              <span>Trades</span>
                              <span>Access</span>
                              <span>Projects</span>
                          </div>
                          ${this._filteredContacts.map((c) => this._renderContactRow(c))}
                      `}
            </div>

            <fb-contact-detail
                .contact=${this._selectedContact}
                .open=${this._showDetail}
                @fb-contact-detail-close=${this._handleDetailClose}
            ></fb-contact-detail>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-contact-directory': FBContactDirectory;
    }
}
