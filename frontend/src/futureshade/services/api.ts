/**
 * FutureShade API Service
 * Client for FutureShade backend endpoints (Admin only)
 */

import type { HealthResponse, ShadowDoc, Decision } from '../types';

const BASE_URL = '/api/v1/futureshade';

/**
 * FutureShadeService provides API access to FutureShade endpoints.
 * Note: All endpoints require Admin authentication.
 */
export class FutureShadeService {
  private token: string | null = null;

  setToken(token: string): void {
    this.token = token;
  }

  private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...(this.token ? { Authorization: `Bearer ${this.token}` } : {}),
    };

    const response = await fetch(`${BASE_URL}${path}`, {
      ...options,
      headers: { ...headers, ...(options.headers || {}) },
    });

    if (!response.ok) {
      throw new Error(`FutureShade API error: ${response.status} ${response.statusText}`);
    }

    return response.json();
  }

  /**
   * Check FutureShade service health.
   */
  async health(): Promise<HealthResponse> {
    return this.request<HealthResponse>('/health');
  }

  /**
   * List all ShadowDocs (paginated).
   * Stub for Step 65+.
   */
  async listShadowDocs(_limit = 50, _offset = 0): Promise<ShadowDoc[]> {
    // TODO: Implement when backend endpoint is ready
    return [];
  }

  /**
   * Get a specific ShadowDoc by ID.
   * Stub for Step 65+.
   */
  async getShadowDoc(_id: string): Promise<ShadowDoc | null> {
    // TODO: Implement when backend endpoint is ready
    return null;
  }

  /**
   * List Tribunal decisions (paginated).
   * Stub for Step 65+.
   */
  async listDecisions(_limit = 50, _offset = 0): Promise<Decision[]> {
    // TODO: Implement when backend endpoint is ready
    return [];
  }

  /**
   * Get a specific Tribunal decision by ID.
   * Stub for Step 65+.
   */
  async getDecision(_id: string): Promise<Decision | null> {
    // TODO: Implement when backend endpoint is ready
    return null;
  }
}

// Singleton instance
export const futureShadeService = new FutureShadeService();
