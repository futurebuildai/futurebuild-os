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
                color: var(--fb-text-primary, #e0e0e0);
                margin: 0;
            }

            .search-bar {
                margin-bottom: 20px;
            }

            .search-input {
                width: 100%;
                padding: 10px 14px;
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 8px;
                background: var(--fb-surface-1, #1a1a2e);
                color: var(--fb-text-primary, #e0e0e0);
                font-size: 14px;
                outline: none;
                box-sizing: border-box;
            }

            .search-input:focus {
                border-color: var(--fb-accent, #6366f1);
            }

            .search-input::placeholder {
                color: var(--fb-text-tertiary, #707080);
            }

            .btn-add {
                padding: 10px 20px;
                border-radius: 8px;
                background: var(--fb-accent, #6366f1);
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
                background: var(--fb-surface-1, #1a1a2e);
                border: 1px solid var(--fb-border, #2a2a3e);
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
                background: var(--fb-surface-2, #252540);
                color: var(--fb-text-secondary, #a0a0b0);
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
                color: var(--fb-text-primary, #e0e0e0);
            }

            .contact-detail {
                font-size: 13px;
                color: var(--fb-text-secondary, #a0a0b0);
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
                background: var(--fb-surface-2, #252540);
                color: var(--fb-text-tertiary, #707080);
            }

            .badge-portal {
                background: rgba(99, 102, 241, 0.15);
                color: var(--fb-accent, #6366f1);
            }

            .empty {
                text-align: center;
                padding: 60px 24px;
                color: var(--fb-text-secondary, #a0a0b0);
            }

            .empty-title {
                font-size: 18px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
                margin-bottom: 8px;
            }

            .count {
                font-size: 13px;
                color: var(--fb-text-tertiary, #707080);
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
                background: var(--fb-surface-1, #1a1a2e);
                border: 1px solid var(--fb-accent, #6366f1);
                border-radius: 10px;
            }

            .add-form-title {
                font-size: 14px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
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
    @state() private _search = '';
    @state() private _showAddForm = false;
    private _searchTimeout = 0;

    override onViewActive(): void {
        this._loadContacts();
    }

    private async _loadContacts() {
        this._loading = true;
        try {
            this._contacts = await api.contacts.list(this._search || undefined);
        } catch {
            this._contacts = [];
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

            ${this._loading ? this._renderLoading() : this._renderContacts()}
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
