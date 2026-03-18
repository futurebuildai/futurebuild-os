/**
 * Tribunal Case Detail Component
 * Shows detailed information about a tribunal decision including model votes.
 * See SHADOW_VIEWER_specs.md Section 5.1
 */

import { html, css, type TemplateResult, type PropertyValues } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { store } from '../../store/store';
import { futureShadeService } from '../../futureshade/services/api';
import type { DecisionDetail, ModelVote } from '../../futureshade/types';

@customElement('tribunal-case-detail')
export class TribunalCaseDetail extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: flex;
                flex-direction: column;
                height: 100%;
                padding: 16px;
                background: #0d0d15;
                overflow-y: auto;
            }

            .header {
                margin-bottom: 16px;
            }

            .header h3 {
                margin: 0 0 8px 0;
                font-size: 16px;
                font-weight: 600;
                color: #e5e7eb;
            }

            .consensus-score {
                font-size: 24px;
                font-weight: 600;
                color: #33FFB8;
            }

            .context {
                margin-bottom: 16px;
                padding: 12px;
                background: #161821;
                border-radius: 6px;
                color: #9ca3af;
                font-size: 14px;
                line-height: 1.5;
            }

            .votes-section h4 {
                margin: 0 0 12px 0;
                font-size: 14px;
                font-weight: 600;
                color: #9ca3af;
            }

            .vote-card {
                background: #161821;
                border-radius: 8px;
                padding: 12px;
                margin-bottom: 8px;
            }

            .vote-header {
                display: flex;
                justify-content: space-between;
                align-items: center;
                margin-bottom: 8px;
            }

            .model-name {
                font-weight: 600;
                color: #e5e7eb;
                font-size: 14px;
            }

            .vote-badge {
                padding: 2px 8px;
                border-radius: 4px;
                font-size: 11px;
                font-weight: 600;
            }

            .vote-badge.yea {
                background: #166534;
                color: #4ade80;
            }

            .vote-badge.nay {
                background: #7f1d1d;
                color: #f87171;
            }

            .vote-badge.abstain {
                background: #374151;
                color: #9ca3af;
            }

            .reasoning {
                font-size: 13px;
                color: #9ca3af;
                line-height: 1.5;
            }

            .reasoning a {
                color: #33FFB8;
                text-decoration: none;
            }

            .reasoning a:hover {
                text-decoration: underline;
            }

            .meta {
                display: flex;
                gap: 12px;
                margin-top: 8px;
                font-size: 12px;
                color: #6b7280;
            }

            .empty-state {
                display: flex;
                align-items: center;
                justify-content: center;
                height: 100%;
                color: #6b7280;
                font-size: 14px;
            }

            .loading {
                display: flex;
                align-items: center;
                justify-content: center;
                height: 100%;
                color: #6b7280;
            }
        `,
    ];

    @property({ type: String }) decisionId: string | null = null;

    @state() private _decision: DecisionDetail | null = null;
    @state() private _loading = false;

    override updated(changedProps: PropertyValues): void {
        if (changedProps.has('decisionId') && this.decisionId) {
            this._loadDecision();
        }
    }

    private async _loadDecision(): Promise<void> {
        if (!this.decisionId) return;

        this._loading = true;
        try {
            this._decision = await futureShadeService.getDecision(this.decisionId);
        } catch (e) {
            console.error('Failed to load decision:', e);
            this._decision = null;
        } finally {
            this._loading = false;
        }
    }

    private _handleDocLink(e: Event, path: string): void {
        e.preventDefault();
        store.actions.setShadowActiveView('docs');
        store.actions.selectDoc(path);
    }

    private _renderReasoning(reasoning: string): TemplateResult {
        // Convert markdown links to clickable doc links
        const linkRegex = /\[([^\]]+)\]\(([^)]+)\)/g;
        const parts: (string | TemplateResult)[] = [];
        let lastIndex = 0;
        let match;

        while ((match = linkRegex.exec(reasoning)) !== null) {
            if (match.index > lastIndex) {
                parts.push(reasoning.slice(lastIndex, match.index));
            }
            const text = match[1] ?? '';
            const href = match[2] ?? '';
            if (href.startsWith('docs/') || href.startsWith('specs/')) {
                parts.push(
                    html`<a href="#" @click=${(e: Event) => { this._handleDocLink(e, href); }}>${text}</a>`
                );
            } else {
                parts.push(html`<a href="${href}" target="_blank" rel="noopener">${text}</a>`);
            }
            lastIndex = match.index + match[0].length;
        }

        if (lastIndex < reasoning.length) {
            parts.push(reasoning.slice(lastIndex));
        }

        return html`${parts}`;
    }

    private _getVoteClass(vote: ModelVote['vote']): string {
        const normalized = vote.toLowerCase();
        if (normalized === 'yea') return 'vote-badge yea';
        if (normalized === 'nay') return 'vote-badge nay';
        return 'vote-badge abstain';
    }

    override render(): TemplateResult {
        if (this._loading) {
            return html`<div class="loading">Loading decision...</div>`;
        }

        if (!this._decision) {
            return html`<div class="empty-state">Select a decision to view details</div>`;
        }

        return html`
            <div class="header">
                <h3>${this._decision.case_id}</h3>
                <div class="consensus-score">
                    Consensus: ${(this._decision.consensus_score * 100).toFixed(0)}%
                </div>
            </div>

            <div class="context">${this._decision.context}</div>

            <div class="votes-section">
                <h4>Model Votes</h4>
                ${this._decision.votes.map(
                    (vote) => html`
                        <div class="vote-card">
                            <div class="vote-header">
                                <span class="model-name">${vote.model}</span>
                                <span class=${this._getVoteClass(vote.vote)}>${vote.vote}</span>
                            </div>
                            <div class="reasoning">${this._renderReasoning(vote.reasoning)}</div>
                            <div class="meta">
                                <span>${vote.latency_ms}ms</span>
                                <span>$${vote.cost_usd.toFixed(4)}</span>
                            </div>
                        </div>
                    `
                )}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'tribunal-case-detail': TribunalCaseDetail;
    }
}
