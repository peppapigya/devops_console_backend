-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS devops_console CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用创建的数据库
USE devops_console;

-- 设置SQL模式以避免严格模式问题
SET sql_mode = 'STRICT_TRANS_TABLES,NO_ZERO_DATE,NO_ZERO_IN_DATE,ERROR_FOR_DIVISION_BY_ZERO';

-- 禁用外键检查，确保初始化过程顺利
SET FOREIGN_KEY_CHECKS = 0;

-- 实例类型枚举表：用于定义ELK技术栈中不同组件的类型
CREATE TABLE IF NOT EXISTS instance_types
(
    id          INT UNSIGNED AUTO_INCREMENT PRIMARY KEY, -- 自增主键ID（使用UNSIGNED确保与外键兼容）
    type_name   VARCHAR(100) UNIQUE NOT NULL,        -- 实例类型名称（如elasticsearch、kibana等）
    description TEXT,                                -- 类型的详细描述信息
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP -- 更新时间
);

-- 实例主表：存储ELK技术栈中各个实例的基本信息和状态
CREATE TABLE IF NOT EXISTS instances
(
    id               INT UNSIGNED AUTO_INCREMENT PRIMARY KEY, -- 自增主键ID（使用UNSIGNED确保与外键兼容）
    instance_type_id INT UNSIGNED NOT NULL,          -- 实例类型ID，关联instance_types表（使用UNSIGNED确保与主键兼容）
    name             VARCHAR(255) NOT NULL,          -- 实例名称（如集群名称、节点名称等）
    address          VARCHAR(500) NOT NULL,          -- 实例的访问地址（IP:PORT或域名）
    https_enabled    TINYINT(1) DEFAULT 0,           -- 是否启用HTTPS：0-否，1-是
    skip_ssl_verify  TINYINT(1) DEFAULT 0,           -- 是否跳过SSL证书验证：0-否，1-是
    status           ENUM('active', 'inactive', 'error') DEFAULT 'active', -- 实例状态：活跃、非活跃、错误
    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP -- 更新时间
);

-- 统一认证配置表：支持ES实例和K8s集群的认证配置
CREATE TABLE IF NOT EXISTS auth_configs
(
    id           INT UNSIGNED AUTO_INCREMENT PRIMARY KEY, -- 自增主键ID（使用UNSIGNED确保一致性）
    resource_type ENUM('instance', 'cluster') NOT NULL, -- 资源类型：instance-ES实例，cluster-K8s集群
    resource_id  INT UNSIGNED NOT NULL,             -- 资源ID（使用UNSIGNED确保与主键兼容）
    resource_name VARCHAR(255) NOT NULL,             -- 资源名称（实例名称或集群名称）
    auth_type    ENUM('none', 'basic', 'api_key', 'aws_iam', 'token', 'certificate', 'kubeconfig') NOT NULL, -- 认证类型
    config_key   VARCHAR(100) NOT NULL,              -- 配置键名（如username, password, api_key, kubeconfig_path等）
    config_value TEXT,                               -- 配置值（敏感信息建议加密存储）
    is_encrypted TINYINT(1) DEFAULT 1,               -- 是否加密存储：1-加密，0-不加密
    status       ENUM('active', 'inactive') DEFAULT 'active', -- 配置状态
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, -- 更新时间
    UNIQUE KEY unique_resource_config (resource_type, resource_id, config_key) -- 联合唯一约束：同一资源的配置键不能重复
);

-- 连接测试记录表：记录对各实例的连通性测试结果
CREATE TABLE IF NOT EXISTS connection_tests
(
    id            INT UNSIGNED AUTO_INCREMENT PRIMARY KEY, -- 自增主键ID（使用UNSIGNED确保一致性）
    resource_type ENUM('instance', 'cluster') NOT NULL, -- 资源类型
    resource_id   INT UNSIGNED NOT NULL,            -- 被测试的资源ID（使用UNSIGNED确保与外键兼容）
    test_result   ENUM('success', 'failure', 'timeout'), -- 测试结果：成功、失败、超时
    response_time INT,                               -- 响应时间（毫秒）
    error_message TEXT,                              -- 错误信息（测试失败时的详细描述）
    tested_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- 测试时间
);

-- 用户账号表：来自kubernate-server的用户管理系统
CREATE TABLE IF NOT EXISTS account
(
    id         INT UNSIGNED AUTO_INCREMENT PRIMARY KEY, -- 自增主键ID（使用UNSIGNED确保一致性）
    user_id    VARCHAR(100) DEFAULT '',              -- 用户id
    password   VARCHAR(255) DEFAULT '',              -- 密码
    nickname   VARCHAR(100) DEFAULT '',              -- 昵称
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- 创建时间
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP -- 更新时间
);

-- 初始化实例类型数据：插入ELK技术栈中常用的组件类型（忽略重复插入错误）
INSERT IGNORE INTO instance_types (type_name, description)
VALUES ('elasticsearch', 'Elasticsearch搜索和分析引擎'), -- 分布式搜索和分析引擎
       ('filebeat', 'Filebeat日志收集器'),            -- 轻量级日志数据收集器
       ('logstash', 'Logstash数据处理管道'),          -- 数据处理和转换管道
       ('kibana', 'Kibana数据可视化平台'),           -- 数据可视化和分析平台
       ('apm', 'APM应用性能监控'),                   -- 应用性能监控系统
       ('metricbeat', 'Metricbeat指标收集器'),       -- 系统和服务指标收集器
       ('kubernetes', 'Kubernetes集群'),             -- Kubernetes集群管理
       ('docker', 'Docker容器');                    -- Docker容器管理

-- 创建外键约束（如果不存在）
-- 1. instances表的外键约束
SET @constraint_count = (SELECT COUNT(*) FROM information_schema.table_constraints 
     WHERE constraint_schema = DATABASE() 
       AND table_name = 'instances' 
       AND constraint_name = 'fk_instances_instance_type');

SET @sql = IF(@constraint_count = 0,
    'ALTER TABLE instances ADD CONSTRAINT fk_instances_instance_type FOREIGN KEY (instance_type_id) REFERENCES instance_types (id) ON DELETE RESTRICT ON UPDATE CASCADE',
    'SELECT "instances表外键约束已存在" AS message'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 2. connection_tests表的外键约束
SET @constraint_count = (SELECT COUNT(*) FROM information_schema.table_constraints 
     WHERE constraint_schema = DATABASE() 
       AND table_name = 'connection_tests' 
       AND constraint_name = 'fk_connection_tests_auth');

SET @sql = IF(@constraint_count = 0,
    'ALTER TABLE connection_tests ADD CONSTRAINT fk_connection_tests_auth FOREIGN KEY (resource_type, resource_id) REFERENCES auth_configs (resource_type, resource_id) ON DELETE CASCADE',
    'SELECT "connection_tests表外键约束已存在" AS message'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 重新启用外键检查
SET FOREIGN_KEY_CHECKS = 1;

-- 创建索引：优化查询性能（使用动态SQL避免重复创建错误）
-- 实例表索引
SET @index_count = (SELECT COUNT(*) FROM information_schema.statistics 
     WHERE table_schema = DATABASE() AND table_name = 'instances' AND index_name = 'idx_instances_type');

SET @sql = IF(@index_count = 0,
    'CREATE INDEX idx_instances_type ON instances (instance_type_id)',
    'SELECT "idx_instances_type索引已存在" AS message'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET @index_count = (SELECT COUNT(*) FROM information_schema.statistics 
     WHERE table_schema = DATABASE() AND table_name = 'instances' AND index_name = 'idx_instances_status');

SET @sql = IF(@index_count = 0,
    'CREATE INDEX idx_instances_status ON instances (status)',
    'SELECT "idx_instances_status索引已存在" AS message'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_count = (SELECT COUNT(*) FROM information_schema.statistics 
     WHERE table_schema = DATABASE() AND table_name = 'instances' AND index_name = 'idx_instances_https');

SET @sql = IF(@index_count = 0,
    'CREATE INDEX idx_instances_https ON instances (https_enabled)',
    'SELECT "idx_instances_https索引已存在" AS message'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 认证配置表索引
SET @index_count = (SELECT COUNT(*) FROM information_schema.statistics 
     WHERE table_schema = DATABASE() AND table_name = 'auth_configs' AND index_name = 'idx_auth_configs_resource');

SET @sql = IF(@index_count = 0,
    'CREATE INDEX idx_auth_configs_resource ON auth_configs (resource_type, resource_id)',
    'SELECT "idx_auth_configs_resource索引已存在" AS message'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_count = (SELECT COUNT(*) FROM information_schema.statistics 
     WHERE table_schema = DATABASE() AND table_name = 'auth_configs' AND index_name = 'idx_auth_configs_type');

SET @sql = IF(@index_count = 0,
    'CREATE INDEX idx_auth_configs_type ON auth_configs (auth_type)',
    'SELECT "idx_auth_configs_type索引已存在" AS message'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_count = (SELECT COUNT(*) FROM information_schema.statistics 
     WHERE table_schema = DATABASE() AND table_name = 'auth_configs' AND index_name = 'idx_auth_configs_status');

SET @sql = IF(@index_count = 0,
    'CREATE INDEX idx_auth_configs_status ON auth_configs (status)',
    'SELECT "idx_auth_configs_status索引已存在" AS message'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 连接测试表索引
SET @index_count = (SELECT COUNT(*) FROM information_schema.statistics 
     WHERE table_schema = DATABASE() AND table_name = 'connection_tests' AND index_name = 'idx_connection_tests_resource');

SET @sql = IF(@index_count = 0,
    'CREATE INDEX idx_connection_tests_resource ON connection_tests (resource_type, resource_id)',
    'SELECT "idx_connection_tests_resource索引已存在" AS message'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @index_count = (SELECT COUNT(*) FROM information_schema.statistics 
     WHERE table_schema = DATABASE() AND table_name = 'connection_tests' AND index_name = 'idx_connection_tests_time');

SET @sql = IF(@index_count = 0,
    'CREATE INDEX idx_connection_tests_time ON connection_tests (tested_at)',
    'SELECT "idx_connection_tests_time索引已存在" AS message'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 用户账号表索引
SET @index_count = (SELECT COUNT(*) FROM information_schema.statistics 
     WHERE table_schema = DATABASE() AND table_name = 'account' AND index_name = 'idx_account_user_id');

SET @sql = IF(@index_count = 0,
    'CREATE INDEX idx_account_user_id ON account (user_id)',
    'SELECT "idx_account_user_id索引已存在" AS message'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 通用资源详情视图：整合所有类型资源的详细信息
DROP VIEW IF EXISTS resource_details;

CREATE VIEW resource_details AS
SELECT 
    ac.resource_type,
    ac.resource_id,
    ac.resource_name,
    ac.status,
    ac.created_at,
    ac.updated_at,
    -- 认证配置信息（JSON格式，便于程序解析）
    JSON_OBJECTAGG(
        ac.config_key, 
        JSON_OBJECT(
            'config_value', ac.config_value,
            'auth_type', ac.auth_type,
            'is_encrypted', ac.is_encrypted
        )
    ) as auth_configs,
    -- 主要认证类型描述
    CASE 
        WHEN MAX(CASE WHEN ac.auth_type = 'kubeconfig' THEN 1 ELSE 0 END) = 1 THEN '配置文件认证'
        WHEN MAX(CASE WHEN ac.auth_type = 'token' THEN 1 ELSE 0 END) = 1 THEN '令牌认证'
        WHEN MAX(CASE WHEN ac.auth_type = 'basic' THEN 1 ELSE 0 END) = 1 THEN '基础认证'
        WHEN MAX(CASE WHEN ac.auth_type = 'api_key' THEN 1 ELSE 0 END) = 1 THEN 'API密钥认证'
        WHEN MAX(CASE WHEN ac.auth_type = 'certificate' THEN 1 ELSE 0 END) = 1 THEN '证书认证'
        WHEN MAX(CASE WHEN ac.auth_type = 'aws_iam' THEN 1 ELSE 0 END) = 1 THEN 'AWS IAM认证'
        ELSE '无认证'
    END as auth_type_desc,
    
    -- 通用连接信息（适用于多种资源类型）
    MAX(i.address) as connection_endpoint,
    MAX(i.https_enabled) as secure_connection,
    MAX(i.skip_ssl_verify) as ssl_verification_disabled,
    
    -- 资源类型信息（仅对instance类型有效）
    MAX(it.type_name) as resource_subtype,
    MAX(it.description) as subtype_description,
    
    -- 最新连接测试结果
    (SELECT ct.test_result FROM connection_tests ct 
     WHERE ct.resource_type = ac.resource_type AND ct.resource_id = ac.resource_id 
     ORDER BY ct.tested_at DESC LIMIT 1) as last_test_result,
    (SELECT ct.response_time FROM connection_tests ct 
     WHERE ct.resource_type = ac.resource_type AND ct.resource_id = ac.resource_id 
     ORDER BY ct.tested_at DESC LIMIT 1) as last_response_time,
    (SELECT ct.tested_at FROM connection_tests ct 
     WHERE ct.resource_type = ac.resource_type AND ct.resource_id = ac.resource_id 
     ORDER BY ct.tested_at DESC LIMIT 1) as last_test_time

FROM auth_configs ac
         LEFT JOIN instances i ON ac.resource_id = i.id AND ac.resource_type = 'instance' AND ac.status = 'active'
         LEFT JOIN instance_types it ON i.instance_type_id = it.id
WHERE ac.status = 'active' -- 只显示活跃的资源配置
GROUP BY ac.resource_type, ac.resource_id, ac.resource_name, ac.status, ac.created_at, ac.updated_at
ORDER BY ac.resource_type, ac.resource_name;

-- 验证外键约束是否创建成功
SELECT 
    '外键约束验证' AS 检查项,
    TABLE_NAME AS 表名,
    COLUMN_NAME AS 列名,
    CONSTRAINT_NAME AS 约束名称,
    REFERENCED_TABLE_NAME AS 引用表,
    REFERENCED_COLUMN_NAME AS 引用列
FROM 
    information_schema.KEY_COLUMN_USAGE 
WHERE 
    constraint_schema = DATABASE() 
    AND referenced_table_name IS NOT NULL
ORDER BY 
    TABLE_NAME, 
    CONSTRAINT_NAME;