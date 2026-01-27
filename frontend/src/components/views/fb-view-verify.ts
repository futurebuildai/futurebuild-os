/**
 * FBViewVerify - Magic Link Verification View
 * See PRODUCTION_PLAN.md B1: Frontend Auth Wiring
 *
 * Handles magic link token verification:
 * 1. Extracts token from URL query parameter
 * 2. Calls api.auth.verifyToken()
 * 3. On success: stores credentials and redirects to dashboard
 * 4. On error: shows error with option to request new link
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { store } from '../../store/store';
import { api } from '../../services/api';

type VerifyState = 'verifying' | 'success' | 'error' | 'no_token';

@customElement('fb-view-verify')
export class FBViewVerify extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                display: flex;
                align-items: center;
                justify-content: center;
                background: var(--fb-bg-primary);
            }

            .verify-card {
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-2xl);
                max-width: 400px;
                width: 100%;
                text-align: center;
            }

            .logo {
                font-size: var(--fb-text-3xl);
                font-weight: 700;
                color: var(--fb-primary);
                margin-bottom: var(--fb-spacing-lg);
            }

            .icon {
                font-size: 48px;
                margin-bottom: var(--fb-spacing-md);
            }

            .title {
                font-size: var(--fb-text-xl);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin-bottom: var(--fb-spacing-md);
            }

            .message {
                color: var(--fb-text-secondary);
                margin-bottom: var(--fb-spacing-lg);
                line-height: 1.5;
            }

            .spinner {
                width: 48px;
                height: 48px;
                border: 3px solid var(--fb-border);
                border-top-color: var(--fb-primary);
                border-radius: 50%;
                animation: spin 1s linear infinite;
                margin: 0 auto var(--fb-spacing-md);
            }

            @keyframes spin {
                to { transform: rotate(360deg); }
            }

            .error {
                color: var(--fb-error);
            }

            .btn-primary {
                display: inline-block;
                padding: var(--fb-spacing-md) var(--fb-spacing-lg);
                background: var(--fb-primary);
                color: white;
                border: none;
                border-radius: var(--fb-radius-md);
                font-size: var(--fb-text-base);
                font-weight: 600;
                cursor: pointer;
                transition: background 0.2s ease;
                text-decoration: none;
            }

            .btn-primary:hover {
                background: var(--fb-primary-hover);
            }

            .btn-secondary {
                display: inline-block;
                padding: var(--fb-spacing-sm) var(--fb-spacing-md);
                background: transparent;
                color: var(--fb-text-secondary);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-md);
                font-size: var(--fb-text-sm);
                cursor: pointer;
                transition: all 0.2s ease;
                text-decoration: none;
                margin-top: var(--fb-spacing-md);
            }

            .btn-secondary:hover {
                background: var(--fb-bg-tertiary);
                color: var(--fb-text-primary);
            }
        `,
    ];

    @state() private _state: VerifyState = 'verifying';
    @state() private _errorMessage = '';

    override connectedCallback(): void {
        super.connectedCallback();
        void this._verifyToken();
    }

    private async _verifyToken(): Promise<void> {
        // Extract token from URL
        const urlParams = new URLSearchParams(window.location.search);
        const token = urlParams.get('token');

        if (!token) {
            this._state = 'no_token';
            return;
        }

        try {
            const response = await api.auth.verifyToken(token);

            // Map API response (principal/access_token) to store format
            const user = {
                id: response.principal.id,
                email: response.principal.email,
                name: response.principal.name,
                role: response.principal.role,
                orgId: response.principal.org_id,
            };

            store.actions.login(user, response.access_token);
            this._state = 'success';

            // Clean up URL and redirect to dashboard
            window.history.replaceState({}, '', '/');

            // Short delay to show success message before redirect
            setTimeout(() => {
                // Reload the page to ensure all components pick up authenticated state
                window.location.href = '/';
            }, 1500);
        } catch (err) {
            this._state = 'error';
            this._errorMessage = err instanceof Error ? err.message : 'Verification failed. The link may have expired.';
        }
    }

    private _goToLogin(): void {
        window.location.href = '/';
    }

    override render(): TemplateResult {
        return html`
            <div class="verify-card">
                <div class="logo">FutureBuild</div>
                ${this._renderContent()}
            </div>
        `;
    }

    private _renderContent(): TemplateResult {
        switch (this._state) {
            case 'verifying':
                return html`
                    <div class="spinner"></div>
                    <div class="title">Verifying your link...</div>
                    <p class="message">Please wait while we sign you in.</p>
                `;

            case 'success':
                return html`
                    <div class="icon">✓</div>
                    <div class="title">You're in!</div>
                    <p class="message">Redirecting to your dashboard...</p>
                `;

            case 'error':
                return html`
                    <div class="icon">✕</div>
                    <div class="title error">Verification Failed</div>
                    <p class="message">${this._errorMessage}</p>
                    <button class="btn-primary" @click=${this._goToLogin.bind(this)}>
                        Request New Link
                    </button>
                `;

            case 'no_token':
                return html`
                    <div class="icon">🔗</div>
                    <div class="title">Invalid Link</div>
                    <p class="message">This link doesn't contain a valid token. Please request a new magic link.</p>
                    <button class="btn-primary" @click=${this._goToLogin.bind(this)}>
                        Go to Login
                    </button>
                `;

            default:
                // Exhaustive check - should never reach here
                return html``;
        }
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-verify': FBViewVerify;
    }
}
