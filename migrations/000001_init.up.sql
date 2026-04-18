CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255)        NOT NULL,
    created_at    TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS wishlists (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    event_date  VARCHAR(20)  NOT NULL DEFAULT '',
    token       VARCHAR(64)  UNIQUE NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wishlists_user_id ON wishlists(user_id);
CREATE INDEX idx_wishlists_token   ON wishlists(token);

CREATE TABLE IF NOT EXISTS wishlist_items (
    id          BIGSERIAL PRIMARY KEY,
    wishlist_id BIGINT       NOT NULL REFERENCES wishlists(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    url         TEXT         NOT NULL DEFAULT '',
    priority    SMALLINT     NOT NULL DEFAULT 0,
    is_reserved BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wishlist_items_wishlist_id ON wishlist_items(wishlist_id);
