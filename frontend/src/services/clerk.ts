/**
 * Clerk Identity Service - Singleton wrapper for @clerk/clerk-js
 * See PHASE_12_PRD.md Section 3.1 and STEP_78_AUTH_PROVIDER.md
 *
 * Provides:
 * - Initialization and lifecycle management
 * - Synchronous token cache (Clerk's getToken is async; http.ts needs sync)
 * - Session change observer for store integration
 * - Mount helpers for Clerk UI components
 *
 * IMPORTANT: This module does NOT import from store to prevent circular dependencies.
 * Auth state changes are communicated via the onAuthChange callback.
 */

import { Clerk } from '@clerk/clerk-js';

// ============================================================================
// Types
// ============================================================================

/**
 * User info extracted from a Clerk session for store consumption.
 * Maps to the existing store User type.
 */
export interface ClerkUser {
    id: string;
    email: string;
    name: string;
    role: string;
    orgId: string;
}

/**
 * Callback fired when Clerk auth state changes.
 */
export type AuthChangeCallback = (user: ClerkUser | null, token: string | null) => void;

// ============================================================================
// Clerk Appearance (Construction Professional Dark Theme)
// ============================================================================

const CLERK_APPEARANCE = {
    variables: {
        colorPrimary: '#3b82f6',       // --fb-primary
        colorBackground: '#111111',     // --fb-bg-card
        colorInputBackground: '#1a1a1a', // --fb-bg-tertiary
        colorInputText: '#ffffff',      // --fb-text-primary
        colorText: '#ffffff',
        colorTextSecondary: '#aaaaaa',  // --fb-text-secondary
        borderRadius: '8px',            // --fb-radius-lg
        fontFamily: 'Inter, system-ui, sans-serif',
    },
    elements: {
        card: {
            backgroundColor: '#111111',
            border: '1px solid #333333',
            borderRadius: '12px',
        },
        formButtonPrimary: {
            backgroundColor: '#3b82f6',
        },
        footerActionLink: {
            color: '#3b82f6',
        },
    },
};

// ============================================================================
// Singleton State
// ============================================================================

let clerk: Clerk | null = null;
let cachedToken: string | null = null;
let authChangeCallback: AuthChangeCallback | null = null;
let tokenRefreshTimer: ReturnType<typeof setInterval> | null = null;
let isLoaded = false;

// ============================================================================
// Internal Helpers
// ============================================================================

/**
 * Extract user info from the active Clerk session.
 */
function extractUser(): ClerkUser | null {
    if (!clerk?.user || !clerk.session) return null;

    const user = clerk.user;
    const org = clerk.organization;

    // Find the membership for the active organization to determine role
    let role = 'Builder';
    if (org) {
        const membership = user.organizationMemberships.find(
            (m) => m.organization.id === org.id
        );
        if (membership?.role === 'org:admin') {
            role = 'Admin';
        }
    }

    return {
        id: user.id,
        email: user.primaryEmailAddress?.emailAddress ?? '',
        name: [user.firstName, user.lastName].filter(Boolean).join(' ') || user.primaryEmailAddress?.emailAddress || '',
        role,
        orgId: org?.id ?? '',
    };
}

/**
 * Refresh the cached token from Clerk's session.
 * Called on session change and periodically.
 */
async function refreshTokenCache(): Promise<void> {
    if (!clerk?.session) {
        cachedToken = null;
        return;
    }

    try {
        cachedToken = await clerk.session.getToken() ?? null;
    } catch {
        cachedToken = null;
    }
}

/**
 * Handle Clerk session/user state changes.
 * Fires the registered callback with current user and token.
 */
async function handleStateChange(): Promise<void> {
    await refreshTokenCache();
    const user = extractUser();

    if (authChangeCallback) {
        authChangeCallback(user, cachedToken);
    }
}

// ============================================================================
// Public API
// ============================================================================

export const clerkService = {
    /**
     * Initialize Clerk with the publishable key.
     * Must be called before any other method.
     * @param publishableKey - Clerk publishable key (pk_test_... or pk_live_...)
     */
    async init(publishableKey: string): Promise<void> {
        if (clerk) return; // Already initialized

        clerk = new Clerk(publishableKey);

        await clerk.load({
            appearance: CLERK_APPEARANCE,
            signInUrl: '/',
            afterSignInUrl: '/',
            afterSignUpUrl: '/',
        });

        isLoaded = true;

        // Listen for all state changes (session, user, org)
        clerk.addListener(() => {
            void handleStateChange();
        });

        // Initial state sync
        await handleStateChange();

        // Refresh token every 50 seconds (Clerk tokens expire at ~60s)
        tokenRefreshTimer = setInterval(() => {
            void refreshTokenCache();
        }, 50_000);
    },

    /**
     * Whether Clerk has finished loading.
     */
    get loaded(): boolean {
        return isLoaded;
    },

    /**
     * Whether a user is currently signed in.
     */
    get isSignedIn(): boolean {
        return clerk?.isSignedIn ?? false;
    },

    /**
     * Get the cached JWT token synchronously.
     * Returns null if not authenticated or token not yet cached.
     * http.ts tokenGetter wires directly to this.
     */
    getToken(): string | null {
        return cachedToken;
    },

    /**
     * Register a callback for auth state changes.
     * Called with (user, token) on sign-in and (null, null) on sign-out.
     */
    onAuthChange(callback: AuthChangeCallback): void {
        authChangeCallback = callback;
    },

    /**
     * Mount the Clerk Sign In component into a container element.
     * @param element - Target HTMLDivElement
     */
    mountSignIn(element: HTMLDivElement): void {
        if (!clerk) return;
        clerk.mountSignIn(element, {
            appearance: CLERK_APPEARANCE,
        });
    },

    /**
     * Unmount the Clerk Sign In component from a container element.
     * @param element - Target HTMLDivElement
     */
    unmountSignIn(element: HTMLDivElement): void {
        if (!clerk) return;
        clerk.unmountSignIn(element);
    },

    /**
     * Mount the Clerk User Button component.
     * @param element - Target HTMLDivElement
     */
    mountUserButton(element: HTMLDivElement): void {
        if (!clerk) return;
        clerk.mountUserButton(element, {
            appearance: CLERK_APPEARANCE,
        });
    },

    /**
     * Unmount the Clerk User Button component.
     * @param element - Target HTMLDivElement
     */
    unmountUserButton(element: HTMLDivElement): void {
        if (!clerk) return;
        clerk.unmountUserButton(element);
    },

    /**
     * Sign out the current user.
     */
    async signOut(): Promise<void> {
        if (!clerk) return;
        cachedToken = null;
        await clerk.signOut();
    },

    /**
     * Clean up resources (for testing or unmount).
     */
    destroy(): void {
        if (tokenRefreshTimer) {
            clearInterval(tokenRefreshTimer);
            tokenRefreshTimer = null;
        }
        cachedToken = null;
        authChangeCallback = null;
        isLoaded = false;
        clerk = null;
    },
};
