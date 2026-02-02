/**
 * FBViewLogin - Clerk Authentication View
 * See STEP_78_AUTH_PROVIDER.md Section 1.3
 *
 * Replaces the magic-link login form with Clerk's pre-built Sign-In component.
 * Clerk handles email/password, social login (Google), MFA, and session management.
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { clerkService } from '../../services/clerk';

@customElement('fb-view-login')
export class FBViewLogin extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                display: flex;
                align-items: center;
                justify-content: center;
                background: var(--fb-bg-primary);
            }

            .login-container {
                display: flex;
                flex-direction: column;
                align-items: center;
                gap: var(--fb-spacing-xl);
                max-width: 440px;
                width: 100%;
            }

            .logo {
                display: flex;
                justify-content: center;
            }

            .logo svg {
                width: 200px;
                height: auto;
                color: var(--fb-primary);
            }

            .tagline {
                color: var(--fb-text-secondary);
                text-align: center;
                margin: 0;
            }

            .clerk-container {
                width: 100%;
                min-height: 300px;
                display: flex;
                align-items: center;
                justify-content: center;
            }

            .loading {
                color: var(--fb-text-muted);
                font-size: var(--fb-text-sm);
            }

            .error {
                color: var(--fb-error, #e74c3c);
                font-size: var(--fb-text-sm);
                text-align: center;
            }
        `,
    ];

    @state() private _clerkReady = false;
    @state() private _clerkError = false;
    private _signInContainer: HTMLDivElement | null = null;
    private _clerkPollInterval: ReturnType<typeof setInterval> | null = null;

    override firstUpdated(): void {
        this._mountClerkSignIn();
    }

    override disconnectedCallback(): void {
        if (this._clerkPollInterval) {
            clearInterval(this._clerkPollInterval);
            this._clerkPollInterval = null;
        }
        this._unmountClerkSignIn();
        super.disconnectedCallback();
    }

    private _mountClerkSignIn(): void {
        if (!clerkService.loaded) {
            // Clerk not yet loaded — wait and retry (max 10s)
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
                }
            }, 100);
            return;
        }

        const container = this.shadowRoot?.getElementById('clerk-sign-in') as HTMLDivElement | null;
        if (!container) return;

        this._signInContainer = container;
        clerkService.mountSignIn(container);
        this._clerkReady = true;
    }

    private _unmountClerkSignIn(): void {
        if (this._signInContainer) {
            clerkService.unmountSignIn(this._signInContainer);
            this._signInContainer = null;
        }
    }

    override render(): TemplateResult {
        return html`
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

                <div class="clerk-container">
                    ${this._clerkError
                        ? html`<span class="error">Unable to load authentication. Please refresh the page and try again.</span>`
                        : !this._clerkReady
                            ? html`<span class="loading">Loading...</span>`
                            : ''}
                    <div id="clerk-sign-in"></div>
                </div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-login': FBViewLogin;
    }
}
