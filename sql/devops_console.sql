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

 Date: 05/02/2026 09:33:20
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for auth_configs
-- ----------------------------
DROP TABLE IF EXISTS `auth_configs`;
CREATE TABLE `auth_configs`  (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `resource_type` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `resource_id` bigint UNSIGNED NOT NULL,
  `resource_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `auth_type` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `config_key` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `config_value` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL,
  `is_encrypted` tinyint(1) NULL DEFAULT 1,
  `status` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'active',
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `unique_resource_config`(`resource_type` ASC, `resource_id` ASC, `config_key` ASC) USING BTREE,
  INDEX `idx_auth_configs_resource`(`resource_type` ASC, `resource_id` ASC) USING BTREE,
  INDEX `idx_auth_configs_type`(`auth_type` ASC) USING BTREE,
  INDEX `idx_auth_configs_status`(`status` ASC) USING BTREE,
  INDEX `idx_auth_configs_resource_type`(`resource_type` ASC) USING BTREE,
  INDEX `idx_auth_configs_resource_id`(`resource_id` ASC) USING BTREE,
  INDEX `idx_auth_configs_auth_type`(`auth_type` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of auth_configs
-- ----------------------------
INSERT INTO `auth_configs` VALUES (2, 'instance', 2, 'test-kubeconfig', 'kubeconfig', 'kubeconfig', '{\"kubeconfigContent\":\"apiVersion: v1\\nclusters:\\n- cluster:\\n    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMvakNDQWVhZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJMU1USXhOVEV5TkRVd01Wb1hEVE0xTVRJeE16RXlORFV3TVZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBT0ZvCityMTl5SktQRGV3MFplb3Q0OXUraUdvK2htbHI3enM4UjhJeTBFVXdYTU40QlBheFJSSE4waFJiNnJVWHluUmEKQUU2OCtrSjMvRDVvVkYycmFMOVBNZkhIWm4wK29NTCt2dkljUG9pWGh3c0tTT1RsVENjL2Z0QUlFMU9BSGprTgpuVUJUTWc5Q1dOeGJkRWdzV2NldGJWWnA5MG9Sd3A2NWpWVXdKM1VDdHUyTFVFSWY2RmJydVlDT09JN1VRTWU2CjVLdERMdTBiZ3FuYkdRM09POGp0Q2Jod2RwSElGNjhnRVh6RlFVQXRIaWorNWtQUkVpQU9IN1dZVXEvekxYSkMKQUczQWtpRGtQYUpMQmJLSk8xeGlyTW1HMFE3cFNJdVhhcjU1NmxIa3NQdThwMk9kQy9kQ3BoTHVrUENsVGFnUApiT1laNzQ5ekpCbzhTMitxUi9VQ0F3RUFBYU5aTUZjd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZFMXdOWUowTVZIditkdnhqZUpoTTVVT2NlV1FNQlVHQTFVZEVRUU8KTUF5Q0NtdDFZbVZ5Ym1WMFpYTXdEUVlKS29aSWh2Y05BUUVMQlFBRGdnRUJBS1FDOWZ0TEdieTREaGFmbWNvNwpwMURnZmtnSHY4MWVoTzN0NXNhTHZmMEhjeXVHSDVWeG9SRzF5SFVDRzJha1ovaU02S2N2eW5adWg3NkExdjRRClRoZ0h0Z1BoVXcyMElsNGRJN0k1NG14N3QzU3ZYZ1I5YkY3c2d4UEtuQlR2eXZmMkl0MzJnejNFbjFnVDFLRlQKTkIxcTZtSGc3Sk5MUFhEOWNyaUNaTG9mYWRJaTBobjJWV2lnWU90TjVhelpCcXdhTTV3VW1sVVNFMWtoRTRVLwpsWC9YSngvdTdreGYycHVNcnluYi9wQWRuZjh6Rkp5dkZiZ1NLMjZqYUJ0ekxZZ2Z4Z3MxenJOOCswOE9ySU1kClNtTURQUDRPbVVHUEUrRFFJYkxybFZSaHFuWGhJQVlkOTk1NUF5dkhNa3RmdmdrQ293NEU4UTRrczdQMThMa0kKYVR3PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==\\n    server: https://10.0.0.175:6443\\n  name: kubernetes\\n- cluster:\\n    server: \\\"\\\"\\n  name: my-cluster\\ncontexts:\\n- context:\\n    cluster: kubernetes\\n    user: kubernetes-admin\\n  name: kubernetes-admin@kubernetes\\ncurrent-context: kubernetes-admin@kubernetes\\nkind: Config\\npreferences: {}\\nusers:\\n- name: kubernetes-admin\\n  user:\\n    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURJVENDQWdtZ0F3SUJBZ0lJVzF0YjlPUUgwaTR3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB5TlRFeU1UVXhNalExTURGYUZ3MHlOakV5TVRVeE1qUTFNREphTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQTRRaGg0Qmp2NENNdVlJd0EKR0dEQVYwckVBMGgzc2pha0NsZ1R1bUl0Yy9oTGRJN2gvNytWNWVIeVNHaUl4dDJXT1U2OE4rME1EQi9sckZQbgp2T3JpcEhUYSt3dUM2M0t4UkcxVW5iL1Y2bGVNMnZqYW9pQmlRaVl2OXVXaTJoclpCQmY4NEVlQXZBL0pCQjdZCklhU2VoaUV2aG43U1FRUGpUSkNFMTBnN2FObHZyOFhnR3liQXZFelpBbzRpdzYxYlNBcTRBV3RwNFVBZTlCWFEKMzhJYUZldmNGY3RIVURSUUtsUnZQcDdlclJCWUNucXRHU3c4cWl3VUUrWTh6MGxRcysxQ2V4cDZ1TFBqOG4zSApnQzNMZUVidVhFVkxaUnQvU3J2a29uUFo1VzRPSDZhUGJvM1F3Wk1odmUwVUMxRzEyRFJ1L1FOM21iN0JhaFo0ClF0NVBvd0lEQVFBQm8xWXdWREFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RBWURWUjBUQVFIL0JBSXdBREFmQmdOVkhTTUVHREFXZ0JSTmNEV0NkREZSNy9uYjhZM2lZVE9WRG5IbAprREFOQmdrcWhraUc5dzBCQVFzRkFBT0NBUUVBVmZsTWFNTEVrdVRlMzRJaWVuRFJVSXE5aGNBUEwxQTdCOHdYCkNXMWxsUnBUcFdKamE5djBLc1VJS1dZNzJJUEExdkNqdGh0OWNxT0IyRndDS2oxemQ4QTRBTlJRUEtCYUlLbkMKU0hsRTVNTWF6US93NXBvdjl0eTllcGFsK2RQdWFhRkRYK2dQYll2dHlucWRHckpBcHR2OVEvK2dtQ3RnTU8yVwpGYnZaOWE3Y3hEWWxFUTRCaW5WTjdnUjZsRTlFbnA3V3NzSjNXUStPeFN6c1U5T2xTVE9YZkwxMEZOL1JZbENQCmdFekV4aEdnTXUyNHFmbE1pVy8wM3BVRlNQeGJpK1U5bEtpbUgvbzM3cWhPeG5wNU5CRVR3Zzd5WUhlNm9HL2IKODhzSzR0aWpJUDE4VnBuWDdCSkhNR0VsSm1PVkgyVUpVV3FmeXZHdElVZUd0QzRFVEE9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==\\n    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBNFFoaDRCanY0Q011WUl3QUdHREFWMHJFQTBoM3NqYWtDbGdUdW1JdGMvaExkSTdoCi83K1Y1ZUh5U0dpSXh0MldPVTY4TiswTURCL2xyRlBudk9yaXBIVGErd3VDNjNLeFJHMVVuYi9WNmxlTTJ2amEKb2lCaVFpWXY5dVdpMmhyWkJCZjg0RWVBdkEvSkJCN1lJYVNlaGlFdmhuN1NRUVBqVEpDRTEwZzdhTmx2cjhYZwpHeWJBdkV6WkFvNGl3NjFiU0FxNEFXdHA0VUFlOUJYUTM4SWFGZXZjRmN0SFVEUlFLbFJ2UHA3ZXJSQllDbnF0CkdTdzhxaXdVRStZOHowbFFzKzFDZXhwNnVMUGo4bjNIZ0MzTGVFYnVYRVZMWlJ0L1NydmtvblBaNVc0T0g2YVAKYm8zUXdaTWh2ZTBVQzFHMTJEUnUvUU4zbWI3QmFoWjRRdDVQb3dJREFRQUJBb0lCQUV5Q1Z4U2tKZHBrMjczRAptN3l1R0hjVlduTnJUaGJ2Y1BKN1k2bTQrNDgwV2lNMCtTM0U2NmdQSEJyMlA1cXRlQWZmOXlwa2svWURXa2t1CjlkbExXdWRqTzVpakgwNEIzcmRQSExmTm8yTmJoTzVtTVo5eHR6YWFXVEJ1ZnVIOHd0QWJmOFNaU3ZHbHhFaDgKWEN1RUZzbXZ1c0xWbDVLM1NhNmNiQzN0eHhVc1JQMEo2dmhWMkJ2NGorQ2VMYXMzOTlTWDE5emNWQzFuQm1vNQprMkdZVEtLNjlrTDhJYzJ2ZzBGRzU0N3FQSGF2TEdIRDRyU3dOYXpUS1ZIUkN2QjBoRjhkN05QODNWbDkycHd6CnBvL1dIZTBkTVpKdlFjNzcyTitlcm5KQk96Z2JuVGgzSjZhSExibVpnRHR6eGVKUzd2ejVUajlsR1Y2WWdaVTQKbXFES1FhRUNnWUVBOTQ5R0ZuMmNjcjNwcmxYK0FkN2FrcVVmeDFHL3JnV0JnRnMvZldTWkZQT2c1dlprRDFlSApIa0tJK2UwWGg1Sk1sckMzU3Mxd2lGVVpvSUxWR3BMRC90RW5LWHZ5bFRXZm5jcVEwSGZidFBONThNM2tqSDdMCmRneDFkRlY4NksvL0NiU2M5eTlWbVlaS2V6Y0cya0YzVCtOMkFVMGxSNlJ2Z000czRiZDFtK3NDZ1lFQTZMUjkKdysvV3VacEpZd3lLMFRaZ2pBODN1bUZ6Yk9LcmtweHNBck1HSFJDOStaUy9qNHNaOW45dERlVk53NWpUd2tKTQpwbmNzcVRqU2dHQ1VLMjcyMHBkRG11cVlDYWNTZXdjem94bGMyREFST2xDVWZXTFYzTnNEdFpHeU1wRDRIT09UCnBYYUtvcmJ3TlpubTkvcFQ1RkJzWjhBazNBVzlpNjZSajNPR1JTa0NnWUVBb2RUZFFuS1d4VU4wOFd4eGdoT0cKMnZwcXpjZVpBRS9GR24yTUFaS3pwOGlqMUpnWlRSWXcxQTAyc2ZyVnVPQmdoTm04Mkg5NEl3ZE9tMmtybWhWNwpYcWFuYlMwRHBacktYMEkrYktrTnpUcWs2bEFPS2ZIeFc1aEZaK2xDb0hIOHpRRnU0di9rZTFvWWNuZkVXUVVXCjAvaWorYkhPdndpMWc0UkVQc0hKZGtVQ2dZRUE0YXl3endWWW8xVGFXT0YvK3BjV21KM2xlSzRyWjJ5SDRiNHIKRFk4YW5iTnYyWXlGSGl0VGVYZG9obkpic1JZVVB5OVc4SlZnelpmYXBUK0VVbjdoaGFmR200Vm8vdXQxQTdVZgpRY3hGK3k3YWRraFJTU3hCcFZjTlNOZk1EamdETnRrSmhnenBOQlhmN011ZGI5M24zK0tTenlkTFY4bUZZZUpoCkxkSm1ZOGtDZ1lFQW1pQ0c1clVlWVNZcVRNcm5HVkNzRHZVeUhVaUFiSm5PSnJZYWFTSzE1QU5UbGlpcWFteXkKSHV3M29VTXUvWEFBZmIrQ3FCanljbTRCTm1HRk96K0tXRTB4c0ZNakVFUXpPdEN6V1BMem1QRzVYMTU1bm9ENQptNjZWRzI0aVhBTzkraG80d3phejhLbFJwVUIyRlFzZ2F2R2VBTXhKRCs4OUFWTmtMVkpzU1lzPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=\\n\",\"fileName\":\"config\"}', 1, 'active', '2026-01-25 18:16:36.201', '2026-01-25 18:16:36.201');

-- ----------------------------
-- Table structure for connection_tests
-- ----------------------------
DROP TABLE IF EXISTS `connection_tests`;
CREATE TABLE `connection_tests`  (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `resource_type` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `resource_id` bigint UNSIGNED NOT NULL,
  `test_result` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL,
  `response_time` bigint NULL DEFAULT NULL,
  `error_message` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL,
  `tested_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_connection_tests_resource`(`resource_type` ASC, `resource_id` ASC) USING BTREE,
  INDEX `idx_connection_tests_time`(`tested_at` ASC) USING BTREE,
  INDEX `idx_connection_tests_resource_type`(`resource_type` ASC) USING BTREE,
  INDEX `idx_connection_tests_resource_id`(`resource_id` ASC) USING BTREE,
  INDEX `idx_connection_tests_tested_at`(`tested_at` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of connection_tests
-- ----------------------------
INSERT INTO `connection_tests` VALUES (1, 'instance', 2, 'success', 16, 'Kubernetes 配置验证通过', '2026-01-25 18:16:53.422');

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
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uni_instance_types_type_name`(`type_name` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 25 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = DYNAMIC;

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
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of instances
-- ----------------------------
INSERT INTO `instances` VALUES (2, 7, 'test-kubeconfig', '10.0.0.175:6443', 1, 0, 'active', '2026-01-25 18:16:35.976', '2026-01-25 18:16:35.976');

-- ----------------------------
-- Table structure for pipeline_runs
-- ----------------------------
DROP TABLE IF EXISTS `pipeline_runs`;
CREATE TABLE `pipeline_runs`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `pipeline_id` int UNSIGNED NOT NULL COMMENT '关联流水线ID',
  `workflow_name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'K8s中生成的Workflow名称',
  `status` enum('Pending','Running','Succeeded','Failed','Error','Omitted') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'Pending' COMMENT '运行状态',
  `operator` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '执行人',
  `branch` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '本次构建实际使用的分支',
  `commit_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '代码CommitID',
  `start_time` timestamp NULL DEFAULT NULL COMMENT '开始时间',
  `end_time` timestamp NULL DEFAULT NULL COMMENT '结束时间',
  `duration` int UNSIGNED NULL DEFAULT 0 COMMENT '执行耗时，单位秒',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_workflow_name`(`workflow_name` ASC) USING BTREE,
  INDEX `idx_pipeline_id`(`pipeline_id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '流水线执行记录表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of pipeline_runs
-- ----------------------------

-- ----------------------------
-- Table structure for pipeline_steps
-- ----------------------------
DROP TABLE IF EXISTS `pipeline_steps`;
CREATE TABLE `pipeline_steps`  (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `pipeline_id` int UNSIGNED NOT NULL COMMENT '所属流水线ID',
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '步骤名称',
  `image` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '使用的镜像',
  `commands` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '运行命令',
  `depends_on` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '依赖的步骤名称',
  `sort` int NULL DEFAULT 0 COMMENT '排序',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` datetime NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_pipeline_id`(`pipeline_id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '流水线步骤详情表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of pipeline_steps
-- ----------------------------
INSERT INTO `pipeline_steps` VALUES (1, 1, '构建', 'go:1.25', 'echo \"test\"', NULL, 1, '2026-01-31 17:54:51', NULL);
INSERT INTO `pipeline_steps` VALUES (2, 1, '拉取代码', 'ubuntu:latest', 'echo \"hello world\"', NULL, 0, '2026-01-31 18:38:56', NULL);

-- ----------------------------
-- Table structure for pipelines
-- ----------------------------
DROP TABLE IF EXISTS `pipelines`;
CREATE TABLE `pipelines`  (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `project_id` int UNSIGNED NOT NULL COMMENT '关联项目ID',
  `k8s_instance_id` int UNSIGNED NOT NULL COMMENT '目标K8s集群实例ID',
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '流水线名称',
  `git_url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '代码仓库地址',
  `branch` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'main' COMMENT '默认构建分支',
  `argo_template` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '关联的 Argo WorkflowTemplate 名称',
  `params_config` json NULL COMMENT '自定义参数配置(环境变量、构建参数等)',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_project_id`(`project_id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '流水线定义表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of pipelines
-- ----------------------------
INSERT INTO `pipelines` VALUES (1, 0, 0, 'test02', 'https://github.com/peppapigya/devops.git', 'master', 'test', NULL, '2026-01-31 16:37:51', '2026-01-31 18:39:06', NULL);
INSERT INTO `pipelines` VALUES (2, 0, 0, 'test02', 'res.data.data', 'main', '1', NULL, '2026-01-31 17:23:01', '2026-01-31 17:31:22', '2026-01-31 17:31:22');

-- ----------------------------
-- Table structure for projects
-- ----------------------------
DROP TABLE IF EXISTS `projects`;
CREATE TABLE `projects`  (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '项目名称',
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT '' COMMENT '项目描述',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_name`(`name` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '项目表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of projects
-- ----------------------------

-- ----------------------------
-- Table structure for system_user_token
-- ----------------------------
DROP TABLE IF EXISTS `system_user_token`;
CREATE TABLE `system_user_token`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `user_id` bigint NOT NULL COMMENT '用户id',
  `refresh_token` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '刷新token',
  `expires_at` datetime NOT NULL COMMENT '超时时间',
  `last_login_ip` varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '最后登录的ip地址',
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_at` datetime NULL DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime NULL DEFAULT NULL COMMENT '删除时间',
  `access_token` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '登录token',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 77 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of system_user_token
-- ----------------------------
INSERT INTO `system_user_token` VALUES (9, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODcyODUxLCJuYmYiOjE3NjkyNjgwNTEsImlhdCI6MTc2OTI2ODA1MX0.CAXpMbKG2-IBaslP5iKfq0bvtxjaQGQ9Is-OJz0veqk', '2026-01-24 23:20:51', NULL, '2026-01-24 23:20:49', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MjcxNjUxLCJuYmYiOjE3NjkyNjgwNTEsImlhdCI6MTc2OTI2ODA1MX0.8DveZaG7RqdfXK1X3SGmX6-JzYTC6SnobdrDsNyfgrA');
INSERT INTO `system_user_token` VALUES (11, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODczODExLCJuYmYiOjE3NjkyNjkwMTEsImlhdCI6MTc2OTI2OTAxMX0.8iu_inA3-q3hI6SDWFWcYxTI3AApRPx6IaV84T-Suvo', '2026-01-24 23:36:52', NULL, '2026-01-24 23:36:49', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MjcyNjExLCJuYmYiOjE3NjkyNjkwMTEsImlhdCI6MTc2OTI2OTAxMX0.yLJQOabzdcL96qNcgrC8vv1IBqfC-OsQpyaB-y2ic3c');
INSERT INTO `system_user_token` VALUES (12, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODczODI2LCJuYmYiOjE3NjkyNjkwMjYsImlhdCI6MTc2OTI2OTAyNn0.82eWueAJwVUWNqcQsg6qSRCWUjBmPF3fWMj1iLIkYyI', '2026-01-24 23:37:06', NULL, '2026-01-24 23:37:04', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MjcyNjI2LCJuYmYiOjE3NjkyNjkwMjYsImlhdCI6MTc2OTI2OTAyNn0.cnczNyD4wXx2XuHqsKN6Sla62-6KgnrNimxGv-VOwZ8');
INSERT INTO `system_user_token` VALUES (13, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODczOTUwLCJuYmYiOjE3NjkyNjkxNTAsImlhdCI6MTc2OTI2OTE1MH0.aZjTzw4bieXT3TvR3cZIHqdCFuzXf72YdOmD38AN7nw', '2026-01-24 23:39:10', NULL, '2026-01-24 23:39:08', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MjcyNzUwLCJuYmYiOjE3NjkyNjkxNTAsImlhdCI6MTc2OTI2OTE1MH0._mmECtMw2yZTrekKvmjLWmBC9uD_0uqPiFYUTQ12IdE');
INSERT INTO `system_user_token` VALUES (14, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc0MDM0LCJuYmYiOjE3NjkyNjkyMzQsImlhdCI6MTc2OTI2OTIzNH0.y_rlCREHbV6GkUYvrK7BX6GpgFSzK8I2vC_MR_iHi3A', '2026-01-24 23:40:35', NULL, '2026-01-24 23:40:32', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MjcyODM0LCJuYmYiOjE3NjkyNjkyMzQsImlhdCI6MTc2OTI2OTIzNH0.PSMdO3Viwho9yaEz9vwP9cOjSP4r--tZdIP36UVrolQ');
INSERT INTO `system_user_token` VALUES (15, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc0MDQ2LCJuYmYiOjE3NjkyNjkyNDYsImlhdCI6MTc2OTI2OTI0Nn0.fyrFYhAPOTQbjDdKvNt3K0kM7MycTeZNBD6XkVP0GBE', '2026-01-24 23:40:46', NULL, '2026-01-24 23:40:44', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MjcyODQ2LCJuYmYiOjE3NjkyNjkyNDYsImlhdCI6MTc2OTI2OTI0Nn0.Kv2USd7rf0HAcDOLO_fuaSQtXQHBbxDnnpw6Wq3mlSM');
INSERT INTO `system_user_token` VALUES (16, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc0NDIwLCJuYmYiOjE3NjkyNjk2MjAsImlhdCI6MTc2OTI2OTYyMH0.I1CCZ_6-f-VK0nz3hPd7XtBCTMxRCqxL2_kSZlumVe8', '2026-01-24 23:47:01', NULL, '2026-01-24 23:46:58', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MjczMjIwLCJuYmYiOjE3NjkyNjk2MjAsImlhdCI6MTc2OTI2OTYyMH0.L73lk3G3pir6Z0wmZ_U8HnNLtqNcdjNpM0ueLhP6ZgM');
INSERT INTO `system_user_token` VALUES (17, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc0NjM4LCJuYmYiOjE3NjkyNjk4MzgsImlhdCI6MTc2OTI2OTgzOH0.tbasIXozwxNoSuZEVMnnYqxSf_NF6ciJbW5-5YE4HhY', '2026-01-24 23:50:38', NULL, '2026-01-24 23:50:36', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MjczNDM4LCJuYmYiOjE3NjkyNjk4MzgsImlhdCI6MTc2OTI2OTgzOH0.ZSjRGWC6I_np6riL_E-P6dL91upr7h9Dqn9yhtiDmyc');
INSERT INTO `system_user_token` VALUES (18, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc0NjU0LCJuYmYiOjE3NjkyNjk4NTQsImlhdCI6MTc2OTI2OTg1NH0.0Mr7bhZ1YTUCtZKBX4j7tEjtl6Oa-1MvbNItKNWZakg', '2026-01-24 23:50:55', NULL, '2026-01-24 23:50:52', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MjczNDU0LCJuYmYiOjE3NjkyNjk4NTQsImlhdCI6MTc2OTI2OTg1NH0.MjN3lgdrHbVpxxw9Uj4qaGtr1VRw82wdmJnPTrUEloc');
INSERT INTO `system_user_token` VALUES (19, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc0Njk1LCJuYmYiOjE3NjkyNjk4OTUsImlhdCI6MTc2OTI2OTg5NX0.Mn_CyDA_iP7vs8pV1mEKbEJV01cr3L6eyVkDXCo4tuA', '2026-01-24 23:51:36', NULL, '2026-01-24 23:51:33', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MjczNDk1LCJuYmYiOjE3NjkyNjk4OTUsImlhdCI6MTc2OTI2OTg5NX0.GHBnpkVAQP73P5wMn0im4odWcSmkqnyFsbxNQX2iiwc');
INSERT INTO `system_user_token` VALUES (20, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc0ODY5LCJuYmYiOjE3NjkyNzAwNjksImlhdCI6MTc2OTI3MDA2OX0.mOpVPnfiltFgdBz9Hlv6DRBTFt6y-U_u_WF6O4q0TJ0', '2026-01-24 23:54:30', NULL, '2026-01-24 23:54:27', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MjczNjY5LCJuYmYiOjE3NjkyNzAwNjksImlhdCI6MTc2OTI3MDA2OX0.NTIwWcSGo195nmpjsetZtSOfwZPu9fb5ISeoTz3wSkc');
INSERT INTO `system_user_token` VALUES (21, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc1MjM2LCJuYmYiOjE3NjkyNzA0MzYsImlhdCI6MTc2OTI3MDQzNn0.IqHaZcrdBJ5QfF-vuo6THQ0JDaw_1G7eyHdkVZPYtmI', '2026-01-25 00:00:36', NULL, '2026-01-25 00:00:34', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5Mjc0MDM2LCJuYmYiOjE3NjkyNzA0MzYsImlhdCI6MTc2OTI3MDQzNn0._om_EI0zG3WUUCHjkCmKEDYl66qJti2aAc-CrhzL_0U');
INSERT INTO `system_user_token` VALUES (22, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc1MjQ1LCJuYmYiOjE3NjkyNzA0NDUsImlhdCI6MTc2OTI3MDQ0NX0.7doF2FKbCRP47lU0SDYLLBjRpsbx3egVSzahAf1evIg', '2026-01-25 00:00:46', NULL, '2026-01-25 00:00:43', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5Mjc0MDQ1LCJuYmYiOjE3NjkyNzA0NDUsImlhdCI6MTc2OTI3MDQ0NX0.Vmghtak4zPnkacGeqA-RbXGR8VrPysqa3YsCgjPPROI');
INSERT INTO `system_user_token` VALUES (23, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc1NjE4LCJuYmYiOjE3NjkyNzA4MTgsImlhdCI6MTc2OTI3MDgxOH0.O3_HYTpsXcVNBE9n3YolXdEzSb4lj2-2WBLLQyyveFQ', '2026-01-25 00:06:59', NULL, '2026-01-25 00:06:57', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5Mjc0NDE4LCJuYmYiOjE3NjkyNzA4MTgsImlhdCI6MTc2OTI3MDgxOH0.NQ887fhoi8ocN8cgUoFeiaIEzklallkVr2MIDEEVUSE');
INSERT INTO `system_user_token` VALUES (24, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc2MDU3LCJuYmYiOjE3NjkyNzEyNTcsImlhdCI6MTc2OTI3MTI1N30.a5OyCY79ZSyS6alGZJk7TuUlxjMjWuOKjtl7E43cDk0', '2026-01-25 00:14:18', NULL, '2026-01-25 00:14:15', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5Mjc0ODU3LCJuYmYiOjE3NjkyNzEyNTcsImlhdCI6MTc2OTI3MTI1N30.YdcdKfzJ_E-rtrSqwF6POR-_vGSbxboaF0JLPYw9Bos');
INSERT INTO `system_user_token` VALUES (25, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc2NjEwLCJuYmYiOjE3NjkyNzE4MTAsImlhdCI6MTc2OTI3MTgxMH0.dRM_CKBlubeyvB3dGVolY7s0HFM7AedLK37aZO5aMwg', '2026-01-25 00:23:31', NULL, '2026-01-25 00:23:29', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5Mjc1NDEwLCJuYmYiOjE3NjkyNzE4MTAsImlhdCI6MTc2OTI3MTgxMH0._VfUMKBXLu5SPZOisPstZ1_y49wMmKQsKXiqQAe0AVs');
INSERT INTO `system_user_token` VALUES (26, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc2NjM0LCJuYmYiOjE3NjkyNzE4MzQsImlhdCI6MTc2OTI3MTgzNH0.q3Kspq6i0csMZADUFIWP2QHB6iy_QL6NodFO-V9KnfQ', '2026-01-25 00:23:54', NULL, '2026-01-25 00:23:53', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5Mjc1NDM0LCJuYmYiOjE3NjkyNzE4MzQsImlhdCI6MTc2OTI3MTgzNH0.yPBziqDm-yZB3Lqo1RHFhiV2faheAxZZj3AxJ5T4orE');
INSERT INTO `system_user_token` VALUES (27, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc2ODQ1LCJuYmYiOjE3NjkyNzIwNDUsImlhdCI6MTc2OTI3MjA0NX0.cLrZ0MiA0i7qofVoIz2U-WUMYsCQ5mSMCtK9cH6ok9w', '2026-01-25 00:27:25', NULL, '2026-01-25 00:27:24', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5Mjc1NjQ1LCJuYmYiOjE3NjkyNzIwNDUsImlhdCI6MTc2OTI3MjA0NX0.1TdgXpJfNGF_RgdHo1tC1IRUiYC5ibZI8SOlFf4WOJo');
INSERT INTO `system_user_token` VALUES (28, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc2ODk1LCJuYmYiOjE3NjkyNzIwOTUsImlhdCI6MTc2OTI3MjA5NX0.ssc6RaTXtESw2kmbDd8A9KNQP71gwKrYErrexPozCdQ', '2026-01-25 00:28:16', NULL, '2026-01-25 00:28:15', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5Mjc1Njk1LCJuYmYiOjE3NjkyNzIwOTUsImlhdCI6MTc2OTI3MjA5NX0.Yaqu7mb34lhYfn_lhGSr0Wp0MdjDnTp_hbJ0vjaiolQ');
INSERT INTO `system_user_token` VALUES (29, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc2OTQ0LCJuYmYiOjE3NjkyNzIxNDQsImlhdCI6MTc2OTI3MjE0NH0.6CmZF-jwITfGiv7upg1QiaYpGgROu61Uv9ZF3vCWjgI', '2026-01-25 00:29:04', NULL, '2026-01-25 00:29:03', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5Mjc1NzQ0LCJuYmYiOjE3NjkyNzIxNDQsImlhdCI6MTc2OTI3MjE0NH0.QSyfPi9B1JF5sGo_rBlNfW_7Szm8_cNO_cFSj32d4tg');
INSERT INTO `system_user_token` VALUES (30, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc2OTg5LCJuYmYiOjE3NjkyNzIxODksImlhdCI6MTc2OTI3MjE4OX0.QKtU2w7Uga8K6CU8_Bg4_p94oH4xzZB1Myz_asnGQus', '2026-01-25 00:29:49', NULL, '2026-01-25 00:29:48', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5Mjc1Nzg5LCJuYmYiOjE3NjkyNzIxODksImlhdCI6MTc2OTI3MjE4OX0.c9ifKDgGi1znesg-Vef2-G2e039w0zGa7iTGmEtR3AQ');
INSERT INTO `system_user_token` VALUES (31, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc3MDA0LCJuYmYiOjE3NjkyNzIyMDQsImlhdCI6MTc2OTI3MjIwNH0.bhFDToo-FCVkt_DTtXHtyISrIcdoxfuZg0ZxcO8GheE', '2026-01-25 00:30:04', NULL, '2026-01-25 00:30:03', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5Mjc1ODA0LCJuYmYiOjE3NjkyNzIyMDQsImlhdCI6MTc2OTI3MjIwNH0.EeYd0TvTmmfHBXqwIVzFOgP9EIV6igaMWMI_OXCQBXY');
INSERT INTO `system_user_token` VALUES (32, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc3NTgwLCJuYmYiOjE3NjkyNzI3ODAsImlhdCI6MTc2OTI3Mjc4MH0.z6r2_bVMB0O1VGVttL33NgYZSdt_GyW_YIX4ceS_Css', '2026-01-25 00:39:40', NULL, '2026-01-25 00:39:39', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5Mjc2MzgwLCJuYmYiOjE3NjkyNzI3ODAsImlhdCI6MTc2OTI3Mjc4MH0.BIV3HmudilrdACD75IGI77dVAHos_FlHyuNQoD3FEbQ');
INSERT INTO `system_user_token` VALUES (33, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc3NjU4LCJuYmYiOjE3NjkyNzI4NTgsImlhdCI6MTc2OTI3Mjg1OH0.wivViM4poar7PtdWTS5oN46f0HF0dEBldCJKHGJBjm4', '2026-01-25 00:40:59', NULL, '2026-01-25 00:40:58', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5Mjc2NDU4LCJuYmYiOjE3NjkyNzI4NTgsImlhdCI6MTc2OTI3Mjg1OH0.qiJbWzU1q2FdkSJY24QJYKW-LfrpgJgb38tvUUyGtQg');
INSERT INTO `system_user_token` VALUES (34, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5ODc3OTMyLCJuYmYiOjE3NjkyNzMxMzIsImlhdCI6MTc2OTI3MzEzMn0.ToUYd5EQzmoypp66hXoUffcPcgpf61IfTq3Vrn6HXgo', '2026-01-25 00:45:32', NULL, '2026-01-25 00:45:31', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5Mjc2NzMyLCJuYmYiOjE3NjkyNzMxMzIsImlhdCI6MTc2OTI3MzEzMn0.9kABvyLNOcgnLhrui5s7CsfnoOFMvCO0CrGsDRwKqIE');
INSERT INTO `system_user_token` VALUES (35, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5OTI2MzM0LCJuYmYiOjE3NjkzMjE1MzQsImlhdCI6MTc2OTMyMTUzNH0.bEeCmIZfVJqymdtt1n1AkpwRZ9jP-UYuyCA7fndnS7k', '2026-01-25 14:12:14', NULL, '2026-01-25 14:12:12', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MzI1MTM0LCJuYmYiOjE3NjkzMjE1MzQsImlhdCI6MTc2OTMyMTUzNH0.hcUr3y_M8a3ovKbyuzMMIXjgC3GhIcudfSCas0xKv9g');
INSERT INTO `system_user_token` VALUES (36, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5OTI2MzU5LCJuYmYiOjE3NjkzMjE1NTksImlhdCI6MTc2OTMyMTU1OX0.ceLmz5wzu-bc-J4TeLynTV8TN4mAh286gGwCLhI4_70', '2026-01-25 14:12:40', NULL, '2026-01-25 14:12:37', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MzI1MTU5LCJuYmYiOjE3NjkzMjE1NTksImlhdCI6MTc2OTMyMTU1OX0.CuEBsf9hjtuPlsvAKcD5D2Xwc1EFoD271IRKCBszxoA');
INSERT INTO `system_user_token` VALUES (37, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5OTI2ODIyLCJuYmYiOjE3NjkzMjIwMjIsImlhdCI6MTc2OTMyMjAyMn0.ggwjJYbqOnL0xpN_M8PhlG68x7xJTfSI9l9lPgMlVfo', '2026-01-25 14:20:23', NULL, '2026-01-25 14:20:20', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MzI1NjIyLCJuYmYiOjE3NjkzMjIwMjIsImlhdCI6MTc2OTMyMjAyMn0.VgAXRx7rFOMHtSk_YCr0DTcV4rkQo_hpAHIvY3TKOXE');
INSERT INTO `system_user_token` VALUES (38, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5OTI3MDIwLCJuYmYiOjE3NjkzMjIyMjAsImlhdCI6MTc2OTMyMjIyMH0.wUF70F0bfPdmb65ogpcjpx7xJNEhlli9hy0wMibWaxY', '2026-01-25 14:23:40', NULL, '2026-01-25 14:23:38', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MzI1ODIwLCJuYmYiOjE3NjkzMjIyMjAsImlhdCI6MTc2OTMyMjIyMH0.3jWh3QBBU_9oh93oeAbnYyRFebw9xAxHlcfzeB1LXIQ');
INSERT INTO `system_user_token` VALUES (39, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5OTI3MDUxLCJuYmYiOjE3NjkzMjIyNTEsImlhdCI6MTc2OTMyMjI1MX0.pM-KtpzMBXUIZ_kZOJMNnYs8ue7XibY8VdlGBVXHMVo', '2026-01-25 14:24:12', NULL, '2026-01-25 14:24:09', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MzI1ODUxLCJuYmYiOjE3NjkzMjIyNTEsImlhdCI6MTc2OTMyMjI1MX0.I7tvRhgttQk0x7MXm5CXJR2IuveLEVekqaI7IiPwU_o');
INSERT INTO `system_user_token` VALUES (40, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5OTMxMzU3LCJuYmYiOjE3NjkzMjY1NTcsImlhdCI6MTc2OTMyNjU1N30.fKNs9IiqtKfUgrcgBtmnMHyaDy28AE8X-tNMfd_ACLw', '2026-01-25 15:35:57', NULL, '2026-01-25 15:35:55', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MzMwMTU3LCJuYmYiOjE3NjkzMjY1NTcsImlhdCI6MTc2OTMyNjU1N30.cSyQFqwbkCUmhWNBfSy_tZgIiYzsxGLNU59uwgG_qjc');
INSERT INTO `system_user_token` VALUES (41, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5OTM2Mjk3LCJuYmYiOjE3NjkzMzE0OTcsImlhdCI6MTc2OTMzMTQ5N30.CbRbnKxQrFf7Muplzb9oacMA80E-85MiNCGrvkGjD84', '2026-01-25 16:58:17', NULL, '2026-01-25 16:58:15', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MzM1MDk3LCJuYmYiOjE3NjkzMzE0OTcsImlhdCI6MTc2OTMzMTQ5N30.8PdXeTraWa7yanppXtWbJatn1ekTCtkX7brY1eGt110');
INSERT INTO `system_user_token` VALUES (42, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5OTQwMTY2LCJuYmYiOjE3NjkzMzUzNjYsImlhdCI6MTc2OTMzNTM2Nn0.1P0xjEbdoOrdSlgr6WD6FKj7Iep9cQk9ScakkP52Ssg', '2026-01-25 18:02:47', NULL, '2026-01-25 18:02:45', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MzM4OTY2LCJuYmYiOjE3NjkzMzUzNjYsImlhdCI6MTc2OTMzNTM2Nn0.l_ZIeSMAHQA92XUz7kHnW6RrudchpC0jCXPRY-S7OAA');
INSERT INTO `system_user_token` VALUES (43, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5OTQ3NjA0LCJuYmYiOjE3NjkzNDI4MDQsImlhdCI6MTc2OTM0MjgwNH0._t_TszQkVkyn0VlTaJZlPpu2C2BEGsJk33vc4wEP8rU', '2026-01-25 20:06:44', NULL, '2026-01-25 20:06:45', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MzQ2NDA0LCJuYmYiOjE3NjkzNDI4MDQsImlhdCI6MTc2OTM0MjgwNH0.cDAHU8vmVheyvEB4U1DEd-XAPiHGN3jTOw8Cx_Ovulo');
INSERT INTO `system_user_token` VALUES (44, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzY5OTU0OTgzLCJuYmYiOjE3NjkzNTAxODMsImlhdCI6MTc2OTM1MDE4M30.Z3sSpOWjP_-Y8vR5WKuj-ifpUMVaCUly-UTVp_Ow170', '2026-01-25 22:09:44', NULL, '2026-01-25 22:09:43', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5MzUzNzgzLCJuYmYiOjE3NjkzNTAxODMsImlhdCI6MTc2OTM1MDE4M30.Q2_A8zOCUaZv33q8EHY3c9-BTVwewvuzmLqV1VY4wkg');
INSERT INTO `system_user_token` VALUES (45, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjYxODk4LCJuYmYiOjE3Njk2NTcwOTgsImlhdCI6MTc2OTY1NzA5OH0.0VGxo9vf8B2_QkswCfx8Zc8aSkNZYf0e_VpircFHntY', '2026-01-29 11:24:59', NULL, '2026-01-29 11:24:58', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5NjYwNjk4LCJuYmYiOjE3Njk2NTcwOTgsImlhdCI6MTc2OTY1NzA5OH0.HhiOEYWG6mrvYTCmXem2YD7y3v678m063kL8kYO3fbo');
INSERT INTO `system_user_token` VALUES (46, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjYxOTEzLCJuYmYiOjE3Njk2NTcxMTMsImlhdCI6MTc2OTY1NzExM30.B9OfUP-1SFxWdU8FRC0NWxzYSKZtCpnNcUvC8djRjYs', '2026-01-29 11:25:14', NULL, '2026-01-29 11:25:14', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5NjYwNzEzLCJuYmYiOjE3Njk2NTcxMTMsImlhdCI6MTc2OTY1NzExM30.ffbZsj1ivrCVDQ0emH_y9jLbxfUvAdwiYjvWfxc5dug');
INSERT INTO `system_user_token` VALUES (47, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjYxOTQ3LCJuYmYiOjE3Njk2NTcxNDcsImlhdCI6MTc2OTY1NzE0N30.UApNLYxBbWqmWPua_seMJFYORbROV_4aTWG6qgvp0Ik', '2026-01-29 11:25:47', NULL, '2026-01-29 11:25:47', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5NjYwNzQ3LCJuYmYiOjE3Njk2NTcxNDcsImlhdCI6MTc2OTY1NzE0N30.UqgoiiPRxnnTUJTsr8IyZPRwOAHmuhsGfR0uTyj2H4E');
INSERT INTO `system_user_token` VALUES (48, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjYyMTI5LCJuYmYiOjE3Njk2NTczMjksImlhdCI6MTc2OTY1NzMyOX0.ZkuJimgbCLIP_zouTHsWFuL0ty--AoTu-AM9y0DaZTM', '2026-01-29 11:28:49', NULL, '2026-01-29 11:28:49', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5NjYwOTI5LCJuYmYiOjE3Njk2NTczMjksImlhdCI6MTc2OTY1NzMyOX0._wgXCoXayJGPg3k0l1KKucbMEBR_cbwTdZXNKFzLB0A');
INSERT INTO `system_user_token` VALUES (49, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjYyMjY4LCJuYmYiOjE3Njk2NTc0NjgsImlhdCI6MTc2OTY1NzQ2OH0.ErKAfg18Fnx3flGOM5oUVjZoL1CciaQyjy11i7R_C5k', '2026-01-29 11:31:09', NULL, '2026-01-29 11:31:09', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5NjYxMDY4LCJuYmYiOjE3Njk2NTc0NjgsImlhdCI6MTc2OTY1NzQ2OH0.Blr1I1Yi_bfx3CcDMbeww-tehX8BBqUtNFrC58EWaUc');
INSERT INTO `system_user_token` VALUES (50, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjYyNDI3LCJuYmYiOjE3Njk2NTc2MjcsImlhdCI6MTc2OTY1NzYyN30.eq0vuO0I2nmam7Wxve6oo9TcX2f8TbYjDXZilmh-dWY', '2026-01-29 11:34:40', NULL, '2026-01-29 11:34:40', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMDE3NjI3LCJuYmYiOjE3Njk2NTc2MjcsImlhdCI6MTc2OTY1NzYyN30.XusUrAIm6HSYwp0WGA20NpenLwbRcJ5UFfZd9pMV3To');
INSERT INTO `system_user_token` VALUES (51, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjYyNTU3LCJuYmYiOjE3Njk2NTc3NTcsImlhdCI6MTc2OTY1Nzc1N30.CxJO75RQZoNb4SpiEGzxP61u1uqt2j_JxePVkRAoErY', '2026-01-29 11:35:58', NULL, '2026-01-29 11:35:58', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMDE3NzU3LCJuYmYiOjE3Njk2NTc3NTcsImlhdCI6MTc2OTY1Nzc1N30.yXl9DcYUurLPy10-0RfATAwatsJ0Ywqr67sMkujMhVs');
INSERT INTO `system_user_token` VALUES (52, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjYyODcxLCJuYmYiOjE3Njk2NTgwNzEsImlhdCI6MTc2OTY1ODA3MX0.SNhpvcCkESm7vi3hqTIk0lReU8fFDNuI13MGqfke340', '2026-01-29 11:41:12', NULL, '2026-01-29 11:41:12', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMDE4MDcxLCJuYmYiOjE3Njk2NTgwNzEsImlhdCI6MTc2OTY1ODA3MX0.pAcwIhCWYG_0qB93mch3x-5aCpcUFXfKyL_EWWEYZSE');
INSERT INTO `system_user_token` VALUES (53, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjYyOTM3LCJuYmYiOjE3Njk2NTgxMzcsImlhdCI6MTc2OTY1ODEzN30.M0suL4heeYnXCATEhpJOIrS277JVdNFdRSat79zXDjU', '2026-01-29 11:42:17', NULL, '2026-01-29 11:42:17', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMDE4MTM3LCJuYmYiOjE3Njk2NTgxMzcsImlhdCI6MTc2OTY1ODEzN30.Fwo6bgRFWkKDbWDARFfeEi_CXSVyVgatF2T9ZwZ4xm4');
INSERT INTO `system_user_token` VALUES (54, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjYzMTM4LCJuYmYiOjE3Njk2NTgzMzgsImlhdCI6MTc2OTY1ODMzOH0.KL0IDxNM9BhOQIdeA-uDi7zQpxSChsBBQZtsgp8XhZI', '2026-01-29 11:45:41', NULL, '2026-01-29 11:45:41', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMDE4MzM4LCJuYmYiOjE3Njk2NTgzMzgsImlhdCI6MTc2OTY1ODMzOH0.0IbjUH91tq6MH2I3gpX-kqn-gy8bl7sEtwgW45qLTo4');
INSERT INTO `system_user_token` VALUES (55, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjYzMTUxLCJuYmYiOjE3Njk2NTgzNTEsImlhdCI6MTc2OTY1ODM1MX0.2_lJXx4EpozIJTgpK6OZceYRH491kVcJnNMoTkYcgFU', '2026-01-29 11:45:51', NULL, '2026-01-29 11:45:51', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMDE4MzUxLCJuYmYiOjE3Njk2NTgzNTEsImlhdCI6MTc2OTY1ODM1MX0.m1WGzg7xfR616kaG0_1yXsJzuSZs33Lj9vv-i-yE5oI');
INSERT INTO `system_user_token` VALUES (56, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjYzMjE2LCJuYmYiOjE3Njk2NTg0MTYsImlhdCI6MTc2OTY1ODQxNn0.tKGxW74_pP5lTFx_MYPOqlbPCyjkotdt7szdcMyvOaM', '2026-01-29 11:46:56', NULL, '2026-01-29 11:46:56', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMDE4NDE2LCJuYmYiOjE3Njk2NTg0MTYsImlhdCI6MTc2OTY1ODQxNn0.mE3sR9dMsuY0Gux_xfvjpPXSEKIyndQP1_RGmUIT9jw');
INSERT INTO `system_user_token` VALUES (57, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjYzMjkyLCJuYmYiOjE3Njk2NTg0OTIsImlhdCI6MTc2OTY1ODQ5Mn0.fIhHs8wHxj5hvN0-B1o53KWU5pm8AC2bu4LB1XlZZ44', '2026-01-29 11:48:13', NULL, '2026-01-29 11:48:12', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMDE4NDkyLCJuYmYiOjE3Njk2NTg0OTIsImlhdCI6MTc2OTY1ODQ5Mn0.l5DpCvOiVC8-_xgHpqFac3cvBxvrzlQs48TemadEW1c');
INSERT INTO `system_user_token` VALUES (58, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjYzMzAyLCJuYmYiOjE3Njk2NTg1MDIsImlhdCI6MTc2OTY1ODUwMn0.GlyaNxlmEfNBo6_5_azQt50xPhIG-a8nfQUMuWFhIEc', '2026-01-29 11:48:23', NULL, '2026-01-29 11:48:23', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMDE4NTAyLCJuYmYiOjE3Njk2NTg1MDIsImlhdCI6MTc2OTY1ODUwMn0.huO3w994jtn2VwVmZc86LhFemjOAE88R2IfjJHzOkZI');
INSERT INTO `system_user_token` VALUES (59, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjYzMzQ0LCJuYmYiOjE3Njk2NTg1NDQsImlhdCI6MTc2OTY1ODU0NH0.DyRqUso2yHAR_H0G_eb5lyXp-0NoPEERGf0aj6LHdHk', '2026-01-29 11:49:04', NULL, '2026-01-29 11:49:04', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMDE4NTQ0LCJuYmYiOjE3Njk2NTg1NDQsImlhdCI6MTc2OTY1ODU0NH0.tmxj8BHoa97v4j_Bj-K1a24d-lAcQO4QX6gBPVk-lSM');
INSERT INTO `system_user_token` VALUES (60, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjY0MDI0LCJuYmYiOjE3Njk2NTkyMjQsImlhdCI6MTc2OTY1OTIyNH0.Ew9V7bBlwiD_bSaMYrEqXKjadxNxamzqJRYStNC-ufo', '2026-01-29 12:00:25', NULL, '2026-01-29 12:00:25', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMDE5MjI0LCJuYmYiOjE3Njk2NTkyMjQsImlhdCI6MTc2OTY1OTIyNH0.8ZDNcqePeEmB_W5fJ9it2DyLMEqv31yBO4amqDPoxsY');
INSERT INTO `system_user_token` VALUES (61, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjY0MjQzLCJuYmYiOjE3Njk2NTk0NDMsImlhdCI6MTc2OTY1OTQ0M30.xiDBEsBtd5vBQ2lslcHU8x2b8SU_5Ibfg2oVc-7adxs', '2026-01-29 12:04:04', NULL, '2026-01-29 12:04:04', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMDE5NDQzLCJuYmYiOjE3Njk2NTk0NDMsImlhdCI6MTc2OTY1OTQ0M30.vNwdZLPVj_5ukMXJWKZ1mBiRcEeBya1DXkwWKBrukCc');
INSERT INTO `system_user_token` VALUES (62, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjY0MzYwLCJuYmYiOjE3Njk2NTk1NjAsImlhdCI6MTc2OTY1OTU2MH0.yvZBQnAgjO0gw3LvCKGaIouvKQqV1Hgux6w9ntVDaPg', '2026-01-29 12:06:00', NULL, '2026-01-29 12:06:00', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMDE5NTYwLCJuYmYiOjE3Njk2NTk1NjAsImlhdCI6MTc2OTY1OTU2MH0.kiCa-O9HCrsKsmrrfdAtaY3k-SDEYxdUceGJb5EyjSM');
INSERT INTO `system_user_token` VALUES (63, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjY0Mzk1LCJuYmYiOjE3Njk2NTk1OTUsImlhdCI6MTc2OTY1OTU5NX0.YQKH1DSiCd-91UHHyOYlTicbsJBJOWT_pmZKTcdvx0I', '2026-01-29 12:06:35', NULL, '2026-01-29 12:06:35', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMDE5NTk1LCJuYmYiOjE3Njk2NTk1OTUsImlhdCI6MTc2OTY1OTU5NX0.Yxs0fbBLl9pqc5wD-vGEfvdpqucewgfoCSC5CyTJX_w');
INSERT INTO `system_user_token` VALUES (64, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwMjY0NDU4LCJuYmYiOjE3Njk2NTk2NTgsImlhdCI6MTc2OTY1OTY1OH0.Ps0_5ciZaZplF8p6TnwCuo1sS2mfTF8a5XHDnovR2Z0', '2026-01-29 12:07:39', NULL, '2026-01-29 12:07:39', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMDE5NjU4LCJuYmYiOjE3Njk2NTk2NTgsImlhdCI6MTc2OTY1OTY1OH0._4ymSXJYOBb5uZXc3-vt40fl-ALis-ue_ewBO-VYUjg');
INSERT INTO `system_user_token` VALUES (65, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwNDM1NDYwLCJuYmYiOjE3Njk4MzA2NjAsImlhdCI6MTc2OTgzMDY2MH0.VdUQj8ZoOcSeRA83NhKk5YvQj8m_Jy_v5tIU3xXfN34', '2026-01-31 11:37:41', NULL, '2026-01-31 11:37:42', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5ODM0MjYwLCJuYmYiOjE3Njk4MzA2NjAsImlhdCI6MTc2OTgzMDY2MH0.J48TesEbVh-FWWsNOjoRUOX5MtTbqZjLdxq0FmkAGQQ');
INSERT INTO `system_user_token` VALUES (66, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwNDQwNjkxLCJuYmYiOjE3Njk4MzU4OTEsImlhdCI6MTc2OTgzNTg5MX0.zb0lk7_n8fU5wq66SdIT2KSSdWu1aHhCqYc4IHLDYQw', '2026-01-31 13:04:51', NULL, '2026-01-31 13:04:51', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5ODM5NDkxLCJuYmYiOjE3Njk4MzU4OTEsImlhdCI6MTc2OTgzNTg5MX0.QFnXh-sZ7RfyUvd2KuNtchatzth_K2p0JBF6BpdVmTI');
INSERT INTO `system_user_token` VALUES (67, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwNDQ5MzAyLCJuYmYiOjE3Njk4NDQ1MDIsImlhdCI6MTc2OTg0NDUwMn0.7hGvbKonXai9DOogfegiWqldyCpXL_cJI-BLzjzGWfs', '2026-01-31 15:28:23', NULL, '2026-01-31 15:28:23', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5ODQ4MTAyLCJuYmYiOjE3Njk4NDQ1MDIsImlhdCI6MTc2OTg0NDUwMn0.4h0_5W5Q5pJpCZXcrEmcL2AiCwpwc0K4sLTob90pwp8');
INSERT INTO `system_user_token` VALUES (68, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwNDUzNDI5LCJuYmYiOjE3Njk4NDg2MjksImlhdCI6MTc2OTg0ODYyOX0.5QWuppU81pq33t7xvvFi9rJdngp19V6C29ZueoxX8xs', '2026-01-31 16:37:10', NULL, '2026-01-31 16:37:10', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5ODUyMjI5LCJuYmYiOjE3Njk4NDg2MjksImlhdCI6MTc2OTg0ODYyOX0.3YN3qcYmkm1ubUCK5ZGjMFQlVv7WgqRMjlIqdHGHbz4');
INSERT INTO `system_user_token` VALUES (69, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwNDU3MDQzLCJuYmYiOjE3Njk4NTIyNDMsImlhdCI6MTc2OTg1MjI0M30.AxlyuXLY-oJemrAyLtTTwpKltoAj61pn-PnOHcu3BFE', '2026-01-31 17:37:24', NULL, '2026-01-31 17:37:24', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5ODU1ODQzLCJuYmYiOjE3Njk4NTIyNDMsImlhdCI6MTc2OTg1MjI0M30.YaG25hk5dG9xtKJOsNgGrMj_aARvDC-x2O3Cn_KUHqs');
INSERT INTO `system_user_token` VALUES (70, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwNDYwNzAwLCJuYmYiOjE3Njk4NTU5MDAsImlhdCI6MTc2OTg1NTkwMH0.64_RYjJqm-Ii8ZS4Ig6xV_4HJ5MwNFm7n6N04G-0RuQ', '2026-01-31 18:38:21', NULL, '2026-01-31 18:38:21', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzY5ODU5NTAwLCJuYmYiOjE3Njk4NTU5MDAsImlhdCI6MTc2OTg1NTkwMH0.U12X_wPm6vtbRjW0T5f3AJNAgpva9IjL-bbuWF37NIs');
INSERT INTO `system_user_token` VALUES (71, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwNjAyNDI4LCJuYmYiOjE3Njk5OTc2MjgsImlhdCI6MTc2OTk5NzYyOH0.LmBME_jD600UhmI_-inMJ8M9hldyR-KhlA4ifO5MCTE', '2026-02-02 10:00:28', NULL, '2026-02-02 10:00:28', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMDAxMjI4LCJuYmYiOjE3Njk5OTc2MjgsImlhdCI6MTc2OTk5NzYyOH0.C0xD7u3SipGxPmHlRY7ypRukJ6RmAJ4DDXLgZHJnxmQ');
INSERT INTO `system_user_token` VALUES (72, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwNjEwMzg2LCJuYmYiOjE3NzAwMDU1ODYsImlhdCI6MTc3MDAwNTU4Nn0.c1AhN_k-uvGAK71rUrJWfxwQR5Ho3CHAeD6WkZmaHUM', '2026-02-02 12:13:06', NULL, '2026-02-02 12:13:06', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMDA5MTg2LCJuYmYiOjE3NzAwMDU1ODYsImlhdCI6MTc3MDAwNTU4Nn0.GatRGReX2hIq2GyEkVeiTXo_4VB3eVcznmV2uE_scgI');
INSERT INTO `system_user_token` VALUES (73, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwNjI5NTY0LCJuYmYiOjE3NzAwMjQ3NjQsImlhdCI6MTc3MDAyNDc2NH0.g4YI0AfsjsYG_VYuO85ofx4d2GKfHG3ObRIeelgPJhw', '2026-02-02 17:32:45', NULL, '2026-02-02 17:32:45', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMDI4MzY0LCJuYmYiOjE3NzAwMjQ3NjQsImlhdCI6MTc3MDAyNDc2NH0.q3M_L9WJe2TW8GnUPi3MLLXtGVKNm_lcQPFpL6nJHJA');
INSERT INTO `system_user_token` VALUES (74, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwODE0Mjg1LCJuYmYiOjE3NzAyMDk0ODUsImlhdCI6MTc3MDIwOTQ4NX0.psuDFR6gShSx5uOu2Vtk3rwhJdWD8_BRvOsgKKBJqhg', '2026-02-04 20:51:26', NULL, '2026-02-04 20:51:25', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMjEzMDg1LCJuYmYiOjE3NzAyMDk0ODUsImlhdCI6MTc3MDIwOTQ4NX0.6z6-cUkgmy1g05AWBGbQCcBjixFDaOIDYSAhdTKMoSA');
INSERT INTO `system_user_token` VALUES (75, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwODE1MTA1LCJuYmYiOjE3NzAyMTAzMDUsImlhdCI6MTc3MDIxMDMwNX0.bEMix_nTooRlOc8R9b9ODd3rNiO-CEYjr7di_ukvIpA', '2026-02-04 21:05:06', NULL, '2026-02-04 21:05:05', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMjEzOTA1LCJuYmYiOjE3NzAyMTAzMDUsImlhdCI6MTc3MDIxMDMwNX0.lnPmXn8z2TUgA6VTMcyxb3CiXcGbpgv5gr2BGBXCUe8');
INSERT INTO `system_user_token` VALUES (76, 1, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiIiLCJSb2xlcyI6bnVsbCwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwic3ViIjoicmVmcmVzaF8xIiwiZXhwIjoxNzcwODE1MTMyLCJuYmYiOjE3NzAyMTAzMzIsImlhdCI6MTc3MDIxMDMzMn0.qyXNuYUk-NG4ZwY_OK1mRbvQbg9kJQ88X7EtDpmOssM', '2026-02-04 21:05:32', NULL, '2026-02-04 21:05:32', NULL, NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiVXNlcm5hbWUiOiJhZG1pbiIsIlJvbGVzIjpbXSwiaXNzIjoiazhzLXBsYXRmb3JtLWdvIiwiZXhwIjoxNzcwMjEzOTMyLCJuYmYiOjE3NzAyMTAzMzIsImlhdCI6MTc3MDIxMDMzMn0.Q_iJihskyhZ4mbkwhGaJ4h9G1P0HcLKkw0KPuwvwDMI');

-- ----------------------------
-- Table structure for system_users
-- ----------------------------
DROP TABLE IF EXISTS `system_users`;
CREATE TABLE `system_users`  (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `username` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
  `password` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
  `status` tinyint UNSIGNED NOT NULL,
  `nickname` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT '',
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_account_user_id`(`username` ASC) USING BTREE,
  INDEX `idx_account_deleted_at`(`deleted_at` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of system_users
-- ----------------------------
INSERT INTO `system_users` VALUES (1, 'admin', '240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9', 1, 'admin', '2026-01-24 21:17:21.000', '2026-01-24 21:17:25.000', NULL);

-- ----------------------------
-- View structure for resource_details
-- ----------------------------
DROP VIEW IF EXISTS `resource_details`;
CREATE ALGORITHM = UNDEFINED SQL SECURITY DEFINER VIEW `resource_details` AS select `ac`.`resource_type` AS `resource_type`,`ac`.`resource_id` AS `resource_id`,`ac`.`resource_name` AS `resource_name`,`ac`.`status` AS `status`,`ac`.`created_at` AS `created_at`,`ac`.`updated_at` AS `updated_at`,json_objectagg(`ac`.`config_key`,json_object('config_value',`ac`.`config_value`,'auth_type',`ac`.`auth_type`,'is_encrypted',`ac`.`is_encrypted`)) AS `auth_configs`,(case when (max((case when (`ac`.`auth_type` = 'kubeconfig') then 1 else 0 end)) = 1) then '配置文件认证' when (max((case when (`ac`.`auth_type` = 'token') then 1 else 0 end)) = 1) then '令牌认证' when (max((case when (`ac`.`auth_type` = 'basic') then 1 else 0 end)) = 1) then '基础认证' when (max((case when (`ac`.`auth_type` = 'api_key') then 1 else 0 end)) = 1) then 'API密钥认证' when (max((case when (`ac`.`auth_type` = 'certificate') then 1 else 0 end)) = 1) then '证书认证' when (max((case when (`ac`.`auth_type` = 'aws_iam') then 1 else 0 end)) = 1) then 'AWS IAM认证' else '无认证' end) AS `auth_type_desc`,max(`i`.`address`) AS `connection_endpoint`,max(`i`.`https_enabled`) AS `secure_connection`,max(`i`.`skip_ssl_verify`) AS `ssl_verification_disabled`,max(`it`.`type_name`) AS `resource_subtype`,max(`it`.`description`) AS `subtype_description`,(select `ct`.`test_result` from `connection_tests` `ct` where ((`ct`.`resource_type` = `ac`.`resource_type`) and (`ct`.`resource_id` = `ac`.`resource_id`)) order by `ct`.`tested_at` desc limit 1) AS `last_test_result`,(select `ct`.`response_time` from `connection_tests` `ct` where ((`ct`.`resource_type` = `ac`.`resource_type`) and (`ct`.`resource_id` = `ac`.`resource_id`)) order by `ct`.`tested_at` desc limit 1) AS `last_response_time`,(select `ct`.`tested_at` from `connection_tests` `ct` where ((`ct`.`resource_type` = `ac`.`resource_type`) and (`ct`.`resource_id` = `ac`.`resource_id`)) order by `ct`.`tested_at` desc limit 1) AS `last_test_time` from ((`auth_configs` `ac` left join `instances` `i` on(((`ac`.`resource_id` = `i`.`id`) and (`ac`.`resource_type` = 'instance') and (`ac`.`status` = 'active')))) left join `instance_types` `it` on((`i`.`instance_type_id` = `it`.`id`))) where (`ac`.`status` = 'active') group by `ac`.`resource_type`,`ac`.`resource_id`,`ac`.`resource_name`,`ac`.`status`,`ac`.`created_at`,`ac`.`updated_at` order by `ac`.`resource_type`,`ac`.`resource_name`;

SET FOREIGN_KEY_CHECKS = 1;
