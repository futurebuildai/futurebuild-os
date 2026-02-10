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
import { customElement, property, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { api } from '../../services/api';
import { store, type ChatCardContext } from '../../store/store';
import { feedSSE } from '../../services/feed-sse';
import type { FeedCard, FeedSSEEvent, PortfolioSummary, FeedCardHorizon } from '../../types/feed';
import './fb-feed-section';
import './fb-greeting-banner';
import './fb-empty-home';

interface GroupedCards {
    today: FeedCard[];
    this_week: FeedCard[];
    horizon: FeedCard[];
}

// Horizon labels moved to fb-feed-section component

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

            .empty-cta:focus-visible {
                outline: 2px solid var(--fb-accent, #6366f1);
                outline-offset: 2px;
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

            .error-retry {
                display: inline-flex;
                align-items: center;
                margin-top: 12px;
                padding: 8px 20px;
                border-radius: 6px;
                border: 1px solid #ef4444;
                background: transparent;
                color: #ef4444;
                font-size: 13px;
                font-weight: 600;
                cursor: pointer;
                transition: all 0.15s ease;
            }

            .error-retry:hover {
                background: #ef4444;
                color: #fff;
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

    /**
     * External project filter — set by parent (e.g., app-shell project route).
     * When changed, triggers a filtered feed reload.
     */
    @property({ type: String, attribute: 'project-filter' }) projectFilter: string | null = null;

    @state() private _greeting = '';
    @state() private _summary: PortfolioSummary | null = null;
    @state() private _cards: FeedCard[] = [];
    @state() private _loading = true;
    @state() private _error: string | null = null;
    @state() private _filterProjectId: string | null = null;
    private _unsubSSE: (() => void) | null = null;

    override connectedCallback() {
        super.connectedCallback();
        // Apply external filter if provided
        if (this.projectFilter) {
            this._filterProjectId = this.projectFilter;
        }
        this._loadFeed();

        // Listen for filter changes from top bar
        this.addEventListener('fb-filter-change', this._onFilterChange as EventListener);

        // Subscribe to SSE feed stream for live updates
        this._unsubSSE = feedSSE.subscribe(this._handleSSEEvent);
        feedSSE.connect();
    }

    /** React to external projectFilter property changes */
    override willUpdate(changedProperties: Map<string, unknown>): void {
        if (changedProperties.has('projectFilter')) {
            const newFilter = this.projectFilter;
            if (newFilter !== this._filterProjectId) {
                this._filterProjectId = newFilter;
                this._loadFeed();
            }
        }
    }

    override disconnectedCallback() {
        super.disconnectedCallback();
        this.removeEventListener('fb-filter-change', this._onFilterChange as EventListener);
        if (this._unsubSSE) {
            this._unsubSSE();
            this._unsubSSE = null;
        }
        feedSSE.disconnect();
    }

    /** Handle live feed SSE events */
    private _handleSSEEvent = (event: FeedSSEEvent): void => {
        switch (event.type) {
            case 'card_added': {
                // Apply project filter if active
                if (this._filterProjectId && event.card.project_id !== this._filterProjectId) return;
                // Insert sorted by priority (lower = higher priority)
                const cards = [...this._cards];
                const idx = cards.findIndex((c) => c.priority > event.card.priority);
                if (idx === -1) {
                    cards.push(event.card);
                } else {
                    cards.splice(idx, 0, event.card);
                }
                this._cards = cards;
                break;
            }
            case 'card_updated': {
                this._cards = this._cards.map((c) =>
                    c.id === event.card.id ? event.card : c
                );
                break;
            }
            case 'card_removed': {
                this._cards = this._cards.filter((c) => c.id !== event.card_id);
                break;
            }
        }
    };

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
        const { cardId, actionId, projectId } = e.detail;

        // Client-side navigation actions — no API call needed
        switch (actionId) {
            case 'view_briefing':
            case 'view_details':
                this.emit('fb-navigate', { view: 'project', id: projectId });
                return;
            case 'view_schedule':
                this.emit('fb-navigate', { view: 'project-schedule', id: projectId });
                return;
            case 'add_contacts':
                this.emit('fb-navigate', { view: 'contacts' });
                return;
            case 'tell_me_more': {
                // Find the card to get full context
                const card = this._cards.find((c) => c.id === cardId);
                if (card) {
                    const ctx: ChatCardContext = {
                        cardId: card.id,
                        cardType: card.card_type,
                        headline: card.headline,
                        body: card.body,
                        consequence: card.consequence ?? '',
                        projectId: card.project_id,
                        projectName: card.project_name ?? '',
                        taskId: card.task_id ?? '',
                    };
                    store.actions.setChatCardContext(ctx);
                    store.actions.setActiveProject(card.project_id);
                    this.emit('fb-navigate', { view: 'project-chat', projectId: card.project_id });
                }
                return;
            }
        }

        // Dismiss — optimistic removal via dedicated endpoint
        if (actionId === 'dismiss') {
            this._cards = this._cards.filter((c) => c.id !== cardId);
            api.portfolio.dismissCard(cardId).catch(() => {
                this._loadFeed(); // Reload on failure
            });
            return;
        }

        // Snooze — optimistic removal via dedicated endpoint
        if (actionId === 'snooze') {
            this._cards = this._cards.filter((c) => c.id !== cardId);
            api.portfolio.snoozeCard(cardId, 24).catch(() => {
                this._loadFeed();
            });
            return;
        }

        // All other actions — call executeAction and handle response
        api.portfolio.executeAction(cardId, actionId).then((resp) => {
            if (resp.effect === 'dismiss') {
                this._cards = this._cards.filter((c) => c.id !== cardId);
            }
            if (resp.effect === 'navigate' && resp.navigate_to) {
                this.emit('fb-navigate', { path: resp.navigate_to });
            }
            if (resp.message) {
                this.emit('fb-toast', { message: resp.message });
            }
        }).catch(() => {
            this.emit('fb-toast', { message: 'Action failed. Please try again.', type: 'error' });
        });
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

    // _renderEmpty replaced by fb-empty-home component

    override render() {
        if (this._loading) {
            return html`
                <main role="main" aria-label="Portfolio Feed" aria-busy="true">
                    <fb-greeting-banner loading></fb-greeting-banner>
                    ${this._renderLoading()}
                </main>
            `;
        }

        if (this._error) {
            return html`
                <main role="main" aria-label="Portfolio Feed">
                    <fb-greeting-banner greeting="Something went wrong"></fb-greeting-banner>
                    <div class="error">
                        ${this._error}
                        <br />
                        <button class="error-retry" @click=${() => this._loadFeed()}>Retry</button>
                    </div>
                </main>
            `;
        }

        // No projects: show full-screen empty state
        if (this._cards.length === 0 && (!this._summary || this._summary.active_project_count === 0)) {
            return html`<fb-empty-home></fb-empty-home>`;
        }

        const groups = this._groupCards();
        const horizons: FeedCardHorizon[] = ['today', 'this_week', 'horizon'];

        return html`
            <main role="main" aria-label="Portfolio Feed">
                <fb-greeting-banner
                    greeting=${this._greeting}
                    .summary=${this._summary}
                ></fb-greeting-banner>

                ${this._cards.length === 0
                    ? html`
                          <div class="empty" role="status">
                              <div class="empty-body">All clear. No items need your attention right now.</div>
                          </div>
                      `
                    : horizons.map((h) => {
                          const cards = groups[h];
                          if (!cards || cards.length === 0) return nothing;
                          return html`
                              <fb-feed-section
                                  horizon=${h}
                                  card-count=${cards.length}
                                  @fb-card-action=${this._handleCardAction}
                              >
                                  ${cards.map(
                                      (card) => html`
                                          <fb-feed-card .card=${card} role="article"></fb-feed-card>
                                      `
                                  )}
                              </fb-feed-section>
                          `;
                      })}
            </main>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-home-feed': FBHomeFeed;
    }
}
