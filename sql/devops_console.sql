/*
 Navicat Premium Dump SQL

 Source Server         : 47.104.247.159
 Source Server Type    : MySQL
 Source Server Version : 80044 (8.0.44)
 Source Host           : 47.104.247.159:8002
 Source Schema         : devops_console

 Target Server Type    : MySQL
 Target Server Version : 80044 (8.0.44)
 File Encoding         : 65001

 Date: 22/01/2026 17:02:20
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for account
-- ----------------------------
DROP TABLE IF EXISTS `account`;
CREATE TABLE `account`  (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `username` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '用户名',
  `password` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '密码',
  `status` tinyint UNSIGNED NOT NULL COMMENT '状态，0可用，1不可用',
  `nickname` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT '' COMMENT '昵称',
  `created_at` datetime(3) NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) NULL DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_account_user_id`(`username` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of account
-- ----------------------------

-- ----------------------------
-- Table structure for auth_configs
-- ----------------------------
DROP TABLE IF EXISTS `auth_configs`;
CREATE TABLE `auth_configs`  (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `resource_type` enum('instance','cluster') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `resource_id` int UNSIGNED NOT NULL,
  `resource_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `auth_type` enum('none','basic','api_key','aws_iam','token','certificate','kubeconfig') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `config_key` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `config_value` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL,
  `is_encrypted` tinyint(1) NULL DEFAULT 1,
  `status` enum('active','inactive') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'active',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `unique_resource_config`(`resource_type` ASC, `resource_id` ASC, `config_key` ASC) USING BTREE,
  INDEX `idx_auth_configs_resource`(`resource_type` ASC, `resource_id` ASC) USING BTREE,
  INDEX `idx_auth_configs_type`(`auth_type` ASC) USING BTREE,
  INDEX `idx_auth_configs_status`(`status` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of auth_configs
-- ----------------------------

-- ----------------------------
-- Table structure for connection_tests
-- ----------------------------
DROP TABLE IF EXISTS `connection_tests`;
CREATE TABLE `connection_tests`  (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `resource_type` enum('instance','cluster') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `resource_id` int UNSIGNED NOT NULL,
  `test_result` enum('success','failure','timeout') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `response_time` int NULL DEFAULT NULL,
  `error_message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL,
  `tested_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_connection_tests_resource`(`resource_type` ASC, `resource_id` ASC) USING BTREE,
  INDEX `idx_connection_tests_time`(`tested_at` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of connection_tests
-- ----------------------------

-- ----------------------------
-- Table structure for instance_types
-- ----------------------------
DROP TABLE IF EXISTS `instance_types`;
CREATE TABLE `instance_types`  (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `type_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `description` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 25 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of instance_types
-- ----------------------------
INSERT INTO `instance_types` VALUES (1, 'elasticsearch', 'Elasticsearch搜索和分析引擎', '2026-01-22 14:38:19.000', '2026-01-22 14:38:19.000');
INSERT INTO `instance_types` VALUES (2, 'filebeat', 'Filebeat日志收集器', '2026-01-22 14:38:19.000', '2026-01-22 14:38:19.000');
INSERT INTO `instance_types` VALUES (3, 'logstash', 'Logstash数据处理管道', '2026-01-22 14:38:19.000', '2026-01-22 14:38:19.000');
INSERT INTO `instance_types` VALUES (4, 'kibana', 'Kibana数据可视化平台', '2026-01-22 14:38:19.000', '2026-01-22 14:38:19.000');
INSERT INTO `instance_types` VALUES (5, 'apm', 'APM应用性能监控', '2026-01-22 14:38:19.000', '2026-01-22 14:38:19.000');
INSERT INTO `instance_types` VALUES (6, 'metricbeat', 'Metricbeat指标收集器', '2026-01-22 14:38:19.000', '2026-01-22 14:38:19.000');
INSERT INTO `instance_types` VALUES (7, 'kubernetes', 'Kubernetes集群', '2026-01-22 14:38:19.000', '2026-01-22 14:38:19.000');
INSERT INTO `instance_types` VALUES (8, 'docker', 'Docker容器', '2026-01-22 14:38:19.000', '2026-01-22 14:38:19.000');

-- ----------------------------
-- Table structure for instances
-- ----------------------------
DROP TABLE IF EXISTS `instances`;
CREATE TABLE `instances`  (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `instance_type_id` int UNSIGNED NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `address` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `https_enabled` tinyint(1) NULL DEFAULT 0,
  `skip_ssl_verify` tinyint(1) NULL DEFAULT 0,
  `status` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'active',
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_instances_type`(`instance_type_id` ASC) USING BTREE,
  INDEX `idx_instances_status`(`status` ASC) USING BTREE,
  INDEX `idx_instances_https`(`https_enabled` ASC) USING BTREE,
  INDEX `idx_instances_instance_type_id`(`instance_type_id` ASC) USING BTREE,
  INDEX `idx_instances_https_enabled`(`https_enabled` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of instances
-- ----------------------------

-- ----------------------------
-- View structure for resource_details
-- ----------------------------
DROP VIEW IF EXISTS `resource_details`;
CREATE ALGORITHM = UNDEFINED SQL SECURITY DEFINER VIEW `resource_details` AS select `ac`.`resource_type` AS `resource_type`,`ac`.`resource_id` AS `resource_id`,`ac`.`resource_name` AS `resource_name`,`ac`.`status` AS `status`,`ac`.`created_at` AS `created_at`,`ac`.`updated_at` AS `updated_at`,json_objectagg(`ac`.`config_key`,json_object('config_value',`ac`.`config_value`,'auth_type',`ac`.`auth_type`,'is_encrypted',`ac`.`is_encrypted`)) AS `auth_configs`,(case when (max((case when (`ac`.`auth_type` = 'kubeconfig') then 1 else 0 end)) = 1) then '配置文件认证' when (max((case when (`ac`.`auth_type` = 'token') then 1 else 0 end)) = 1) then '令牌认证' when (max((case when (`ac`.`auth_type` = 'basic') then 1 else 0 end)) = 1) then '基础认证' when (max((case when (`ac`.`auth_type` = 'api_key') then 1 else 0 end)) = 1) then 'API密钥认证' when (max((case when (`ac`.`auth_type` = 'certificate') then 1 else 0 end)) = 1) then '证书认证' when (max((case when (`ac`.`auth_type` = 'aws_iam') then 1 else 0 end)) = 1) then 'AWS IAM认证' else '无认证' end) AS `auth_type_desc`,max(`i`.`address`) AS `connection_endpoint`,max(`i`.`https_enabled`) AS `secure_connection`,max(`i`.`skip_ssl_verify`) AS `ssl_verification_disabled`,max(`it`.`type_name`) AS `resource_subtype`,max(`it`.`description`) AS `subtype_description`,(select `ct`.`test_result` from `connection_tests` `ct` where ((`ct`.`resource_type` = `ac`.`resource_type`) and (`ct`.`resource_id` = `ac`.`resource_id`)) order by `ct`.`tested_at` desc limit 1) AS `last_test_result`,(select `ct`.`response_time` from `connection_tests` `ct` where ((`ct`.`resource_type` = `ac`.`resource_type`) and (`ct`.`resource_id` = `ac`.`resource_id`)) order by `ct`.`tested_at` desc limit 1) AS `last_response_time`,(select `ct`.`tested_at` from `connection_tests` `ct` where ((`ct`.`resource_type` = `ac`.`resource_type`) and (`ct`.`resource_id` = `ac`.`resource_id`)) order by `ct`.`tested_at` desc limit 1) AS `last_test_time` from ((`auth_configs` `ac` left join `instances` `i` on(((`ac`.`resource_id` = `i`.`id`) and (`ac`.`resource_type` = 'instance') and (`ac`.`status` = 'active')))) left join `instance_types` `it` on((`i`.`instance_type_id` = `it`.`id`))) where (`ac`.`status` = 'active') group by `ac`.`resource_type`,`ac`.`resource_id`,`ac`.`resource_name`,`ac`.`status`,`ac`.`created_at`,`ac`.`updated_at` order by `ac`.`resource_type`,`ac`.`resource_name`;

SET FOREIGN_KEY_CHECKS = 1;
