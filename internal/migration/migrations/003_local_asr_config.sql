-- +migrate Up
CREATE TABLE `local_asr_config` (
    `id`              BIGINT       NOT NULL AUTO_INCREMENT,
    `app_id`          VARCHAR(100) NOT NULL COMMENT '应用标识',
    `subject_id`      VARCHAR(100) NOT NULL COMMENT '用户标识',
    `scope_type`      VARCHAR(50)  NOT NULL COMMENT '作用域类型',
    `scope_id`        VARCHAR(100) NOT NULL COMMENT '作用域 ID',
    `enabled`         TINYINT      NOT NULL DEFAULT 1 COMMENT '是否启用本地 ASR',
    `timeout_ms`      INT          NULL COMMENT '本地模型超时(ms)',
    `probe_url`       VARCHAR(500) NULL COMMENT '健康检查 URL',
    `transcribe_url`  VARCHAR(500) NULL COMMENT '转写 URL',
    `created_at`      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_app_subject_scope` (`app_id`, `subject_id`, `scope_type`, `scope_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- +migrate Down
DROP TABLE IF EXISTS `local_asr_config`;
