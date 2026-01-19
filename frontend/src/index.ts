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

// Artifact Components (Step 55)
import { FBArtifactGantt } from './components/artifacts/fb-artifact-gantt';
import { FBArtifactBudget } from './components/artifacts/fb-artifact-budget';
import { FBArtifactInvoice } from './components/artifacts/fb-artifact-invoice';

// Feature Components (Step 56)
import { FBFileDrop } from './components/features/fb-file-drop';

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
});

console.log(`[FutureBuild] Registered ${String(registered)} component(s)`);
console.log('[FutureBuild] Frontend initialized');
