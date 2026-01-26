/**
 * FutureShade API Service
 * Client for FutureShade backend endpoints (Admin only)
 * See SHADOW_VIEWER_specs.md
 */

import type {
  HealthResponse,
  ShadowDoc,
  Decision,
  DecisionDetail,
  ListDecisionsResponse,
  ListDecisionsFilter,
  TreeResponse,
  ContentResponse,
} from '../types';

const BASE_URL = '/api/v1/futureshade';
const TRIBUNAL_URL = '/api/v1/tribunal';
const SHADOW_URL = '/api/v1/shadow';

/**
 * FutureShadeService provides API access to FutureShade endpoints.
 * Note: All endpoints require Admin authentication.
 */
export class FutureShadeService {
  private token: string | null = null;

  setToken(token: string): void {
    this.token = token;
  }

  private async request<T>(url: string, options: RequestInit = {}): Promise<T> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...(this.token ? { Authorization: `Bearer ${this.token}` } : {}),
    };

    const response = await fetch(url, {
      ...options,
      headers: { ...headers, ...(options.headers || {}) },
    });

    if (!response.ok) {
      const errorBody = await response.text();
      throw new Error(`API error: ${response.status} ${response.statusText} - ${errorBody}`);
    }

    return response.json();
  }

  /**
   * Check FutureShade service health.
   */
  async health(): Promise<HealthResponse> {
    return this.request<HealthResponse>(`${BASE_URL}/health`);
  }

  /**
   * List all ShadowDocs (paginated).
   * Stub for future implementation.
   */
  async listShadowDocs(_limit = 50, _offset = 0): Promise<ShadowDoc[]> {
    // TODO: Implement when backend endpoint is ready
    return [];
  }

  /**
   * Get a specific ShadowDoc by ID.
   * Stub for future implementation.
   */
  async getShadowDoc(_id: string): Promise<ShadowDoc | null> {
    // TODO: Implement when backend endpoint is ready
    return null;
  }

  // ============================================================
  // Tribunal Endpoints (See SHADOW_VIEWER_specs.md Section 3.1)
  // ============================================================

  /**
   * List Tribunal decisions with filtering.
   * GET /api/v1/tribunal/decisions
   */
  async listDecisions(filter: ListDecisionsFilter = {}): Promise<ListDecisionsResponse> {
    const params = new URLSearchParams();
    if (filter.limit) params.set('limit', filter.limit.toString());
    if (filter.offset) params.set('offset', filter.offset.toString());
    if (filter.status) params.set('status', filter.status);
    if (filter.model) params.set('model', filter.model);
    if (filter.start_date) params.set('start_date', filter.start_date);
    if (filter.end_date) params.set('end_date', filter.end_date);
    if (filter.search) params.set('search', filter.search);

    const queryString = params.toString();
    const url = queryString ? `${TRIBUNAL_URL}/decisions?${queryString}` : `${TRIBUNAL_URL}/decisions`;

    return this.request<ListDecisionsResponse>(url);
  }

  /**
   * Get a specific Tribunal decision by ID with full details.
   * GET /api/v1/tribunal/decisions/{id}
   */
  async getDecision(id: string): Promise<DecisionDetail> {
    return this.request<DecisionDetail>(`${TRIBUNAL_URL}/decisions/${id}`);
  }

  /**
   * @deprecated Use listDecisions(filter) instead
   */
  async listDecisionsLegacy(_limit = 50, _offset = 0): Promise<Decision[]> {
    const response = await this.listDecisions({ limit: _limit, offset: _offset });
    // Map to legacy format - this is approximate, use new methods instead
    return response.decisions.map((d) => ({
      id: d.id,
      source_type: 'Tribunal',
      source_id: d.case_id,
      question: d.context,
      votes: [],
      final_verdict: d.status,
      created_at: d.timestamp,
    }));
  }

  // ============================================================
  // ShadowDocs Endpoints (See SHADOW_VIEWER_specs.md Section 3.2)
  // ============================================================

  /**
   * Get the docs/specs file tree.
   * GET /api/v1/shadow/docs/tree
   */
  async getDocsTree(): Promise<TreeResponse> {
    return this.request<TreeResponse>(`${SHADOW_URL}/docs/tree`);
  }

  /**
   * Get file content by path.
   * GET /api/v1/shadow/docs/content?path={path}
   */
  async getDocContent(path: string): Promise<ContentResponse> {
    return this.request<ContentResponse>(
      `${SHADOW_URL}/docs/content?path=${encodeURIComponent(path)}`
    );
  }
}

// Singleton instance
export const futureShadeService = new FutureShadeService();
