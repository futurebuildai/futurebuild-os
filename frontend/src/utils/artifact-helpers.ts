/**
 * Artifact Helpers - Shared utilities for artifact type handling
 * See FRONTEND_SCOPE.md Section 8.3
 * 
 * Centralizes type normalization to prevent DRY violations across
 * store.ts, fb-panel-right.ts, and other consumers.
 */

import { ArtifactType } from '../types/enums';

/**
 * Normalized artifact type strings for UI consumption.
 */
export type NormalizedArtifactType = 'gantt' | 'budget' | 'invoice' | 'table' | 'chart';

/**
 * Normalize ArtifactType enum values to lowercase strings for UI components.
 * Handles both enum values (e.g., ArtifactType.Invoice) and string values
 * that may come from the backend (e.g., "Budget_View").
 */
export function normalizeArtifactType(type: ArtifactType | string): NormalizedArtifactType {
    const typeStr = typeof type === 'string' ? type : String(type);
    const normalized = typeStr.toLowerCase().replace('_view', '');

    switch (normalized) {
        case 'gantt': return 'gantt';
        case 'budget': return 'budget';
        case 'invoice': return 'invoice';
        case 'chart': return 'chart';
        default: return 'table';
    }
}

/**
 * Get emoji icon for artifact type.
 */
export function getArtifactIcon(type: NormalizedArtifactType): string {
    const icons: Record<string, string> = {
        budget: '💰',
        gantt: '📊',
        invoice: '📄',
        table: '📋',
        chart: '📈',
    };
    return icons[type] ?? '📎';
}
