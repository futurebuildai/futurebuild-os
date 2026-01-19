/**
 * FBHeader - Context Header Component
 * See FRONTEND_SCOPE.md Section 3.3, PRODUCTION_PLAN.md Step 51.3
 *
 * Minimal header (56px) for the Command Center layout.
 * - Left: Breadcrumbs/Context
 * - Right: Status indicator
 * - Mobile: Includes user profile and theme toggle
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';
import type { Theme, User } from '../../store/types';

/**
 * Context Header - Minimal top bar
 * @element fb-header
 */
@customElement('fb-header')
export class FBHeader extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: flex;
                align-items: center;
                justify-content: space-between;
                height: var(--fb-header-height);
                padding: 0 var(--fb-spacing-lg);
                background: var(--fb-bg-secondary);
                border-bottom: 1px solid var(--fb-border);
                grid-area: header;
                z-index: var(--fb-z-header);
            }

            .left {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm);
            }

            .breadcrumb {
                font-size: var(--fb-text-sm);
                color: var(--fb-text-secondary);
            }

            .breadcrumb-separator {
                color: var(--fb-text-muted);
                margin: 0 var(--fb-spacing-xs);
            }

            .breadcrumb-active {
                color: var(--fb-text-primary);
                font-weight: 500;
            }

            .right {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-md);
            }

            .status {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-xs);
                font-size: var(--fb-text-xs);
                color: var(--fb-text-muted);
            }

            .status-dot {
                width: 8px;
                height: 8px;
                border-radius: var(--fb-radius-full);
                background: var(--fb-success);
            }

            .status-dot.offline {
                background: var(--fb-error);
            }

            /* Mobile controls (hidden on desktop) */
            .mobile-controls {
                display: none;
                align-items: center;
                gap: var(--fb-spacing-sm);
            }

            .avatar {
                width: 32px;
                height: 32px;
                border-radius: var(--fb-radius-full);
                background: var(--fb-dawn-gradient);
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: var(--fb-text-xs);
                font-weight: 600;
                color: var(--fb-text-primary);
            }

            .theme-toggle {
                width: 32px;
                height: 32px;
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
                width: 18px;
                height: 18px;
                fill: none;
                stroke: currentColor;
                stroke-width: 2;
            }

            @media (max-width: 768px) {
                :host {
                    padding: 0 var(--fb-spacing-md);
                }

                .mobile-controls {
                    display: flex;
                }

                .status {
                    display: none;
                }
            }
        `,
    ];

    @state() private activeView: string = 'chat';
    @state() private theme: Theme = 'system';
    @state() private userInitials: string = '';
    @state() private isOnline: boolean = true;

    private _disposeEffect: (() => void) | null = null;

    override connectedCallback(): void {
        super.connectedCallback();

        // Subscribe to store signals
        this._disposeEffect = effect(() => {
            this.activeView = store.activeView$.value;
            this.theme = store.theme$.value;
            this.userInitials = this._computeInitials(store.user$.value);
        });

        // Online/offline detection
        window.addEventListener('online', this._handleOnline);
        window.addEventListener('offline', this._handleOffline);
        this.isOnline = navigator.onLine;
    }

    override disconnectedCallback(): void {
        this._disposeEffect?.();
        this._disposeEffect = null;
        window.removeEventListener('online', this._handleOnline);
        window.removeEventListener('offline', this._handleOffline);
        super.disconnectedCallback();
    }

    private _handleOnline = (): void => {
        this.isOnline = true;
    };

    private _handleOffline = (): void => {
        this.isOnline = false;
    };

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

    private _handleThemeToggle(): void {
        const next: Theme = this.theme === 'dark' ? 'light' : 'dark';
        store.actions.setTheme(next);
    }

    private _formatViewName(view: string): string {
        return view.charAt(0).toUpperCase() + view.slice(1);
    }

    private _renderSunIcon(): TemplateResult {
        return html`<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="5"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></svg>`;
    }

    private _renderMoonIcon(): TemplateResult {
        return html`<svg viewBox="0 0 24 24"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>`;
    }

    override render(): TemplateResult {
        const isDark = this.theme === 'dark' || this.theme === 'system';

        return html`
            <div class="left">
                <span class="breadcrumb">
                    FutureBuild
                    <span class="breadcrumb-separator">›</span>
                    <span class="breadcrumb-active">${this._formatViewName(this.activeView)}</span>
                </span>
            </div>
            <div class="right">
                <div class="status">
                    <span class="status-dot ${this.isOnline ? '' : 'offline'}"></span>
                    <span>${this.isOnline ? 'Connected' : 'Offline'}</span>
                </div>
                <div class="mobile-controls">
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
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-header': FBHeader;
    }
}
