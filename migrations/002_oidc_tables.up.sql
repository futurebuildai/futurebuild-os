-- 002: OIDC Provider tables
-- Supports authorization code flow with PKCE, refresh tokens, and key rotation.

-- OIDC Clients (registered relying parties)
CREATE TABLE oidc_clients (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id           TEXT NOT NULL UNIQUE,            -- Public identifier (e.g. "fb-os")
    client_secret_hash  TEXT,                            -- bcrypt hash (NULL for public clients using PKCE)
    client_name         TEXT NOT NULL,
    redirect_uris       TEXT[] NOT NULL,
    grant_types         TEXT[] NOT NULL DEFAULT '{authorization_code}',
    response_types      TEXT[] NOT NULL DEFAULT '{code}',
    token_endpoint_auth TEXT NOT NULL DEFAULT 'none',    -- none (public+PKCE), client_secret_post
    scopes              TEXT[] NOT NULL DEFAULT '{openid,profile,email,org}',
    is_active           BOOLEAN NOT NULL DEFAULT true,
    dev_mode            BOOLEAN NOT NULL DEFAULT false,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- OIDC Authorization Requests (ephemeral)
CREATE TABLE oidc_auth_requests (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id             TEXT NOT NULL REFERENCES oidc_clients(client_id),
    redirect_uri          TEXT NOT NULL,
    scopes                TEXT[] NOT NULL,
    state                 TEXT NOT NULL,
    nonce                 TEXT,
    code_challenge        TEXT,                    -- PKCE S256
    code_challenge_method TEXT DEFAULT 'S256',
    response_type         TEXT NOT NULL DEFAULT 'code',
    response_mode         TEXT,
    user_id               UUID,                    -- Set after authentication
    auth_code             TEXT UNIQUE,             -- Set after consent
    authenticated         BOOLEAN DEFAULT false,
    auth_time             TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at            TIMESTAMPTZ NOT NULL     -- Short-lived (10 min)
);
CREATE INDEX idx_auth_requests_code ON oidc_auth_requests(auth_code) WHERE auth_code IS NOT NULL;
CREATE INDEX idx_auth_requests_expires ON oidc_auth_requests(expires_at);

-- OIDC Access Tokens
CREATE TABLE oidc_access_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id       TEXT NOT NULL REFERENCES oidc_clients(client_id),
    user_id         UUID NOT NULL REFERENCES hub_users(id),
    refresh_token_id UUID,
    audience        TEXT[] NOT NULL,
    scopes          TEXT[] NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_access_tokens_user ON oidc_access_tokens(user_id);
CREATE INDEX idx_access_tokens_expires ON oidc_access_tokens(expires_at);

-- OIDC Refresh Tokens
CREATE TABLE oidc_refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash      TEXT NOT NULL UNIQUE,     -- SHA-256 of refresh token
    client_id       TEXT NOT NULL REFERENCES oidc_clients(client_id),
    user_id         UUID NOT NULL REFERENCES hub_users(id),
    access_token_id UUID,
    scopes          TEXT[] NOT NULL,
    amr             TEXT[],
    audience        TEXT[] NOT NULL,
    auth_time       TIMESTAMPTZ NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,             -- Null = active
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_refresh_tokens_user ON oidc_refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_hash ON oidc_refresh_tokens(token_hash);

-- OIDC Signing Keys (RSA key pairs for JWT signing)
CREATE TABLE oidc_signing_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kid             TEXT NOT NULL UNIQUE,     -- Key ID for JWKS (e.g. "brain-key-2026-04")
    algorithm       TEXT NOT NULL DEFAULT 'RS256',
    private_key_pem TEXT NOT NULL,            -- PEM-encoded RSA private key
    public_key_pem  TEXT NOT NULL,            -- PEM-encoded RSA public key
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    rotated_at      TIMESTAMPTZ              -- Set when key is rotated
);

-- Seed FB-OS as the default OIDC client
INSERT INTO oidc_clients (client_id, client_name, redirect_uris, grant_types, response_types, token_endpoint_auth, scopes)
VALUES (
    'fb-os',
    'FutureBuild OS',
    ARRAY['http://localhost:8080/auth/callback'],
    ARRAY['authorization_code', 'refresh_token'],
    ARRAY['code'],
    'none',
    ARRAY['openid', 'profile', 'email', 'org']
);
