import { GanttData } from '../types/models';
import { TaskStatus } from '../types/enums';

/**
 * Mock Schedule Service
 * Provides realistic Gantt chart data for demo/testing.
 */
export const mockScheduleService = {
    async get(projectId: string): Promise<GanttData> {
        // Simulate network delay
        await new Promise(resolve => setTimeout(resolve, 800));

        const today = new Date();
        const formatDate = (daysFromNow: number): string => {
            const d = new Date(today);
            d.setDate(d.getDate() + daysFromNow);
            return d.toISOString().substring(0, 10);
        };

        if (projectId === 'p1') {
            // Willow Creek - Early Phase (Excavation & Foundation)
            return {
                project_id: 'p1',
                calculated_at: new Date().toISOString(),
                projected_end_date: formatDate(120),
                critical_path: ['t3', 't4', 't5'],
                tasks: [
                    {
                        wbs_code: '1.0',
                        name: 'Mobilization',
                        status: TaskStatus.Completed,
                        early_start: formatDate(-14),
                        early_finish: formatDate(-10),
                        duration_days: 4,
                        is_critical: false
                    },
                    {
                        wbs_code: '2.0',
                        name: 'Site Clearing',
                        status: TaskStatus.Completed,
                        early_start: formatDate(-10),
                        early_finish: formatDate(-5),
                        duration_days: 5,
                        is_critical: false
                    },
                    {
                        wbs_code: '3.0',
                        name: 'Excavation',
                        status: TaskStatus.InProgress,
                        early_start: formatDate(-5),
                        early_finish: formatDate(5),
                        duration_days: 10,
                        is_critical: true
                    },
                    {
                        wbs_code: '3.1',
                        name: 'Grading',
                        status: TaskStatus.Pending,
                        early_start: formatDate(5),
                        early_finish: formatDate(8),
                        duration_days: 3,
                        is_critical: true
                    },
                    {
                        wbs_code: '4.0',
                        name: 'Foundation Pour',
                        status: TaskStatus.Pending,
                        early_start: formatDate(8),
                        early_finish: formatDate(12),
                        duration_days: 4,
                        is_critical: true
                    }
                ],
                dependencies: [
                    { from: '1.0', to: '2.0' },
                    { from: '2.0', to: '3.0' },
                    { from: '3.0', to: '3.1' },
                    { from: '3.1', to: '4.0' }
                ]
            };
        }

        if (projectId === 'p2') {
            // Hudson Yards - Late Phase (Finishes & Closeout)
            return {
                project_id: 'p2',
                calculated_at: new Date().toISOString(),
                projected_end_date: formatDate(30),
                critical_path: ['t8', 't9'],
                tasks: [
                    {
                        wbs_code: '5.0',
                        name: 'MEP Rough-in',
                        status: TaskStatus.Completed,
                        early_start: formatDate(-45),
                        early_finish: formatDate(-30),
                        duration_days: 15,
                        is_critical: false
                    },
                    {
                        wbs_code: '6.0',
                        name: 'Drywall & Texture',
                        status: TaskStatus.Completed,
                        early_start: formatDate(-30),
                        early_finish: formatDate(-15),
                        duration_days: 15,
                        is_critical: false
                    },
                    {
                        wbs_code: '7.0',
                        name: 'Interior Paint',
                        status: TaskStatus.Completed,
                        early_start: formatDate(-15),
                        early_finish: formatDate(-5),
                        duration_days: 10,
                        is_critical: false
                    },
                    {
                        wbs_code: '8.0',
                        name: 'Flooring Installation',
                        status: TaskStatus.InProgress,
                        early_start: formatDate(-5),
                        early_finish: formatDate(5),
                        duration_days: 10,
                        is_critical: true
                    },
                    {
                        wbs_code: '9.0',
                        name: 'Final Clean & Punch',
                        status: TaskStatus.Pending,
                        early_start: formatDate(5),
                        early_finish: formatDate(10),
                        duration_days: 5,
                        is_critical: true
                    }
                ],
                dependencies: [
                    { from: '5.0', to: '6.0' },
                    { from: '6.0', to: '7.0' },
                    { from: '7.0', to: '8.0' },
                    { from: '8.0', to: '9.0' }
                ]
            };
        }

        // Default empty state
        return {
            project_id: projectId,
            calculated_at: new Date().toISOString(),
            projected_end_date: new Date().toISOString(),
            critical_path: [],
            tasks: []
        };
    },

    async recalculate(projectId: string): Promise<GanttData> {
        return this.get(projectId);
    }
};
