/**
 * FBViewInviteAccept - Invite Acceptance View
 * See LAUNCH_STRATEGY.md Task B2: User Invite Flow
 *
 * Handles invitation acceptance:
 * 1. Extracts token from URL query parameter
 * 2. Fetches invite info (email, role)
 * 3. User enters their name
 * 4. Submits to create account
 * 5. Redirects to login page
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { api } from '../../services/api';

type InviteState = 'loading' | 'ready' | 'submitting' | 'success' | 'error';

@customElement('fb-view-invite-accept')
export class FBViewInviteAccept extends FBViewElement {
    static override styles = [
        FBViewElement.styles,
        css`
            :host {
                display: flex;
                align-items: center;
                justify-content: center;
                background: var(--fb-bg-primary);
            }

            .invite-card {
                background: var(--fb-bg-card);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-lg);
                padding: var(--fb-spacing-2xl);
                max-width: 440px;
                width: 100%;
            }

            .logo {
                font-size: var(--fb-text-3xl);
                font-weight: 700;
                color: var(--fb-primary);
                margin-bottom: var(--fb-spacing-lg);
                text-align: center;
            }

            .title {
                font-size: var(--fb-text-xl);
                font-weight: 600;
                color: var(--fb-text-primary);
                margin-bottom: var(--fb-spacing-md);
                text-align: center;
            }

            .subtitle {
                color: var(--fb-text-secondary);
                margin-bottom: var(--fb-spacing-xl);
                text-align: center;
                line-height: 1.5;
            }

            .email-display {
                background: var(--fb-bg-tertiary);
                border-radius: var(--fb-radius-md);
                padding: var(--fb-spacing-md);
                margin-bottom: var(--fb-spacing-lg);
                text-align: center;
            }

            .email-label {
                font-size: var(--fb-text-sm);
                color: var(--fb-text-muted);
                margin-bottom: var(--fb-spacing-xs);
            }

            .email-value {
                color: var(--fb-primary);
                font-weight: 500;
            }

            .role-badge {
                display: inline-block;
                background: var(--fb-primary);
                color: white;
                padding: var(--fb-spacing-xs) var(--fb-spacing-sm);
                border-radius: var(--fb-radius-sm);
                font-size: var(--fb-text-xs);
                text-transform: uppercase;
                margin-top: var(--fb-spacing-xs);
            }

            .form-group {
                margin-bottom: var(--fb-spacing-lg);
            }

            label {
                display: block;
                font-size: var(--fb-text-sm);
                color: var(--fb-text-secondary);
                margin-bottom: var(--fb-spacing-xs);
            }

            input {
                width: 100%;
                padding: var(--fb-spacing-md);
                background: var(--fb-bg-tertiary);
                border: 1px solid var(--fb-border);
                border-radius: var(--fb-radius-md);
                color: var(--fb-text-primary);
                font-size: var(--fb-text-base);
                box-sizing: border-box;
            }

            input:focus {
                outline: none;
                border-color: var(--fb-primary);
            }

            input:disabled {
                opacity: 0.6;
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

            .btn-primary:hover:not(:disabled) {
                background: var(--fb-primary-hover);
            }

            .btn-primary:disabled {
                opacity: 0.6;
                cursor: not-allowed;
            }

            .error {
                color: var(--fb-error);
                text-align: center;
                margin-bottom: var(--fb-spacing-md);
            }

            .spinner {
                width: 48px;
                height: 48px;
                border: 3px solid var(--fb-border);
                border-top-color: var(--fb-primary);
                border-radius: 50%;
                animation: spin 1s linear infinite;
                margin: var(--fb-spacing-xl) auto;
            }

            @keyframes spin {
                to { transform: rotate(360deg); }
            }

            .icon {
                font-size: 48px;
                text-align: center;
                margin-bottom: var(--fb-spacing-md);
            }

            .success-message {
                text-align: center;
                color: var(--fb-text-secondary);
                margin-bottom: var(--fb-spacing-lg);
            }
        `,
    ];

    @state() private _state: InviteState = 'loading';
    @state() private _email = '';
    @state() private _role = '';
    @state() private _name = '';
    @state() private _error = '';

    override connectedCallback(): void {
        super.connectedCallback();
        void this._loadInviteInfo();
    }

    private async _loadInviteInfo(): Promise<void> {
        const urlParams = new URLSearchParams(window.location.search);
        const token = urlParams.get('token');

        if (!token) {
            this._state = 'error';
            this._error = 'No invitation token provided';
            return;
        }

        try {
            const info = await api.invites.getInfo(token);
            this._email = info.email;
            this._role = info.role;
            this._state = 'ready';
        } catch (err) {
            this._state = 'error';
            this._error = err instanceof Error ? err.message : 'Invalid or expired invitation';
        }
    }

    private _handleNameInput(e: Event): void {
        this._name = (e.target as HTMLInputElement).value;
    }

    private async _handleSubmit(): Promise<void> {
        if (!this._name.trim()) {
            this._error = 'Please enter your name';
            return;
        }

        const urlParams = new URLSearchParams(window.location.search);
        const token = urlParams.get('token');

        if (!token) {
            this._error = 'Invalid invitation token';
            return;
        }

        this._state = 'submitting';
        this._error = '';

        try {
            await api.invites.accept(token, this._name.trim());
            this._state = 'success';

            // Redirect to login after a short delay
            setTimeout(() => {
                window.location.href = '/';
            }, 2000);
        } catch (err) {
            this._state = 'ready';
            this._error = err instanceof Error ? err.message : 'Failed to accept invitation';
        }
    }

    private _handleKeyUp(e: KeyboardEvent): void {
        if (e.key === 'Enter') {
            void this._handleSubmit();
        }
    }

    override render(): TemplateResult {
        return html`
            <div class="invite-card">
                <div class="logo">FutureBuild</div>
                ${this._renderContent()}
            </div>
        `;
    }

    private _renderContent(): TemplateResult {
        switch (this._state) {
            case 'loading':
                return html`
                    <div class="spinner"></div>
                    <div class="subtitle">Loading invitation...</div>
                `;

            case 'error':
                return html`
                    <div class="icon">&#x2715;</div>
                    <div class="title">Invitation Error</div>
                    <p class="error">${this._error}</p>
                    <button class="btn-primary" @click=${() => { window.location.href = '/'; }}>
                        Go to Login
                    </button>
                `;

            case 'success':
                return html`
                    <div class="icon">&#x2713;</div>
                    <div class="title">Account Created!</div>
                    <p class="success-message">
                        Your account has been created successfully.<br />
                        Redirecting to login...
                    </p>
                `;

            case 'ready':
            case 'submitting':
                return html`
                    <div class="title">Join FutureBuild</div>
                    <p class="subtitle">Complete your account setup to get started.</p>

                    <div class="email-display">
                        <div class="email-label">Email</div>
                        <div class="email-value">${this._email}</div>
                        <span class="role-badge">${this._role}</span>
                    </div>

                    ${this._error ? html`<p class="error">${this._error}</p>` : ''}

                    <div class="form-group">
                        <label for="name">Your Name</label>
                        <input
                            id="name"
                            type="text"
                            placeholder="Enter your full name"
                            .value=${this._name}
                            ?disabled=${this._state === 'submitting'}
                            @input=${this._handleNameInput.bind(this)}
                            @keyup=${this._handleKeyUp.bind(this)}
                        />
                    </div>

                    <button
                        class="btn-primary"
                        ?disabled=${this._state === 'submitting'}
                        @click=${this._handleSubmit.bind(this)}
                    >
                        ${this._state === 'submitting' ? 'Creating Account...' : 'Create Account'}
                    </button>
                `;
        }
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-invite-accept': FBViewInviteAccept;
    }
}
