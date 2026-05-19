-- +migrate Up
CREATE TABLE `app_registry` (
    `id`          BIGINT       NOT NULL AUTO_INCREMENT,
    `app_id`      VARCHAR(100) NOT NULL COMMENT '应用唯一标识',
    `app_name`    VARCHAR(200) NOT NULL COMMENT '应用显示名',
    `api_key`     VARCHAR(255) NOT NULL COMMENT 'API Key（哈希存储）',
    `status`      TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1=启用，0=禁用',
    `created_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_app_id` (`app_id`),
    UNIQUE KEY `uk_api_key` (`api_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- +migrate Down
DROP TABLE IF EXISTS `app_registry`;
