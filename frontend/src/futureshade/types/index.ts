/**
 * FutureShade Types - TypeScript interfaces matching Go types
 * See: internal/futureshade/types.go and tribunal/types.go
 */

// ShadowDoc represents a document analyzed by the Shadow layer.
export interface ShadowDoc {
  id: string;
  source_type: 'PRD' | 'Spec' | 'Code';
  source_id: string;
  content_hash: string;
  analysis: Record<string, unknown>;
  created_at: string; // ISO 8601
}

// DecisionStatus represents the outcome of a Tribunal decision.
export type DecisionStatus = 'pending' | 'approved' | 'rejected' | 'conflict';

// Vote represents a single model's vote in a Tribunal decision.
export interface Vote {
  model_id: string;
  decision: DecisionStatus;
  confidence: number; // 0.0 to 1.0
  reasoning: string;
  cast_at: string; // ISO 8601
}

// Decision represents a consensus decision made by The Tribunal.
export interface Decision {
  id: string;
  source_type: string;
  source_id: string;
  question: string;
  votes: Vote[];
  final_verdict: DecisionStatus;
  created_at: string; // ISO 8601
  resolved_at?: string; // ISO 8601
}

// TribunalConfig holds configuration for the Tribunal consensus system.
export interface TribunalConfig {
  minimum_votes: number;
  consensus_percent: number; // 0.0-1.0
}

// EventType categorizes system events for observation.
export type EventType =
  | 'pr_created'
  | 'pr_updated'
  | 'task_completed'
  | 'document_added'
  | 'schedule_change';

// Event represents a system event observed by the Shadow layer.
export interface ShadowEvent {
  type: EventType;
  source_id: string;
  payload: Record<string, unknown>;
}

// HealthResponse is the response from the FutureShade health endpoint.
export interface HealthResponse {
  status: 'active' | 'disabled' | 'degraded';
  tribunal_count: number;
}
