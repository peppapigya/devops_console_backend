-- ================================================================
-- MCP 根因分析与修复功能 - 数据库初始化脚本
-- 数据库: devops (与现有后端一致)
-- 执行方式: mysql -u root -p devops < sql/repair.sql
-- ================================================================

-- 根因分析会话主表
CREATE TABLE IF NOT EXISTS `repair_sessions` (
    `id`                VARCHAR(36)     NOT NULL COMMENT 'Session UUID',
    `log_event_id`      VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '关联的告警日志 ID',
    `log_source`        VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '日志来源（服务名/文件路径）',
    `log_message`       TEXT                     COMMENT '告警日志内容',
    `log_level`         VARCHAR(20)     NOT NULL DEFAULT 'ERROR' COMMENT '日志级别',
    `log_host`          VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '产生告警的主机',
    `log_service`       VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '产生告警的服务',
    -- AI 分析结果
    `analysis`          TEXT                     COMMENT 'AI 问题分析',
    `root_cause`        TEXT                     COMMENT 'AI 根因判断',
    `severity`          VARCHAR(20)              COMMENT '严重程度: low|medium|high|critical',
    `confidence`        FLOAT           NOT NULL DEFAULT 0 COMMENT 'AI 置信度 0-1',
    -- 执行状态
    `status`            VARCHAR(20)     NOT NULL DEFAULT 'pending' COMMENT '状态: pending|running|waiting_confirm|success|partial|failed|paused',
    `total_actions`     INT             NOT NULL DEFAULT 0 COMMENT '修复动作总数',
    `completed_actions` INT             NOT NULL DEFAULT 0 COMMENT '已完成动作数',
    -- 时间
    `created_at`        DATETIME(3)              COMMENT '创建时间',
    `updated_at`        DATETIME(3)              COMMENT '更新时间',
    `finished_at`       DATETIME(3)              COMMENT '完成时间',
    PRIMARY KEY (`id`),
    INDEX `idx_log_event_id` (`log_event_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='根因分析与修复会话';


-- 每轮 AI 对话消息表（持久化上下文，支持 Token 压缩摘要替换）
CREATE TABLE IF NOT EXISTS `session_messages` (
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `session_id`    VARCHAR(36)     NOT NULL COMMENT '关联的 session id',
    `role`          VARCHAR(20)     NOT NULL COMMENT '角色: system|user|assistant',
    `content`       MEDIUMTEXT      NOT NULL COMMENT '消息内容',
    `is_summary`    TINYINT(1)      NOT NULL DEFAULT 0 COMMENT '是否为压缩摘要行',
    `token_count`   INT             NOT NULL DEFAULT 0 COMMENT '估算 token 数',
    `created_at`    DATETIME(3)              COMMENT '创建时间',
    PRIMARY KEY (`id`),
    INDEX `idx_session_id` (`session_id`),
    INDEX `idx_session_summary` (`session_id`, `is_summary`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI 对话消息历史';


-- 修复动作执行记录表
CREATE TABLE IF NOT EXISTS `repair_actions` (
    `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `session_id`        VARCHAR(36)     NOT NULL COMMENT '关联的 session id',
    `action_order`      INT             NOT NULL COMMENT '执行顺序',
    `description`       VARCHAR(512)             COMMENT '动作描述',
    `thought`           TEXT                     COMMENT 'AI 推理思路',
    `command`           TEXT                     COMMENT '执行的 shell 命令',
    `cwd`               VARCHAR(255)    NOT NULL DEFAULT '/' COMMENT '执行工作目录',
    `target`            VARCHAR(255)    NOT NULL DEFAULT '' COMMENT 'SSH 目标主机',
    `timeout`           INT             NOT NULL DEFAULT 30 COMMENT '超时秒数',
    `risk_level`        VARCHAR(20)              COMMENT '风险等级: low|medium|high',
    `risk_reason`       TEXT                     COMMENT '风险说明',
    `rollback_command`  TEXT                     COMMENT '回滚命令',
    -- 执行结果
    `status`            VARCHAR(20)     NOT NULL DEFAULT 'pending' COMMENT '状态: pending|waiting_confirm|running|success|failed|skipped',
    `output`            MEDIUMTEXT               COMMENT '执行输出',
    `error_msg`         TEXT                     COMMENT '错误信息',
    `exit_code`         INT             NOT NULL DEFAULT 0 COMMENT 'shell 退出码',
    `duration_ms`       INT             NOT NULL DEFAULT 0 COMMENT '执行耗时（毫秒）',
    `executed_at`       DATETIME(3)              COMMENT '执行时间',
    `created_at`        DATETIME(3)              COMMENT '创建时间',
    PRIMARY KEY (`id`),
    INDEX `idx_session_id` (`session_id`),
    INDEX `idx_session_order` (`session_id`, `action_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='修复动作执行记录';
