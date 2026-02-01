/**
 * Onboarding Store - Signals-Based State for Smart Onboarding Wizard
 * See STEP_76_REALTIME_FORM_FILLING.md
 *
 * Implements bidirectional state synchronization between chat and form panels.
 * State changes propagate instantly via Signals, with visual indicators for AI vs user edits.
 */

import { signal, computed } from '@preact/signals-core';
import type { CreateProjectRequest } from '../services/api';

// ============================================================================
// Types
// ============================================================================

export interface OnboardingMessage {
    id: string;
    role: 'user' | 'assistant' | 'system';
    content: string;
    timestamp: Date;
}

export type FieldSource = 'user' | 'ai' | 'default';

// ============================================================================
// Core State Signals
// ============================================================================

/**
 * Form field values extracted from conversation or user input.
 */
export const onboardingValues = signal<Partial<CreateProjectRequest>>({});

/**
 * Tracks the source of each field value (user, ai, or default).
 * Used to determine visual treatment.
 */
export const onboardingSources = signal<Record<string, FieldSource>>({});

/**
 * Confidence scores (0.0-1.0) for AI-populated fields.
 * Fields with confidence < 0.8 show "Verify" badge.
 */
export const onboardingConfidence = signal<Record<string, number>>({});

/**
 * Conversation messages between user and Interrogator agent.
 */
export const onboardingMessages = signal<OnboardingMessage[]>([]);

/**
 * Whether the agent is currently processing a request.
 */
export const isProcessing = signal<boolean>(false);

/**
 * Recently updated fields for glow animation (transient state).
 */
export const recentlyUpdatedFields = signal<Set<string>>(new Set());

// ============================================================================
// Computed Values
// ============================================================================

/**
 * Whether the form has minimum required fields to create a project.
 * Required: name, address
 */
export const isReadyToCreate = computed(() => {
    const v = onboardingValues.value;
    return !!(v.name && v.address);
});

/**
 * List of fields that need user verification (low confidence < 0.8).
 */
export const fieldsNeedingVerification = computed(() => {
    const conf = onboardingConfidence.value;
    return Object.entries(conf)
        .filter(([_, score]) => score < 0.8)
        .map(([field, _]) => field);
});

// ============================================================================
// Actions
// ============================================================================

/**
 * Update a single field value from user input.
 * Marks source as 'user' and clears AI confidence.
 */
export function setFieldValue(field: keyof CreateProjectRequest, value: string | number): void {
    onboardingValues.value = {
        ...onboardingValues.value,
        [field]: value
    };
    onboardingSources.value = {
        ...onboardingSources.value,
        [field]: 'user'
    };
    // Clear confidence when user manually edits
    const conf = { ...onboardingConfidence.value };
    const rest: Record<string, number> = {};
    for (const [k, v] of Object.entries(conf)) {
        if (k !== field) rest[k] = v;
    }
    onboardingConfidence.value = rest;
}

/**
 * Apply AI extraction results from Interrogator agent.
 * Only updates empty fields or fields previously populated by AI.
 * User edits are never overwritten.
 */
// Fix 7: Numeric fields that need type coercion (AI may return strings)
const NUMERIC_FIELDS = new Set<string>([
    'square_footage', 'bedrooms', 'bathrooms', 'stories', 'lot_size'
]);

export function applyAIExtraction(
    extractedValues: Record<string, unknown>,
    confidenceScores: Record<string, number>
): void {
    const currentValues = { ...onboardingValues.value };
    const currentSources = { ...onboardingSources.value };
    const currentConf = { ...onboardingConfidence.value };
    const updatedFields = new Set<string>();

    for (const [field, rawValue] of Object.entries(extractedValues)) {
        // Fix 7: Coerce string values to numbers for numeric fields
        let value: unknown = rawValue;
        if (NUMERIC_FIELDS.has(field) && typeof rawValue === 'string') {
            const num = Number(rawValue);
            if (!isNaN(num)) value = num;
        }

        // Only apply if field is empty OR was previously AI-populated
        const existingSource = currentSources[field];
        if (!currentValues[field as keyof CreateProjectRequest] || existingSource === 'ai' || existingSource === 'default') {
            (currentValues as Record<string, unknown>)[field] = value;
            currentSources[field] = 'ai';
            currentConf[field] = confidenceScores[field] ?? 0.5;
            updatedFields.add(field);
        }
    }

    onboardingValues.value = currentValues;
    onboardingSources.value = currentSources;
    onboardingConfidence.value = currentConf;

    // Mark fields as recently updated for glow animation
    recentlyUpdatedFields.value = updatedFields;

    // Clear animation state after 600ms
    setTimeout(() => {
        recentlyUpdatedFields.value = new Set();
    }, 600);
}

/**
 * Add a message to the conversation history.
 */
export function addMessage(message: OnboardingMessage): void {
    onboardingMessages.value = [...onboardingMessages.value, message];
}

/**
 * Reset all wizard state (for new session or cancellation).
 */
export function resetOnboarding(): void {
    onboardingValues.value = {};
    onboardingSources.value = {};
    onboardingConfidence.value = {};
    onboardingMessages.value = [];
    isProcessing.value = false;
    recentlyUpdatedFields.value = new Set();
}
