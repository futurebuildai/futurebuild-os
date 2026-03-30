/**
 * FBAdminSidebar - Platform Admin Navigation Sidebar
 * Left nav for the /admin/* shell with links to admin sections.
 */
import { html, css, type TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';

@customElement('fb-admin-sidebar')
export class FBAdminSidebar extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: flex;
                flex-direction: column;
                height: 100%;
                background: var(--fb-bg-panel, #0a0a0a);
                border-right: 1px solid var(--fb-border, rgba(255,255,255,0.05));
                color: var(--fb-text-primary, #fff);
            }

            .header {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm, 8px);
                padding: var(--fb-spacing-md, 16px);
                border-bottom: 1px solid var(--fb-border-light, #222);
            }

            .logo {
                width: 28px;
                height: 28px;
                color: var(--fb-text-primary, #fff);
            }

            .brand {
                font-size: var(--fb-text-md, 14px);
                font-weight: 600;
            }

            .brand span {
                font-weight: 300;
            }

            .badge {
                font-size: 10px;
                font-weight: 600;
                text-transform: uppercase;
                letter-spacing: 0.05em;
                color: var(--fb-warning, #f59e0b);
                background: rgba(245, 158, 11, 0.1);
                padding: 2px 6px;
                border-radius: 4px;
                margin-left: auto;
            }

            nav {
                flex: 1;
                padding: var(--fb-spacing-sm, 8px);
                display: flex;
                flex-direction: column;
                gap: 2px;
            }

            .section-label {
                padding: var(--fb-spacing-xs, 4px) var(--fb-spacing-sm, 8px);
                font-size: var(--fb-text-xs, 11px);
                font-weight: 600;
                text-transform: uppercase;
                letter-spacing: 0.05em;
                color: var(--fb-text-muted, #4A4B55);
                margin-top: var(--fb-spacing-sm, 8px);
            }

            .nav-item {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm, 8px);
                padding: var(--fb-spacing-sm, 8px) var(--fb-spacing-md, 16px);
                border-radius: var(--fb-radius-md, 8px);
                cursor: pointer;
                color: var(--fb-text-secondary, #8B8D98);
                font-size: var(--fb-text-sm, 13px);
                border: none;
                background: transparent;
                width: 100%;
                text-align: left;
                transition: background 0.15s ease, color 0.15s ease;
            }

            .nav-item:hover {
                background: var(--fb-bg-tertiary, #1a1a1a);
                color: var(--fb-text-primary, #fff);
            }

            .nav-item:focus-visible {
                outline: 2px solid var(--fb-primary);
                outline-offset: -2px;
            }

            .nav-item.active {
                background: var(--fb-bg-tertiary, #1a1a1a);
                color: var(--fb-warning, #f59e0b);
            }

            .nav-icon {
                width: 16px;
                height: 16px;
                flex-shrink: 0;
            }

            .nav-icon svg {
                width: 100%;
                height: 100%;
                stroke: currentColor;
                fill: none;
                stroke-width: 2;
            }

            .spacer {
                flex: 1;
            }

            .back-link {
                margin-top: auto;
                border-top: 1px solid var(--fb-border-light, #222);
                padding: var(--fb-spacing-sm, 8px);
            }

            .back-link .nav-item {
                color: var(--fb-text-muted, #4A4B55);
            }

            .back-link .nav-item:hover {
                color: var(--fb-text-primary, #fff);
            }

            .user-footer {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm, 8px);
                padding: var(--fb-spacing-md, 16px);
                border-top: 1px solid var(--fb-border-light, #222);
            }

            .avatar {
                width: 28px;
                height: 28px;
                border-radius: 50%;
                background: var(--fb-dawn-gradient, linear-gradient(135deg, #00FFA3 0%, #00CC82 100%));
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 11px;
                font-weight: 600;
                color: white;
                flex-shrink: 0;
            }

            .user-name {
                font-size: var(--fb-text-xs, 11px);
                color: var(--fb-text-secondary, #8B8D98);
                white-space: nowrap;
                overflow: hidden;
                text-overflow: ellipsis;
            }
        `,
    ];

    @state() private _currentPath = '/admin';
    @state() private _userName = '';
    @state() private _userInitials = '';

    private _disposeEffects: (() => void)[] = [];

    private _handlePopState = (): void => {
        this._currentPath = window.location.pathname;
    };

    override connectedCallback(): void {
        super.connectedCallback();
        this._currentPath = window.location.pathname;
        window.addEventListener('popstate', this._handlePopState);

        this._disposeEffects.push(
            effect(() => {
                const user = store.user$.value;
                this._userName = user?.name ?? '';
                this._userInitials = this._computeInitials(user?.name);
            })
        );
    }

    override disconnectedCallback(): void {
        window.removeEventListener('popstate', this._handlePopState);
        this._disposeEffects.forEach((d) => { d(); });
        this._disposeEffects = [];
        super.disconnectedCallback();
    }

    private _computeInitials(name?: string): string {
        if (!name) return '?';
        const parts = name.split(' ');
        if (parts.length >= 2) {
            const first = parts[0]?.[0] ?? '';
            const second = parts[1]?.[0] ?? '';
            return `${first}${second}`.toUpperCase();
        }
        return name.substring(0, 2).toUpperCase();
    }

    private _navigate(path: string): void {
        window.history.pushState({}, '', path);
        window.dispatchEvent(new PopStateEvent('popstate'));
    }

    private _isActive(path: string): boolean {
        return this._currentPath === path;
    }

    override render(): TemplateResult {
        return html`
            <div class="header">
                <svg class="logo" viewBox="0 0 100 100" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <path d="M10 45 L50 15 L90 45"/>
                    <path d="M20 50 L50 25 L80 50"/>
                    <path d="M25 48 L25 85 L75 85 L75 48"/>
                    <path d="M50 50 L40 56 L40 68 L50 74 L60 68 L60 56 Z"/>
                </svg>
                <div class="brand">FUTURE<span>BUILD</span></div>
                <span class="badge">Admin</span>
            </div>

            <nav aria-label="Admin navigation">
                <div class="section-label">Platform</div>
                <button
                    class="nav-item ${this._isActive('/admin') ? 'active' : ''}"
                    @click=${(): void => { this._navigate('/admin'); }}
                    aria-label="Dashboard"
                >
                    <span class="nav-icon" aria-hidden="true">
                        <svg viewBox="0 0 24 24"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>
                    </span>
                    Dashboard
                </button>
                <button
                    class="nav-item ${this._isActive('/admin/invitations') ? 'active' : ''}"
                    @click=${(): void => { this._navigate('/admin/invitations'); }}
                    aria-label="Invitations"
                >
                    <span class="nav-icon" aria-hidden="true">
                        <svg viewBox="0 0 24 24"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="8.5" cy="7" r="4"/><path d="M20 8v6M23 11h-6"/></svg>
                    </span>
                    Invitations
                </button>
                <button
                    class="nav-item ${this._isActive('/admin/shadow') ? 'active' : ''}"
                    @click=${(): void => { this._navigate('/admin/shadow'); }}
                    aria-label="Shadow Mode"
                >
                    <span class="nav-icon" aria-hidden="true">
                        <svg viewBox="0 0 24 24"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
                    </span>
                    Shadow Mode
                </button>

                <div class="section-label">Intelligence</div>
                <button
                    class="nav-item ${this._isActive('/admin/agents') ? 'active' : ''}"
                    @click=${(): void => { this._navigate('/admin/agents'); }}
                    aria-label="Agent Settings"
                >
                    <span class="nav-icon" aria-hidden="true">
                        <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="3"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></svg>
                    </span>
                    Agent Settings
                </button>
                <button
                    class="nav-item ${this._isActive('/admin/brain') ? 'active' : ''}"
                    @click=${(): void => { this._navigate('/admin/brain'); }}
                    aria-label="FB-Brain"
                >
                    <span class="nav-icon" aria-hidden="true">
                        <svg viewBox="0 0 24 24"><path d="M9.5 2A2.5 2.5 0 0 1 12 4.5v15a2.5 2.5 0 0 1-4.96.44 2.5 2.5 0 0 1-2.96-3.08 3 3 0 0 1-.34-5.58 2.5 2.5 0 0 1 1.32-4.24 2.5 2.5 0 0 1 1.98-3A2.5 2.5 0 0 1 9.5 2Z"/><path d="M14.5 2A2.5 2.5 0 0 0 12 4.5v15a2.5 2.5 0 0 0 4.96.44 2.5 2.5 0 0 0 2.96-3.08 3 3 0 0 0 .34-5.58 2.5 2.5 0 0 0-1.32-4.24 2.5 2.5 0 0 0-1.98-3A2.5 2.5 0 0 0 14.5 2Z"/></svg>
                    </span>
                    FB-Brain
                </button>

            </nav>

            <div class="back-link">
                <button
                    class="nav-item"
                    @click=${(): void => { this._navigate('/'); }}
                    aria-label="Back to app"
                >
                    <span class="nav-icon" aria-hidden="true">
                        <svg viewBox="0 0 24 24"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
                    </span>
                    Back to App
                </button>
            </div>

            <footer class="user-footer">
                <div class="avatar" aria-hidden="true">${this._userInitials}</div>
                <span class="user-name">${this._userName || 'Admin'}</span>
            </footer>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-admin-sidebar': FBAdminSidebar;
    }
}
