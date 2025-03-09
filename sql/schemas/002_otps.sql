-- +goose Up
CREATE TYPE otp_type AS ENUM ('password_reset', 'email_verification', 'two_factor_auth', 'other_type');

CREATE TABLE otps(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(255) NOT NULL,
    type otp_type NOT NULL, 
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL, 
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 5
);

CREATE INDEX idx_otps_code ON otps(code);
CREATE INDEX idx_otps_expires_at ON otps(expires_at);
CREATE INDEX idx_otps_user_id_type ON otps(user_id, type);

-- +goose Down
DROP TABLE otps;
DROP TYPE otp_type;
DROP INDEX IF EXISTS idx_otps_code;
DROP INDEX IF EXISTS idx_otps_expires_at;
DROP INDEX IF EXISTS idx_otps_user_id_type;