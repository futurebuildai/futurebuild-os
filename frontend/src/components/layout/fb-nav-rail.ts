/**
 * FBNavRail - Navigation Rail Component
 * See FRONTEND_SCOPE.md Section 3.3, PRODUCTION_PLAN.md Step 51.3
 *
 * Slim vertical navigation (64px) styled after VS Code/Discord.
 * - Top: Navigation icons (Chat, Projects, Schedule, Directory)
 * - Bottom: User avatar, Theme toggle
 * - Mobile: Transforms to bottom navigation bar
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';
import type { Theme, User } from '../../store/types';

/** Navigation item definition */
interface NavItem {
    readonly id: string;
    readonly label: string;
    readonly icon: TemplateResult;
}



/**
 * Navigation Rail - Command Center style vertical nav
 * @element fb-nav-rail
 * @fires fb-navigate - When a nav item is clicked
 */
@customElement('fb-nav-rail')
export class FBNavRail extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: flex;
                flex-direction: column;
                width: var(--fb-rail-width);
                height: 100%;
                background: var(--fb-bg-rail);
                border-right: 1px solid var(--fb-border);
                grid-area: rail;
                z-index: var(--fb-z-rail);
            }

            .nav-top {
                flex: 1;
                display: flex;
                flex-direction: column;
                align-items: center;
                padding-top: var(--fb-spacing-md);
                gap: var(--fb-spacing-xs);
            }

            .nav-bottom {
                display: flex;
                flex-direction: column;
                align-items: center;
                padding-bottom: var(--fb-spacing-md);
                gap: var(--fb-spacing-sm);
            }

            .nav-item {
                position: relative;
                display: flex;
                align-items: center;
                justify-content: center;
                width: 48px;
                height: 48px;
                border: none;
                background: transparent;
                color: var(--fb-text-muted);
                cursor: pointer;
                border-radius: var(--fb-radius-md);
                transition: color var(--fb-transition-fast), background var(--fb-transition-fast);
            }

            .nav-item:hover {
                color: var(--fb-text-secondary);
                background: var(--fb-bg-tertiary);
            }

            .nav-item:focus-visible {
                outline: 2px solid var(--fb-primary);
                outline-offset: 2px;
            }

            .nav-item[aria-current="page"] {
                color: var(--fb-primary);
            }

            /* Active indicator pill */
            .nav-item[aria-current="page"]::before {
                content: '';
                position: absolute;
                left: 0;
                top: 50%;
                transform: translateY(-50%);
                width: 3px;
                height: 24px;
                background: var(--fb-primary);
                border-radius: 0 var(--fb-radius-sm) var(--fb-radius-sm) 0;
            }

            .nav-item svg {
                width: 24px;
                height: 24px;
                fill: none;
                stroke: currentColor;
                stroke-width: 2;
                stroke-linecap: round;
                stroke-linejoin: round;
            }

            .avatar {
                width: 36px;
                height: 36px;
                border-radius: var(--fb-radius-full);
                background: var(--fb-dawn-gradient);
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: var(--fb-text-sm);
                font-weight: 600;
                color: var(--fb-text-primary);
            }

            .theme-toggle {
                width: 40px;
                height: 40px;
                border: none;
                background: transparent;
                color: var(--fb-text-muted);
                cursor: pointer;
                border-radius: var(--fb-radius-md);
                display: flex;
                align-items: center;
                justify-content: center;
                transition: color var(--fb-transition-fast);
            }

            .theme-toggle:hover {
                color: var(--fb-text-secondary);
            }

            .theme-toggle:focus-visible {
                outline: 2px solid var(--fb-primary);
                outline-offset: 2px;
            }

            .theme-toggle svg {
                width: 20px;
                height: 20px;
                fill: none;
                stroke: currentColor;
                stroke-width: 2;
            }

            /* Mobile: Bottom navigation bar */
            @media (max-width: 768px) {
                :host {
                    flex-direction: row;
                    width: 100%;
                    height: auto;
                    border-right: none;
                    border-top: 1px solid var(--fb-border);
                    grid-area: nav;
                }

                .nav-top {
                    flex-direction: row;
                    justify-content: space-around;
                    padding: var(--fb-spacing-sm) 0;
                    gap: 0;
                }

                .nav-bottom {
                    display: none; /* Avatar/theme in header on mobile */
                }

                .nav-item[aria-current="page"]::before {
                    left: 50%;
                    top: 0;
                    transform: translateX(-50%);
                    width: 24px;
                    height: 3px;
                    border-radius: 0 0 var(--fb-radius-sm) var(--fb-radius-sm);
                }
            }
        `,
    ];

    /** Navigation items */
    private readonly navItems: readonly NavItem[] = [
        {
            id: 'chat',
            label: 'Chat',
            icon: html`<svg viewBox="0 0 24 24"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>`,
        },
        {
            id: 'projects',
            label: 'Projects',
            icon: html`<svg viewBox="0 0 24 24"><path d="M3 3h18v18H3zM3 9h18M9 21V9"/></svg>`,
        },
        {
            id: 'schedule',
            label: 'Schedule',
            icon: html`<svg viewBox="0 0 24 24"><rect x="3" y="4" width="18" height="18" rx="2"/><path d="M16 2v4M8 2v4M3 10h18"/></svg>`,
        },
        {
            id: 'directory',
            label: 'Directory',
            icon: html`<svg viewBox="0 0 24 24"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75"/></svg>`,
        },
    ];

    @state() private activeView: string = 'chat';
    @state() private theme: Theme = 'system';
    @state() private userInitials: string = '';

    private _disposeEffect: (() => void) | null = null;

    override connectedCallback(): void {
        super.connectedCallback();
        // Subscribe to store signals
        this._disposeEffect = effect(() => {
            this.activeView = store.activeView$.value;
            this.theme = store.theme$.value;
            this.userInitials = this._computeInitials(store.user$.value);
        });
    }

    override disconnectedCallback(): void {
        this._disposeEffect?.();
        this._disposeEffect = null;
        super.disconnectedCallback();
    }

    private _computeInitials(user: User | null): string {
        if (!user?.name) return '?';
        const parts = user.name.split(' ');
        const first = parts[0];
        const second = parts[1];
        if (first !== undefined && second !== undefined && first.length > 0 && second.length > 0) {
            return `${String(first[0])}${String(second[0])}`.toUpperCase();
        }
        return user.name.substring(0, 2).toUpperCase();
    }

    private _handleNavClick(viewId: string): void {
        store.actions.setActiveView(viewId);
        this.emit('fb-navigate', { view: viewId });
    }

    private _handleThemeToggle(): void {
        const next: Theme = this.theme === 'dark' ? 'light' : 'dark';
        store.actions.setTheme(next);
    }

    private _renderSunIcon(): TemplateResult {
        return html`<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="5"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></svg>`;
    }

    private _renderMoonIcon(): TemplateResult {
        return html`<svg viewBox="0 0 24 24"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>`;
    }

    override render(): TemplateResult {
        const isDark = this.theme === 'dark' || (this.theme === 'system');

        return html`
            <nav class="nav-top" role="navigation" aria-label="Main navigation">
                ${this.navItems.map(
            (item) => html`
                        <button
                            class="nav-item"
                            aria-label="${item.label}"
                            aria-current="${this.activeView === item.id ? 'page' : 'false'}"
                            @click=${(): void => { this._handleNavClick(item.id); }}
                        >
                            ${item.icon}
                        </button>
                    `
        )}       </nav>
            <div class="nav-bottom">
                <div class="avatar" title="User Profile">${this.userInitials}</div>
                <button
                    class="theme-toggle"
                    aria-label="Toggle theme"
                    aria-pressed="${isDark}"
                    @click=${this._handleThemeToggle.bind(this)}
                >
                    ${isDark ? this._renderSunIcon() : this._renderMoonIcon()}
                </button>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-nav-rail': FBNavRail;
    }
}
