/**
 * FBViewDirectory - Contact Directory View
 * See FRONTEND_V2_SPEC.md §10.3
 *
 * Full org contact directory with search and quick-add.
 * Accessible from user menu or setup_team feed card.
 */
import { html, css, nothing, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { api } from '../../services/api';
import type { Contact } from '../../types/models';
import '../feed/fb-contact-inline-add';
import { mockContactsService } from '../../services/mock-contacts-service';

@customElement('fb-view-directory')
export class FBViewDirectory extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                display: block;
                max-width: 800px;
                margin: 0 auto;
                padding: 24px 16px 80px;
            }

            .header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                margin-bottom: 24px;
            }

            h1 {
                font-size: 24px;
                font-weight: 700;
                color: var(--fb-text-primary, #F0F0F5);
                margin: 0;
            }

            .search-bar {
                margin-bottom: 20px;
            }

            .search-input {
                width: 100%;
                padding: 10px 14px;
                border: 1px solid var(--fb-border, rgba(255,255,255,0.05));
                border-radius: 8px;
                background: var(--fb-surface-1, #161821);
                color: var(--fb-text-primary, #F0F0F5);
                font-size: 14px;
                outline: none;
                box-sizing: border-box;
            }

            .search-input:focus {
                border-color: var(--fb-accent, #00FFA3);
            }

            .search-input::placeholder {
                color: var(--fb-text-tertiary, #5A5B66);
            }

            .btn-add {
                padding: 10px 20px;
                border-radius: 8px;
                background: var(--fb-accent, #00FFA3);
                color: #fff;
                font-size: 14px;
                font-weight: 600;
                border: none;
                cursor: pointer;
                white-space: nowrap;
            }

            .btn-add:hover { opacity: 0.9; }

            .contact-list {
                display: flex;
                flex-direction: column;
                gap: 8px;
            }

            .contact-card {
                display: flex;
                align-items: center;
                padding: 14px 16px;
                background: var(--fb-surface-1, #161821);
                border: 1px solid var(--fb-border, rgba(255,255,255,0.05));
                border-radius: 10px;
                transition: border-color 0.15s ease;
            }

            .contact-card:hover {
                border-color: var(--fb-accent-dim, #4a4d8a);
            }

            .contact-avatar {
                width: 36px;
                height: 36px;
                border-radius: 50%;
                background: var(--fb-surface-2, #1E2029);
                color: var(--fb-text-secondary, #8B8D98);
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 14px;
                font-weight: 600;
                flex-shrink: 0;
                margin-right: 12px;
            }

            .contact-info {
                flex: 1;
                min-width: 0;
            }

            .contact-name {
                font-size: 14px;
                font-weight: 600;
                color: var(--fb-text-primary, #F0F0F5);
            }

            .contact-detail {
                font-size: 13px;
                color: var(--fb-text-secondary, #8B8D98);
                margin-top: 2px;
            }

            .contact-badges {
                display: flex;
                gap: 6px;
                flex-shrink: 0;
            }

            .badge {
                font-size: 11px;
                padding: 2px 8px;
                border-radius: 10px;
                font-weight: 500;
            }

            .badge-pref {
                background: var(--fb-surface-2, #1E2029);
                color: var(--fb-text-tertiary, #5A5B66);
            }

            .badge-portal {
                background: rgba(0, 255, 163, 0.15);
                color: var(--fb-accent, #00FFA3);
            }

            .empty {
                text-align: center;
                padding: 60px 24px;
                color: var(--fb-text-secondary, #8B8D98);
            }

            .empty-title {
                font-size: 18px;
                font-weight: 600;
                color: var(--fb-text-primary, #F0F0F5);
                margin-bottom: 8px;
            }

            .count {
                font-size: 13px;
                color: var(--fb-text-tertiary, #5A5B66);
                margin-bottom: 16px;
            }

            .loading {
                display: flex;
                flex-direction: column;
                gap: 12px;
                margin-top: 20px;
            }

            .loading-card {
                height: 64px;
                border-radius: 10px;
            }

            .add-form {
                margin-bottom: 20px;
                padding: 16px;
                background: var(--fb-surface-1, #161821);
                border: 1px solid var(--fb-accent, #00FFA3);
                border-radius: 10px;
            }

            .add-form-title {
                font-size: 14px;
                font-weight: 600;
                color: var(--fb-text-primary, #F0F0F5);
                margin-bottom: 12px;
            }

            @media (max-width: 768px) {
                :host { padding: 16px 12px 80px; }
                h1 { font-size: 20px; }
            }
        `,
    ];

    @state() private _contacts: Contact[] = [];
    @state() private _loading = true;
    @state() private _error: string | null = null;
    @state() private _search = '';
    @state() private _showAddForm = false;
    private _searchTimeout = 0;

    override connectedCallback() {
        super.connectedCallback();
        this._loadContacts();
    }

    override onViewActive(): void {
        this._loadContacts();
    }

    private async _loadContacts() {
        this._loading = true;
        this._error = null;
        try {
            this._contacts = await api.contacts.list(this._search || undefined);
        } catch (err) {
            console.warn('[FBViewDirectory] Failed to load contacts from API, using mock service', err);
            try {
                // Use static import for reliability in demo
                this._contacts = await mockContactsService.list(this._search || undefined);
            } catch (mockErr) {
                this._contacts = [];
                this._error = err instanceof Error ? err.message : 'Failed to load contacts';
            }
        } finally {
            this._loading = false;
        }
    }

    private _handleSearchInput(e: InputEvent) {
        this._search = (e.target as HTMLInputElement).value;
        clearTimeout(this._searchTimeout);
        this._searchTimeout = window.setTimeout(() => {
            this._loadContacts();
        }, 300);
    }

    private _handleAddClick() {
        this._showAddForm = !this._showAddForm;
    }

    private _handleContactSaved() {
        this._showAddForm = false;
        this._loadContacts();
    }

    private _getInitials(name: string): string {
        const parts = name.trim().split(/\s+/);
        if (parts.length >= 2) {
            return (parts[0]![0]! + parts[parts.length - 1]![0]!).toUpperCase();
        }
        return name.substring(0, 2).toUpperCase();
    }

    override render(): TemplateResult {
        return html`
            <div class="header">
                <h1>Contacts</h1>
                <button class="btn-add" @click=${this._handleAddClick}>
                    ${this._showAddForm ? 'Cancel' : '+ Add Contact'}
                </button>
            </div>

            ${this._showAddForm
                ? html`
                      <div class="add-form">
                          <div class="add-form-title">New Contact</div>
                          <fb-contact-inline-add
                              @fb-contact-saved=${this._handleContactSaved}
                              @fb-contact-cancelled=${() => { this._showAddForm = false; }}
                          ></fb-contact-inline-add>
                      </div>
                  `
                : nothing}

            <div class="search-bar">
                <input
                    class="search-input"
                    type="text"
                    placeholder="Search contacts by name, phone, or email..."
                    .value=${this._search}
                    @input=${this._handleSearchInput}
                />
            </div>

            ${this._loading
                ? this._renderLoading()
                : this._error
                    ? html`
                        <div class="empty" role="alert">
                            <div class="empty-title" style="color: #F43F5E;">Something went wrong</div>
                            <div>${this._error}</div>
                            <button
                                style="margin-top: 12px; padding: 8px 20px; border-radius: 6px; border: 1px solid #F43F5E; background: transparent; color: #F43F5E; font-size: 13px; font-weight: 600; cursor: pointer;"
                                @click=${() => this._loadContacts()}
                            >Retry</button>
                        </div>
                    `
                    : this._renderContacts()}
        `;
    }

    private _renderLoading(): TemplateResult {
        return html`
            <div class="loading">
                <div class="loading-card skeleton"></div>
                <div class="loading-card skeleton"></div>
                <div class="loading-card skeleton"></div>
            </div>
        `;
    }

    private _renderContacts(): TemplateResult {
        if (this._contacts.length === 0) {
            return html`
                <div class="empty">
                    <div class="empty-title">${this._search ? 'No matches' : 'No contacts yet'}</div>
                    <div>${this._search ? 'Try a different search term.' : 'Add your subs and vendors to get started.'}</div>
                </div>
            `;
        }

        return html`
            <div class="count">${this._contacts.length} contact${this._contacts.length !== 1 ? 's' : ''}</div>
            <div class="contact-list">
                ${this._contacts.map(
            (c) => html`
                        <div class="contact-card">
                            <div class="contact-avatar">${this._getInitials(c.name)}</div>
                            <div class="contact-info">
                                <div class="contact-name">${c.name}${c.company ? ` \u2014 ${c.company}` : ''}</div>
                                <div class="contact-detail">
                                    ${c.phone ?? ''}${c.phone && c.email ? ' \u00B7 ' : ''}${c.email ?? ''}
                                </div>
                            </div>
                            <div class="contact-badges">
                                <span class="badge badge-pref">${c.contact_preference}</span>
                                ${c.portal_enabled
                    ? html`<span class="badge badge-portal">Portal</span>`
                    : nothing}
                            </div>
                        </div>
                    `
        )}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-directory': FBViewDirectory;
    }
}
