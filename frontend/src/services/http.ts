/**
 * HTTP Service - Typed Fetch Wrapper
 * See FRONTEND_SCOPE.md Section 5.2
 *
 * Provides a strongly-typed HTTP client with:
 * - Automatic Authorization header injection
 * - Generic JSON response parsing
 * - 401 Unauthorized handling via configurable callback
 * - Request timeout support
 *
 * IMPORTANT: This module does NOT import from store to prevent circular dependencies.
 * Token getter and 401 handler are injected at bootstrap time.
 */

// ============================================================================
// Types
// ============================================================================

/**
 * Standard API response wrapper.
 * All backend endpoints return this shape.
 */
export interface ApiResponse<T> {
    data: T;
    message?: string;
}

/**
 * API error response from backend.
 */
export interface ApiErrorResponse {
    error: string;
    code?: string;
    details?: Record<string, unknown>;
}

/**
 * Custom error class for API failures.
 */
export class ApiError extends Error {
    constructor(
        message: string,
        public readonly status: number,
        public readonly code?: string,
        public readonly details?: Record<string, unknown>
    ) {
        super(message);
        this.name = 'ApiError';
    }
}

/**
 * HTTP request options.
 */
export interface RequestOptions {
    method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
    body?: unknown;
    headers?: Record<string, string>;
    timeout?: number;
    skipAuth?: boolean;
}

// ============================================================================
// Configuration (Dependency Injection)
// ============================================================================

const API_BASE = '/api/v1';
const DEFAULT_TIMEOUT = 30000; // 30 seconds

/**
 * Token getter function - injected at bootstrap.
 * Returns the current auth token or null.
 */
let tokenGetter: (() => string | null) | null = null;

/**
 * Unauthorized handler - injected at bootstrap.
 * Called when a 401 response is received.
 */
let onUnauthorized: (() => void) | null = null;

/**
 * Configures the token getter function.
 * Called once at application bootstrap to avoid circular dependencies.
 *
 * @param getter - Function that returns the current auth token
 */
export function setTokenGetter(getter: () => string | null): void {
    tokenGetter = getter;
}

/**
 * Configures the 401 handler callback.
 * Called once at application bootstrap to avoid circular dependencies.
 *
 * @param handler - Function to call when 401 is received (typically logout)
 */
export function setUnauthorizedHandler(handler: () => void): void {
    onUnauthorized = handler;
}

// ============================================================================
// HTTP Client
// ============================================================================

/**
 * Performs an HTTP request with automatic auth header injection.
 *
 * @typeParam T - The expected response data type
 * @param endpoint - API endpoint path (without base URL)
 * @param options - Request configuration
 * @returns Promise resolving to the typed response data
 * @throws {ApiError} When the request fails or returns an error status
 *
 * @example
 * ```typescript
 * // GET request
 * const projects = await request<Project[]>('/projects');
 *
 * // POST request with body
 * const result = await request<AuthResponse>('/auth/login', {
 *   method: 'POST',
 *   body: { email: 'user@example.com' }
 * });
 * ```
 */
export async function request<T>(
    endpoint: string,
    options: RequestOptions = {}
): Promise<T> {
    const {
        method = 'GET',
        body,
        headers = {},
        timeout = DEFAULT_TIMEOUT,
        skipAuth = false,
    } = options;

    // Build headers
    const requestHeaders: Record<string, string> = {
        'Content-Type': 'application/json',
        Accept: 'application/json',
        ...headers,
    };

    // Inject auth token if available and not skipped
    if (!skipAuth && tokenGetter) {
        const token = tokenGetter();
        if (token) {
            requestHeaders['Authorization'] = `Bearer ${token}`;
        }
    }

    // Build request init
    const init: RequestInit = {
        method,
        headers: requestHeaders,
    };

    if (body !== undefined) {
        init.body = JSON.stringify(body);
    }

    // Create abort controller for timeout
    const controller = new AbortController();
    init.signal = controller.signal;

    const timeoutId = setTimeout(() => {
        controller.abort();
    }, timeout);

    try {
        const response = await fetch(`${API_BASE}${endpoint}`, init);

        clearTimeout(timeoutId);

        // Handle 401 Unauthorized
        if (response.status === 401) {
            if (onUnauthorized) {
                onUnauthorized();
            }
            throw new ApiError('Unauthorized', 401, 'UNAUTHORIZED');
        }

        // Handle non-OK responses
        if (!response.ok) {
            const errorBody = (await response.json().catch(() => ({
                error: 'Unknown error',
            }))) as ApiErrorResponse;

            throw new ApiError(
                errorBody.error || `HTTP ${String(response.status)}`,
                response.status,
                errorBody.code,
                errorBody.details
            );
        }

        // Handle 204 No Content
        if (response.status === 204) {
            return undefined as T;
        }

        // Parse JSON response
        const data = (await response.json()) as T;
        return data;
    } catch (error) {
        clearTimeout(timeoutId);

        if (error instanceof ApiError) {
            throw error;
        }

        if (error instanceof Error) {
            if (error.name === 'AbortError') {
                throw new ApiError('Request timeout', 408, 'TIMEOUT');
            }
            throw new ApiError(error.message, 0, 'NETWORK_ERROR');
        }

        throw new ApiError('Unknown error', 0, 'UNKNOWN');
    }
}

/**
 * Convenience method for GET requests.
 */
export function get<T>(
    endpoint: string,
    options?: Omit<RequestOptions, 'method' | 'body'>
): Promise<T> {
    return request<T>(endpoint, { ...options, method: 'GET' });
}

/**
 * Convenience method for POST requests.
 */
export function post<T>(
    endpoint: string,
    body?: unknown,
    options?: Omit<RequestOptions, 'method' | 'body'>
): Promise<T> {
    return request<T>(endpoint, { ...options, method: 'POST', body });
}

/**
 * Convenience method for PUT requests.
 */
export function put<T>(
    endpoint: string,
    body?: unknown,
    options?: Omit<RequestOptions, 'method' | 'body'>
): Promise<T> {
    return request<T>(endpoint, { ...options, method: 'PUT', body });
}

/**
 * Convenience method for PATCH requests.
 */
export function patch<T>(
    endpoint: string,
    body?: unknown,
    options?: Omit<RequestOptions, 'method' | 'body'>
): Promise<T> {
    return request<T>(endpoint, { ...options, method: 'PATCH', body });
}

/**
 * Convenience method for DELETE requests.
 */
export function del<T>(
    endpoint: string,
    options?: Omit<RequestOptions, 'method' | 'body'>
): Promise<T> {
    return request<T>(endpoint, { ...options, method: 'DELETE' });
}
