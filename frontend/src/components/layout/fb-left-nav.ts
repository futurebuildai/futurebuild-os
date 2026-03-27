/**
 * fb-left-nav — Sticky left navigation sidebar.
 * Refactor Step 6.
 *
 * Context-aware navigation:
 * - Global context (no project selected): Shows global views (Feed, Schedule, Budget, Chat, Contacts).
 * - Project context: Shows project-specific views.
 */
import { html, css, nothing } from 'lit';
import { customElement, state, property } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';
import type { ContextScope } from '../../store/types';

@customElement('fb-left-nav')
export class FBLeftNav extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                width: 64px; /* Icon-only width */
                height: 100%;
                /* Glassmorphism panel */
                background: rgba(10, 11, 16, 0.8);
                backdrop-filter: blur(48px);
                -webkit-backdrop-filter: blur(48px);
                display: flex;
                flex-direction: column;
                align-items: center; /* Center icons */
                padding: 16px 0;
                font-family: var(--fb-font-family);
                z-index: 20; /* Ensure on top for tooltips if needed */
            }

            .nav-group {
                display: flex;
                flex-direction: column;
                gap: 8px;
                width: 100%;
                align-items: center;
            }

            .nav-item {
                display: flex;
                align-items: center;
                justify-content: center;
                width: 48px;
                height: 48px;
                padding: 0;
                border-radius: 12px;
                color: var(--fb-text-secondary, #8B8D98);
                text-decoration: none;
                transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
                cursor: pointer;
                border: none;
                background: transparent;
                position: relative;
            }

            /* Tooltip on hover */
            .nav-item::after {
                content: attr(aria-label);
                position: absolute;
                left: 60px;
                background: var(--fb-surface-2, #1E2029);
                color: var(--fb-text-primary, #F0F0F5);
                padding: 4px 8px;
                border-radius: 4px;
                font-size: 12px;
                white-space: nowrap;
                opacity: 0;
                pointer-events: none;
                transition: opacity 0.2s ease;
                box-shadow: 0 2px 8px rgba(0,0,0,0.2);
                z-index: 100;
            }

            .nav-item:hover::after {
                opacity: 1;
            }

            /* Minimal hover effect */
            .nav-item:hover {
                background: rgba(255, 255, 255, 0.05);
                color: var(--fb-text-primary, #F0F0F5);
                transform: scale(1.05);
            }

            /* Active state with green indicator bar */
            .nav-item.active {
                background: rgba(0, 255, 163, 0.1);
                color: #00FFA3;
                border: 1px solid rgba(0, 255, 163, 0.15);
                box-shadow: 0 0 12px rgba(0, 255, 163, 0.15);
            }

            .nav-item.active::before {
                content: '';
                position: absolute;
                left: -8px;
                top: 50%;
                transform: translateY(-50%);
                width: 3px;
                height: 60%;
                background: #00FFA3;
                border-radius: 0 3px 3px 0;
            }

            .nav-item svg {
                width: 24px;
                height: 24px;
                flex-shrink: 0;
            }
            
            .context-label {
                display: none; /* Hidden in icon-only mode; can be shown in expanded nav */
            }

            /* Context indicator dot */
            .context-dot {
                width: 6px;
                height: 6px;
                border-radius: 50%;
                margin-top: 4px;
                transition: background 0.2s ease;
            }

            .context-dot.global {
                background: var(--fb-accent, #00FFA3);
            }

            .context-dot.project {
                background: #00FFA3;
            }

            /* Logo styles */
            .logo-container {
                margin-bottom: 24px;
                cursor: pointer;
            }

            .logo-icon {
                width: 40px;
                height: 40px;
                background: linear-gradient(135deg, #00FFA3 0%, #00CC82 100%);
                border-radius: 10px;
                display: flex;
                align-items: center;
                justify-content: center;
                color: white;
                box-shadow: 0 4px 12px rgba(0, 255, 163, 0.3);
                transition: transform 0.2s ease;
            }

            .logo-icon:hover {
                transform: scale(1.05);
            }

            .logo-svg {
                width: 24px;
                height: 24px;
            }

            /* Spacer to push user menu to bottom */
            .spacer {
                flex: 1;
            }

            /* User Avatar styles */
            .user-section {
                margin-top: auto;
                position: relative;
                padding-bottom: 16px;
                display: flex;
                justify-content: center;
                width: 100%;
            }

            .avatar {
                width: 44px;
                height: 44px;
                border-radius: 50%;
                background: var(--fb-accent, #00FFA3);
                color: #fff;
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 14px;
                font-weight: 600;
                cursor: pointer;
                border: 2px solid transparent;
                transition: all 0.2s ease;
            }

            .avatar:hover {
                border-color: rgba(255, 255, 255, 0.2);
                transform: scale(1.05);
            }

            /* User menu dropdown - glassmorphism */
            .user-menu {
                position: absolute;
                bottom: 12px;
                left: 56px; /* Pop out to the right */
                width: 200px;
                background: rgba(22, 24, 33, 0.85);
                backdrop-filter: blur(24px);
                -webkit-backdrop-filter: blur(24px);
                border: 1px solid rgba(255, 255, 255, 0.08);
                border-radius: 12px;
                box-shadow: 0 4px 24px rgba(0, 0, 0, 0.4);
                z-index: 100;
                overflow: hidden;
                animation: fadeIn 0.15s ease;
            }

            .menu-header {
                padding: 12px 16px;
                border-bottom: 1px solid var(--fb-border, rgba(255,255,255,0.05));
            }

            .menu-name {
                font-size: 14px;
                font-weight: 600;
                color: var(--fb-text-primary, #F0F0F5);
            }

            .menu-role {
                font-size: 12px;
                color: var(--fb-text-tertiary, #5A5B66);
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
                color: var(--fb-text-secondary, #8B8D98);
                background: none;
                border: none;
                text-align: left;
                cursor: pointer;
                transition: background 0.1s ease;
            }

            .menu-item:hover {
                background: var(--fb-surface-2, #1E2029);
                color: var(--fb-text-primary, #F0F0F5);
            }

            .menu-divider {
                height: 1px;
                background: var(--fb-border, rgba(255,255,255,0.05));
                margin: 4px 0;
            }

            .menu-item--danger {
                color: #F43F5E;
            }

            .menu-item--danger:hover {
                background: rgba(239, 68, 68, 0.1);
                color: #F43F5E;
            }
            
            @keyframes fadeIn {
                from { opacity: 0; transform: translateX(-10px); }
                to { opacity: 1; transform: translateX(0); }
            }

            @media(max-width: 768px) {
                :host {
                    display: none;
                }
            }
        `,
    ];

    @state() private _projectId: string | null = null;
    @state() private _scope: ContextScope = 'global';
    @state() private _currentView = 'home';
    @property({ type: String, attribute: 'user-name' }) userName = '';
    @property({ type: String, attribute: 'user-role' }) userRole = '';

    @state() private _menuOpen = false;

    private _disposeEffects: (() => void)[] = [];

    override connectedCallback() {
        super.connectedCallback();

        this._disposeEffects.push(
            effect(() => {
                const ctx = store.contextState$.value;
                this._scope = ctx.scope;
                this._projectId = ctx.projectId;
            })
        );

        // Listen for route changes
        this._syncRoute();
        window.addEventListener('popstate', this._syncRoute);
        window.addEventListener('fb-route-change', this._syncRoute);

        // Close menu on outside click
        document.addEventListener('click', this._handleOutsideClick);
    }

    override disconnectedCallback() {
        window.removeEventListener('popstate', this._syncRoute);
        window.removeEventListener('fb-route-change', this._syncRoute);
        document.removeEventListener('click', this._handleOutsideClick);

        this._disposeEffects.forEach(dispose => dispose());
        super.disconnectedCallback();
    }

    private _handleOutsideClick = (e: Event) => {
        // If clicking outside component, close menu
        // Ideally we'd check if path includes this component, but sidebar isolates the click.
        // We can check if the target is inside the user-section
        const path = e.composedPath();
        const isInside = path.some(node => node === this.shadowRoot?.querySelector('.user-section'));

        if (this._menuOpen && !isInside) {
            this._menuOpen = false;
        }
    };

    private _syncRoute = () => {
        const path = window.location.pathname;
        if (path === '/' || path.match(/^\/project\/[^/]+$/)) {
            this._currentView = 'feed';
        } else if (path.includes('/schedule')) {
            this._currentView = 'schedule';
        } else if (path.includes('/budget')) {
            this._currentView = 'budget';
        } else if (path.includes('/chat')) {
            this._currentView = 'chat';
        } else if (path.includes('/contacts')) {
            this._currentView = 'contacts';
        } else {
            this._currentView = 'other';
        }
    };

    private _navigate(view: string) {
        const pid = this._projectId;

        // Sprint 1.2: Branch on explicit scope, not projectId truthiness
        if (this._scope === 'project' && pid) {
            switch (view) {
                case 'feed': this.emit('fb-navigate', { view: 'project', projectId: pid }); break;
                case 'schedule': this.emit('fb-navigate', { view: 'project-schedule', projectId: pid }); break;
                case 'chat': this.emit('fb-navigate', { view: 'project-chat', projectId: pid }); break;
                case 'budget': this.emit('fb-navigate', { view: 'project-budget', projectId: pid }); break;
                case 'contacts': this.emit('fb-navigate', { view: 'project-contacts', projectId: pid }); break;
                default: this.emit('fb-navigate', { view: 'project', projectId: pid });
            }
        } else {
            switch (view) {
                case 'feed': this.emit('fb-navigate', { view: 'home' }); break;
                case 'contacts': this.emit('fb-navigate', { view: 'contacts' }); break;
                case 'schedule': this.emit('fb-navigate', { view: 'schedule' }); break;
                case 'budget': this.emit('fb-navigate', { view: 'budget' }); break;
                case 'chat': this.emit('fb-navigate', { view: 'chat' }); break;
                default: console.warn('Global view not implemented:', view);
            }
        }
    }

    private _toggleMenu(e: Event) {
        e.stopPropagation();
        this._menuOpen = !this._menuOpen;
    }

    private _handleSignOut() {
        this._menuOpen = false;
        this.emit('fb-sign-out', {});
    }

    private _toggleTheme() {
        this._menuOpen = false;
        this.emit('fb-toggle-theme', {});
    }

    private _getInitials(): string {
        if (!this.userName) return '?';
        const parts = this.userName.trim().split(/\s+/);
        if (parts.length >= 2) {
            return (parts[0]![0]! + parts[parts.length - 1]![0]!).toUpperCase();
        }
        return this.userName.substring(0, 2).toUpperCase();
    }

    private _goHome() {
        this.emit('fb-navigate', { view: 'home' });
    }

    override render() {
        return html`
            <div class="logo-container" @click=${this._goHome} title="Home">
                <div class="logo-icon">
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" fill="none" class="logo-svg">
                        <g stroke="currentColor" stroke-width="8" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M10 45 L50 15 L90 45"/>
                            <path d="M20 50 L50 25 L80 50"/>
                            <path d="M25 48 L25 85"/>
                            <path d="M75 48 L75 85"/>
                            <path d="M25 85 L75 85"/>
                        </g>
                        <!-- Simplified Hexagon for Icon Scale -->
                        <path d="M50 50 L40 56 L40 68 L50 74 L60 68 L60 56 Z" fill="currentColor"/>
                    </svg>
                </div>
                <div class="context-dot ${this._scope}"></div>
            </div>

            <div class="nav-group">
                <button 
                    class="nav-item ${this._currentView === 'feed' ? 'active' : ''}"
                    @click=${() => this._navigate('feed')}
                    aria-label=${this._scope === 'project' ? 'Project Focus' : 'Daily Focus'}
                    title=${this._scope === 'project' ? 'Project Focus' : 'Daily Focus'}
                >
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
                        <polyline points="9 22 9 12 15 12 15 22"/>
                    </svg>
                </button>

                <button 
                    class="nav-item ${this._currentView === 'chat' ? 'active' : ''}"
                    @click=${() => this._navigate('chat')}
                    aria-label="Chat"
                    title="Chat"
                >
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
                    </svg>
                </button>

                <button 
                    class="nav-item ${this._currentView === 'schedule' ? 'active' : ''}"
                    @click=${() => this._navigate('schedule')}
                    aria-label=${this._scope === 'project' ? 'Project Schedule' : 'All Schedules'}
                    title=${this._scope === 'project' ? 'Project Schedule' : 'All Schedules'}
                >
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
                        <line x1="16" y1="2" x2="16" y2="6"/>
                        <line x1="8" y1="2" x2="8" y2="6"/>
                        <line x1="3" y1="10" x2="21" y2="10"/>
                    </svg>
                </button>

                ${this.userRole !== 'Subcontractor' && this.userRole !== 'Viewer' ? html`
                <button 
                    class="nav-item ${this._currentView === 'budget' ? 'active' : ''}"
                    @click=${() => this._navigate('budget')}
                    aria-label=${this._scope === 'project' ? 'Project Budget' : 'Company Budget'}
                    title=${this._scope === 'project' ? 'Project Budget' : 'Company Budget'}
                >
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <line x1="12" y1="1" x2="12" y2="23"/>
                        <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/>
                    </svg>
                </button>
                ` : nothing}


                <button 
                    class="nav-item ${this._currentView === 'contacts' ? 'active' : ''}"
                    @click=${() => this._navigate('contacts')}
                    aria-label="Contacts"
                    title="Contacts"
                >
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                        <circle cx="9" cy="7" r="4"/>
                        <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
                        <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
                    </svg>
                </button>
            </div>
            <div class="spacer"></div>
            
            <div class="user-section">
                <div 
                    class="avatar" 
                    @click=${this._toggleMenu}
                    title="Account Settings"
                >
                    ${this._getInitials()}
                </div>

                ${this._menuOpen ? html`
                    <div class="user-menu">
                         <div class="menu-header">
                            <div class="menu-name">${this.userName || 'User'}</div>
                            <div class="menu-role">${this.userRole || 'Member'}</div>
                        </div>
                        <div class="menu-items">
                            <button class="menu-item" @click=${() => { this._menuOpen = false; this._navigate('settings-profile'); }}>
                                Profile
                            </button>
                            <button class="menu-item" @click=${this._toggleTheme}>
                                Toggle Theme
                            </button>
                            <div class="menu-divider"></div>
                             <button class="menu-item menu-item--danger" @click=${this._handleSignOut}>
                                Sign Out
                            </button>
                        </div>
                    </div>
                ` : nothing}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-left-nav': FBLeftNav;
    }
}
