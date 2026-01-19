import { GanttArtifactData } from '../types/artifacts';
import { TaskStatus } from '../types/enums';

export const MOCK_GANTT_DATA: GanttArtifactData = {
    project_id: 'proj-123',
    calculated_at: '2024-01-20T12:00:00Z',
    projected_end_date: '2024-06-15',
    critical_path: ['wbs-1.1', 'wbs-1.2'],
    tasks: [
        {
            wbs_code: '1.1',
            name: 'Foundation',
            status: TaskStatus.Completed,
            early_start: '2024-01-01',
            early_finish: '2024-01-10',
            duration_days: 10,
            is_critical: true
        },
        {
            wbs_code: '1.2',
            name: 'Framing',
            status: TaskStatus.InProgress,
            early_start: '2024-01-12',
            early_finish: '2024-02-15',
            duration_days: 34,
            is_critical: true
        },
        {
            wbs_code: '1.3',
            name: 'Rough Plumbing',
            status: TaskStatus.Pending,
            early_start: '2024-02-16',
            early_finish: '2024-02-28',
            duration_days: 12,
            is_critical: false
        },
        {
            wbs_code: '1.4',
            name: 'Electrical',
            status: TaskStatus.Pending,
            early_start: '2024-03-01',
            early_finish: '2024-03-14',
            duration_days: 14,
            is_critical: false
        }
    ]
};
