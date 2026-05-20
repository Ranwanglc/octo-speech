-- +migrate Up
ALTER TABLE `app_registry`
    MODIFY COLUMN `api_key` VARCHAR(255) NOT NULL COMMENT 'API Key (SHA-256 hash)';

-- +migrate Down
ALTER TABLE `app_registry`
    MODIFY COLUMN `api_key` VARCHAR(255) NOT NULL COMMENT 'API Key（哈希存储）';
