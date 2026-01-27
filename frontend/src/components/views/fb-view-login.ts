/**
 * FBViewLogin - Authentication View
 * See PRODUCTION_PLAN.md Step 51.4
 *
 * The "Trap" - unauthenticated users are forced here by the Auth Guard.
 * Uses magic link authentication - user enters email, receives link via SendGrid.
 */
import { html, css, TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { api } from '../../services/api';

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

            .login-card {
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-2xl);
                max-width: 400px;
                width: 100%;
                text-align: center;
            }

            .logo {
                display: flex;
                justify-content: center;
                margin-bottom: var(--fb-spacing-lg);
            }

            .logo svg {
                width: 200px;
                height: auto;
                color: var(--fb-primary);
            }

            .tagline {
                color: var(--fb-text-secondary);
                margin-bottom: var(--fb-spacing-xl);
            }

            .input-group {
                margin-bottom: var(--fb-spacing-md);
            }

            input {
                width: 100%;
                padding: var(--fb-spacing-md);
                background: var(--fb-bg-tertiary);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-md);
                color: var(--fb-text-primary);
                font-size: var(--fb-text-base);
            }

            input:focus {
                outline: none;
                border-color: var(--fb-primary);
            }

            .btn-primary {
                width: 100%;
                padding: var(--fb-spacing-md);
                background: var(--fb-primary);
                color: white;
                border: none;
                border-radius: var(--fb-radius-md);
                font-size: var(--fb-text-base);
                font-weight: 600;
                cursor: pointer;
                transition: background 0.2s ease;
            }

            .btn-primary:hover {
                background: var(--fb-primary-hover);
            }

            .dev-note {
                margin-top: var(--fb-spacing-lg);
                padding: var(--fb-spacing-md);
                background: var(--fb-bg-tertiary);
                border-radius: var(--fb-radius-sm);
                font-size: var(--fb-text-sm);
                color: var(--fb-text-muted);
            }

            .error {
                color: var(--fb-error);
                font-size: var(--fb-text-sm);
                margin-bottom: var(--fb-spacing-md);
            }

            .success-card {
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-2xl);
                max-width: 400px;
                width: 100%;
                text-align: center;
            }

            .success-icon {
                font-size: 48px;
                margin-bottom: var(--fb-spacing-md);
            }

            .success-title {
                font-size: var(--fb-text-xl);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin-bottom: var(--fb-spacing-md);
            }

            .success-message {
                color: var(--fb-text-secondary);
                margin-bottom: var(--fb-spacing-lg);
                line-height: 1.5;
            }

            .email-highlight {
                color: var(--fb-primary);
                font-weight: 500;
            }

            .btn-secondary {
                padding: var(--fb-spacing-sm) var(--fb-spacing-md);
                background: transparent;
                color: var(--fb-text-secondary);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-md);
                font-size: var(--fb-text-sm);
                cursor: pointer;
                transition: all 0.2s ease;
            }

            .btn-secondary:hover {
                background: var(--fb-bg-tertiary);
                color: var(--fb-text-primary);
            }

            .btn-primary:disabled {
                opacity: 0.6;
                cursor: not-allowed;
            }
        `,
    ];

    @state() private _email = '';
    @state() private _error = '';
    @state() private _isLoading = false;
    @state() private _emailSent = false;

    override render(): TemplateResult {
        // Show success state after email is sent
        if (this._emailSent) {
            return html`
                <div class="success-card">
                    <div class="success-icon">📧</div>
                    <div class="success-title">Check your email</div>
                    <p class="success-message">
                        We've sent a magic link to<br />
                        <span class="email-highlight">${this._email}</span><br />
                        Click the link to sign in.
                    </p>
                    <button class="btn-secondary" @click=${this._resetForm.bind(this)}>
                        Use a different email
                    </button>
                </div>
            `;
        }

        return html`
            <div class="login-card">
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

                ${this._error ? html`<p class="error">${this._error}</p>` : nothing}

                <div class="input-group">
                    <input
                        type="email"
                        placeholder="Enter your email"
                        aria-label="Email address"
                        .value=${this._email}
                        ?disabled=${this._isLoading}
                        @input=${this._handleEmailInput.bind(this)}
                        @keyup=${this._handleKeyUp.bind(this)}
                    />
                </div>

                <button
                    class="btn-primary"
                    ?disabled=${this._isLoading}
                    @click=${this._handleLogin.bind(this)}
                >
                    ${this._isLoading ? 'Sending...' : 'Request Magic Link'}
                </button>
            </div>
        `;
    }

    private _handleEmailInput(e: Event): void {
        this._email = (e.target as HTMLInputElement).value;
        this._error = '';
    }

    private _handleKeyUp(e: KeyboardEvent): void {
        if (e.key === 'Enter') {
            this._handleLogin();
        }
    }

    private async _handleLogin(): Promise<void> {
        // Validate email
        if (!this._email || !this._email.includes('@')) {
            this._error = 'Please enter a valid email address';
            return;
        }

        this._isLoading = true;
        this._error = '';

        try {
            await api.auth.requestMagicLink(this._email);
            this._emailSent = true;
        } catch (err) {
            this._error = err instanceof Error ? err.message : 'Failed to send magic link. Please try again.';
        } finally {
            this._isLoading = false;
        }
    }

    private _resetForm(): void {
        this._emailSent = false;
        this._email = '';
        this._error = '';
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-login': FBViewLogin;
    }
}
