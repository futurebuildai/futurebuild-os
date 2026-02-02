/**
 * NotificationService - System Notification Stream
 * See STEP_91_NOTIFICATION_UI.md
 *
 * Provides a persistent notification stream for the bell icon notification center.
 * This is separate from the toast notification system (store/notifications.ts).
 *
 * Currently uses mock data. API-ready: replace mock with GET /api/v1/notifications.
 */

import { signal, computed, type ReadonlySignal } from '@preact/signals-core';

// ============================================================================
// Types
// ============================================================================

/**
 * Notification type determines the visual style and icon color.
 */
export type SystemNotificationType = 'system' | 'mention' | 'alert' | 'success';

/**
 * A persistent system notification in the notification center.
 */
export interface SystemNotification {
    id: string;
    type: SystemNotificationType;
    title: string;
    message: string;
    /** Deep-link URL to navigate to when clicked */
    link?: string;
    /** Whether the user has read this notification */
    isRead: boolean;
    /** ISO timestamp */
    createdAt: string;
    /** Optional metadata for extensibility */
    metadata?: Record<string, unknown>;
}

// ============================================================================
// Mock Data
// ============================================================================

const MOCK_NOTIFICATIONS: SystemNotification[] = [
    {
        id: 'notif-1',
        type: 'alert',
        title: 'Schedule Slip Detected',
        message: 'Project Alpha critical path delayed by 3 days. Foundation work behind schedule.',
        link: '/projects',
        isRead: false,
        createdAt: new Date(Date.now() - 5 * 60 * 1000).toISOString(), // 5m ago
    },
    {
        id: 'notif-2',
        type: 'success',
        title: 'Invoice Approved',
        message: 'Invoice #INV-2024-0847 from ABC Concrete approved for $12,450.',
        isRead: false,
        createdAt: new Date(Date.now() - 32 * 60 * 1000).toISOString(), // 32m ago
    },
    {
        id: 'notif-3',
        type: 'system',
        title: 'Weather Alert',
        message: 'Rain expected Thursday–Friday. Pre-dry-in tasks may be affected.',
        isRead: true,
        createdAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(), // 2h ago
    },
    {
        id: 'notif-4',
        type: 'mention',
        title: 'New Message',
        message: 'Subcontractor replied to your inquiry about electrical rough-in timeline.',
        link: '/chat',
        isRead: true,
        createdAt: new Date(Date.now() - 5 * 60 * 60 * 1000).toISOString(), // 5h ago
    },
];

// ============================================================================
// State
// ============================================================================

const isDev = (import.meta as unknown as { env?: { DEV?: boolean } }).env?.DEV === true;
const _notifications$ = signal<SystemNotification[]>(isDev ? MOCK_NOTIFICATIONS : []);

/**
 * Readonly signal of all system notifications.
 */
export const systemNotifications$: ReadonlySignal<SystemNotification[]> = _notifications$;

/**
 * Computed: count of unread notifications.
 */
export const unreadCount$: ReadonlySignal<number> = computed(
    () => _notifications$.value.filter((n) => !n.isRead).length
);

// ============================================================================
// Actions
// ============================================================================

/**
 * Mark a single notification as read.
 */
export function markAsRead(id: string): void {
    _notifications$.value = _notifications$.value.map((n) =>
        n.id === id ? { ...n, isRead: true } : n
    );
}

/**
 * Mark all notifications as read.
 */
export function markAllAsRead(): void {
    _notifications$.value = _notifications$.value.map((n) => ({ ...n, isRead: true }));
}

/**
 * Add a new notification to the stream.
 * Used by realtime handlers when the backend pushes a notification.
 */
export function addSystemNotification(notification: SystemNotification): void {
    _notifications$.value = [notification, ..._notifications$.value];
}

/**
 * Remove a notification by ID.
 */
export function dismissNotification(id: string): void {
    _notifications$.value = _notifications$.value.filter((n) => n.id !== id);
}
