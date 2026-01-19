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

    override render(): TemplateResult {
        return html`
            <div class="header">
                Agent Activity
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
