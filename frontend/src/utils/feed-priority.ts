import { FeedCard, FeedCardType } from '../types/feed';

export type FeedPriority = 'critical' | 'urgent' | 'routine';

function hoursAgo(dateStr: string): number {
    const ms = Date.now() - new Date(dateStr).getTime();
    return ms / (1000 * 60 * 60);
}

/**
 * Evaluates a feed card and returns its priority tier and a numeric score for sorting.
 * Higher score = higher priority (sorted descending).
 */
export function scorePriority(card: FeedCard): { priority: FeedPriority; score: number } {
    // P1: Critical (Safety, blockers, or explicitly marked priority 0)
    if (
        card.card_type === 'procurement_critical' ||
        card.card_type === 'weather_risk' ||
        card.tags?.includes('blocking') ||
        card.priority === 0
    ) {
        return { priority: 'critical', score: 100 };
    }

    // P2: Urgent (Pending approvals/actions older than 48 hours, or priority 1)
    const urgentTypes: FeedCardType[] = ['invoice_ready', 'sub_unconfirmed', 'budget_alert', 'procurement_warning'];
    if (urgentTypes.includes(card.card_type)) {
        if (card.created_at) {
            const hours = hoursAgo(card.created_at);
            if (hours > 48) {
                return { priority: 'urgent', score: 80 };
            }
        }
    }

    if (card.priority === 1) {
        return { priority: 'urgent', score: 70 };
    }

    // P3: Routine (Everything else)
    return { priority: 'routine', score: 20 };
}
