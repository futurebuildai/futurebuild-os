/**
 * fb-settings-brain — FB-Brain Connection + A2A Agent Management.
 *
 * Route: /admin/brain
 * Tabs: Active Agents | Integrations | Execution Logs
 * Access: Admin only
 *
 * Phase 18: Enhanced with tabbed interface per FRONTEND_SCOPE.md §15.1
 */
import { html, css, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { api } from '../../services/api';
import type { BrainConnectionResponse } from '../../services/api';
import type { ActiveAgentConnection } from '../../types/a2a';
import type { A2AExecutionLog } from '../../types/a2a';

type BrainTab = 'agents' | 'integrations' | 'logs';

@customElement('fb-settings-brain')
export class FBSettingsBrain extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                max-width: 720px;
                margin: 0 auto;
                padding: 32px 16px;
            }

            .header { margin-bottom: 24px; }

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

            /* Tab bar */
            .tab-bar {
                display: flex;
                gap: 4px;
                margin-bottom: 24px;
                padding: 4px;
                background: var(--fb-surface-1, #161821);
                border-radius: 10px;
                border: 1px solid var(--fb-border, rgba(255,255,255,0.05));
            }

            .tab-btn {
                flex: 1;
                padding: 10px 16px;
                border: none;
                border-radius: 8px;
                font-size: 13px;
                font-weight: 600;
                cursor: pointer;
                background: transparent;
                color: var(--fb-text-secondary, #8B8D98);
                transition: all 0.15s;
            }

            .tab-btn:hover { color: var(--fb-text-primary, #F0F0F5); }

            .tab-btn.active {
                background: var(--fb-surface-2, #1E2029);
                color: var(--fb-text-primary, #F0F0F5);
                border-bottom: 2px solid var(--fb-accent, #00FFA3);
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
                color: #0A0B10;
            }

            .btn-primary:hover:not(:disabled) { opacity: 0.9; }
            .btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }

            .btn-danger {
                background: rgba(239, 68, 68, 0.15);
                color: #F43F5E;
                border: 1px solid rgba(239, 68, 68, 0.2);
            }

            .btn-danger:hover { background: rgba(239, 68, 68, 0.25); }

            .btn-sm {
                padding: 4px 12px;
                font-size: 12px;
            }

            .btn-toggle {
                background: rgba(255,255,255,0.05);
                color: var(--fb-text-secondary, #8B8D98);
                border: 1px solid var(--fb-border, rgba(255,255,255,0.05));
            }

            .btn-toggle.active {
                background: rgba(0, 255, 163, 0.1);
                color: #00FFA3;
                border-color: rgba(0, 255, 163, 0.2);
            }

            .btn-toggle.paused {
                background: rgba(245, 158, 11, 0.1);
                color: #f59e0b;
                border-color: rgba(245, 158, 11, 0.2);
            }

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

            /* Agent list */
            .agent-item {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: 12px;
                background: var(--fb-surface-2, #1E2029);
                border-radius: 8px;
                font-size: 13px;
                margin-bottom: 8px;
            }

            .agent-info { display: flex; flex-direction: column; gap: 4px; }
            .agent-name { color: var(--fb-text-primary, #F0F0F5); font-weight: 600; }
            .agent-meta { color: var(--fb-text-secondary, #8B8D98); font-size: 12px; }

            /* Logs table */
            .logs-table {
                width: 100%;
                border-collapse: collapse;
                font-size: 12px;
            }

            .logs-table th {
                text-align: left;
                padding: 8px 10px;
                color: var(--fb-text-secondary, #8B8D98);
                font-weight: 500;
                border-bottom: 1px solid var(--fb-border, rgba(255,255,255,0.05));
            }

            .logs-table td {
                padding: 8px 10px;
                color: var(--fb-text-primary, #F0F0F5);
                border-bottom: 1px solid var(--fb-border, rgba(255,255,255,0.03));
            }

            .logs-table td.mono {
                font-family: monospace;
                font-size: 11px;
            }

            .badge {
                font-size: 11px;
                padding: 2px 8px;
                border-radius: 4px;
                font-weight: 600;
            }

            .badge-completed {
                background: rgba(0, 255, 163, 0.1);
                color: #00FFA3;
            }

            .badge-failed {
                background: rgba(244, 63, 94, 0.1);
                color: #F43F5E;
            }

            .badge-pending {
                background: rgba(245, 158, 11, 0.1);
                color: #f59e0b;
            }

            .empty-state {
                text-align: center;
                padding: 32px;
                color: var(--fb-text-secondary, #8B8D98);
                font-size: 14px;
            }
        `,
    ];

    @state() private _activeTab: BrainTab = 'integrations';
    @state() private _loading = true;
    @state() private _saving = false;
    @state() private _dirty = false;
    @state() private _success = '';
    @state() private _error = '';
    @state() private _conn: BrainConnectionResponse | null = null;
    @state() private _brainURL = '';
    @state() private _agents: ActiveAgentConnection[] = [];
    @state() private _logs: A2AExecutionLog[] = [];

    override connectedCallback() {
        super.connectedCallback();
        this._loadConnection();
        this._loadAgents();
        this._loadLogs();
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

    private async _loadAgents() {
        try {
            this._agents = await api.settings.getBrainAgents();
        } catch {
            // A2A service may not be configured yet
        }
    }

    private async _loadLogs() {
        try {
            this._logs = await api.settings.getBrainLogs(50);
        } catch {
            // A2A service may not be configured yet
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

    private async _toggleAgent(agentId: string, currentStatus: string) {
        try {
            if (currentStatus === 'active') {
                await api.settings.pauseBrainAgent(agentId);
            } else {
                await api.settings.resumeBrainAgent(agentId);
            }
            await this._loadAgents();
        } catch (err) {
            this._error = err instanceof Error ? err.message : 'Failed to toggle agent';
        }
    }

    override render() {
        if (this._loading || !this._conn) {
            return html`<div class="header"><div class="title">FB-Brain Connection</div></div><p style="color: var(--fb-text-secondary)">Loading...</p>`;
        }

        return html`
            <div class="header">
                <div class="title">FB-Brain Connection</div>
                <div class="subtitle">Manage cross-system integration with the FutureBuild ecosystem</div>
            </div>

            ${this._success ? html`<div class="message message-success">${this._success}</div>` : nothing}
            ${this._error ? html`<div class="message message-error">${this._error}</div>` : nothing}

            <div class="tab-bar">
                <button class="tab-btn ${this._activeTab === 'agents' ? 'active' : ''}"
                        @click=${() => { this._activeTab = 'agents'; }}>Active Agents</button>
                <button class="tab-btn ${this._activeTab === 'integrations' ? 'active' : ''}"
                        @click=${() => { this._activeTab = 'integrations'; }}>Integrations</button>
                <button class="tab-btn ${this._activeTab === 'logs' ? 'active' : ''}"
                        @click=${() => { this._activeTab = 'logs'; }}>Execution Logs</button>
            </div>

            ${this._activeTab === 'agents' ? this._renderAgentsTab() : nothing}
            ${this._activeTab === 'integrations' ? this._renderIntegrationsTab() : nothing}
            ${this._activeTab === 'logs' ? this._renderLogsTab() : nothing}
        `;
    }

    private _renderAgentsTab() {
        if (this._agents.length === 0) {
            return html`<div class="card"><div class="empty-state">No active agent connections. Configure agents in FB-Brain to see them here.</div></div>`;
        }

        return html`
            <div class="card">
                <div class="card-title">Active Agent Connections</div>
                ${this._agents.map(a => html`
                    <div class="agent-item">
                        <div class="agent-info">
                            <span class="agent-name">${a.agent_name}</span>
                            <span class="agent-meta">
                                ${a.agent_type} &middot;
                                ${a.execution_count} runs &middot;
                                ${a.error_count} errors
                                ${a.last_execution_at ? html` &middot; Last: ${new Date(a.last_execution_at).toLocaleString()}` : nothing}
                            </span>
                        </div>
                        <button class="btn btn-sm btn-toggle ${a.status}"
                                @click=${() => this._toggleAgent(a.id, a.status)}>
                            ${a.status === 'active' ? 'Pause' : 'Resume'}
                        </button>
                    </div>
                `)}
            </div>
        `;
    }

    private _renderIntegrationsTab() {
        const c = this._conn!;
        const statusClass = c.status === 'connected' ? 'connected' : c.status === 'connecting' ? 'connecting' : 'disconnected';

        return html`
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
                    <button class="btn btn-danger btn-sm" @click=${this._handleRegenerateKey}>Regenerate</button>
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

    private _renderLogsTab() {
        if (this._logs.length === 0) {
            return html`<div class="card"><div class="empty-state">No execution logs yet. Agent actions will appear here once connected.</div></div>`;
        }

        return html`
            <div class="card">
                <div class="card-title">Execution Logs (Last 50)</div>
                <table class="logs-table">
                    <thead>
                        <tr>
                            <th>Time</th>
                            <th>Action</th>
                            <th>Source / Target</th>
                            <th>Status</th>
                            <th>Duration</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${this._logs.map(l => html`
                            <tr>
                                <td class="mono">${new Date(l.executed_at).toLocaleString()}</td>
                                <td>${l.action_type}</td>
                                <td>${l.source_system} → ${l.target_system}</td>
                                <td><span class="badge ${this._statusBadgeClass(l.status)}">${l.status}</span></td>
                                <td class="mono">${l.duration_ms != null ? `${l.duration_ms}ms` : '—'}</td>
                            </tr>
                        `)}
                    </tbody>
                </table>
            </div>
        `;
    }

    private _statusBadgeClass(status: string): string {
        if (status === 'completed' || status === 'success') return 'badge-completed';
        if (status === 'failed' || status === 'error') return 'badge-failed';
        return 'badge-pending';
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-settings-brain': FBSettingsBrain;
    }
}
