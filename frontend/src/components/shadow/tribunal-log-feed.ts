/**
 * Tribunal Log Feed Component
 * Lists tribunal decisions with status badges and filtering.
 * See SHADOW_VIEWER_specs.md Section 5.1
 */

import { html, css, type TemplateResult } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';
import { futureShadeService } from '../../futureshade/services/api';
import type { DecisionSummary, DecisionStatus, ListDecisionsFilter } from '../../futureshade/types';

@customElement('tribunal-log-feed')
export class TribunalLogFeed extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: flex;
                flex-direction: column;
                height: 100%;
                background: #0a0a12;
            }

            .header {
                padding: 16px;
                border-bottom: 1px solid #161821;
            }

            .header h2 {
                margin: 0 0 12px 0;
                font-size: 18px;
                font-weight: 600;
                color: #e5e7eb;
            }

            .filters {
                display: flex;
                gap: 8px;
                flex-wrap: wrap;
            }

            .filter-input {
                padding: 6px 10px;
                background: #161821;
                border: 1px solid rgba(255,255,255,0.05);
                border-radius: 4px;
                color: #F0F0F5;
                font-size: 13px;
                font-family: inherit;
            }

            .filter-input:focus {
                outline: none;
                border-color: #00CC82;
            }

            .list {
                flex: 1;
                overflow-y: auto;
            }

            .item {
                display: flex;
                align-items: center;
                gap: 12px;
                padding: 12px 16px;
                border-bottom: 1px solid #161821;
                cursor: pointer;
                transition: background 0.1s ease;
            }

            .item:hover {
                background: #161821;
            }

            .item.selected {
                background: #1a1a3e;
            }

            .status-badge {
                padding: 2px 8px;
                border-radius: 4px;
                font-size: 11px;
                font-weight: 600;
                text-transform: uppercase;
                flex-shrink: 0;
            }

            .status-badge.approved {
                background: #166534;
                color: #4ade80;
            }

            .status-badge.rejected {
                background: #7f1d1d;
                color: #f87171;
            }

            .status-badge.conflict {
                background: #854d0e;
                color: #fbbf24;
            }

            .status-badge.pending {
                background: #374151;
                color: #9ca3af;
            }

            .context {
                flex: 1;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
                color: #e5e7eb;
                font-size: 14px;
            }

            .timestamp {
                color: #6b7280;
                font-size: 12px;
                flex-shrink: 0;
            }

            .empty-state {
                padding: 32px 16px;
                text-align: center;
                color: #6b7280;
            }

            .loading {
                padding: 16px;
                text-align: center;
                color: #6b7280;
            }
        `,
    ];

    @state() private _decisions: DecisionSummary[] = [];
    @state() private _loading = false;
    @state() private _filter: ListDecisionsFilter = { limit: 50 };
    @state() private _selectedId: string | null = null;

    override connectedCallback(): void {
        super.connectedCallback();
        this._loadDecisions();
    }

    private async _loadDecisions(): Promise<void> {
        this._loading = true;
        try {
            const response = await futureShadeService.listDecisions(this._filter);
            this._decisions = response.decisions;
        } catch (e) {
            console.error('Failed to load decisions:', e);
        } finally {
            this._loading = false;
        }
    }

    private _handleSelect(id: string): void {
        this._selectedId = id;
        store.actions.selectDecision(id);
    }

    private _handleStatusFilter(e: Event): void {
        const value = (e.target as HTMLSelectElement).value;
        const newFilter = { ...this._filter };
        if (value) {
            newFilter.status = value as DecisionStatus;
        } else {
            delete newFilter.status;
        }
        this._filter = newFilter;
        this._loadDecisions();
    }

    private _handleSearchInput(e: Event): void {
        const value = (e.target as HTMLInputElement).value;
        const newFilter = { ...this._filter };
        if (value) {
            newFilter.search = value;
        } else {
            delete newFilter.search;
        }
        this._filter = newFilter;
        // Debounce would be ideal here, but keeping it simple for now
        this._loadDecisions();
    }

    private _getStatusClass(status: DecisionStatus): string {
        const normalized = status.toLowerCase();
        if (normalized === 'approved') return 'status-badge approved';
        if (normalized === 'rejected') return 'status-badge rejected';
        if (normalized === 'conflict') return 'status-badge conflict';
        return 'status-badge pending';
    }

    private _formatTimestamp(iso: string): string {
        try {
            return new Date(iso).toLocaleString();
        } catch {
            return iso;
        }
    }

    override render(): TemplateResult {
        return html`
            <div class="header">
                <h2>Tribunal Decisions</h2>
                <div class="filters">
                    <select class="filter-input" @change=${this._handleStatusFilter}>
                        <option value="">All Status</option>
                        <option value="APPROVED">Approved</option>
                        <option value="REJECTED">Rejected</option>
                        <option value="CONFLICT">Conflict</option>
                    </select>
                    <input
                        type="text"
                        class="filter-input"
                        placeholder="Search..."
                        @input=${this._handleSearchInput}
                    />
                </div>
            </div>

            <div class="list">
                ${this._loading
                    ? html`<div class="loading">Loading decisions...</div>`
                    : this._decisions.length === 0
                      ? html`<div class="empty-state">No decisions found</div>`
                      : this._decisions.map(
                            (d) => html`
                                <div
                                    class="item ${this._selectedId === d.id ? 'selected' : ''}"
                                    @click=${() => { this._handleSelect(d.id); }}
                                >
                                    <span class=${this._getStatusClass(d.status)}>${d.status}</span>
                                    <span class="context">${d.context}</span>
                                    <span class="timestamp">${this._formatTimestamp(d.timestamp)}</span>
                                </div>
                            `
                        )}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'tribunal-log-feed': TribunalLogFeed;
    }
}
