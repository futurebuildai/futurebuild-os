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

// Layout components (Step 51.3)
import { FBNavRail } from './components/layout/fb-nav-rail';
import { FBHeader } from './components/layout/fb-header';
import { FBAppShell } from './components/layout/fb-app-shell';

// App shell
import './app-root';

// Register all components
const registered = registerComponents({
    'fb-demo-button': FBDemoButton,
    'fb-nav-rail': FBNavRail,
    'fb-header': FBHeader,
    'fb-app-shell': FBAppShell,
});

console.log(`[FutureBuild] Registered ${String(registered)} component(s)`);
console.log('[FutureBuild] Frontend initialized');

