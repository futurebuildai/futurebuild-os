/**
 * Feed Types — Portfolio feed card system.
 * Rosetta Stone parity with Go internal/models/feed_card.go.
 * See FRONTEND_V2_SPEC.md §3
 */

/**
 * FeedCardType maps to Go FeedCardType constants.
 */
export type FeedCardType =
    | 'daily_briefing'
    | 'procurement_warning'
    | 'procurement_critical'
    | 'task_starting'
    | 'task_completed'
    | 'inspection_upcoming'
    | 'inspection_result'
    | 'schedule_recalc'
    | 'weather_risk'
    | 'weather_window'
    | 'sub_confirmation'
    | 'sub_unconfirmed'
    | 'invoice_ready'
    | 'budget_alert'
    | 'setup_team'
    | 'setup_contacts'
    | 'calibration_drift'
    | 'milestone'
    | 'welcome'
    // Agent approval card types (human-in-the-loop)
    | 'agent_approval'
    | 'agent_recommendation'
    | 'change_order'
    | 'delay_mitigation'
    | 'draft_message'
    // Integration card types (FB-Brain cross-system flows)
    | 'material_quote_prompt'
    | 'material_quote_review'
    | 'material_order_confirm'
    | 'labor_bid_prompt'
    | 'labor_bid_review'
    | 'labor_bid_confirm'
    | 'delivery_confirm';

/**
 * FeedCardHorizon — temporal grouping.
 */
export type FeedCardHorizon = 'today' | 'this_week' | 'horizon';

/**
 * FeedCardAction — inline action button on a card.
 */
export interface FeedCardAction {
    id: string;
    label: string;
    style: 'primary' | 'secondary' | 'danger';
}

/**
 * FeedCard — a single card in the portfolio feed.
 * Matches Go models.FeedCard JSON output.
 */
export interface FeedCard {
    id: string;
    org_id: string;
    project_id: string;
    card_type: FeedCardType;
    priority: number;
    headline: string;
    body: string;
    consequence?: string;
    horizon: FeedCardHorizon;
    deadline?: string;
    actions: FeedCardAction[];
    engine_data?: Record<string, unknown>;
    agent_source?: string;
    task_id?: string;
    tags?: string[]; // Added for priority scoring logic
    created_at: string;
    expires_at?: string;
    // Denormalized from JOIN
    project_name?: string;
}

/**
 * PortfolioSummary — high-level overview of user's projects.
 * Matches Go service.PortfolioSummary.
 */
export interface PortfolioSummary {
    active_project_count: number;
    total_tasks: number;
    critical_alerts: number;
    projected_completions: ProjectCompletionSummary[];
}

/**
 * ProjectCompletionSummary — single project completion status.
 */
export interface ProjectCompletionSummary {
    project_id: string;
    project_name: string;
    end_date: string;
    on_track: boolean;
    slip_days: number;
}

/**
 * ProjectPill — minimal project info for the top-bar pills.
 */
export interface ProjectPill {
    id: string;
    name: string;
    address: string;
    status: string;
}

/**
 * ActionResponse — structured response from POST /api/v1/portfolio/feed/action.
 */
export interface ActionResponse {
    success: boolean;
    effect: 'dismiss' | 'navigate' | 'none';
    message?: string;
    navigate_to?: string;
    payload?: Record<string, unknown>;
}

/**
 * PortfolioFeedResponse — full response from GET /api/v1/portfolio/feed.
 */
export interface PortfolioFeedResponse {
    greeting: string;
    summary: PortfolioSummary;
    cards: FeedCard[];
    projects: ProjectPill[];
}

// ============================================================================
// SSE Feed Stream Events (Phase 7 Step 40)
// See FRONTEND_V2_SPEC.md §6.5
// ============================================================================

/**
 * SSE event: a new card was added to the feed.
 */
export interface FeedCardAddedEvent {
    type: 'card_added';
    card: FeedCard;
}

/**
 * SSE event: an existing card was updated (e.g., priority change).
 */
export interface FeedCardUpdatedEvent {
    type: 'card_updated';
    card: FeedCard;
}

/**
 * SSE event: a card was removed (dismissed, expired, or snoozed).
 */
export interface FeedCardRemovedEvent {
    type: 'card_removed';
    card_id: string;
}

/**
 * Union of all feed SSE event types.
 */
export type FeedSSEEvent = FeedCardAddedEvent | FeedCardUpdatedEvent | FeedCardRemovedEvent;
