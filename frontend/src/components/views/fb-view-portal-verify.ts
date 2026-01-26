/**
 * FBViewPortalVerify - Portal Token Verification View
 * See LAUNCH_PLAN.md P2: Field Portal (Mobile)
 *
 * Handles magic link verification for portal accounts.
 * Stores JWT and redirects to portal dashboard.
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { notify } from '../../store/notifications';

import '../portal/fb-portal-shell';

/**
 * Portal token verification view component.
 * @element fb-view-portal-verify
 */
@customElement('fb-view-portal-verify')
export class FBViewPortalVerify extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                display: block;
            }

            .content {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                min-height: 300px;
                text-align: center;
                padding: 32px;
            }

            .spinner {
                width: 48px;
                height: 48px;
                border: 4px solid var(--fb-border, #333);
                border-top-color: var(--fb-primary, #667eea);
                border-radius: 50%;
                animation: spin 1s linear infinite;
                margin-bottom: 24px;
            }

            @keyframes spin {
                to { transform: rotate(360deg); }
            }

            .title {
                color: var(--fb-text-primary, #fff);
                font-size: 20px;
                font-weight: 600;
                margin: 0 0 8px 0;
            }

            .subtitle {
                color: var(--fb-text-secondary, #aaa);
                font-size: 14px;
                margin: 0;
            }

            .error-icon {
                width: 64px;
                height: 64px;
                color: var(--fb-error, #c62828);
                margin-bottom: 16px;
            }

            .error-message {
                color: var(--fb-text-secondary, #aaa);
                font-size: 14px;
                margin: 0 0 24px 0;
            }

            .retry-btn {
                padding: 12px 24px;
                font-size: 14px;
                font-weight: 600;
                background: var(--fb-primary, #667eea);
                color: white;
                border: none;
                border-radius: 8px;
                cursor: pointer;
                transition: background 0.2s ease;
            }

            .retry-btn:hover {
                background: var(--fb-primary-hover, #5a6fd6);
            }

            .success-icon {
                width: 64px;
                height: 64px;
                color: var(--fb-success, #2e7d32);
                margin-bottom: 16px;
            }
        `,
    ];

    @state() private _verifying = true;
    @state() private _error: string | null = null;
    @state() private _success = false;

    override connectedCallback(): void {
        super.connectedCallback();
        void this._verifyToken();
    }

    private async _verifyToken(): Promise<void> {
        const params = new URLSearchParams(window.location.search);
        const token = params.get('token');

        if (!token) {
            this._error = 'Invalid verification link. No token provided.';
            this._verifying = false;
            return;
        }

        try {
            const response = await fetch('/api/v1/portal/auth/verify', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ token }),
            });

            if (!response.ok) {
                const data = await response.json() as { error?: string };
                throw new Error(data.error ?? 'Verification failed');
            }

            const data = await response.json() as { token: string; contact: unknown };

            // Store portal JWT (different from main app JWT)
            localStorage.setItem('portal_token', data.token);
            localStorage.setItem('portal_contact', JSON.stringify(data.contact));

            this._success = true;
            this._verifying = false;

            notify.success('Successfully verified! Redirecting to your dashboard...');

            // Redirect to dashboard after brief delay
            setTimeout(() => {
                window.location.href = '/portal/dashboard';
            }, 1500);
        } catch (err) {
            this._error = err instanceof Error ? err.message : 'Verification failed. Please try again.';
            this._verifying = false;
        }
    }

    private _handleRetry(): void {
        window.location.href = '/portal/login';
    }

    private _renderVerifying(): TemplateResult {
        return html`
            <div class="content">
                <div class="spinner"></div>
                <h1 class="title">Verifying your link...</h1>
                <p class="subtitle">Please wait while we verify your login.</p>
            </div>
        `;
    }

    private _renderError(): TemplateResult {
        return html`
            <div class="content">
                <svg class="error-icon" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
                </svg>
                <h1 class="title">Verification Failed</h1>
                <p class="error-message">${this._error}</p>
                <button class="retry-btn" @click=${this._handleRetry.bind(this)}>
                    Request New Link
                </button>
            </div>
        `;
    }

    private _renderSuccess(): TemplateResult {
        return html`
            <div class="content">
                <svg class="success-icon" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/>
                </svg>
                <h1 class="title">Verified!</h1>
                <p class="subtitle">Redirecting to your dashboard...</p>
            </div>
        `;
    }

    override render(): TemplateResult {
        let content: TemplateResult;

        if (this._verifying) {
            content = this._renderVerifying();
        } else if (this._error) {
            content = this._renderError();
        } else if (this._success) {
            content = this._renderSuccess();
        } else {
            content = this._renderVerifying();
        }

        return html`
            <fb-portal-shell minimal>
                ${content}
            </fb-portal-shell>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-portal-verify': FBViewPortalVerify;
    }
}
