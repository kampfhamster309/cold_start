-- AUTH-1: foundational tables. Deliberately minimal — OIDC-subject
-- mapping, magic-link tokens, and RBAC grants belong to AUTH-2/3/4, not
-- here.
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Sessions are looked up by their token hash on every request, so that's
-- the primary key directly rather than a separate surrogate id. The raw
-- token only ever exists in the cookie and in transient memory during
-- creation/validation — only its SHA-256 hash is stored, so a database
-- leak alone can't be used to hijack a session (same reasoning as
-- password hashing, applied to session tokens).
CREATE TABLE sessions (
    token_hash TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX sessions_user_id_idx ON sessions (user_id);
