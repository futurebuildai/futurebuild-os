/**
 * FBAppShell - Root Application Container
 * See FRONTEND_SCOPE.md Section 3.3, PRODUCTION_PLAN.md Step 51.3
 *
 * CSS Grid container for the Command Center layout.
 * - Desktop: Grid "rail header" / "rail stage"
 * - Mobile: Grid "header" / "stage" / "nav"
 * - Viewport locked: height 100vh, overflow hidden
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { store, initializeStore } from '../../store/store';

// Import child layout components
import './fb-nav-rail';
import './fb-header';

/**
 * Application Shell - Root layout container
 * @element fb-app-shell
 * @slot - Default slot for main stage content
 */
@customElement('fb-app-shell')
export class FBAppShell extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: grid;
                grid-template-columns: var(--fb-rail-width) 1fr;
                grid-template-rows: var(--fb-header-height) 1fr;
                grid-template-areas:
                    "rail header"
                    "rail stage";
                height: 100vh;
                width: 100vw;
                overflow: hidden;
                background: var(--fb-bg-primary);
                color: var(--fb-text-primary);
                font-family: var(--fb-font-family);
            }

            fb-nav-rail {
                grid-area: rail;
            }

            fb-header {
                grid-area: header;
            }

            .stage {
                grid-area: stage;
                overflow: auto;
                background: var(--fb-bg-stage);
            }

            /* Mobile layout */
            @media (max-width: 768px) {
                :host {
                    grid-template-columns: 1fr;
                    grid-template-rows: var(--fb-header-height) 1fr auto;
                    grid-template-areas:
                        "header"
                        "stage"
                        "nav";
                }

                fb-nav-rail {
                    grid-area: nav;
                }
            }

            /* Theme: Light */
            .shell[data-theme="light"] {
                --fb-bg-primary: #ffffff;
                --fb-bg-secondary: #f5f5f5;
                --fb-bg-tertiary: #eeeeee;
                --fb-bg-card: #ffffff;
                --fb-bg-rail: #f0f0f0;
                --fb-bg-stage: #fafafa;
                --fb-text-primary: #111111;
                --fb-text-secondary: #666666;
                --fb-text-muted: #999999;
                --fb-border: #e0e0e0;
                --fb-border-light: #f0f0f0;
            }

            /* Theme: Dark (default) */
            .shell[data-theme="dark"] {
                --fb-bg-primary: #000000;
                --fb-bg-secondary: #0a0a0a;
                --fb-bg-tertiary: #1a1a1a;
                --fb-bg-card: #111111;
                --fb-bg-rail: #050505;
                --fb-bg-stage: #0a0a0a;
                --fb-text-primary: #ffffff;
                --fb-text-secondary: #aaaaaa;
                --fb-text-muted: #666666;
                --fb-border: #333333;
                --fb-border-light: #222222;
            }

            .shell {
                display: contents;
            }
        `,
    ];

    @state() private _resolvedTheme: 'light' | 'dark' = 'dark';

    private _disposeEffect: (() => void) | null = null;

    override connectedCallback(): void {
        super.connectedCallback();

        // Initialize store on shell mount
        initializeStore();

        // Subscribe to theme changes
        this._disposeEffect = effect(() => {
            const theme = store.theme$.value;
            this._resolvedTheme = this._resolveTheme(theme);
        });
    }

    override disconnectedCallback(): void {
        this._disposeEffect?.();
        this._disposeEffect = null;
        super.disconnectedCallback();
    }

    private _resolveTheme(theme: string): 'light' | 'dark' {
        if (theme === 'system') {
            const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
            return prefersDark ? 'dark' : 'light';
        }
        return theme === 'light' ? 'light' : 'dark';
    }

    override render(): TemplateResult {
        return html`
            <div class="shell" data-theme="${this._resolvedTheme}">
                <fb-nav-rail></fb-nav-rail>
                <fb-header></fb-header>
                <main class="stage">
                    <slot></slot>
                </main>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-app-shell': FBAppShell;
    }
}

