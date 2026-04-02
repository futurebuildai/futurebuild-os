-- 001: Identity & Organization tables
-- FB-Brain is the identity provider for the FutureBuild ecosystem.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users (Brain-managed identity)
CREATE TABLE hub_users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT NOT NULL UNIQUE,
    display_name    TEXT NOT NULL,
    magic_link_hash TEXT,                    -- SHA-256 of magic link token
    magic_link_exp  TIMESTAMPTZ,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Organizations
CREATE TABLE organizations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    plan_tier   TEXT NOT NULL DEFAULT 'free'
        CHECK (plan_tier IN ('free', 'pro', 'enterprise')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Org Memberships
CREATE TABLE org_memberships (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES hub_users(id),
    org_id      UUID NOT NULL REFERENCES organizations(id),
    role        TEXT NOT NULL DEFAULT 'member'
        CHECK (role IN ('owner', 'admin', 'member')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, org_id)
);
CREATE INDEX idx_org_memberships_user ON org_memberships(user_id);
CREATE INDEX idx_org_memberships_org ON org_memberships(org_id);

-- Account Links (cross-system identity mapping)
CREATE TABLE account_links (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id),
    user_id             UUID REFERENCES hub_users(id),
    gable_customer_id   TEXT,
    localblue_site_id   TEXT,
    localblue_user_id   TEXT,
    quickbooks_realm_id TEXT,
    link_type           TEXT NOT NULL DEFAULT 'full'
        CHECK (link_type IN ('supplier', 'labor', 'full')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_account_links_org ON account_links(org_id);
