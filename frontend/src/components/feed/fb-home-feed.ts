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
import { store, type ChatCardContext } from '../../store/store';
import { feedSSE } from '../../services/feed-sse';
import { api } from '../../services/api';
import type { FeedCard, FeedSSEEvent, PortfolioSummary, FeedCardHorizon } from '../../types/feed';
import { effect } from '@preact/signals-core';
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

            /* Sprint 5.1: Slide-in animation for new cards */
            @keyframes slideIn {
                from {
                    opacity: 0;
                    transform: translateY(-12px);
                }
                to {
                    opacity: 1;
                    transform: translateY(0);
                }
            }

            fb-feed-card.new-card {
                animation: slideIn 0.35s ease-out;
            }

            /* Sprint 5.1: SSE connection status indicator */
            .connection-status {
                display: inline-flex;
                align-items: center;
                gap: 6px;
                font-size: 11px;
                color: var(--fb-text-tertiary, #707080);
                margin-bottom: 16px;
            }

            .status-dot {
                width: 7px;
                height: 7px;
                border-radius: 50%;
                background: #ef4444;
                transition: background 0.3s ease;
            }

            .status-dot.connected {
                background: #22c55e;
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

            .empty-state-container {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                padding: 60px 24px;
                text-align: center;
                color: var(--fb-text-secondary, #a0a0b0);
                gap: 32px;
                min-height: 400px;
            }

            .widget-clock {
                font-size: 64px;
                font-weight: 200;
                color: var(--fb-text-primary, #e0e0e0);
                font-variant-numeric: tabular-nums;
                letter-spacing: -2px;
            }

            .widget-date {
                font-size: 18px;
                font-weight: 500;
                color: var(--fb-accent, #6366f1);
                margin-top: -24px;
                margin-bottom: 24px;
                text-transform: uppercase;
                letter-spacing: 1px;
            }

            .widget-weather {
                display: flex;
                align-items: center;
                gap: 12px;
                font-size: 16px;
                color: var(--fb-text-primary, #e0e0e0);
                background: rgba(255, 255, 255, 0.05);
                padding: 8px 16px;
                border-radius: 20px;
            }

            .widget-haiku {
                max-width: 400px;
                font-style: italic;
                line-height: 1.6;
                position: relative;
                padding: 20px;
                border-left: 3px solid var(--fb-accent, #6366f1);
                background: linear-gradient(90deg, rgba(99, 102, 241, 0.1) 0%, transparent 100%);
                border-radius: 0 8px 8px 0;
                text-align: left;
            }

            .haiku-text {
                white-space: pre-line;
                font-size: 16px;
                color: var(--fb-text-primary, #e0e0e0);
            }

            .haiku-meta {
                margin-top: 12px;
                font-size: 12px;
                color: var(--fb-text-tertiary, #707080);
                text-transform: uppercase;
                letter-spacing: 0.5px;
            }

            .empty-cta {
                display: inline-flex;
                align-items: center;
                gap: 8px;
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
    @state() private _currentTime = new Date();
    @state() private _currentHaiku = this._getRandomHaiku();
    @state() private _currentWeather = this._getMockWeather();
    @state() private _sseConnected = false;

    private _unsubSSE: (() => void) | null = null;
    private _unsubStatus: (() => void) | null = null;
    private _unsubContext: (() => void) | null = null;
    private _timer: number | null = null;
    private _newCardIds: Set<string> = new Set();

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

        // Sprint 5.1: Subscribe to SSE connection status changes
        this._unsubStatus = feedSSE.onStatusChange((connected) => {
            this._sseConnected = connected;
        });

        // Sprint 5.1: React to contextState$ changes for scope-aware filtering
        this._unsubContext = effect(() => {
            const ctx = store.contextState$.value;
            const newFilter = ctx.scope === 'project' ? ctx.projectId : null;
            if (newFilter !== this._filterProjectId) {
                this._filterProjectId = newFilter;
                this._loadFeed();
            }
        });

        // Clock timer
        this._timer = window.setInterval(() => {
            this._currentTime = new Date();
        }, 1000);
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
        if (this._unsubStatus) {
            this._unsubStatus();
            this._unsubStatus = null;
        }
        if (this._unsubContext) {
            this._unsubContext();
            this._unsubContext = null;
        }
        feedSSE.disconnect();
        if (this._timer) {
            clearInterval(this._timer);
            this._timer = null;
        }
    }

    private _getRandomHaiku() {
        const haikus = [
            "Code flows like river,\nBugs are stones within the stream,\nTesting smooths the path.",
            "Servers hum softly,\nData travelers at rest,\nSystem writes its logs.",
            "Pixels on the screen,\nPainting logic in the light,\nUser finds their way.",
            "Build script starts to run,\nDependencies are fetched now,\nGreen checkmarks delight.",
            "Silence in the feed,\nTasks completed, mind at ease,\nFocus on the now.",
            "Screens glow in the dark,\nLines of code build silent worlds,\nLogic finds its home."
        ];
        return haikus[Math.floor(Math.random() * haikus.length)];
    }

    private _getMockWeather() {
        const weathers = [
            { temp: 72, condition: 'Sunny', icon: 'sunny' },
            { temp: 65, condition: 'Cloudy', icon: 'cloud' },
            { temp: 68, condition: 'Partly Cloudy', icon: 'partly_cloudy_day' },
            { temp: 58, condition: 'Rain', icon: 'rainy' }
        ];
        return weathers[Math.floor(Math.random() * weathers.length)];
    }

    /** Handle live feed SSE events */
    private _handleSSEEvent = (event: FeedSSEEvent): void => {
        switch (event.type) {
            case 'card_added': {
                // Apply project filter if active
                if (this._filterProjectId && event.card.project_id !== this._filterProjectId) return;
                // Sprint 5.1: Track new cards for slide-in animation
                this._newCardIds.add(event.card.id);
                // Insert sorted by priority (lower = higher priority)
                const cards = [...this._cards];
                const idx = cards.findIndex((c) => c.priority > event.card.priority);
                if (idx === -1) {
                    cards.push(event.card);
                } else {
                    cards.splice(idx, 0, event.card);
                }
                this._cards = cards;
                // Clear animation class after animation completes
                setTimeout(() => {
                    this._newCardIds.delete(event.card.id);
                }, 400);
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

    private async _handleCardAction(e: CustomEvent<{ cardId: string; actionId: string; projectId: string }>) {
        const { cardId, actionId, projectId } = e.detail;

        // Client-side navigation actions — no API call needed
        switch (actionId) {
            case 'view_briefing':
            case 'view_details':
                this.emit('fb-navigate', { view: 'project', projectId });
                return;
            case 'view_schedule':
                this.emit('fb-navigate', { view: 'project-schedule', projectId });
                return;
            case 'review_budget':
                this.emit('fb-navigate', { view: 'budget' });
                return;
            case 'add_contacts':
                this.emit('fb-navigate', { view: 'contacts' });
                return;
            case 'show_details': {
                // Navigate to project detail view (shows project context + feed)
                this.emit('fb-navigate', { view: 'project', projectId });
                return;
            }
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
        console.log('[FBHomeFeed] Executing action:', actionId, cardId);
        try {
            const resp = await api.portfolio.executeAction(cardId, actionId);
            console.log('[FBHomeFeed] Action response:', resp);

            if (resp.effect === 'dismiss') {
                this._cards = this._cards.filter((c) => c.id !== cardId);
            }
            if (resp.effect === 'navigate' && resp.navigate_to) {
                console.log('[FBHomeFeed] Navigating to:', resp.navigate_to);
                this.emit('fb-navigate', { path: resp.navigate_to });
            }
            if (resp.message) {
                this.emit('fb-toast', { message: resp.message });
            }
        } catch (err) {
            console.error('[FBHomeFeed] Action failed:', err);
            this.emit('fb-toast', { message: 'Action failed. Please try again.', type: 'error' });
        }
    }

    private _handleCreateProject() {
        this.emit('fb-navigate', { view: 'project-create' });
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


        const groups = this._groupCards();
        const horizons: FeedCardHorizon[] = ['today', 'this_week', 'horizon'];

        return html`
            <main role="main" aria-label="Portfolio Feed">
                <fb-greeting-banner
                    greeting=${this._greeting}
                    .summary=${this._summary}
                ></fb-greeting-banner>

                <!-- Sprint 5.1: SSE connection status indicator -->
                <div class="connection-status">
                    <span class="status-dot ${this._sseConnected ? 'connected' : ''}"></span>
                    ${this._sseConnected ? 'Live' : 'Connecting...'}
                </div>

                ${this._cards.length === 0
                ? html`
                          <div class="empty-state-container" role="status">
                              <div class="widget-clock">
                                  ${this._currentTime.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                              </div>
                              <div class="widget-date">
                                  ${this._currentTime.toLocaleDateString([], { weekday: 'long', month: 'long', day: 'numeric' })}
                              </div>
                              
                              <div class="widget-weather">
                                  <span class="material-symbols-outlined">${this._currentWeather?.icon}</span>
                                  <span>San Francisco, ${this._currentWeather?.temp}°F</span>
                              </div>

                              <div class="widget-haiku">
                                  <div class="haiku-text">${this._currentHaiku}</div>
                                  <div class="haiku-meta">Daily Inspiration</div>
                              </div>

                              ${(!this._summary || this._summary.active_project_count === 0)
                        ? html`
                                    <button class="empty-cta" @click=${this._handleCreateProject}>
                                        <span class="material-symbols-outlined">add</span>
                                        Start First Project
                                    </button>
                                  `
                        : nothing}
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
                                          <fb-feed-card
                                              .card=${card}
                                              class=${this._newCardIds.has(card.id) ? 'new-card' : ''}
                                              role="article"
                                          ></fb-feed-card>
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
