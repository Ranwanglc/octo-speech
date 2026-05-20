-- +migrate Up
CREATE TABLE `audit_log` (
    `id`         BIGINT       NOT NULL AUTO_INCREMENT,
    `action`     VARCHAR(50)  NOT NULL COMMENT '操作类型: create/delete/disable/enable/reset_key',
    `app_id`     VARCHAR(100) NOT NULL COMMENT '被操作的应用 ID',
    `app_name`   VARCHAR(200) NOT NULL DEFAULT '' COMMENT '应用名称快照',
    `operator`   VARCHAR(100) NOT NULL COMMENT '操作者（admin 用户名）',
    `detail`     JSON         NULL COMMENT '额外信息',
    `created_at` TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_app_id` (`app_id`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- +migrate Down
DROP TABLE IF EXISTS `audit_log`;
