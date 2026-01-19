/**
 * FBViewLogin - Authentication View
 * See PRODUCTION_PLAN.md Step 51.4
 *
 * The "Trap" - unauthenticated users are forced here by the Auth Guard.
 * Provides simulated login for development.
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBViewElement } from '../base/FBViewElement';
import { store } from '../../store/store';
import { UserRole } from '../../types/enums';

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
                font-size: var(--fb-text-3xl);
                font-weight: 700;
                color: var(--fb-primary);
                margin-bottom: var(--fb-spacing-lg);
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
        `,
    ];

    @state() private _email = '';
    @state() private _error = '';

    override render(): TemplateResult {
        return html`
            <div class="login-card">
                <div class="logo">FutureBuild</div>
                <p class="tagline">AI-Powered Construction Management</p>

                ${this._error ? html`<p class="error">${this._error}</p>` : null}

                <div class="input-group">
                    <input
                        type="email"
                        placeholder="Enter your email"
                        .value=${this._email}
                        @input=${this._handleEmailInput.bind(this)}
                        @keyup=${this._handleKeyUp.bind(this)}
                    />
                </div>

                <button class="btn-primary" @click=${this._handleLogin.bind(this)}>
                    Request Magic Link
                </button>

                <div class="dev-note">
                    <strong>Dev Mode:</strong> Click the button to simulate login.
                </div>
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

    private _handleLogin(): void {
        // Simulated login for development
        // In production, this would call api.auth.requestMagicLink()
        if (!this._email || !this._email.includes('@')) {
            this._error = 'Please enter a valid email address';
            return;
        }

        // Simulate successful login
        const mockUser = {
            id: 'user-dev-001',
            email: this._email,
            name: this._email.split('@')[0] ?? 'Dev User',
            role: UserRole.Builder,
            orgId: 'org-dev-001',
        };
        const mockToken = 'dev-jwt-token-' + Date.now().toString();

        store.actions.login(mockUser, mockToken);

        // Setup mock data for demo
        store.actions.setProjects([
            { id: 'proj-1', name: 'Sunrise Villa', address: '123 Main St', status: 'active', completionPercentage: 45, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
            { id: 'proj-2', name: 'Oak Ridge Home', address: '456 Oak Ave', status: 'active', completionPercentage: 20, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
        ]);
        store.actions.setFocusTasks([
            { id: 'ft-1', title: 'Review Framing Invoice', description: '', priority: 'high', projectId: 'proj-1', projectName: 'Sunrise Villa', actionType: 'approval', createdAt: new Date().toISOString() },
            { id: 'ft-2', title: 'Confirm Plumber Arrival', description: '', priority: 'medium', projectId: 'proj-1', projectName: 'Sunrise Villa', actionType: 'confirmation', createdAt: new Date().toISOString() },
        ]);
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-view-login': FBViewLogin;
    }
}
