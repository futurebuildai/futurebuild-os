/**
 * FBAppShell - Root Application Container (3-Panel Agent Command Center)
 * See FRONTEND_SCOPE.md Section 3.3 (Updated v1.3.0)
 *
 * CSS Grid container for 3-panel layout:
 * - Left Panel (280px): Projects, Threads, Daily Focus, Agent Activity
 * - Center Panel (flex: 1): Conversation / Main Content
 * - Right Panel (320px, collapsible): Artifacts
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { store, initializeStore } from '../../store/store';
import { clerkService } from '../../services/clerk';

// Import panel components
import './fb-panel-left';
import './fb-panel-center';
import './fb-panel-right';

// Import feature components (Step 56)
import '../features/fb-file-drop';

// Import resize handle (Step 59.5)
import './fb-resize-handle';

// Import artifact modal (Step 59.5)
import '../artifacts/fb-artifact-modal';

// Import toast container (LAUNCH_PLAN.md P2)
import '../feedback/fb-toast-container';

// Import mobile navigation (Step 90: Mobile Navigation)
import './fb-mobile-nav';

// Import shadow layout (FutureShade)
import '../shadow/shadow-layout';

/**
 * Application Shell - 3-Panel Layout Container
 * @element fb-app-shell
 */
@customElement('fb-app-shell')
export class FBAppShell extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: grid;
                grid-template-columns: var(--fb-left-panel-width, 280px) 1fr var(--fb-right-panel-width, 320px);
                grid-template-rows: 1fr;
                grid-template-areas: "left center right";
                height: 100vh;
                width: 100vw;
                overflow: hidden;
                background: var(--fb-bg-primary);
                color: var(--fb-text-primary);
                font-family: var(--fb-font-family);
            }

            /* Left panel closed */
            :host([left-closed]) {
                grid-template-columns: 0 1fr var(--fb-right-panel-width, 320px);
            }

            /* Right panel closed */
            :host([right-closed]) {
                grid-template-columns: var(--fb-left-panel-width, 280px) 1fr 0;
            }

            /* Both panels closed */
            :host([left-closed][right-closed]) {
                grid-template-columns: 0 1fr 0;
            }

            fb-panel-left {
                grid-area: left;
            }

            fb-panel-center {
                grid-area: center;
            }

            fb-panel-right {
                grid-area: right;
            }

            /* Panel visibility */
            :host([left-closed]) fb-panel-left {
                display: none;
            }

            :host([right-closed]) fb-panel-right {
                display: none;
            }

            /* Mobile: Only center panel visible + bottom padding for mobile nav */
            @media (max-width: 767px) {
                :host {
                    grid-template-columns: 1fr;
                    grid-template-areas: "center";
                    /* Step 90: Reserve space for fixed mobile nav bar */
                    padding-bottom: calc(64px + env(safe-area-inset-bottom, 0px));
                }

                fb-panel-left,
                fb-panel-right {
                    position: fixed;
                    top: 0;
                    bottom: 0;
                    z-index: var(--fb-z-panel, 100);
                    transition: transform 0.3s ease;
                }

                fb-panel-left {
                    left: 0;
                    width: 280px;
                    transform: translateX(-100%);
                }

                fb-panel-right {
                    right: 0;
                    width: 320px;
                    transform: translateX(100%);
                }

                :host(:not([left-closed])) fb-panel-left {
                    display: block;
                    transform: translateX(0);
                }

                :host(:not([right-closed])) fb-panel-right {
                    display: block;
                    transform: translateX(0);
                }
            }

            /* Tablet: Hide right panel by default */
            @media (min-width: 769px) and (max-width: 1024px) {
                :host {
                    grid-template-columns: var(--fb-left-panel-width, 280px) 1fr 0;
                }

                :host(:not([right-closed])) {
                    grid-template-columns: var(--fb-left-panel-width, 280px) 1fr var(--fb-right-panel-width, 320px);
                }
            }

            /* Login mode: Full screen center */
            :host([login-mode]) {
                grid-template-columns: 1fr;
                grid-template-areas: "center";
                padding-bottom: 0;
            }

            :host([login-mode]) fb-panel-left,
            :host([login-mode]) fb-panel-right {
                display: none;
            }

            .backdrop {
                position: fixed;
                top: 0;
                left: 0;
                width: 100vw;
                height: 100vh;
                background: rgba(0, 0, 0, 0.5);
                backdrop-filter: blur(2px);
                z-index: calc(var(--fb-z-panel, 100) - 1);
                animation: fadeIn 0.2s ease;
            }

            @keyframes fadeIn {
                from { opacity: 0; }
                to { opacity: 1; }
            }

            /* Theme: Light */
            .shell[data-theme="light"] {
                --fb-bg-primary: #ffffff;
                --fb-bg-secondary: #f5f5f5;
                --fb-bg-tertiary: #eeeeee;
                --fb-bg-card: #ffffff;
                --fb-bg-panel: #fafafa;
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
                --fb-bg-panel: #0a0a0a;
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
    @state() private _isAuthenticated = false;
    @state() private _clerkLoaded = false;
    @state() private _leftPanelOpen = true;
    @state() private _rightPanelOpen = true;
    @state() private _isMobile = false;
    @state() private _rightPanelWidth = 320;
    @state() private _hasPopoutArtifact = false;
    @state() private _shadowModeEnabled = false;

    private _disposeEffects: (() => void)[] = [];

    /**
     * Counter for drag enter/leave events.
     * Prevents flicker when cursor moves over child elements.
     * See: https://stackoverflow.com/q/7110353
     * Step 56: Drag-and-Drop Ingestion
     */
    private _dragCounter = 0;

    // ---- Drag-and-Drop Handlers (Step 56) ----

    private _handleDragEnter = (e: DragEvent): void => {
        e.preventDefault();
        e.stopPropagation();
        this._dragCounter++;
        if (this._dragCounter === 1) {
            store.actions.setDragging(true);
        }
    };

    private _handleDragOver = (e: DragEvent): void => {
        e.preventDefault();
        e.stopPropagation();
    };

    private _handleDragLeave = (e: DragEvent): void => {
        e.preventDefault();
        e.stopPropagation();
        this._dragCounter--;
        if (this._dragCounter === 0) {
            store.actions.setDragging(false);
        }
    };

    private _handleDrop = (e: DragEvent): void => {
        e.preventDefault();
        e.stopPropagation();
        this._dragCounter = 0;

        const files = e.dataTransfer?.files;
        if (files && files.length > 0) {
            store.actions.handleFileDrop(files);
        } else {
            store.actions.setDragging(false);
        }
    };

    override connectedCallback(): void {
        super.connectedCallback();

        // Phase 12: Initialize Clerk before the store.
        // Clerk must load first so the store can wire its auth observer.
        void this._initClerkAndStore();

        // Drag-and-drop event listeners (Step 56)
        this.addEventListener('dragenter', this._handleDragEnter);
        this.addEventListener('dragover', this._handleDragOver);
        this.addEventListener('dragleave', this._handleDragLeave);
        this.addEventListener('drop', this._handleDrop);

        // Subscribe to theme changes
        this._disposeEffects.push(
            effect(() => {
                const theme = store.theme$.value;
                this._resolvedTheme = this._resolveTheme(theme);
            })
        );

        // Subscribe to auth state changes
        this._disposeEffects.push(
            effect(() => {
                this._isAuthenticated = store.isAuthenticated$.value;
                if (this._isAuthenticated) {
                    this.removeAttribute('login-mode');
                } else {
                    this.setAttribute('login-mode', '');
                }
            })
        );

        // Subscribe to panel visibility
        this._disposeEffects.push(
            effect(() => {
                this._leftPanelOpen = store.leftPanelOpen$.value;
                if (this._leftPanelOpen) {
                    this.removeAttribute('left-closed');
                } else {
                    this.setAttribute('left-closed', '');
                }
            })
        );

        this._disposeEffects.push(
            effect(() => {
                this._rightPanelOpen = store.rightPanelOpen$.value;
                if (this._rightPanelOpen) {
                    this.removeAttribute('right-closed');
                } else {
                    this.setAttribute('right-closed', '');
                }
            })
        );

        // Subscribe to mobile state
        this._disposeEffects.push(
            effect(() => {
                this._isMobile = store.isMobile$.value;
            })
        );

        // Subscribe to panel width (Step 59.5: UX Enhancements)
        this._disposeEffects.push(
            effect(() => {
                this._rightPanelWidth = store.rightPanelWidth$.value;
                this.style.setProperty('--fb-right-panel-width', `${String(this._rightPanelWidth)}px`);
            })
        );

        // Subscribe to popout artifact
        this._disposeEffects.push(
            effect(() => {
                this._hasPopoutArtifact = store.popoutArtifact$.value !== null;
            })
        );

        // Subscribe to shadow mode (FutureShade)
        this._disposeEffects.push(
            effect(() => {
                this._shadowModeEnabled = store.shadowModeEnabled$.value;
            })
        );
    }

    /**
     * Initialize Clerk identity provider, then bootstrap the store.
     * Clerk must be loaded before initializeStore() so the Clerk auth
     * observer is available when the store wires its callbacks.
     * See STEP_78_AUTH_PROVIDER.md Section 1.2
     */
    private async _initClerkAndStore(): Promise<void> {
        const publishableKey = import.meta.env.VITE_CLERK_PUBLISHABLE_KEY as string | undefined;
        if (!publishableKey) {
            console.error('[FBAppShell] VITE_CLERK_PUBLISHABLE_KEY not set. Auth will not work.');
            this._clerkLoaded = true;
            initializeStore();
            return;
        }

        try {
            await clerkService.init(publishableKey);
        } catch (err) {
            console.error('[FBAppShell] Clerk initialization failed:', err);
        }

        this._clerkLoaded = true;
        initializeStore();
    }

    override disconnectedCallback(): void {
        // Remove drag listeners (Step 56)
        this.removeEventListener('dragenter', this._handleDragEnter);
        this.removeEventListener('dragover', this._handleDragOver);
        this.removeEventListener('dragleave', this._handleDragLeave);
        this.removeEventListener('drop', this._handleDrop);

        this._disposeEffects.forEach((dispose) => { dispose(); });
        this._disposeEffects = [];
        super.disconnectedCallback();
    }

    private _resolveTheme(theme: string): 'light' | 'dark' {
        if (theme === 'system') {
            const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
            return prefersDark ? 'dark' : 'light';
        }
        return theme === 'light' ? 'light' : 'dark';
    }

    private _closePanels(): void {
        if (this._leftPanelOpen) store.actions.setLeftPanelOpen(false);
        if (this._rightPanelOpen) store.actions.setRightPanelOpen(false);
    }

    override render(): TemplateResult {
        // Phase 12: Gate rendering on Clerk initialization
        if (!this._clerkLoaded) {
            return html`
                <div class="shell" data-theme="dark" style="position: relative; display: flex; align-items: center; justify-content: center; height: 100vh;">
                    <span style="color: var(--fb-text-muted, #666);">Loading...</span>
                </div>
            `;
        }

        const resizeHandleOffset = this._rightPanelOpen && !this._isMobile ? this._rightPanelWidth : 0;

        // Shadow mode replaces the standard layout with FutureShade
        if (this._isAuthenticated && this._shadowModeEnabled) {
            return html`
                <div class="shell" data-theme="dark" style="position: relative;">
                    <shadow-layout></shadow-layout>
                    <fb-toast-container></fb-toast-container>
                </div>
            `;
        }

        return html`
            <div class="shell" data-theme="${this._resolvedTheme}" style="position: relative;">
                <fb-file-drop></fb-file-drop>
                ${this._isAuthenticated ? html`<fb-panel-left></fb-panel-left>` : nothing}
                <fb-panel-center .isAuthenticated=${this._isAuthenticated}></fb-panel-center>
                ${this._isAuthenticated ? html`<fb-panel-right></fb-panel-right>` : nothing}

                ${this._isAuthenticated && this._rightPanelOpen && !this._isMobile ? html`
                    <fb-resize-handle style="--resize-handle-offset: ${resizeHandleOffset}px;"></fb-resize-handle>
                ` : nothing}

                ${this._isMobile && (this._leftPanelOpen || this._rightPanelOpen) ? html`
                    <div class="backdrop" @click=${this._closePanels.bind(this)}></div>
                ` : nothing}

                ${this._hasPopoutArtifact ? html`<fb-artifact-modal></fb-artifact-modal>` : nothing}

                <!-- Step 90: Mobile bottom navigation bar (authenticated only) -->
                ${this._isAuthenticated ? html`<fb-mobile-nav></fb-mobile-nav>` : nothing}

                <fb-toast-container></fb-toast-container>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-app-shell': FBAppShell;
    }
}
