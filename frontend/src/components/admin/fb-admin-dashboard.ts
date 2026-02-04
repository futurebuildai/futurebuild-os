/**
 * FBAdminDashboard - Platform Admin Landing View
 * Quick links to admin sections and system health card.
 */
import { html, css, type TemplateResult, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';

interface HealthStatus {
    status: string;
    checks?: Record<string, { status: string; message?: string }>;
}

@customElement('fb-admin-dashboard')
export class FBAdminDashboard extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                padding: var(--fb-spacing-2xl, 32px);
                overflow-y: auto;
                height: 100%;
                color: var(--fb-text-primary, #fff);
            }

            h1 {
                font-size: 24px;
                font-weight: 600;
                margin: 0 0 8px;
            }

            .subtitle {
                font-size: var(--fb-text-sm, 13px);
                color: var(--fb-text-muted, #666);
                margin-bottom: var(--fb-spacing-xl, 24px);
            }

            .grid {
                display: grid;
                grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
                gap: var(--fb-spacing-md, 16px);
                margin-bottom: var(--fb-spacing-xl, 24px);
            }

            .card {
                background: var(--fb-bg-card, #111);
                border: 1px solid var(--fb-border, #333);
                border-radius: var(--fb-radius-lg, 12px);
                padding: var(--fb-spacing-lg, 20px);
                transition: border-color 0.15s ease;
            }

            .card-link {
                cursor: pointer;
            }

            .card-link:hover {
                border-color: var(--fb-warning, #f59e0b);
            }

            .card-header {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm, 8px);
                margin-bottom: var(--fb-spacing-sm, 8px);
            }

            .card-icon {
                width: 20px;
                height: 20px;
                color: var(--fb-warning, #f59e0b);
            }

            .card-icon svg {
                width: 100%;
                height: 100%;
                stroke: currentColor;
                fill: none;
                stroke-width: 2;
            }

            .card-title {
                font-size: var(--fb-text-md, 14px);
                font-weight: 600;
            }

            .card-desc {
                font-size: var(--fb-text-sm, 13px);
                color: var(--fb-text-secondary, #aaa);
                line-height: 1.5;
            }

            .health-section {
                margin-top: var(--fb-spacing-lg, 20px);
            }

            .health-section h2 {
                font-size: var(--fb-text-md, 14px);
                font-weight: 600;
                margin: 0 0 var(--fb-spacing-sm, 8px);
            }

            .health-status {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm, 8px);
                font-size: var(--fb-text-sm, 13px);
                color: var(--fb-text-secondary, #aaa);
            }

            .health-dot {
                width: 8px;
                height: 8px;
                border-radius: 50%;
                flex-shrink: 0;
            }

            .health-dot.ok { background: var(--fb-success, #22c55e); }
            .health-dot.degraded { background: var(--fb-warning, #f59e0b); }
            .health-dot.error { background: var(--fb-danger, #ef4444); }
            .health-dot.loading { background: var(--fb-text-muted, #666); }

            .health-checks {
                margin-top: var(--fb-spacing-sm, 8px);
                display: flex;
                flex-direction: column;
                gap: 4px;
            }

            .health-check-item {
                display: flex;
                align-items: center;
                gap: var(--fb-spacing-sm, 8px);
                font-size: var(--fb-text-xs, 11px);
                color: var(--fb-text-muted, #666);
            }
        `,
    ];

    @state() private _health: HealthStatus | null = null;
    @state() private _healthLoading = true;
    @state() private _healthError: string | null = null;

    override connectedCallback(): void {
        super.connectedCallback();
        void this._fetchHealth();
    }

    private async _fetchHealth(): Promise<void> {
        this._healthLoading = true;
        this._healthError = null;
        try {
            const res = await fetch('/api/v1/readiness');
            if (!res.ok) {
                this._healthError = `HTTP ${String(res.status)}`;
                return;
            }
            const json = await res.json() as { data?: HealthStatus } & HealthStatus;
            this._health = json.data ?? json;
        } catch (err) {
            this._healthError = err instanceof Error ? err.message : 'Failed to fetch';
        } finally {
            this._healthLoading = false;
        }
    }

    private _navigate(path: string): void {
        window.history.pushState({}, '', path);
        window.dispatchEvent(new PopStateEvent('popstate'));
    }

    private _healthDotClass(): string {
        if (this._healthLoading) return 'loading';
        if (this._healthError) return 'error';
        const status = this._health?.status?.toLowerCase() ?? '';
        if (status === 'ok' || status === 'healthy') return 'ok';
        if (status === 'degraded') return 'degraded';
        return 'error';
    }

    private _healthLabel(): string {
        if (this._healthLoading) return 'Checking...';
        if (this._healthError) return `Error: ${this._healthError}`;
        return this._health?.status ?? 'Unknown';
    }

    override render(): TemplateResult {
        return html`
            <h1>Platform Administration</h1>
            <p class="subtitle">Manage platform-wide settings, invitations, and system health.</p>

            <div class="grid">
                <div class="card card-link" @click=${(): void => { this._navigate('/admin/invitations'); }}>
                    <div class="card-header">
                        <span class="card-icon" aria-hidden="true">
                            <svg viewBox="0 0 24 24"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="8.5" cy="7" r="4"/><path d="M20 8v6M23 11h-6"/></svg>
                        </span>
                        <span class="card-title">Invitations</span>
                    </div>
                    <p class="card-desc">Manage user invitations and access to the platform.</p>
                </div>

                <div class="card card-link" @click=${(): void => { this._navigate('/admin/shadow'); }}>
                    <div class="card-header">
                        <span class="card-icon" aria-hidden="true">
                            <svg viewBox="0 0 24 24"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
                        </span>
                        <span class="card-title">Shadow Mode</span>
                    </div>
                    <p class="card-desc">View FutureShade tribunal logs and shadow documentation.</p>
                </div>
            </div>

            <div class="card">
                <div class="health-section">
                    <h2>System Health</h2>
                    <div class="health-status">
                        <span class="health-dot ${this._healthDotClass()}"></span>
                        <span>${this._healthLabel()}</span>
                    </div>
                    ${this._health?.checks ? html`
                        <div class="health-checks">
                            ${Object.entries(this._health.checks).map(([name, check]) => html`
                                <div class="health-check-item">
                                    <span class="health-dot ${check.status?.toLowerCase() === 'ok' || check.status?.toLowerCase() === 'healthy' ? 'ok' : 'error'}"></span>
                                    <span>${name}</span>
                                    ${check.message ? html`<span>- ${check.message}</span>` : nothing}
                                </div>
                            `)}
                        </div>
                    ` : nothing}
                </div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-admin-dashboard': FBAdminDashboard;
    }
}
