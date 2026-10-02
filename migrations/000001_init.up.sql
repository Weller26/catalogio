CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    email           TEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    token_hash BYTEA NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE item_statuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID
        REFERENCES users(id)
        ON DELETE CASCADE,

    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE item_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID
        REFERENCES users(id)
        ON DELETE CASCADE,

    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,

    title TEXT NOT NULL,
    description TEXT,
    
    status_id UUID REFERENCES item_statuses(id) ON DELETE SET NULL,
    type_id UUID REFERENCES item_types(id) ON DELETE SET NULL,

    rating DOUBLE PRECISION
        CHECK (rating >= 0 AND rating <= 10),
    notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_items_user_id_created_at 
    ON items(user_id, created_at DESC);

CREATE INDEX idx_refresh_tokens_user_id 
    ON refresh_tokens(user_id);

CREATE UNIQUE INDEX uq_item_statuses_global_name
    ON item_statuses (LOWER(name))
    WHERE user_id IS NULL;

CREATE UNIQUE INDEX uq_item_statuses_user_name
    ON item_statuses (user_id, LOWER(name))
    WHERE user_id IS NOT NULL;

CREATE UNIQUE INDEX uq_item_types_global_name
    ON item_types (LOWER(name))
    WHERE user_id IS NULL;

CREATE UNIQUE INDEX uq_item_types_user_name
    ON item_types (user_id, LOWER(name))
    WHERE user_id IS NOT NULL;