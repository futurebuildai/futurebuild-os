/**
 * Correction Types — Sprint 3.2: Interactive Learning
 *
 * Defines the CorrectionEvent interface used to capture user corrections
 * to AI-extracted artifact fields. These events are sent to the backend
 * for model improvement and audit logging.
 */

/**
 * A single correction event, representing a user edit to an AI-extracted field.
 * Created when the user saves changes to a field that was originally populated
 * by the VisionService extraction pipeline.
 */
export interface CorrectionEvent {
    /** Artifact UUID (e.g. invoice ID) */
    artifactId: string;
    /** Artifact type discriminator */
    artifactType: 'invoice' | 'budget' | 'schedule';
    /** Dot-notation field path, e.g. "line_items[2].unit_price_cents" */
    fieldPath: string;
    /** The original AI-extracted value */
    oldValue: unknown;
    /** The user-corrected value */
    newValue: unknown;
    /** AI confidence score at extraction time (0.0–1.0) */
    originalConfidence: number;
    /** ISO 8601 timestamp of the correction */
    timestamp: string;
    /** User identifier — backend resolves actual ID from auth token */
    userId: string;
}
