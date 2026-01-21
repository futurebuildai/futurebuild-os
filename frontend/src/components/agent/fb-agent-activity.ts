import { html, css, TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { effect } from '@preact/signals-core';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';
import type { AgentActivity } from '../../store/types';

@customElement('fb-agent-activity')
export class FBAgentActivity extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                border-top: 1px solid var(--fb-border-light);
                padding: 0 var(--fb-spacing-md);
            }

            .header {
                display: flex;
                align-items: center;
                justify-content: space-between;
                padding: var(--fb-spacing-md) 0;
                font-size: var(--fb-text-xs);
                font-weight: 600;
                color: var(--fb-text-secondary);
                text-transform: uppercase;
                letter-spacing: 0.05em;
            }

            /* Step 60.2.2: DEV-only debug controls */
            .debug-controls {
                display: flex;
                gap: 4px;
            }

            .debug-btn {
                padding: 2px 6px;
                font-size: 10px;
                border: 1px solid var(--fb-border-light);
                border-radius: 4px;
                background: transparent;
                color: var(--fb-text-muted);
                cursor: pointer;
                transition: all 0.15s ease;
            }

            .debug-btn:hover {
                background: var(--fb-surface-hover);
                color: var(--fb-text-primary);
            }

            .debug-btn--danger:hover {
                background: var(--fb-error);
                color: white;
                border-color: var(--fb-error);
            }

            .activity-list {
                display: flex;
                flex-direction: column;
                gap: 12px;
                padding-bottom: var(--fb-spacing-lg);
            }

            .activity-item {
                display: grid;
                grid-template-columns: 24px 1fr;
                gap: 8px;
                font-size: var(--fb-text-sm);
            }

            .icon-wrapper {
                display: flex;
                align-items: flex-start;
                justify-content: center;
                padding-top: 2px;
            }

            /* Status Dots */
            .status-dot {
                width: 8px;
                height: 8px;
                border-radius: 50%;
                background: var(--fb-text-muted);
            }

            .status-running {
                background: var(--fb-success);
                box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.7);
                animation: pulse-green 2s infinite;
            }

            .status-completed {
                background: var(--fb-text-secondary);
            }

            .status-failed {
                background: var(--fb-error);
            }

            @keyframes pulse-green {
                0% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.7); }
                70% { transform: scale(1); box-shadow: 0 0 0 6px rgba(34, 197, 94, 0); }
                100% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(34, 197, 94, 0); }
            }

            .content {
                display: flex;
                flex-direction: column;
                gap: 2px;
            }

            .action {
                font-weight: 500;
                color: var(--fb-text-primary);
            }

            .description {
                font-size: 11px;
                color: var(--fb-text-muted);
                line-height: 1.4;
            }

            .timestamp {
                font-size: 10px;
                color: var(--fb-text-muted);
                opacity: 0.6;
                margin-top: 2px;
            }

            .empty-state {
                font-size: var(--fb-text-xs);
                color: var(--fb-text-muted);
                text-align: center;
                padding: var(--fb-spacing-lg) 0;
                font-style: italic;
            }
        `
    ];

    @state() private _activities: AgentActivity[] = [];
    private _disposeEffects: (() => void)[] = [];

    override connectedCallback(): void {
        super.connectedCallback();

        this._disposeEffects.push(
            effect(() => {
                // Show last 5 items to keep it compact
                this._activities = store.agentActivity$.value.slice(0, 5);
            })
        );
    }

    override disconnectedCallback(): void {
        this._disposeEffects.forEach((d) => { d(); });
        this._disposeEffects = [];
        super.disconnectedCallback();
    }

    private _formatTime(isoString: string): string {
        try {
            return new Date(isoString).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
        } catch {
            return '';
        }
    }

    // Step 60.2.2: Debug button handlers (arrow functions to satisfy unbound-method lint)
    private _handleFlood = (): void => {
        const loadTest = (window as unknown as { fb?: { loadTest?: { flood: (n: number) => void } } }).fb?.loadTest;
        if (loadTest) {
            loadTest.flood(1000);
        } else {
            console.warn('[FBAgentActivity] LoadTestService not available');
        }
    };

    private _handleStream = (): void => {
        const loadTest = (window as unknown as { fb?: { loadTest?: { stream: (n: number) => void } } }).fb?.loadTest;
        if (loadTest) {
            loadTest.stream(20);
        } else {
            console.warn('[FBAgentActivity] LoadTestService not available');
        }
    };

    private _handleStop = (): void => {
        const loadTest = (window as unknown as { fb?: { loadTest?: { stop: () => void } } }).fb?.loadTest;
        if (loadTest) {
            loadTest.stop();
        } else {
            console.warn('[FBAgentActivity] LoadTestService not available');
        }
    };

    /** Check if running in DEV mode */
    private _isDevMode(): boolean {
        return (import.meta as unknown as { env?: { DEV?: boolean } }).env?.DEV === true;
    }

    /** Renders DEV-only debug controls */
    private _renderDebugControls(): TemplateResult | null {
        // Only show in DEV mode
        if (!this._isDevMode()) return null;

        return html`
            <div class="debug-controls">
                <button
                    class="debug-btn"
                    id="debug-flood-btn"
                    @click=${this._handleFlood}
                    title="Inject 1000 messages"
                >⚡ Flood 1000</button>
                <button
                    class="debug-btn"
                    id="debug-stream-btn"
                    @click=${this._handleStream}
                    title="Stream 20 msg/sec"
                >🌊 Stream 20/s</button>
                <button
                    class="debug-btn debug-btn--danger"
                    id="debug-stop-btn"
                    @click=${this._handleStop}
                    title="Stop all tests"
                >🛑 Stop</button>
            </div>
        `;
    }

    override render(): TemplateResult {
        return html`
            <div class="header">
                <span>Agent Activity</span>
                ${this._renderDebugControls()}
            </div>

            <div class="activity-list">
                ${this._activities.length === 0 ? html`
                    <div class="empty-state">Agent idle. Waiting for tasks...</div>
                ` : this._activities.map(activity => html`
                    <div class="activity-item">
                        <div class="icon-wrapper">
                            <div class="status-dot status-${activity.status}"></div>
                        </div>
                        <div class="content">
                            <div class="action">${activity.action}</div>
                            <div class="description">${activity.description}</div>
                            <div class="timestamp">${this._formatTime(activity.timestamp)}</div>
                        </div>
                    </div>
                `)}
            </div>
        `;
    }
}
