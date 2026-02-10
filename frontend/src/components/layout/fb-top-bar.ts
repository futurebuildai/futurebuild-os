/**
 * fb-top-bar — V2 Top navigation bar.
 * See FRONTEND_V2_SPEC.md §4.2
 *
 * Renders: [Logo] [All] [Proj1] [Proj2] [+New] ... [Bell] [Avatar→Menu]
 * Project pills filter the feed. "All" shows cross-project feed.
 * Avatar click opens a role-gated dropdown menu.
 */
import { html, css, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import type { ProjectPill } from '../../types/feed';

@customElement('fb-top-bar')
export class FBTopBar extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                height: 56px;
                background: var(--fb-surface-1, #1a1a2e);
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
                padding: 0 16px;
                z-index: 10;
            }

            .bar {
                display: flex;
                align-items: center;
                height: 100%;
                gap: 12px;
            }

            .logo {
                font-size: 18px;
                font-weight: 700;
                color: var(--fb-text-primary, #e0e0e0);
                white-space: nowrap;
                cursor: pointer;
                letter-spacing: -0.3px;
            }

            .logo span {
                color: var(--fb-accent, #6366f1);
            }

            .pills {
                display: flex;
                align-items: center;
                gap: 6px;
                overflow-x: auto;
                flex: 1;
                scrollbar-width: none;
                padding: 0 8px;
            }

            .pills::-webkit-scrollbar {
                display: none;
            }

            .pill {
                display: inline-flex;
                align-items: center;
                padding: 6px 14px;
                border-radius: 20px;
                font-size: 13px;
                font-weight: 500;
                white-space: nowrap;
                cursor: pointer;
                border: 1px solid var(--fb-border, #2a2a3e);
                background: transparent;
                color: var(--fb-text-secondary, #a0a0b0);
                transition: all 0.15s ease;
            }

            .pill:hover {
                background: var(--fb-surface-2, #252540);
                color: var(--fb-text-primary, #e0e0e0);
            }

            .pill[data-active] {
                background: var(--fb-accent, #6366f1);
                color: #fff;
                border-color: var(--fb-accent, #6366f1);
            }

            .pill-new {
                border-style: dashed;
                color: var(--fb-text-tertiary, #707080);
            }

            .pill-new:hover {
                border-color: var(--fb-accent, #6366f1);
                color: var(--fb-accent, #6366f1);
            }

            .actions {
                display: flex;
                align-items: center;
                gap: 8px;
                margin-left: auto;
            }

            .avatar-wrap {
                position: relative;
            }

            .avatar {
                width: 32px;
                height: 32px;
                border-radius: 50%;
                background: var(--fb-accent, #6366f1);
                color: #fff;
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 13px;
                font-weight: 600;
                cursor: pointer;
                border: none;
            }

            .avatar:hover {
                opacity: 0.85;
            }

            /* User menu dropdown */
            .user-menu {
                position: absolute;
                top: 40px;
                right: 0;
                min-width: 200px;
                background: var(--fb-surface-1, #1a1a2e);
                border: 1px solid var(--fb-border, #2a2a3e);
                border-radius: 8px;
                box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
                z-index: 100;
                overflow: hidden;
            }

            .menu-header {
                padding: 12px 16px;
                border-bottom: 1px solid var(--fb-border, #2a2a3e);
            }

            .menu-name {
                font-size: 14px;
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .menu-role {
                font-size: 12px;
                color: var(--fb-text-tertiary, #707080);
                margin-top: 2px;
            }

            .menu-items {
                padding: 4px 0;
            }

            .menu-item {
                display: block;
                width: 100%;
                padding: 10px 16px;
                font-size: 14px;
                color: var(--fb-text-secondary, #a0a0b0);
                background: none;
                border: none;
                text-align: left;
                cursor: pointer;
                transition: background 0.1s ease;
            }

            .menu-item:hover {
                background: var(--fb-surface-2, #252540);
                color: var(--fb-text-primary, #e0e0e0);
            }

            .menu-divider {
                height: 1px;
                background: var(--fb-border, #2a2a3e);
                margin: 4px 0;
            }

            .menu-item--danger {
                color: #ef4444;
            }

            .menu-item--danger:hover {
                background: rgba(239, 68, 68, 0.1);
                color: #ef4444;
            }

            @media (max-width: 768px) {
                :host {
                    padding: 0 12px;
                }

                .logo {
                    font-size: 16px;
                }

                .pill {
                    font-size: 12px;
                    padding: 5px 10px;
                }
            }
        `,
    ];

    @property({ type: Array }) projects: ProjectPill[] = [];
    @property({ type: String, attribute: 'active-filter' }) activeFilter: string | null = null;
    @property({ type: String, attribute: 'user-name' }) userName = '';
    @property({ type: String, attribute: 'user-role' }) userRole = '';

    @state() private _menuOpen = false;

    override connectedCallback() {
        super.connectedCallback();
        document.addEventListener('click', this._handleOutsideClick);
    }

    override disconnectedCallback() {
        super.disconnectedCallback();
        document.removeEventListener('click', this._handleOutsideClick);
    }

    private _handleOutsideClick = (e: Event) => {
        if (this._menuOpen && !this.contains(e.target as Node)) {
            this._menuOpen = false;
        }
    };

    private _handlePillClick(projectId: string | null) {
        this.emit('fb-filter-change', { projectId });
    }

    private _handleNewProject() {
        this.emit('fb-navigate', { view: 'onboard' });
    }

    private _handleLogoClick() {
        this.emit('fb-navigate', { view: 'home' });
    }

    private _toggleMenu() {
        this._menuOpen = !this._menuOpen;
    }

    private _navigate(view: string) {
        this._menuOpen = false;
        this.emit('fb-navigate', { view });
    }

    private _handleSignOut() {
        this._menuOpen = false;
        this.emit('fb-sign-out', {});
    }

    private _getInitials(): string {
        if (!this.userName) return '?';
        const parts = this.userName.trim().split(/\s+/);
        if (parts.length >= 2) {
            return (parts[0]![0]! + parts[parts.length - 1]![0]!).toUpperCase();
        }
        return this.userName.substring(0, 2).toUpperCase();
    }

    /** Check if user has admin/builder role for write-gated menu items */
    private _isAdminOrBuilder(): boolean {
        return this.userRole === 'Admin' || this.userRole === 'Builder';
    }

    private _isAdmin(): boolean {
        return this.userRole === 'Admin';
    }

    override render() {
        return html`
            <div class="bar">
                <div class="logo" @click=${this._handleLogoClick}>
                    Future<span>Build</span>
                </div>

                <div class="pills">
                    <button
                        class="pill"
                        ?data-active=${this.activeFilter === null}
                        @click=${() => this._handlePillClick(null)}
                    >
                        All
                    </button>

                    ${this.projects.map(
                        (p) => html`
                            <button
                                class="pill"
                                ?data-active=${this.activeFilter === p.id}
                                @click=${() => this._handlePillClick(p.id)}
                                title=${p.address}
                            >
                                ${p.name}
                            </button>
                        `
                    )}

                    <button class="pill pill-new" @click=${this._handleNewProject}>
                        + New
                    </button>
                </div>

                <div class="actions">
                    <div class="avatar-wrap">
                        <button
                            class="avatar"
                            @click=${this._toggleMenu}
                            title="Account menu"
                            aria-haspopup="true"
                            aria-expanded=${this._menuOpen}
                        >
                            ${this._getInitials()}
                        </button>

                        ${this._menuOpen ? this._renderMenu() : nothing}
                    </div>
                </div>
            </div>
        `;
    }

    private _renderMenu() {
        return html`
            <div class="user-menu" role="menu">
                <div class="menu-header">
                    <div class="menu-name">${this.userName || 'User'}</div>
                    <div class="menu-role">${this.userRole || 'Member'}</div>
                </div>
                <div class="menu-items">
                    <button class="menu-item" role="menuitem" @click=${() => this._navigate('settings-profile')}>
                        My Profile
                    </button>
                    <button class="menu-item" role="menuitem" @click=${() => this._navigate('contacts')}>
                        Contacts
                    </button>
                    ${this._isAdminOrBuilder()
                        ? html`
                              <button class="menu-item" role="menuitem" @click=${() => this._navigate('settings-org')}>
                                  Organization
                              </button>
                          `
                        : nothing}
                    ${this._isAdmin()
                        ? html`
                              <button class="menu-item" role="menuitem" @click=${() => this._navigate('settings-team')}>
                                  Team & Invites
                              </button>
                          `
                        : nothing}
                    <div class="menu-divider"></div>
                    <button class="menu-item menu-item--danger" role="menuitem" @click=${this._handleSignOut}>
                        Sign Out
                    </button>
                </div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-top-bar': FBTopBar;
    }
}
