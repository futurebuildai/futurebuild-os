/**
 * CPM Engine — Critical Path Method Schedule Calculator
 * Sprint 2.3: Physics Trigger
 *
 * Pure TypeScript implementation of the Critical Path Method.
 * Computes early/late start-finish, total float, and critical path
 * from a directed acyclic graph of construction tasks.
 */
import type { GanttData, GanttTask, GanttDependency } from '../types/models';
import { TaskStatus } from '../types/enums';

// ============================================================================
// Types
// ============================================================================

/**
 * Input task for CPM calculation.
 * Dependencies reference other task IDs (WBS codes).
 */
export interface CPMTask {
    id: string;
    name: string;
    duration: number;        // calendar days
    dependencies: string[];  // predecessor task IDs
}

/**
 * Computed CPM task with schedule data.
 */
export interface CPMTaskResult extends CPMTask {
    earlyStart: number;
    earlyFinish: number;
    lateStart: number;
    lateFinish: number;
    totalFloat: number;
    isCritical: boolean;
}

/**
 * Full CPM calculation result.
 */
export interface CPMResult {
    tasks: CPMTaskResult[];
    criticalPath: string[];
    projectDuration: number;
}

// ============================================================================
// Algorithm
// ============================================================================

/**
 * Topological sort using Kahn's algorithm.
 * Throws if the dependency graph contains a cycle.
 */
function topologicalSort(tasks: CPMTask[]): CPMTask[] {
    const taskMap = new Map<string, CPMTask>();
    const inDegree = new Map<string, number>();
    const adjacency = new Map<string, string[]>();

    for (const t of tasks) {
        taskMap.set(t.id, t);
        inDegree.set(t.id, 0);
        adjacency.set(t.id, []);
    }

    for (const t of tasks) {
        for (const dep of t.dependencies) {
            if (!taskMap.has(dep)) {
                throw new Error(`Task "${t.id}" depends on unknown task "${dep}"`);
            }
            adjacency.get(dep)!.push(t.id);
            inDegree.set(t.id, (inDegree.get(t.id) ?? 0) + 1);
        }
    }

    const queue: string[] = [];
    for (const [id, deg] of inDegree) {
        if (deg === 0) queue.push(id);
    }

    const sorted: CPMTask[] = [];
    while (queue.length > 0) {
        const id = queue.shift()!;
        sorted.push(taskMap.get(id)!);
        for (const succ of adjacency.get(id) ?? []) {
            const newDeg = (inDegree.get(succ) ?? 1) - 1;
            inDegree.set(succ, newDeg);
            if (newDeg === 0) queue.push(succ);
        }
    }

    if (sorted.length !== tasks.length) {
        throw new Error('Dependency graph contains a cycle');
    }

    return sorted;
}

/**
 * Compute Critical Path Method schedule.
 *
 * 1. Topological sort
 * 2. Forward pass → early start / early finish
 * 3. Backward pass → late start / late finish
 * 4. Float → total float = lateStart − earlyStart
 * 5. Critical → float === 0
 */
export function calculate(tasks: CPMTask[]): CPMResult {
    if (tasks.length === 0) {
        return { tasks: [], criticalPath: [], projectDuration: 0 };
    }

    const sorted = topologicalSort(tasks);

    // Forward pass
    const earlyStart = new Map<string, number>();
    const earlyFinish = new Map<string, number>();

    for (const t of sorted) {
        let es = 0;
        for (const dep of t.dependencies) {
            const depEF = earlyFinish.get(dep) ?? 0;
            if (depEF > es) es = depEF;
        }
        earlyStart.set(t.id, es);
        earlyFinish.set(t.id, es + t.duration);
    }

    // Project duration = max early finish
    let projectDuration = 0;
    for (const ef of earlyFinish.values()) {
        if (ef > projectDuration) projectDuration = ef;
    }

    // Backward pass
    const lateStart = new Map<string, number>();
    const lateFinish = new Map<string, number>();

    // Build successor map
    const successors = new Map<string, string[]>();
    for (const t of tasks) {
        successors.set(t.id, []);
    }
    for (const t of tasks) {
        for (const dep of t.dependencies) {
            successors.get(dep)!.push(t.id);
        }
    }

    // Process in reverse topological order
    for (let i = sorted.length - 1; i >= 0; i--) {
        const t = sorted[i]!;
        const succs = successors.get(t.id) ?? [];
        let lf = projectDuration;
        for (const succ of succs) {
            const succLS = lateStart.get(succ) ?? projectDuration;
            if (succLS < lf) lf = succLS;
        }
        lateFinish.set(t.id, lf);
        lateStart.set(t.id, lf - t.duration);
    }

    // Build results
    const results: CPMTaskResult[] = sorted.map((t) => {
        const es = earlyStart.get(t.id) ?? 0;
        const ef = earlyFinish.get(t.id) ?? 0;
        const ls = lateStart.get(t.id) ?? 0;
        const lf = lateFinish.get(t.id) ?? 0;
        const totalFloat = ls - es;
        return {
            ...t,
            earlyStart: es,
            earlyFinish: ef,
            lateStart: ls,
            lateFinish: lf,
            totalFloat,
            isCritical: totalFloat === 0,
        };
    });

    const criticalPath = results
        .filter((t) => t.isCritical)
        .map((t) => t.id);

    return { tasks: results, criticalPath, projectDuration };
}

// ============================================================================
// Conversion Helpers
// ============================================================================

/**
 * Convert CPM result to GanttData format for the schedule view.
 * Maps day offsets (from project start date) to ISO date strings.
 */
export function toGanttData(
    projectId: string,
    startDate: Date,
    result: CPMResult
): GanttData {
    const addDays = (d: Date, days: number): string => {
        const out = new Date(d);
        out.setDate(out.getDate() + days);
        return out.toISOString().substring(0, 10);
    };

    const tasks: GanttTask[] = result.tasks.map((t) => ({
        wbs_code: t.id,
        name: t.name,
        status: TaskStatus.Pending,
        early_start: addDays(startDate, t.earlyStart),
        early_finish: addDays(startDate, t.earlyFinish),
        duration_days: t.duration,
        is_critical: t.isCritical,
    }));

    // Build dependency edges from the input tasks
    const dependencies: GanttDependency[] = [];
    for (const t of result.tasks) {
        for (const dep of t.dependencies) {
            dependencies.push({ from: dep, to: t.id });
        }
    }

    return {
        project_id: projectId,
        calculated_at: new Date().toISOString(),
        projected_end_date: addDays(startDate, result.projectDuration),
        critical_path: result.criticalPath,
        tasks,
        dependencies,
    };
}
