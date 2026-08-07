-- Bootstrap extension: gen_random_uuid() is expected to back most future
-- primary keys (users, grants, plan templates, etc. — architecture §2.3).
-- Enabling it here, in its own migration, means every later migration can
-- rely on it existing rather than each re-checking.
CREATE EXTENSION IF NOT EXISTS pgcrypto;
