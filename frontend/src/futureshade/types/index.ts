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

// ============================================================
// Shadow Viewer Types (See SHADOW_VIEWER_specs.md)
// ============================================================

// VoteType represents an individual model's vote.
// Maps to tribunal_vote_type enum in database.
export type VoteType = 'YEA' | 'NAY' | 'ABSTAIN';

// ModelVote represents a single model's vote in a Tribunal decision.
// See SHADOW_VIEWER_specs.md Section 3.1
export interface ModelVote {
  model: string;
  vote: VoteType;
  reasoning: string;
  latency_ms: number;
  cost_usd: number;
}

// DecisionSummary is the list view response for tribunal decisions.
// See SHADOW_VIEWER_specs.md Section 3.1 GET /api/v1/tribunal/decisions
export interface DecisionSummary {
  id: string;
  case_id: string;
  status: DecisionStatus;
  context: string;
  timestamp: string; // ISO 8601
  models_consulted: string[];
}

// DecisionDetail is the full detail view response including individual model votes.
// See SHADOW_VIEWER_specs.md Section 3.1 GET /api/v1/tribunal/decisions/{id}
export interface DecisionDetail {
  id: string;
  case_id: string;
  status: DecisionStatus;
  context: string;
  consensus_score: number;
  votes: ModelVote[];
  policy_links?: string[];
  timestamp: string; // ISO 8601
}

// ListDecisionsFilter holds query parameters for filtering decisions.
export interface ListDecisionsFilter {
  limit?: number;
  offset?: number;
  status?: DecisionStatus;
  model?: string;
  start_date?: string; // ISO 8601
  end_date?: string; // ISO 8601
  search?: string;
}

// ListDecisionsResponse is the paginated response for the list endpoint.
export interface ListDecisionsResponse {
  decisions: DecisionSummary[];
  total: number;
  has_more: boolean;
}

// ============================================================
// ShadowDocs Types (See SHADOW_VIEWER_specs.md Section 3.2)
// ============================================================

// FileType represents the type of a file system node.
export type FileType = 'dir' | 'file';

// TreeNode represents a file or directory in the docs tree.
export interface TreeNode {
  name: string;
  type: FileType;
  path?: string;
  children?: TreeNode[];
}

// TreeResponse is the response for the docs tree endpoint.
// GET /api/v1/shadow/docs/tree
export interface TreeResponse {
  roots: TreeNode[];
}

// ContentResponse is the response for the docs content endpoint.
// GET /api/v1/shadow/docs/content
export interface ContentResponse {
  path: string;
  content: string;
}
