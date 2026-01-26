/**
 * Notification Store - Signal-Based Toast Notification System
 * See LAUNCH_PLAN.md P2 (Notifications/Toast UI)
 *
 * Provides a centralized notification system for visual feedback.
 * Uses Preact Signals for reactive state management.
 */

import { signal, type ReadonlySignal } from '@preact/signals-core';

/**
 * Notification type determines the visual style and icon.
 */
export type NotificationType = 'success' | 'error' | 'warning' | 'info';

/**
 * Notification represents a single toast message.
 */
export interface Notification {
    id: string;
    type: NotificationType;
    message: string;
    /** Duration in milliseconds before auto-dismiss. Default: 5000 */
    duration: number;
    /** Optional action button */
    action?: {
        label: string;
        callback: () => void;
    };
    /** Timestamp for ordering */
    createdAt: number;
}

/**
 * Options for creating a notification.
 */
export interface NotifyOptions {
    /** Duration in milliseconds. Default: 5000 for success/info, 8000 for error/warning */
    duration?: number;
    /** Optional action button */
    action?: {
        label: string;
        callback: () => void;
    };
}

// Default durations by type
const DEFAULT_DURATIONS: Record<NotificationType, number> = {
    success: 5000,
    info: 5000,
    warning: 8000,
    error: 8000,
};

// Internal writable signal
const _notifications$ = signal<Notification[]>([]);

/**
 * Readonly signal exposing current notifications.
 * Subscribe to this in toast container component.
 */
export const notifications$: ReadonlySignal<Notification[]> = _notifications$;

/**
 * Creates a notification and adds it to the queue.
 */
function createNotification(
    type: NotificationType,
    message: string,
    options?: NotifyOptions
): string {
    const id = crypto.randomUUID();
    const notification: Notification = {
        id,
        type,
        message,
        duration: options?.duration ?? DEFAULT_DURATIONS[type],
        createdAt: Date.now(),
    };

    // Only set action if provided (exactOptionalPropertyTypes compliance)
    if (options?.action) {
        notification.action = options.action;
    }

    _notifications$.value = [..._notifications$.value, notification];
    return id;
}

/**
 * Dismisses a notification by ID.
 */
function dismiss(id: string): void {
    _notifications$.value = _notifications$.value.filter((n) => n.id !== id);
}

/**
 * Clears all notifications.
 */
function clearAll(): void {
    _notifications$.value = [];
}

/**
 * Notification API for creating toast notifications.
 *
 * @example
 * ```typescript
 * import { notify } from '../store/notifications';
 *
 * // Simple notifications
 * notify.success('Project saved successfully');
 * notify.error('Failed to upload file');
 * notify.warning('Network connection unstable');
 * notify.info('New update available');
 *
 * // With custom duration
 * notify.success('Saved!', { duration: 3000 });
 *
 * // With action button
 * notify.error('Failed to save', {
 *   action: {
 *     label: 'Retry',
 *     callback: () => saveProject(),
 *   },
 * });
 * ```
 */
export const notify = {
    /**
     * Shows a success notification.
     * @param message - The message to display
     * @param options - Optional duration and action
     * @returns The notification ID
     */
    success: (message: string, options?: NotifyOptions): string =>
        createNotification('success', message, options),

    /**
     * Shows an error notification.
     * @param message - The message to display
     * @param options - Optional duration and action
     * @returns The notification ID
     */
    error: (message: string, options?: NotifyOptions): string =>
        createNotification('error', message, options),

    /**
     * Shows a warning notification.
     * @param message - The message to display
     * @param options - Optional duration and action
     * @returns The notification ID
     */
    warning: (message: string, options?: NotifyOptions): string =>
        createNotification('warning', message, options),

    /**
     * Shows an info notification.
     * @param message - The message to display
     * @param options - Optional duration and action
     * @returns The notification ID
     */
    info: (message: string, options?: NotifyOptions): string =>
        createNotification('info', message, options),

    /**
     * Dismisses a notification by ID.
     * @param id - The notification ID to dismiss
     */
    dismiss,

    /**
     * Clears all notifications.
     */
    clearAll,
};
