/**
 * Load Test Service - "The Torture Chamber"
 * See PRODUCTION_PLAN.md Phase 8, Step 60.2.2
 *
 * Stress-tests the virtualized chat list by simulating high-throughput
 * message ingestion. Uses requestIdleCallback for browser-friendly batching.
 */

import { store } from '../../store/store';
import type { ChatMessage } from '../../store/types';

// ============================================================================
// Message Content Library (Data Entropy)
// ============================================================================

/** Short messages (20% of test corpus) */
const SHORT_MESSAGES = [
    'Got it.',
    'OK',
    'Thanks!',
    'Done.',
    'Confirmed.',
    'Will do.',
    '✓',
    'Sounds good.',
    'Noted.',
    'Approved.',
];

/** Medium messages (60% of test corpus) */
const MEDIUM_MESSAGES = [
    'I reviewed the latest invoice from ABC Lumber. The line items match our PO, but there is a slight variance in the tax calculation. I recommend approving with a note to reconcile.',
    'The framing crew is on schedule for tomorrow. Weather forecast looks clear until Thursday. I have notified the electrical sub to be ready for their walk-through.',
    'Budget update: We are currently at 68% of the projected cost with 72% of the work complete. This puts us slightly under budget. The main risk is the pending HVAC equipment delivery.',
    'Material lead times for the custom windows have been confirmed at 6 weeks. I have updated the Gantt chart and notified the finish carpentry crew of the adjusted timeline.',
    'Site visit completed. Foundation inspection passed with no flags. Photos have been uploaded to the project docs. Ready to proceed with framing.',
];

/** Long messages - markdown code blocks (10% of test corpus) */
const LONG_MESSAGES = [
    `Here is the updated project schedule breakdown:

\`\`\`
Phase 1: Foundation (Week 1-2)
  - Excavation: 3 days
  - Footings: 2 days
  - Foundation walls: 4 days

Phase 2: Framing (Week 3-5)
  - Floor system: 2 days
  - Wall framing: 5 days
  - Roof trusses: 3 days
\`\`\`

Let me know if you need any adjustments to these durations.`,
    `Invoice reconciliation complete. Summary:

\`\`\`json
{
  "vendor": "BuildRight Supply",
  "po_number": "PO-2026-0142",
  "invoice_total": 12450.00,
  "matched_items": 18,
  "variance": -23.50,
  "status": "APPROVED_WITH_VARIANCE"
}
\`\`\`

The variance is within tolerance. Recommend payment.`,
];

/** Artifact placeholders (10% of test corpus) */
const ARTIFACT_MESSAGES = [
    '📊 I have generated an updated budget forecast. [View Budget Artifact]',
    '📅 Here is the revised project timeline. [View Gantt Chart]',
    '📄 Invoice #INV-2026-0089 is ready for review. [View Invoice]',
    '📈 Monthly progress report attached. [View Report]',
];

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Selects a random item from an array.
 */
function randomFrom<T>(arr: T[]): T {
    return arr[Math.floor(Math.random() * arr.length)] as T;
}

/**
 * Generates a mixed message based on the entropy distribution:
 * - 20% Short
 * - 60% Medium
 * - 10% Long
 * - 10% Artifact
 */
function generateMixedMessage(): string {
    const roll = Math.random();
    if (roll < 0.2) return randomFrom(SHORT_MESSAGES);
    if (roll < 0.8) return randomFrom(MEDIUM_MESSAGES);
    if (roll < 0.9) return randomFrom(LONG_MESSAGES);
    return randomFrom(ARTIFACT_MESSAGES);
}

/**
 * Formats an ISO timestamp to a display time string.
 */
function formatDisplayTime(isoString: string): string {
    try {
        return new Date(isoString).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch {
        return '';
    }
}

/**
 * Creates a ChatMessage with realistic metadata.
 */
function createTestMessage(index: number): ChatMessage {
    const createdAt = new Date().toISOString();
    const indexStr = String(index);
    return {
        id: `load-test-${String(Date.now())}-${indexStr}`,
        role: Math.random() > 0.5 ? 'user' : 'assistant',
        content: generateMixedMessage(),
        createdAt,
        displayTime: formatDisplayTime(createdAt),
    };
}

// ============================================================================
// Load Test Service Class
// ============================================================================

/**
 * LoadTestService provides stress-testing capabilities for the chat UI.
 * Exposes flood() and stream() methods for high-throughput message injection.
 */
export class LoadTestService {
    private _streamIntervalId: number | null = null;
    private _floodActive = false;
    private _lastFrameTime = 0;
    private _jankCount = 0;

    /**
     * Floods the store with a batch of messages using requestIdleCallback.
     * Injects messages in batches of 5-10 per idle frame to avoid main thread freeze.
     * @param count Total number of messages to inject.
     */
    flood(count: number): void {
        if (this._floodActive) {
            console.warn('[LoadTest] Flood already in progress. Call stop() first.');
            return;
        }

        console.log(`[LoadTest] 🔥 Flooding ${String(count)} messages...`);
        this._floodActive = true;
        this._jankCount = 0;
        const startTime = performance.now();
        let injected = 0;

        const batchSize = 8; // Messages per idle callback

        const injectBatch = (deadline: IdleDeadline): void => {
            // Jank detection: check if we're meeting 60fps target
            const now = performance.now();
            if (this._lastFrameTime > 0 && now - this._lastFrameTime > 16) {
                this._jankCount++;
                console.warn(`[LoadTest] ⚠️ Jank detected: ${(now - this._lastFrameTime).toFixed(1)}ms gap`);
            }
            this._lastFrameTime = now;

            // Inject as many as we can within the idle deadline
            let batchCount = 0;
            while (injected < count && batchCount < batchSize && deadline.timeRemaining() > 0) {
                const msg = createTestMessage(injected);
                store.actions.addMessage(msg);
                injected++;
                batchCount++;
            }

            if (injected < count && this._floodActive) {
                requestIdleCallback(injectBatch, { timeout: 50 });
            } else {
                const elapsed = performance.now() - startTime;
                console.log(`[LoadTest] ✅ Flood complete: ${String(injected)} messages in ${elapsed.toFixed(0)}ms`);
                console.log(`[LoadTest] 📊 Jank events: ${String(this._jankCount)}`);
                this._floodActive = false;
            }
        };

        // Fallback for browsers without requestIdleCallback
        if ('requestIdleCallback' in window) {
            requestIdleCallback(injectBatch, { timeout: 50 });
        } else {
            // Fallback: use setTimeout chunking
            const fallbackInject = (): void => {
                for (let i = 0; i < batchSize && injected < count; i++) {
                    const msg = createTestMessage(injected);
                    store.actions.addMessage(msg);
                    injected++;
                }
                if (injected < count && this._floodActive) {
                    setTimeout(fallbackInject, 0);
                } else {
                    const elapsed = performance.now() - startTime;
                    console.log(`[LoadTest] ✅ Flood complete (fallback): ${String(injected)} messages in ${elapsed.toFixed(0)}ms`);
                    this._floodActive = false;
                }
            };
            fallbackInject();
        }
    }

    /**
     * Starts a continuous message stream at the specified rate.
     * @param ratePerSecond Messages injected per second.
     */
    stream(ratePerSecond: number): void {
        if (this._streamIntervalId !== null) {
            console.warn('[LoadTest] Stream already active. Call stop() first.');
            return;
        }

        const intervalMs = 1000 / ratePerSecond;
        console.log(`[LoadTest] 🌊 Starting stream at ${String(ratePerSecond)} msg/sec...`);

        let streamCount = 0;
        this._streamIntervalId = window.setInterval(() => {
            const msg = createTestMessage(streamCount);
            store.actions.addMessage(msg);
            streamCount++;

            // Log progress every 100 messages
            if (streamCount % 100 === 0) {
                console.log(`[LoadTest] 📨 Streamed ${String(streamCount)} messages`);
            }
        }, intervalMs);
    }

    /**
     * Stops any active flood or stream operation.
     */
    stop(): void {
        if (this._streamIntervalId !== null) {
            clearInterval(this._streamIntervalId);
            this._streamIntervalId = null;
            console.log('[LoadTest] 🛑 Stream stopped.');
        }
        if (this._floodActive) {
            this._floodActive = false;
            console.log('[LoadTest] 🛑 Flood aborted.');
        }
    }

    /**
     * Clears all messages from the store (reset for testing).
     */
    clear(): void {
        store.actions.setMessages([]);
        console.log('[LoadTest] 🗑️ Messages cleared.');
    }

    /**
     * Returns current message count in the store.
     */
    get messageCount(): number {
        return store.messages$.value.length;
    }
}

// Singleton instance
export const loadTestService = new LoadTestService();
