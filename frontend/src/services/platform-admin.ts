/**
 * Platform Admin Utility
 * UX guard for platform-level admin access. Backend is the real security boundary.
 */

const PLATFORM_ADMIN_EMAILS: readonly string[] = ['colton@futurebuild.ai'];

export function isPlatformAdmin(email: string): boolean {
    return PLATFORM_ADMIN_EMAILS.includes(email.toLowerCase());
}
