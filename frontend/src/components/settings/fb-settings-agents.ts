/**
 * fb-settings-agents — Agent Settings admin page.
 *
 * Route: /admin/agents
 * Shows: Per-agent configuration (enable/disable, thresholds, AI provider)
 * Access: Admin only
 */
import { html, css, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { api } from '../../services/api';
import type { AgentSettingsResponse } from '../../services/api';

@customElement('fb-settings-agents')
export class FBSettingsAgents extends FBElement {
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
                display: flex;
                align-items: center;
                gap: 8px;
            }

            .card-icon { font-size: 18px; }

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
            }

            select, input[type="number"], input[type="text"] {
                background: var(--fb-surface-2, #1E2029);
                border: 1px solid var(--fb-border, rgba(255,255,255,0.05));
                border-radius: 6px;
                padding: 6px 10px;
                color: var(--fb-text-primary, #F0F0F5);
                font-size: 13px;
                min-width: 120px;
            }

            input[type="number"] { width: 80px; }

            .toggle {
                position: relative;
                width: 40px;
                height: 22px;
                background: var(--fb-surface-2, #1E2029);
                border-radius: 11px;
                cursor: pointer;
                border: 1px solid var(--fb-border, rgba(255,255,255,0.05));
                transition: background 0.2s;
            }

            .toggle.on {
                background: var(--fb-accent, #00FFA3);
            }

            .toggle::after {
                content: '';
                position: absolute;
                top: 2px;
                left: 2px;
                width: 16px;
                height: 16px;
                background: white;
                border-radius: 50%;
                transition: transform 0.2s;
            }

            .toggle.on::after {
                transform: translateX(18px);
            }

            .actions {
                display: flex;
                gap: 12px;
                margin-top: 24px;
                padding-top: 20px;
                border-top: 1px solid var(--fb-border, rgba(255,255,255,0.05));
            }

            .btn {
                padding: 10px 20px;
                border-radius: 8px;
                font-size: 14px;
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
        `,
    ];

    @state() private _loading = true;
    @state() private _saving = false;
    @state() private _dirty = false;
    @state() private _success = '';
    @state() private _error = '';
    @state() private _settings: AgentSettingsResponse | null = null;

    override connectedCallback() {
        super.connectedCallback();
        this._loadSettings();
    }

    private async _loadSettings() {
        this._loading = true;
        try {
            this._settings = await api.settings.getAgents();
            this._dirty = false;
        } catch (err) {
            console.warn('[FBSettingsAgents] Failed to load:', err);
        } finally {
            this._loading = false;
        }
    }

    private _markDirty() { this._dirty = true; this.requestUpdate(); }

    private async _handleSave() {
        if (!this._settings) return;
        this._saving = true;
        this._error = '';
        this._success = '';
        try {
            this._settings = await api.settings.updateAgents(this._settings);
            this._dirty = false;
            this._success = 'Agent settings saved';
            setTimeout(() => { this._success = ''; }, 3000);
        } catch (err) {
            this._error = err instanceof Error ? err.message : 'Failed to save';
        } finally {
            this._saving = false;
        }
    }

    override render() {
        if (this._loading || !this._settings) {
            return html`<div class="header"><div class="title">Agent Settings</div></div><p style="color: var(--fb-text-secondary)">Loading...</p>`;
        }

        const s = this._settings;

        return html`
            <div class="header">
                <div class="title">Agent Settings</div>
                <div class="subtitle">Configure AI agent behavior and thresholds</div>
            </div>

            ${this._success ? html`<div class="message message-success">${this._success}</div>` : nothing}
            ${this._error ? html`<div class="message message-error">${this._error}</div>` : nothing}

            <div class="card">
                <div class="card-title"><span class="card-icon">📋</span> Daily Focus Agent</div>
                <div class="form-row">
                    <span class="form-label">Enabled</span>
                    <div class="toggle ${s.daily_focus.enabled ? 'on' : ''}"
                         @click=${() => { s.daily_focus.enabled = !s.daily_focus.enabled; this._markDirty(); }}></div>
                </div>
                <div class="form-row">
                    <span class="form-label">Run Time</span>
                    <input type="text" .value=${s.daily_focus.run_time}
                           @input=${(e: Event) => { s.daily_focus.run_time = (e.target as HTMLInputElement).value; this._markDirty(); }} />
                </div>
                <div class="form-row">
                    <span class="form-label">AI Provider</span>
                    <select .value=${s.daily_focus.ai_provider}
                            @change=${(e: Event) => { s.daily_focus.ai_provider = (e.target as HTMLSelectElement).value; this._markDirty(); }}>
                        <option value="claude">Claude Opus 4.6</option>
                        <option value="gemini">Gemini Flash</option>
                    </select>
                </div>
                <div class="form-row">
                    <span class="form-label">Max Focus Cards</span>
                    <input type="number" min="1" max="10" .value=${String(s.daily_focus.max_focus_cards)}
                           @input=${(e: Event) => { s.daily_focus.max_focus_cards = parseInt((e.target as HTMLInputElement).value) || 3; this._markDirty(); }} />
                </div>
            </div>

            <div class="card">
                <div class="card-title"><span class="card-icon">📦</span> Procurement Agent</div>
                <div class="form-row">
                    <span class="form-label">Warning Threshold (days)</span>
                    <input type="number" min="1" max="60" .value=${String(s.procurement.lead_time_warning_threshold)}
                           @input=${(e: Event) => { s.procurement.lead_time_warning_threshold = parseInt((e.target as HTMLInputElement).value) || 14; this._markDirty(); }} />
                </div>
                <div class="form-row">
                    <span class="form-label">Staging Buffer (days)</span>
                    <input type="number" min="1" max="14" .value=${String(s.procurement.staging_buffer_days)}
                           @input=${(e: Event) => { s.procurement.staging_buffer_days = parseInt((e.target as HTMLInputElement).value) || 2; this._markDirty(); }} />
                </div>
                <div class="form-row">
                    <span class="form-label">Weather Buffer (days)</span>
                    <input type="number" min="1" max="14" .value=${String(s.procurement.default_weather_buffer_days)}
                           @input=${(e: Event) => { s.procurement.default_weather_buffer_days = parseInt((e.target as HTMLInputElement).value) || 3; this._markDirty(); }} />
                </div>
            </div>

            <div class="card">
                <div class="card-title"><span class="card-icon">📱</span> Sub Liaison Agent</div>
                <div class="form-row">
                    <span class="form-label">Enabled</span>
                    <div class="toggle ${s.sub_liaison.enabled ? 'on' : ''}"
                         @click=${() => { s.sub_liaison.enabled = !s.sub_liaison.enabled; this._markDirty(); }}></div>
                </div>
                <div class="form-row">
                    <span class="form-label">Confirmation Window</span>
                    <select .value=${s.sub_liaison.confirmation_window}
                            @change=${(e: Event) => { s.sub_liaison.confirmation_window = (e.target as HTMLSelectElement).value; this._markDirty(); }}>
                        <option value="48h">48 hours</option>
                        <option value="72h">72 hours</option>
                        <option value="96h">96 hours</option>
                    </select>
                </div>
                <div class="form-row">
                    <span class="form-label">Auto-Resend After</span>
                    <select .value=${s.sub_liaison.auto_resend_after}
                            @change=${(e: Event) => { s.sub_liaison.auto_resend_after = (e.target as HTMLSelectElement).value; this._markDirty(); }}>
                        <option value="12h">12 hours</option>
                        <option value="24h">24 hours</option>
                        <option value="48h">48 hours</option>
                    </select>
                </div>
            </div>

            <div class="card">
                <div class="card-title"><span class="card-icon">💬</span> Chat Intelligence</div>
                <div class="form-row">
                    <span class="form-label">AI Provider</span>
                    <select .value=${s.chat.ai_provider}
                            @change=${(e: Event) => { s.chat.ai_provider = (e.target as HTMLSelectElement).value; this._markDirty(); }}>
                        <option value="claude">Claude Opus 4.6</option>
                        <option value="gemini">Gemini Flash</option>
                    </select>
                </div>
                <div class="form-row">
                    <span class="form-label">Max Tool Calls</span>
                    <input type="number" min="1" max="25" .value=${String(s.chat.max_tool_calls)}
                           @input=${(e: Event) => { s.chat.max_tool_calls = parseInt((e.target as HTMLInputElement).value) || 10; this._markDirty(); }} />
                </div>
            </div>

            ${this._dirty ? html`
                <div class="actions">
                    <button class="btn btn-primary" ?disabled=${this._saving} @click=${this._handleSave}>
                        ${this._saving ? 'Saving...' : 'Save Changes'}
                    </button>
                </div>
            ` : nothing}
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-settings-agents': FBSettingsAgents;
    }
}
