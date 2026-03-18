/**
 * FBViewPortalSignup - Portal Account Signup View
 * See LAUNCH_PLAN.md P2: Field Portal (Mobile)
 *
 * Account creation for contacts who want permanent portal access.
 * Linked to existing contact record via email.
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { notify } from '../../store/notifications';

import '../portal/fb-portal-shell';

/**
 * Portal signup view component.
 * @element fb-view-portal-signup
 */
@customElement('fb-view-portal-signup')
export class FBViewPortalSignup extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                display: block;
            }

            .signup-form {
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

            .info-box {
                background: var(--fb-primary-alpha, rgba(0, 255, 163, 0.1));
                border: 1px solid var(--fb-primary, #00FFA3);
                border-radius: 8px;
                padding: 16px;
                color: var(--fb-text-secondary, #8B8D98);
                font-size: 14px;
                line-height: 1.5;
            }

            .info-box strong {
                color: var(--fb-text-primary, #fff);
            }
        `,
    ];

    @state() private _email = '';
    @state() private _name = '';
    @state() private _submitting = false;
    @state() private _submitted = false;

    private _handleEmailChange(e: Event): void {
        this._email = (e.target as HTMLInputElement).value;
    }

    private _handleNameChange(e: Event): void {
        this._name = (e.target as HTMLInputElement).value;
    }

    private async _handleSubmit(e: Event): Promise<void> {
        e.preventDefault();
        if (!this._email || !this._name || this._submitting) return;

        this._submitting = true;

        try {
            const response = await fetch('/api/v1/portal/auth/signup', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    email: this._email,
                    name: this._name,
                }),
            });

            if (!response.ok) {
                const data = await response.json() as { error?: string };
                throw new Error(data.error ?? 'Failed to create account');
            }

            this._submitted = true;
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to create account. Please try again.';
            notify.error(message);
        } finally {
            this._submitting = false;
        }
    }

    private _renderSuccess(): TemplateResult {
        return html`
            <div class="success-message">
                <svg class="success-icon" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/>
                </svg>
                <h2 class="success-title">Account Created!</h2>
                <p class="success-text">
                    We've sent a verification link to ${this._email}.<br />
                    Click the link to activate your portal account.
                </p>
            </div>
        `;
    }

    private _renderForm(): TemplateResult {
        return html`
            <form class="signup-form" @submit=${this._handleSubmit.bind(this)}>
                <div class="header">
                    <h1 class="title">Create Portal Account</h1>
                    <p class="subtitle">Get permanent access to your assigned tasks</p>
                </div>

                <div class="info-box">
                    <strong>Why create an account?</strong><br />
                    With a portal account, you can view all your assigned tasks in one place
                    instead of using one-time links from text messages.
                </div>

                <div class="field">
                    <label class="label" for="name">Your Name</label>
                    <input
                        class="input"
                        type="text"
                        id="name"
                        name="name"
                        placeholder="John Smith"
                        autocomplete="name"
                        required
                        .value=${this._name}
                        @input=${this._handleNameChange.bind(this)}
                    />
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
                    ?disabled=${this._submitting || !this._email || !this._name}
                >
                    ${this._submitting ? 'Creating Account...' : 'Create Account'}
                </button>

                <div class="footer">
                    Already have an account?
                    <a class="footer-link" href="/portal/login">Sign in here</a>
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
        'fb-view-portal-signup': FBViewPortalSignup;
    }
}
