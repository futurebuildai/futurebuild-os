/**
 * FBViewPortalLogin - Portal Login View
 * See LAUNCH_PLAN.md P2: Field Portal (Mobile)
 *
 * Login for contacts who have created permanent portal accounts.
 * Mobile-first design with large touch targets.
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { notify } from '../../store/notifications';

import '../portal/fb-portal-shell';

/**
 * Portal login view component.
 * @element fb-view-portal-login
 */
@customElement('fb-view-portal-login')
export class FBViewPortalLogin extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                display: block;
            }

            .login-form {
                display: flex;
                flex-direction: column;
                gap: 24px;
                max-width: 400px;
                margin: 0 auto;
            }

            .header {
                text-align: center;
                margin-bottom: 16px;
            }

            .title {
                color: var(--fb-text-primary, #fff);
                font-size: 24px;
                font-weight: 600;
                margin: 0 0 8px 0;
            }

            .subtitle {
                color: var(--fb-text-secondary, #8B8D98);
                font-size: 14px;
                margin: 0;
            }

            .field {
                display: flex;
                flex-direction: column;
                gap: 8px;
            }

            .label {
                color: var(--fb-text-secondary, #8B8D98);
                font-size: 14px;
                font-weight: 500;
            }

            .input {
                padding: 14px 16px;
                font-size: 16px;
                background: var(--fb-bg-card, #161821);
                border: 1px solid var(--fb-border, rgba(255,255,255,0.05));
                border-radius: 8px;
                color: var(--fb-text-primary, #fff);
                outline: none;
                transition: border-color 0.2s ease;
            }

            .input:focus {
                border-color: var(--fb-primary, #00FFA3);
            }

            .input::placeholder {
                color: var(--fb-text-muted, #4A4B55);
            }

            .submit-btn {
                padding: 16px 24px;
                font-size: 16px;
                font-weight: 600;
                background: var(--fb-primary, #00FFA3);
                color: white;
                border: none;
                border-radius: 12px;
                cursor: pointer;
                transition: background 0.2s ease;
                margin-top: 8px;
            }

            .submit-btn:hover:not([disabled]) {
                background: var(--fb-primary-hover, #5a6fd6);
            }

            .submit-btn[disabled] {
                opacity: 0.5;
                cursor: not-allowed;
            }

            .footer {
                text-align: center;
                color: var(--fb-text-secondary, #8B8D98);
                font-size: 14px;
            }

            .footer-link {
                color: var(--fb-primary, #00FFA3);
                text-decoration: none;
            }

            .footer-link:hover {
                text-decoration: underline;
            }

            .success-message {
                text-align: center;
                padding: 32px;
            }

            .success-icon {
                width: 64px;
                height: 64px;
                color: var(--fb-success, #00FFA3);
                margin-bottom: 16px;
            }

            .success-title {
                color: var(--fb-text-primary, #fff);
                font-size: 20px;
                font-weight: 600;
                margin: 0 0 8px 0;
            }

            .success-text {
                color: var(--fb-text-secondary, #8B8D98);
                font-size: 14px;
                margin: 0;
            }
        `,
    ];

    @state() private _email = '';
    @state() private _submitting = false;
    @state() private _submitted = false;

    private _handleEmailChange(e: Event): void {
        this._email = (e.target as HTMLInputElement).value;
    }

    private async _handleSubmit(e: Event): Promise<void> {
        e.preventDefault();
        if (!this._email || this._submitting) return;

        this._submitting = true;

        const controller = new AbortController();
        const timeoutId = setTimeout(() => { controller.abort(); }, 15000);

        try {
            const response = await fetch('/api/v1/portal/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email: this._email }),
                signal: controller.signal,
            });
            clearTimeout(timeoutId);

            if (!response.ok) {
                throw new Error('Failed to send login link');
            }

            this._submitted = true;
        } catch (err) {
            clearTimeout(timeoutId);
            if (err instanceof DOMException && err.name === 'AbortError') {
                notify.error('Request timed out. Please try again.');
            } else {
                notify.error('Failed to send login link. Please try again.');
            }
        } finally {
            this._submitting = false;
        }
    }

    private _renderSuccess(): TemplateResult {
        return html`
            <div class="success-message">
                <svg class="success-icon" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M20 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 14H4V8l8 5 8-5v10zm-8-7L4 6h16l-8 5z"/>
                </svg>
                <h2 class="success-title">Check Your Email</h2>
                <p class="success-text">
                    We've sent a login link to ${this._email}.<br />
                    Click the link to sign in to your portal.
                </p>
            </div>
        `;
    }

    private _renderForm(): TemplateResult {
        return html`
            <form class="login-form" @submit=${this._handleSubmit.bind(this)}>
                <div class="header">
                    <h1 class="title">Portal Login</h1>
                    <p class="subtitle">Enter your email to receive a login link</p>
                </div>

                <div class="field">
                    <label class="label" for="email">Email Address</label>
                    <input
                        class="input"
                        type="email"
                        id="email"
                        name="email"
                        placeholder="you@company.com"
                        autocomplete="email"
                        required
                        .value=${this._email}
                        @input=${this._handleEmailChange.bind(this)}
                    />
                </div>

                <button
                    class="submit-btn"
                    type="submit"
                    ?disabled=${this._submitting || !this._email}
                >
                    ${this._submitting ? 'Sending...' : 'Send Login Link'}
                </button>

                <div class="footer">
                    Don't have an account?
                    <a class="footer-link" href="/portal/signup">Sign up here</a>
                </div>
            </form>
        `;
    }

    override render(): TemplateResult {
        return html`
            <fb-portal-shell minimal>
                ${this._submitted ? this._renderSuccess() : this._renderForm()}
            </fb-portal-shell>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-portal-login': FBViewPortalLogin;
    }
}
