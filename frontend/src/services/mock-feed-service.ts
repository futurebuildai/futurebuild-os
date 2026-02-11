
import type { PortfolioFeedResponse, ActionResponse } from '../types/feed';

/**
 * Mock Feed Service
 * Bypasses backend API calls for development/demo purposes.
 */
export const mockFeedService = {
    /**
     * Get a mock portfolio feed.
     * Returns an empty feed to demonstrate the "Empty State" widgets.
     */
    async getFeed(projectId?: string): Promise<PortfolioFeedResponse> {
        // Simulate network delay
        await new Promise(resolve => setTimeout(resolve, 600));

        // Mock Projects
        const projects: any[] = [
            { id: 'p1', name: 'Willow Creek', status: 'active', client: 'Urban Living', address: '123 Willow Ln' },
            { id: 'p2', name: 'Hudson Yards', status: 'on_hold', client: 'Metropolis Dev', address: '456 Hudson Ave' }
        ];

        // Helper to create a compliant FeedCard
        const createCard = (id: string, type: any, priority: number, headline: string, body: string, consequence: string | undefined, horizon: 'today' | 'this_week' | 'horizon', actions: any[]) => ({

            id,
            org_id: 'org_1',
            project_id: projectId || 'p1',
            card_type: type,
            priority,
            headline,
            body,
            consequence: consequence ?? '',
            horizon,
            actions: actions.map(a => ({ id: a.id, label: a.label, style: a.style || 'primary' })),
            created_at: new Date().toISOString()
        });

        // Willow Creek Cards - TODAY
        const willowToday = [
            createCard(
                'c1',
                'daily_briefing',
                1,
                'Daily Log Missing',
                'Willow Creek - Foundation Phase. Please submit your daily log.',
                undefined,
                'today',
                [{ id: 'view_briefing', label: 'Draft Log', style: 'primary' }, { id: 'snooze', label: 'Snooze', style: 'secondary' }]
            ),
            createCard(
                'c2',
                'weather_risk',
                2,
                'Rain Delay Risk',
                '60% chance of rain tomorrow. Cover foundation pit.',
                '→ Foundation pour may slip by 2 days if not covered.',
                'today',
                [{ id: 'acknowledge', label: 'Acknowledge', style: 'primary' }, { id: 'show_details', label: 'Show Details', style: 'ghost' }]
            )
        ];

        // Willow Creek Cards - THIS WEEK
        const willowThisWeek = [
            createCard(
                'c4',
                'inspection_upcoming',
                2,
                'Electrical Rough-in Inspection',
                'Scheduled for Friday, Oct 27. Ensure all boxes are accessible.',
                '→ Critical path item. Failure delays drywall start.',
                'this_week',
                [{ id: 'acknowledge', label: 'Confirm Ready', style: 'primary' }, { id: 'view_schedule', label: 'Reschedule', style: 'secondary' }, { id: 'show_details', label: 'Show Details', style: 'ghost' }]
            ),
            createCard(
                'c5',
                'procurement_warning',
                3,
                'Window Package Delivery',
                'Milgard windows scheduled for delivery Thursday.',
                undefined,
                'this_week',
                [{ id: 'view_details', label: 'Track Shipment', style: 'ghost' }]
            )
        ];

        // Willow Creek Cards - HORIZON
        const willowHorizon = [
            createCard(
                'c6',
                'procurement_critical',
                1,
                'Roof Trusses Lead Time',
                'Current lead time increased to 8 weeks. Order by Nov 5 to maintain schedule.',
                '→ Framing completion moves from Dec 15 to Jan 10 if delayed.',
                'horizon',
                [{ id: 'approve', label: 'Order Now', style: 'primary' }, { id: 'view_details', label: 'Get Quote', style: 'secondary' }, { id: 'show_details', label: 'Show Details', style: 'ghost' }]
            )
        ];

        // Hudson Yards Cards (Global view only example)
        const hudsonCards = [
            createCard(
                'c3',
                'budget_alert',
                1,
                'Budget Approval Needed',
                'Hudson Yards - Electrical. Variance +5% ($12,500).',
                '→ Project margin decreases by 0.5%.',
                'today',
                [{ id: 'review_budget', label: 'Review', style: 'primary' }, { id: 'reject', label: 'Reject', style: 'danger' }]
            )
        ];

        // If a specific project is selected, filter context
        if (projectId) {
            if (projectId === 'p1') {
                return {
                    greeting: `Willow Creek Update`,
                    summary: {
                        active_project_count: 1,
                        total_tasks: 12,
                        critical_alerts: 1,
                        projected_completions: []
                    },
                    cards: [...willowToday, ...willowThisWeek, ...willowHorizon],
                    projects: projects
                };
            }

            // Empty project (Hudson Yards)
            return {
                greeting: 'Hudson Yards Status',
                summary: {
                    active_project_count: 1,
                    total_tasks: 0,
                    critical_alerts: 0,
                    projected_completions: []
                },
                cards: [], // No cards for this specific project view in this mock scenario
                projects: projects
            };
        }

        // Global Feed (All Projects)
        return {
            greeting: 'Good Afternoon, Colton',
            summary: {
                active_project_count: 2,
                total_tasks: 12,
                critical_alerts: 1,
                projected_completions: []
            },
            cards: [...willowToday, ...willowThisWeek, ...willowHorizon, ...hudsonCards],
            projects: projects
        };
    },

    async executeAction(_cardId: string, actionId: string, _payload?: any): Promise<ActionResponse> {
        // Simulate network delay
        await new Promise(resolve => setTimeout(resolve, 400));

        switch (actionId) {
            case 'view_briefing':
            case 'view_details':
                return { success: true, effect: 'navigate', navigate_to: '/project/p1' };

            case 'view_schedule':
                return { success: true, effect: 'navigate', navigate_to: '/project/p1/schedule' };

            case 'add_contacts':
                return { success: true, effect: 'navigate', navigate_to: '/contacts' };

            case 'review_budget':
            case 'review':
                return { success: true, effect: 'navigate', navigate_to: '/budget' };

            case 'reject':
                return { success: true, effect: 'none', message: 'Request rejected. Project manager notified.' };

            case 'approve':
            case 'acknowledge':
                return { success: true, effect: 'dismiss', message: 'Acknowledged.' };

            default:
                return { success: true, effect: 'none', message: 'Action executed successfully.' };
        }
    },

    async dismissCard(_cardId: string): Promise<{ success: boolean }> {
        return { success: true };
    },

    async snoozeCard(_cardId: string, _hours: number): Promise<{ success: boolean }> {
        return { success: true };
    }
};
