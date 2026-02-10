/**
 * fb-home-feed — Portfolio home feed view.
 * See FRONTEND_V2_SPEC.md §2.1
 *
 * The default authenticated view. Shows:
 * - Greeting banner with portfolio summary
 * - Urgency-sorted feed cards grouped by horizon
 * - Empty state for new users (redirect to onboarding)
 */
import { html, css, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { api } from '../../services/api';
import type { FeedCard, PortfolioSummary, FeedCardHorizon } from '../../types/feed';

interface GroupedCards {
    today: FeedCard[];
    this_week: FeedCard[];
    horizon: FeedCard[];
}

const HORIZON_LABELS: Record<FeedCardHorizon, string> = {
    today: 'Today',
    this_week: 'This Week',
    horizon: 'On the Horizon',
};

@customElement('fb-home-feed')
export class FBHomeFeed extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
                max-width: 720px;
                margin: 0 auto;
                padding: 24px 16px 80px;
            }

            .greeting {
                font-size: 28px;
                font-weight: 700;
                color: var(--fb-text-primary, #e0e0e0);
                margin-bottom: 4px;
            }

            .summary {
                font-size: 14px;
                color: var(--fb-text-secondary, #a0a0b0);
                margin-bottom: 32px;
                line-height: 1.5;
            }

            .summary-stat {
                font-weight: 600;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .summary-alert {
                color: var(--fb-warning, #f59e0b);
                font-weight: 600;
            }

            .horizon-group {
                margin-bottom: 28px;
            }

            .horizon-label {
                font-size: 12px;
                font-weight: 600;
                text-transform: uppercase;
                letter-spacing: 0.8px;
                color: var(--fb-text-tertiary, #707080);
                margin-bottom: 12px;
                padding-left: 4px;
            }

            .cards {
                display: flex;
                flex-direction: column;
                gap: 12px;
            }

            .loading {
                display: flex;
                flex-direction: column;
                gap: 16px;
                margin-top: 24px;
            }

            .loading-card {
                height: 120px;
                border-radius: 12px;
            }

            .empty {
                text-align: center;
                padding: 80px 24px;
            }

            .empty-title {
                font-size: 24px;
                font-weight: 700;
                color: var(--fb-text-primary, #e0e0e0);
                margin-bottom: 8px;
            }

            .empty-body {
                font-size: 15px;
                color: var(--fb-text-secondary, #a0a0b0);
                margin-bottom: 24px;
                line-height: 1.5;
            }

            .empty-cta {
                display: inline-flex;
                align-items: center;
                padding: 12px 28px;
                border-radius: 8px;
                background: var(--fb-accent, #6366f1);
                color: #fff;
                font-size: 15px;
                font-weight: 600;
                cursor: pointer;
                border: none;
                transition: opacity 0.15s ease;
            }

            .empty-cta:hover {
                opacity: 0.9;
            }

            .error {
                padding: 16px;
                background: var(--fb-surface-1, #1a1a2e);
                border: 1px solid #ef4444;
                border-radius: 8px;
                color: #ef4444;
                font-size: 14px;
                margin-top: 24px;
            }

            @media (max-width: 768px) {
                :host {
                    padding: 16px 12px 80px;
                }

                .greeting {
                    font-size: 22px;
                }
            }
        `,
    ];

    @state() private _greeting = '';
    @state() private _summary: PortfolioSummary | null = null;
    @state() private _cards: FeedCard[] = [];
    @state() private _loading = true;
    @state() private _error: string | null = null;
    @state() private _filterProjectId: string | null = null;

    override connectedCallback() {
        super.connectedCallback();
        this._loadFeed();

        // Listen for filter changes from top bar
        this.addEventListener('fb-filter-change', this._onFilterChange as EventListener);
    }

    override disconnectedCallback() {
        super.disconnectedCallback();
        this.removeEventListener('fb-filter-change', this._onFilterChange as EventListener);
    }

    private _onFilterChange = (e: CustomEvent<{ projectId: string | null }>) => {
        this._filterProjectId = e.detail.projectId;
        this._loadFeed();
    };

    async _loadFeed() {
        this._loading = true;
        this._error = null;
        try {
            const resp = await api.portfolio.getFeed(
                this._filterProjectId ?? undefined
            );
            this._greeting = resp.greeting;
            this._summary = resp.summary;
            this._cards = resp.cards;
        } catch (err) {
            this._error = err instanceof Error ? err.message : 'Failed to load feed';
        } finally {
            this._loading = false;
        }
    }

    /** Reload feed — callable from parent */
    public reload() {
        this._loadFeed();
    }

    /** Set project filter — callable from parent */
    public setFilter(projectId: string | null) {
        this._filterProjectId = projectId;
        this._loadFeed();
    }

    private _groupCards(): GroupedCards {
        const groups: GroupedCards = { today: [], this_week: [], horizon: [] };
        for (const card of this._cards) {
            const bucket = groups[card.horizon];
            if (bucket) {
                bucket.push(card);
            }
        }
        return groups;
    }

    private _handleCardAction(e: CustomEvent<{ cardId: string; actionId: string; projectId: string }>) {
        const { cardId, actionId } = e.detail;

        if (actionId === 'dismiss') {
            api.portfolio.dismissCard(cardId).then(() => {
                this._cards = this._cards.filter((c) => c.id !== cardId);
            }).catch(() => { /* silently fail, card stays */ });
            return;
        }

        if (actionId === 'snooze') {
            api.portfolio.snoozeCard(cardId, 24).then(() => {
                this._cards = this._cards.filter((c) => c.id !== cardId);
            }).catch(() => { /* silently fail */ });
            return;
        }

        // Delegate other actions up
        api.portfolio.executeAction(cardId, actionId).catch(() => {
            // TODO: show toast on failure
        });
    }

    private _handleStartProject() {
        this.emit('fb-navigate', { view: 'onboard' });
    }

    private _renderSummary() {
        if (!this._summary) return nothing;
        const s = this._summary;

        const parts: string[] = [];
        if (s.active_project_count > 0) {
            parts.push(`${s.active_project_count} active project${s.active_project_count > 1 ? 's' : ''}`);
        }
        if (s.total_tasks > 0) {
            parts.push(`${s.total_tasks} tasks`);
        }

        return html`
            <div class="summary">
                ${parts.length > 0
                    ? html`<span class="summary-stat">${parts.join(' \u00B7 ')}</span>`
                    : nothing}
                ${s.critical_alerts > 0
                    ? html` \u00B7 <span class="summary-alert">${s.critical_alerts} need${s.critical_alerts > 1 ? '' : 's'} attention</span>`
                    : nothing}
            </div>
        `;
    }

    private _renderLoading() {
        return html`
            <div class="loading">
                <div class="loading-card skeleton"></div>
                <div class="loading-card skeleton"></div>
                <div class="loading-card skeleton"></div>
            </div>
        `;
    }

    private _renderEmpty() {
        return html`
            <div class="empty">
                <div class="empty-title">Your engine is ready</div>
                <div class="empty-body">
                    Add your first project and I'll build your schedule,<br />
                    track your subs, and watch your deadlines.
                </div>
                <button class="empty-cta" @click=${this._handleStartProject}>
                    Start a project
                </button>
            </div>
        `;
    }

    override render() {
        if (this._loading) {
            return html`
                <div class="greeting skeleton skeleton-text" style="width: 200px; height: 32px;"></div>
                ${this._renderLoading()}
            `;
        }

        if (this._error) {
            return html`
                <div class="greeting">Something went wrong</div>
                <div class="error">${this._error}</div>
            `;
        }

        if (this._cards.length === 0 && (!this._summary || this._summary.active_project_count === 0)) {
            return html`
                <div class="greeting">${this._greeting}</div>
                ${this._renderEmpty()}
            `;
        }

        const groups = this._groupCards();
        const horizons: FeedCardHorizon[] = ['today', 'this_week', 'horizon'];

        return html`
            <div class="greeting">${this._greeting}</div>
            ${this._renderSummary()}

            ${this._cards.length === 0
                ? html`
                      <div class="empty">
                          <div class="empty-body">All clear. No items need your attention right now.</div>
                      </div>
                  `
                : horizons.map((h) => {
                      const cards = groups[h];
                      if (!cards || cards.length === 0) return nothing;
                      return html`
                          <div class="horizon-group">
                              <div class="horizon-label">${HORIZON_LABELS[h]}</div>
                              <div class="cards" @fb-card-action=${this._handleCardAction}>
                                  ${cards.map(
                                      (card) => html`
                                          <fb-feed-card .card=${card}></fb-feed-card>
                                      `
                                  )}
                              </div>
                          </div>
                      `;
                  })}
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-home-feed': FBHomeFeed;
    }
}
