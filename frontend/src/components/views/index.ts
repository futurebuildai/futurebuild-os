/**
 * View Registry - Centralized View Component Exports
 * See PRODUCTION_PLAN.md Step 51.4
 *
 * This module exports all view components and provides a registry
 * for the router. Structured for future async loading support.
 *
 * V2: fb-view-dashboard removed (replaced by fb-home-feed).
 */

// ============================================================================
// View Imports (Synchronous)
// ============================================================================

import './fb-view-login';
import './fb-view-projects';
import './fb-view-chat';
import './fb-view-schedule';
import './fb-view-directory';
import './fb-view-onboarding';

// Re-export for convenience
export { FBViewLogin } from './fb-view-login';
export { FBViewProjects } from './fb-view-projects';
export { FBViewChat } from './fb-view-chat';
export { FBViewSchedule } from './fb-view-schedule';
export { FBViewDirectory } from './fb-view-directory';
export { FBViewOnboarding } from './fb-view-onboarding';

// ============================================================================
// View Registry (Future: Async Loaders)
// ============================================================================

/**
 * Type for async view loader function.
 * Returns the constructor of the view component.
 */
export type ViewLoader = () => Promise<CustomElementConstructor>;

/**
 * View registry supporting async loading.
 * Currently returns sync-imported classes wrapped in Promise.resolve.
 */
export const VIEW_REGISTRY = {
    projects: () => import('./fb-view-projects').then((m) => m.FBViewProjects),
    chat: () => import('./fb-view-chat').then((m) => m.FBViewChat),
    schedule: () => import('./fb-view-schedule').then((m) => m.FBViewSchedule),
    directory: () => import('./fb-view-directory').then((m) => m.FBViewDirectory),
    login: () => import('./fb-view-login').then((m) => m.FBViewLogin),
    onboarding: () => import('./fb-view-onboarding').then((m) => m.FBViewOnboarding),
} as const;
