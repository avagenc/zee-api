CREATE TABLE tuya_app_accounts (
    owner_id   UUID         PRIMARY KEY,
    tuya_uid   VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_tuya_app_accounts_tuya_uid_active
    ON tuya_app_accounts (tuya_uid)
    WHERE deleted_at IS NULL;
