-- name: CreateUser :one
INSERT INTO USERS(USERNAME, EMAIL, PASSWORD)
VALUES ($1, $2, $3)
RETURNING ID, USERNAME, EMAIL, EMAIL_VERIFIED, CREATED_AT, UPDATED_AT;

-- name: GetUser :one
SELECT ID, USERNAME, EMAIL, EMAIL_VERIFIED, CREATED_AT, UPDATED_AT FROM USERS WHERE ID = $1;

-- name: ListUsers :many
SELECT ID, USERNAME, EMAIL, EMAIL_VERIFIED, CREATED_AT, UPDATED_AT FROM USERS ORDER BY ID LIMIT $1 OFFSET $2;

-- name: CreateBlog :one
INSERT INTO BLOGS(TITLE, CONTENT,USER_ID)
VALUES ($1, $2, $3)
RETURNING ID, TITLE, CONTENT, USER_ID, CREATED_AT, UPDATED_AT;

-- name: ListBlogs :many
SELECT ID, TITLE, CONTENT, USER_ID, CREATED_AT, UPDATED_AT FROM BLOGS ORDER BY ID;

-- name: GetUserByUsername :one
SELECT ID, USERNAME, EMAIL, PASSWORD, EMAIL_VERIFIED, CREATED_AT, UPDATED_AT FROM USERS WHERE USERNAME = $1;

-- name: GetUserByEmail :one
SELECT ID, USERNAME, EMAIL, PASSWORD, EMAIL_VERIFIED, CREATED_AT, UPDATED_AT FROM USERS WHERE EMAIL = $1;

-- name: CreateUserProfile :one
INSERT INTO USER_PROFILES (USER_ID, PROFILE_IMAGE)
VALUES ($1, $2)
RETURNING ID, USER_ID, PROFILE_IMAGE;

-- name: GetUserProfileByUserId :one
SELECT ID, USER_ID, PROFILE_IMAGE, CREATED_AT, UPDATED_AT
FROM USER_PROFILES WHERE USER_ID = $1;

-- name: CreateOTP :one
INSERT INTO OTP_VERIFICATION (OTP_KEY, USER_ID)
VALUES ($1, $2)
RETURNING ID, OTP_KEY, ISSUED_AT, EXPIRES_AT, USER_ID;

-- name: GetValidOtpForUser :one
SELECT * FROM otp_verification WHERE user_id = $1 AND is_used = false AND expires_at > now() ORDER BY issued_at DESC LIMIT 1 FOR UPDATE;
-- name: MarkOtpUsed :one
UPDATE otp_verification
SET is_used = true,
    updated_at = now()
WHERE id = $1
AND is_used = false
AND expires_at > now()
RETURNING id, otp_key, issued_at, expires_at, user_id, created_at, updated_at;
-- name: MarkUserEmailVerified :one
UPDATE users
SET email_verified = true,
    updated_at = now()
WHERE id = $1
AND email_verified = false
RETURNING id, username, email, email_verified, created_at, updated_at;
-- name: InsertRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash)
VALUES ($1, $2)
RETURNING ID, USER_ID, TOKEN_HASH, EXPIRY, IS_REVOKED, CREATED_AT, UPDATED_AT;
-- name: GetRefreshToken :one
SELECT id, user_id, token_hash, expiry, is_revoked, CREATED_AT, UPDATED_AT
FROM refresh_tokens
WHERE token_hash = $1
AND is_revoked = false
AND expiry > now()
FOR UPDATE;
-- name: MarkRefreshTokenRevoked :one
UPDATE refresh_tokens
SET is_revoked = true,
    updated_at = now()
WHERE token_hash = $1
AND is_revoked = false
AND expiry > now()
RETURNING id, user_id, token_hash, expiry, is_revoked, CREATED_AT, UPDATED_AT;