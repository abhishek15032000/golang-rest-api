-- migrate:up
CREATE TABLE IF NOT EXISTS otp_verification(
    id serial PRIMARY KEY,
    otp_key varchar(255) NOT NULL,
    user_id int NOT NULL,
    issued_at timestamp DEFAULT current_timestamp,
    expires_at timestamp DEFAULT current_timestamp + interval '5 minutes',
    is_used boolean DEFAULT false,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT unique_user_otp UNIQUE (user_id, otp_key),
    created_at timestamp DEFAULT current_timestamp,
    updated_at timestamp DEFAULT current_timestamp
);

-- migrate:down
DROP TABLE IF EXISTS otp_verification;