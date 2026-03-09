/**
 * Global Error Reporter (Sprint 6.3)
 * Captures, categorizes, and routes frontend errors.
 */

export type ErrorCategory = 'network' | 'auth' | 'data' | 'ai' | 'unknown';

export interface AppError {
    category: ErrorCategory;
    message: string;
    originalError?: any;
}

class ErrorReporter {
    private listeners: ((error: AppError) => void)[] = [];

    init() {
        window.addEventListener('unhandledrejection', (event) => {
            this.reportError({
                category: this.categorizeError(event.reason),
                message: this.extractMessage(event.reason),
                originalError: event.reason
            });
        });

        window.addEventListener('error', (event) => {
            this.reportError({
                category: this.categorizeError(event.error),
                message: event.message,
                originalError: event.error
            });
        });
    }

    reportError(error: AppError) {
        console.error('[ErrorReporter] Captured:', error);
        this.listeners.forEach(l => l(error));
    }

    addListener(listener: (error: AppError) => void) {
        this.listeners.push(listener);
    }

    removeListener(listener: (error: AppError) => void) {
        this.listeners = this.listeners.filter(l => l !== listener);
    }

    private categorizeError(error: any): ErrorCategory {
        const msg = String(error?.message || error || '').toLowerCase();
        if (msg.includes('network') || msg.includes('fetch')) return 'network';
        if (msg.includes('unauthorized') || msg.includes('auth') || msg.includes('401') || msg.includes('403')) return 'auth';
        if (msg.includes('ai') || msg.includes('manual_mode') || msg.includes('vertex')) return 'ai';
        if (msg.includes('json') || msg.includes('parse')) return 'data';
        return 'unknown';
    }

    private extractMessage(error: any): string {
        return error?.message || String(error) || 'An unexpected error occurred';
    }
}

export const errorReporter = new ErrorReporter();
