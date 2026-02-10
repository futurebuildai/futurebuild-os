/**
 * FutureBuild Frontend - Main Entry Point
 * See FRONTEND_V2_SPEC.md §4
 *
 * V2: Feed-first layout with top bar. Old 3-panel removed.
 *
 * This file bootstraps the application by:
 * 1. Importing global styles (CSS variables, resets)
 * 2. Registering all web components
 * 3. Importing the app root
 */

// Global styles (order matters: variables must load first)
import './styles/main.css';

// Component registration
import { registerComponents } from './components/registry';

// V2 Layout components
import { FBAppShell } from './components/layout/fb-app-shell';
import { FBTopBar } from './components/layout/fb-top-bar';
import { FBMobileNav } from './components/layout/fb-mobile-nav';

// V2 Feed components
import { FBHomeFeed } from './components/feed/fb-home-feed';
import { FBFeedCard } from './components/feed/fb-feed-card';
import { FBFeedSection } from './components/feed/fb-feed-section';
import { FBGreetingBanner } from './components/feed/fb-greeting-banner';
import { FBEmptyHome } from './components/feed/fb-empty-home';

// V2 Layout - Right panel for artifacts
import { FBPanelRight } from './components/layout/fb-panel-right';

// Chat Components (Step 52, retained for project-chat view)
import { FBMessageList } from './components/chat/fb-message-list';
import { FBActionCard } from './components/chat/fb-action-card';
import { FBInputBar } from './components/chat/fb-input-bar';
import { FBAgentActivity } from './components/agent/fb-agent-activity';
import { FBTypingIndicator } from './components/chat/fb-typing-indicator';

// Artifact Components (Step 55, retained for right panel & modal)
import { FBArtifactGantt } from './components/artifacts/fb-artifact-gantt';
import { FBArtifactBudget } from './components/artifacts/fb-artifact-budget';
import { FBArtifactInvoice } from './components/artifacts/fb-artifact-invoice';

// Feature Components (Step 56)
import { FBFileDrop } from './components/features/fb-file-drop';

// V2 Settings Components
import { FBSettingsProfile } from './components/settings/fb-settings-profile';
import { FBSettingsOrg } from './components/settings/fb-settings-org';
import { FBSettingsTeam } from './components/settings/fb-settings-team';

// V2 Project Components
import { FBProjectHeader } from './components/project/fb-project-header';

// V2 Onboarding Components
import { FBOnboardFlow } from './components/features/onboarding/fb-onboard-flow';
import { FBExtractionStream } from './components/features/onboarding/fb-extraction-stream';

// V2 Schedule Components
import { FBScheduleTaskBar } from './components/schedule/fb-schedule-task-bar';
import { FBScheduleDiff } from './components/schedule/fb-schedule-diff';

// V2 Contact Components
import { FBContactPhaseGrid } from './components/contacts/fb-contact-phase-grid';
import { FBContactBulkInput } from './components/contacts/fb-contact-bulk-input';
import { FBContactDetail } from './components/contacts/fb-contact-detail';
import { FBContactDirectory } from './components/contacts/fb-contact-directory';

// Base Components (Step 58.5)
import { FBErrorBoundary } from './components/base/fb-error-boundary';

// Notification Components (Step 91)
import { FBNotificationBell } from './components/notifications/fb-notification-bell';
import { FBNotificationList } from './components/notifications/fb-notification-list';

// Admin Components (Platform Admin)
import { FBAdminShell } from './components/admin/fb-admin-shell';
import { FBAdminSidebar } from './components/admin/fb-admin-sidebar';
import { FBAdminDashboard } from './components/admin/fb-admin-dashboard';

// Shadow Viewer Components (SHADOW_VIEWER_specs.md)
import { ShadowToggle } from './components/shadow/shadow-toggle';
import { ShadowLayout } from './components/shadow/shadow-layout';
import { ShadowNav } from './components/shadow/shadow-nav';
import { TribunalLogFeed } from './components/shadow/tribunal-log-feed';
import { TribunalCaseDetail } from './components/shadow/tribunal-case-detail';
import { ShadowDocsViewer } from './components/shadow/shadow-docs-viewer';

// App shell
import './app-root';

// Register all components
const registered = registerComponents({
    // V2 layout
    'fb-app-shell': FBAppShell,
    'fb-top-bar': FBTopBar,
    'fb-mobile-nav': FBMobileNav,
    // V2 feed
    'fb-home-feed': FBHomeFeed,
    'fb-feed-card': FBFeedCard,
    'fb-feed-section': FBFeedSection,
    'fb-greeting-banner': FBGreetingBanner,
    'fb-empty-home': FBEmptyHome,
    // V2 layout - right panel for artifacts
    'fb-panel-right': FBPanelRight,
    // Chat (retained for project-chat view)
    'fb-message-list': FBMessageList,
    'fb-action-card': FBActionCard,
    'fb-input-bar': FBInputBar,
    'fb-agent-activity': FBAgentActivity,
    'fb-typing-indicator': FBTypingIndicator,
    // Artifacts
    'fb-artifact-gantt': FBArtifactGantt,
    'fb-artifact-budget': FBArtifactBudget,
    'fb-artifact-invoice': FBArtifactInvoice,
    // Features
    'fb-file-drop': FBFileDrop,
    'fb-error-boundary': FBErrorBoundary,
    // V2 Settings
    'fb-settings-profile': FBSettingsProfile,
    'fb-settings-org': FBSettingsOrg,
    'fb-settings-team': FBSettingsTeam,
    // V2 Project
    'fb-project-header': FBProjectHeader,
    // V2 Onboarding
    'fb-onboard-flow': FBOnboardFlow,
    'fb-extraction-stream': FBExtractionStream,
    // V2 Schedule
    'fb-schedule-task-bar': FBScheduleTaskBar,
    'fb-schedule-diff': FBScheduleDiff,
    // V2 Contacts
    'fb-contact-phase-grid': FBContactPhaseGrid,
    'fb-contact-bulk-input': FBContactBulkInput,
    'fb-contact-detail': FBContactDetail,
    'fb-contact-directory': FBContactDirectory,
    // Notifications
    'fb-notification-bell': FBNotificationBell,
    'fb-notification-list': FBNotificationList,
    // Admin
    'fb-admin-shell': FBAdminShell,
    'fb-admin-sidebar': FBAdminSidebar,
    'fb-admin-dashboard': FBAdminDashboard,
    // Shadow
    'shadow-toggle': ShadowToggle,
    'shadow-layout': ShadowLayout,
    'shadow-nav': ShadowNav,
    'tribunal-log-feed': TribunalLogFeed,
    'tribunal-case-detail': TribunalCaseDetail,
    'shadow-docs-viewer': ShadowDocsViewer,
});

if ((import.meta as unknown as { env?: { DEV?: boolean } }).env?.DEV === true) {
    console.log(`[FutureBuild] Registered ${String(registered)} component(s)`);
    console.log('[FutureBuild] Frontend V2 initialized');
}

// Step 60.2.2: Load Test Harness - DEV only
const isDev = (import.meta as unknown as { env?: { DEV?: boolean } }).env?.DEV === true;
if (isDev) {
    import('./services/debug/load-test').then(({ loadTestService }) => {
        const fbGlobal = (window as unknown as { fb?: Record<string, unknown> }).fb ?? {};
        fbGlobal['loadTest'] = loadTestService;
        (window as unknown as { fb: Record<string, unknown> }).fb = fbGlobal;
    }).catch((err: unknown) => {
        console.warn('[FutureBuild] Failed to load LoadTestService', err);
    });
}
