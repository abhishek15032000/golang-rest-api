-- migrate:up
CREATE TABLE IF NOT EXISTS refresh_tokens(
    id serial PRIMARY KEY,
    user_id int NOT NULL,
    token_hash text NOT NULL UNIQUE,
    expiry timestamp DEFAULT current_timestamp + interval '14 days',
    is_revoked boolean DEFAULT false,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamp DEFAULT current_timestamp,
    updated_at timestamp DEFAULT current_timestamp
);
-- migrate:down
DROP TABLE IF EXISTS refresh_tokens;
