/**
 * fb-settings-brain — FB-Brain Connection settings page.
 *
 * Route: /admin/brain
 * Shows: Brain URL, integration key, status, connected platforms
 * Access: Admin only
 */
import { html, css, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { api } from '../../services/api';
import type { BrainConnectionResponse } from '../../services/api';

@customElement('fb-settings-brain')
export class FBSettingsBrain extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                max-width: 640px;
                margin: 0 auto;
                padding: 32px 16px;
            }

            .header { margin-bottom: 32px; }

            .title {
                font-size: 24px;
                font-weight: 700;
                color: var(--fb-text-primary, #F0F0F5);
            }

            .subtitle {
                font-size: 14px;
                color: var(--fb-text-secondary, #8B8D98);
                margin-top: 4px;
            }

            .card {
                background: var(--fb-surface-1, #161821);
                border: 1px solid var(--fb-border, rgba(255,255,255,0.05));
                border-radius: 12px;
                padding: 24px;
                margin-bottom: 20px;
            }

            .card-title {
                font-size: 16px;
                font-weight: 600;
                color: var(--fb-text-primary, #F0F0F5);
                margin-bottom: 20px;
                padding-bottom: 12px;
                border-bottom: 1px solid var(--fb-border, rgba(255,255,255,0.05));
            }

            .status-row {
                display: flex;
                align-items: center;
                gap: 8px;
                margin-bottom: 16px;
            }

            .status-dot {
                width: 10px;
                height: 10px;
                border-radius: 50%;
            }

            .status-dot.connected { background: #00FFA3; }
            .status-dot.connecting { background: #f59e0b; }
            .status-dot.disconnected { background: #ef4444; }

            .status-label {
                font-size: 14px;
                font-weight: 500;
                text-transform: capitalize;
            }

            .form-row {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: 12px 0;
                border-bottom: 1px solid var(--fb-border, rgba(255,255,255,0.03));
            }

            .form-row:last-child { border-bottom: none; }

            .form-label {
                font-size: 13px;
                color: var(--fb-text-secondary, #8B8D98);
            }

            .form-value {
                font-size: 13px;
                color: var(--fb-text-primary, #F0F0F5);
                font-family: monospace;
            }

            input[type="text"] {
                background: var(--fb-surface-2, #1E2029);
                border: 1px solid var(--fb-border, rgba(255,255,255,0.05));
                border-radius: 6px;
                padding: 6px 10px;
                color: var(--fb-text-primary, #F0F0F5);
                font-size: 13px;
                width: 300px;
            }

            .btn {
                padding: 8px 16px;
                border-radius: 6px;
                font-size: 13px;
                font-weight: 600;
                cursor: pointer;
                border: none;
                transition: opacity 0.15s;
            }

            .btn-primary {
                background: var(--fb-accent, #00FFA3);
                color: #fff;
            }

            .btn-primary:hover:not(:disabled) { opacity: 0.9; }
            .btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }

            .btn-danger {
                background: rgba(239, 68, 68, 0.15);
                color: #F43F5E;
                border: 1px solid rgba(239, 68, 68, 0.2);
            }

            .btn-danger:hover { background: rgba(239, 68, 68, 0.25); }

            .actions {
                display: flex;
                gap: 12px;
                margin-top: 24px;
                padding-top: 20px;
                border-top: 1px solid var(--fb-border, rgba(255,255,255,0.05));
            }

            .platform-list {
                display: flex;
                flex-direction: column;
                gap: 8px;
                margin-top: 8px;
            }

            .platform-item {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: 10px 12px;
                background: var(--fb-surface-2, #1E2029);
                border-radius: 8px;
                font-size: 13px;
            }

            .platform-name { color: var(--fb-text-primary, #F0F0F5); font-weight: 500; }

            .platform-status {
                font-size: 12px;
                padding: 2px 8px;
                border-radius: 4px;
            }

            .platform-status.active {
                background: rgba(0, 255, 163, 0.1);
                color: #00FFA3;
            }

            .platform-status.inactive {
                background: rgba(139, 141, 152, 0.1);
                color: #8B8D98;
            }

            .message {
                padding: 10px 14px;
                border-radius: 8px;
                font-size: 13px;
                margin-bottom: 16px;
            }

            .message-success {
                background: rgba(34, 197, 94, 0.1);
                color: #00FFA3;
            }

            .message-error {
                background: rgba(239, 68, 68, 0.1);
                color: #F43F5E;
            }

            .message-warning {
                background: rgba(245, 158, 11, 0.1);
                color: #f59e0b;
            }
        `,
    ];

    @state() private _loading = true;
    @state() private _saving = false;
    @state() private _dirty = false;
    @state() private _success = '';
    @state() private _error = '';
    @state() private _conn: BrainConnectionResponse | null = null;
    @state() private _brainURL = '';

    override connectedCallback() {
        super.connectedCallback();
        this._loadConnection();
    }

    private async _loadConnection() {
        this._loading = true;
        try {
            this._conn = await api.settings.getBrain();
            this._brainURL = this._conn.brain_url;
            this._dirty = false;
        } catch (err) {
            console.warn('[FBSettingsBrain] Failed to load:', err);
        } finally {
            this._loading = false;
        }
    }

    private async _handleSave() {
        this._saving = true;
        this._error = '';
        this._success = '';
        try {
            this._conn = await api.settings.updateBrain({ brain_url: this._brainURL });
            this._dirty = false;
            this._success = 'Brain connection updated';
            setTimeout(() => { this._success = ''; }, 3000);
        } catch (err) {
            this._error = err instanceof Error ? err.message : 'Failed to save';
        } finally {
            this._saving = false;
        }
    }

    private async _handleRegenerateKey() {
        try {
            const result = await api.settings.regenerateBrainKey();
            this._success = `New key generated: ${result.integration_key.substring(0, 12)}...`;
            this._loadConnection();
            setTimeout(() => { this._success = ''; }, 5000);
        } catch (err) {
            this._error = err instanceof Error ? err.message : 'Failed to regenerate key';
        }
    }

    override render() {
        if (this._loading || !this._conn) {
            return html`<div class="header"><div class="title">FB-Brain Connection</div></div><p style="color: var(--fb-text-secondary)">Loading...</p>`;
        }

        const c = this._conn;
        const statusClass = c.status === 'connected' ? 'connected' : c.status === 'connecting' ? 'connecting' : 'disconnected';

        return html`
            <div class="header">
                <div class="title">FB-Brain Connection</div>
                <div class="subtitle">Manage cross-system integration with the FutureBuild ecosystem</div>
            </div>

            ${this._success ? html`<div class="message message-success">${this._success}</div>` : nothing}
            ${this._error ? html`<div class="message message-error">${this._error}</div>` : nothing}

            <div class="card">
                <div class="card-title">Connection Status</div>
                <div class="status-row">
                    <span class="status-dot ${statusClass}"></span>
                    <span class="status-label">${c.status}</span>
                </div>
                ${c.last_sync_at ? html`
                    <div class="form-row">
                        <span class="form-label">Last Sync</span>
                        <span class="form-value">${new Date(c.last_sync_at).toLocaleString()}</span>
                    </div>
                ` : nothing}
            </div>

            <div class="card">
                <div class="card-title">Configuration</div>
                <div class="form-row">
                    <span class="form-label">Brain URL</span>
                    <input type="text" .value=${this._brainURL} placeholder="https://brain.futurebuild.app"
                           @input=${(e: Event) => { this._brainURL = (e.target as HTMLInputElement).value; this._dirty = true; }} />
                </div>
                <div class="form-row">
                    <span class="form-label">Integration Key</span>
                    <span class="form-value">${c.integration_key || 'Not generated'}</span>
                    <button class="btn btn-danger" @click=${this._handleRegenerateKey}>Regenerate</button>
                </div>

                ${this._dirty ? html`
                    <div class="actions">
                        <button class="btn btn-primary" ?disabled=${this._saving} @click=${this._handleSave}>
                            ${this._saving ? 'Saving...' : 'Save Connection'}
                        </button>
                    </div>
                ` : nothing}
            </div>

            ${c.platforms.length > 0 ? html`
                <div class="card">
                    <div class="card-title">Connected Platforms</div>
                    <div class="platform-list">
                        ${c.platforms.map(p => html`
                            <div class="platform-item">
                                <span class="platform-name">${p.name} (${p.type})</span>
                                <span class="platform-status ${p.status === 'active' ? 'active' : 'inactive'}">${p.status}</span>
                            </div>
                        `)}
                    </div>
                </div>
            ` : nothing}
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-settings-brain': FBSettingsBrain;
    }
}
