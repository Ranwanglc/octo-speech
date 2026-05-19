-- +migrate Up
CREATE TABLE `vocabulary_profile` (
    `id`          BIGINT       NOT NULL AUTO_INCREMENT,
    `app_id`      VARCHAR(100) NOT NULL COMMENT '应用标识',
    `subject_id`  VARCHAR(100) NOT NULL COMMENT '用户标识',
    `scope_type`  VARCHAR(50)  NOT NULL COMMENT '作用域类型',
    `scope_id`    VARCHAR(100) NOT NULL COMMENT '作用域 ID',
    `content`     TEXT         NOT NULL COMMENT '纠错上下文内容',
    `updated_by`  VARCHAR(100) NOT NULL COMMENT '最后更新者',
    `created_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_app_subject_scope` (`app_id`, `subject_id`, `scope_type`, `scope_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- +migrate Down
DROP TABLE IF EXISTS `vocabulary_profile`;
