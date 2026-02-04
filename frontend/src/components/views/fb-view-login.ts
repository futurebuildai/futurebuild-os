/**
 * FBViewLogin - Clerk Authentication View
 * See STEP_78_AUTH_PROVIDER.md Section 1.3
 *
 * Replaces the magic-link login form with Clerk's pre-built Sign-In component.
 * Clerk handles email/password, social login (Google), MFA, and session management.
 *
 * PORTAL PATTERN: Clerk uses Emotion CSS-in-JS which injects <style> tags into
 * document.head. Because this component lives inside multiple nested Shadow DOMs
 * (app-root → fb-app-shell → fb-panel-center), those styles can never reach
 * Clerk's rendered elements. To fix this, we mount the entire login UI as a
 * "portal" div directly in document.body, outside all shadow boundaries.
 * The component cleans up the portal when it disconnects (user logs in).
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { clerkService } from '../../services/clerk';

/** Unique ID for the portal element to prevent duplicates. */
const PORTAL_ID = 'fb-login-portal';

@customElement('fb-view-login')
export class FBViewLogin extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                /* Host is invisible — all visible content is in the portal */
                display: block;
                width: 0;
                height: 0;
                overflow: hidden;
            }
        `,
    ];

    @state() private _clerkError = false;
    private _portal: HTMLDivElement | null = null;
    private _signInContainer: HTMLDivElement | null = null;
    private _clerkPollInterval: ReturnType<typeof setInterval> | null = null;

    override firstUpdated(): void {
        this._createPortal();
        this._mountClerkSignIn();
    }

    override disconnectedCallback(): void {
        if (this._clerkPollInterval) {
            clearInterval(this._clerkPollInterval);
            this._clerkPollInterval = null;
        }
        this._unmountClerkSignIn();
        this._removePortal();
        super.disconnectedCallback();
    }

    // ---- Portal Lifecycle ----

    private _createPortal(): void {
        // Prevent duplicate portals
        const existing = document.getElementById(PORTAL_ID);
        if (existing) existing.remove();

        const portal = document.createElement('div');
        portal.id = PORTAL_ID;
        portal.innerHTML = `
            <style>
                #${PORTAL_ID} {
                    position: fixed;
                    inset: 0;
                    z-index: 10000;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    background: #000;
                    font-family: 'Poppins', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
                    overflow-y: auto;
                }

                #${PORTAL_ID} .login-container {
                    display: flex;
                    flex-direction: column;
                    align-items: center;
                    gap: 2rem;
                    max-width: 440px;
                    width: 100%;
                    padding: 2rem 1rem;
                }

                #${PORTAL_ID} .logo {
                    display: flex;
                    justify-content: center;
                }

                #${PORTAL_ID} .logo svg {
                    width: 200px;
                    height: auto;
                    color: #667eea;
                }

                #${PORTAL_ID} .tagline {
                    color: #aaa;
                    text-align: center;
                    margin: 0;
                    font-size: 0.95rem;
                }

                #${PORTAL_ID} .clerk-mount {
                    width: 100%;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                }

                #${PORTAL_ID} .loading-text {
                    color: #666;
                    font-size: 0.875rem;
                }

                #${PORTAL_ID} .error-text {
                    color: #e74c3c;
                    font-size: 0.875rem;
                    text-align: center;
                }
            </style>
            <div class="login-container">
                <div class="logo">
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 320 150" fill="none">
                        <g transform="translate(110, 5) scale(0.8)" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M10 45 L50 15 L90 45"/>
                            <path d="M20 50 L50 25 L80 50"/>
                            <path d="M25 48 L25 85"/>
                            <path d="M75 48 L75 85"/>
                            <path d="M25 85 L75 85"/>
                            <path d="M65 25 L65 35"/>
                            <path d="M65 25 L75 25"/>
                            <path d="M75 25 L75 40"/>
                        </g>
                        <g transform="translate(110, 5) scale(0.8)" stroke="currentColor" stroke-width="2" fill="none">
                            <path d="M50 38 L42 45 L42 50 L50 45 L58 50 L58 45 Z" fill="currentColor"/>
                            <path d="M50 50 L40 56 L40 68 L50 74 L60 68 L60 56 Z"/>
                            <circle cx="46" cy="60" r="2" fill="currentColor"/>
                            <circle cx="54" cy="60" r="2" fill="currentColor"/>
                            <circle cx="50" cy="42" r="1.5" fill="currentColor"/>
                            <path d="M50 64 L50 72"/>
                            <path d="M44 68 L56 68"/>
                            <path d="M50 74 L50 82"/>
                            <path d="M46 78 L50 82 L54 78"/>
                        </g>
                        <g fill="currentColor">
                            <text x="40" y="130" font-family="Inter, system-ui, sans-serif" font-size="22" font-weight="700" letter-spacing="3">FUTURE</text>
                            <text x="148" y="130" font-family="Inter, system-ui, sans-serif" font-size="22" font-weight="300" letter-spacing="3">BUILD AI</text>
                        </g>
                    </svg>
                </div>
                <p class="tagline">Give Your Construction Project a Mind of Its Own</p>
                <div class="clerk-mount">
                    <span class="loading-text">Loading...</span>
                </div>
            </div>
        `;

        document.body.appendChild(portal);
        this._portal = portal;
    }

    private _removePortal(): void {
        if (this._portal) {
            this._portal.remove();
            this._portal = null;
        }
    }

    // ---- Clerk Mount ----

    private _mountClerkSignIn(): void {
        if (!clerkService.loaded) {
            let attempts = 0;
            const maxAttempts = 100;
            this._clerkPollInterval = setInterval(() => {
                attempts++;
                if (clerkService.loaded) {
                    clearInterval(this._clerkPollInterval!);
                    this._clerkPollInterval = null;
                    this._mountClerkSignIn();
                } else if (attempts >= maxAttempts) {
                    clearInterval(this._clerkPollInterval!);
                    this._clerkPollInterval = null;
                    this._clerkError = true;
                    this._updatePortalState();
                }
            }, 100);
            return;
        }

        // If Clerk says the user is signed in, the auth state will propagate
        // through the store and the app shell will unmount this view.
        if (clerkService.isSignedIn) {
            return;
        }

        const mountPoint = this._portal?.querySelector('.clerk-mount') as HTMLDivElement | null;
        if (!mountPoint) return;

        // Clear loading text
        mountPoint.innerHTML = '';

        this._signInContainer = mountPoint;
        clerkService.mountSignIn(mountPoint);
    }

    private _unmountClerkSignIn(): void {
        if (this._signInContainer) {
            clerkService.unmountSignIn(this._signInContainer);
            this._signInContainer = null;
        }
    }

    private _updatePortalState(): void {
        if (!this._portal) return;
        const mountPoint = this._portal.querySelector('.clerk-mount');
        if (!mountPoint) return;

        if (this._clerkError) {
            mountPoint.innerHTML = '<span class="error-text">Unable to load authentication. Please refresh the page and try again.</span>';
        }
    }

    override render(): TemplateResult {
        // All visible content is in the portal (document.body).
        // This host element is invisible.
        return html``;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-login': FBViewLogin;
    }
}
