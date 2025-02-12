-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS promo_codes (
    code VARCHAR(50) PRIMARY KEY,
    type VARCHAR(20) NOT NULL CHECK (type IN ('referral', 'single', 'free_use')),
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    bonus_redeemer INT NOT NULL DEFAULT 0,
    bonus_referrer INT NOT NULL DEFAULT 0,
    usage_limit INT NOT NULL DEFAULT 1,
    usage_count INT NOT NULL DEFAULT 0,
    valid_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_to TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_transactions
(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bot_id BIGINT REFERENCES bots(id) ON DELETE SET NULL,
    service_id INT REFERENCES services(id) ON DELETE SET NULL,  -- Если есть услуги
    type VARCHAR(20) NOT NULL CHECK (type IN ('deposit', 'withdrawal', 'refund', 'promo_redeem', 'referral_bonus')),
    amount INT NOT NULL CHECK (amount != 0),
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'completed', 'failed', 'expired')),
    description VARCHAR(255),
    provider VARCHAR(255),
    external_id VARCHAR(255),
    promo_code_id VARCHAR(50) REFERENCES promo_codes(code) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS promo_code_usages (
    id SERIAL PRIMARY KEY,
    promo_code_id VARCHAR(50) REFERENCES promo_codes(code) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    used_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_promo_code_usages_unique ON promo_code_usages (promo_code_id, user_id);

CREATE TABLE IF NOT EXISTS user_accounts (
    user_id INT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    balance INT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION update_updated_at_user_accounts()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_user_accounts_updated_at
BEFORE UPDATE ON user_accounts
FOR EACH ROW
EXECUTE PROCEDURE update_updated_at_user_accounts();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trigger_user_accounts_updated_at ON user_accounts;
DROP FUNCTION IF EXISTS update_updated_at_user_accounts();
DROP TABLE IF EXISTS user_accounts;
DROP TABLE IF EXISTS promo_code_usages;
DROP TABLE IF EXISTS user_transactions;
DROP TABLE IF EXISTS promo_codes;
-- +goose StatementEnd