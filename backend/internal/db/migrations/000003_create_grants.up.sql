-- AUTH-2: (user, role, resource_type, resource_id) grants — the RBAC
-- model described in architecture §5 / tech-stack §6. resource_type is a
-- real enum specifically so a later resource type is added by extending
-- it and populating rows, not by migrating this table's shape.
--
-- Phase 0 only ever grants against the 'global' sentinel; the other
-- resource_type values exist now so Phase 2 (doc_space/repo_connection/
-- onboarding_hire scoped grants) is a data change against this same
-- table, not a schema change.
CREATE TYPE resource_type AS ENUM ('global', 'doc_space', 'repo_connection', 'onboarding_hire');
CREATE TYPE grant_role AS ENUM ('viewer', 'editor', 'admin');

CREATE TABLE grants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role grant_role NOT NULL,
    resource_type resource_type NOT NULL,
    -- NULL only for the 'global' sentinel, which has no specific
    -- resource to point at; every other resource_type must carry one.
    -- Not a foreign key: the resource lives in a different table per
    -- resource_type, so this is a polymorphic reference the database
    -- can't enforce referential integrity on directly.
    resource_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT grants_resource_id_matches_type CHECK (
        (resource_type = 'global' AND resource_id IS NULL)
        OR (resource_type <> 'global' AND resource_id IS NOT NULL)
    )
);

CREATE INDEX grants_user_id_idx ON grants (user_id);

-- Two partial unique indexes rather than one unique constraint over all
-- four columns: a plain UNIQUE(user_id, role, resource_type, resource_id)
-- wouldn't dedupe global grants at all, since Postgres treats every NULL
-- resource_id as distinct from every other NULL.
CREATE UNIQUE INDEX grants_global_unique_idx ON grants (user_id, role, resource_type) WHERE resource_type = 'global';
CREATE UNIQUE INDEX grants_scoped_unique_idx ON grants (user_id, role, resource_type, resource_id) WHERE resource_type <> 'global';
