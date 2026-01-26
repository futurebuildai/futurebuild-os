/**
 * FutureBuild Frontend - Main Entry Point
 * See FRONTEND_SCOPE.md Section 4.2
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
import { FBDemoButton } from './components/base/demo-button';

// Layout components (3-Panel Agent Command Center - Step 51.3)
import { FBAppShell } from './components/layout/fb-app-shell';
import { FBPanelLeft } from './components/layout/fb-panel-left';
import { FBPanelCenter } from './components/layout/fb-panel-center';
import { FBPanelRight } from './components/layout/fb-panel-right';

// Chat Components (Step 52)
import { FBMessageList } from './components/chat/fb-message-list';
import { FBActionCard } from './components/chat/fb-action-card';
import { FBInputBar } from './components/chat/fb-input-bar';
import { FBAgentActivity } from './components/agent/fb-agent-activity'; // Step 53
import { FBTypingIndicator } from './components/chat/fb-typing-indicator'; // Step 57

// Artifact Components (Step 55)
import { FBArtifactGantt } from './components/artifacts/fb-artifact-gantt';
import { FBArtifactBudget } from './components/artifacts/fb-artifact-budget';
import { FBArtifactInvoice } from './components/artifacts/fb-artifact-invoice';

// Feature Components (Step 56)
import { FBFileDrop } from './components/features/fb-file-drop';

// Base Components (Step 58.5: Fortress Hardening)
import { FBErrorBoundary } from './components/base/fb-error-boundary';

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
    'fb-demo-button': FBDemoButton,
    'fb-app-shell': FBAppShell,
    'fb-panel-left': FBPanelLeft,
    'fb-panel-center': FBPanelCenter,
    'fb-panel-right': FBPanelRight,
    'fb-message-list': FBMessageList,
    'fb-action-card': FBActionCard,
    'fb-input-bar': FBInputBar,
    'fb-agent-activity': FBAgentActivity,
    'fb-artifact-gantt': FBArtifactGantt,
    'fb-artifact-budget': FBArtifactBudget,
    'fb-artifact-invoice': FBArtifactInvoice,
    'fb-file-drop': FBFileDrop,
    'fb-typing-indicator': FBTypingIndicator,
    'fb-error-boundary': FBErrorBoundary,
    // Shadow Viewer components (SHADOW_VIEWER_specs.md)
    'shadow-toggle': ShadowToggle,
    'shadow-layout': ShadowLayout,
    'shadow-nav': ShadowNav,
    'tribunal-log-feed': TribunalLogFeed,
    'tribunal-case-detail': TribunalCaseDetail,
    'shadow-docs-viewer': ShadowDocsViewer,
});

console.log(`[FutureBuild] Registered ${String(registered)} component(s)`);
console.log('[FutureBuild] Frontend initialized');

// Step 60.2.2: Load Test Harness - DEV only
// Exposes window.fb.loadTest for console stress-testing
const isDev = (import.meta as unknown as { env?: { DEV?: boolean } }).env?.DEV === true;
if (isDev) {
    import('./services/debug/load-test').then(({ loadTestService }) => {
        // Extend existing window.fb or create it
        const fbGlobal = (window as unknown as { fb?: Record<string, unknown> }).fb ?? {};
        fbGlobal['loadTest'] = loadTestService;
        (window as unknown as { fb: Record<string, unknown> }).fb = fbGlobal;
        console.log('[FutureBuild] 🧪 LoadTestService attached to window.fb.loadTest');
    }).catch((err: unknown) => {
        console.warn('[FutureBuild] Failed to load LoadTestService', err);
    });
}
