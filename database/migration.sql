
CREATE DATABASE ofdhq_prod DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `ofdhq_prod`;

CREATE TABLE IF NOT EXISTS `topics` (
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `admin_user_id` INT(11) NOT NULL COMMENT '用户ID',
  `lang` varchar(100) DEFAULT '' COMMENT '语言',
  `title` VARCHAR(400) DEFAULT '' COMMENT '标题',
  `body` MEDIUMTEXT COMMENT '内容',
  `img_url` VARCHAR(800) NOT NULL COMMENT '图片',
  `flag` TINYINT(4) DEFAULT 1 COMMENT '状态：0-无效，1-有效',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `topics_idx_user_id` (`admin_user_id`),
  UNIQUE KEY `idx_title`(`title`),
  KEY `topics_idx_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `customers` (
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `first_name` varchar(80) DEFAULT '' COMMENT '姓',
  `last_name` varchar(80) DEFAULT '' COMMENT '名',
  `email` varchar(120) DEFAULT '' COMMENT '邮箱名',
  `subject` varchar(400) DEFAULT '' COMMENT '主题',
  `messages` MEDIUMTEXT COMMENT '信息',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `admin_users` (
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `real_name` VARCHAR(30) DEFAULT '' COMMENT '姓名',
  `email` VARCHAR(200) DEFAULT '' COMMENT '邮箱',
  `pass` VARCHAR(128) DEFAULT '' COMMENT '密码',
  `status` TINYINT(4) DEFAULT 1 COMMENT '状态:1-正常,0-禁用',
  `role_type` INT(11) DEFAULT 1 COMMENT '角色状态: 1-管理员,2-操作员',
  `last_login_time` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `last_login_ip` CHAR(30) DEFAULT '' COMMENT '最近一次登录ip',
  `login_times` INT(11) DEFAULT 0 COMMENT '累计登录次数',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `admin_users_indx_email` (`email`),
  KEY `users_indx_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `admin_oauth_access_tokens` (
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `admin_user_id` INT(11) DEFAULT 0 COMMENT '外键:admin_users表id',
  `client_id` INT(10) UNSIGNED DEFAULT 1 COMMENT '普通用户的授权，默认为1',
  `token` VARCHAR(500) DEFAULT NULL,
  `action_name` VARCHAR(128) CHARACTER SET utf8 COLLATE utf8_unicode_ci DEFAULT '' COMMENT 'login|refresh|reset表示token生成动作',
  `scopes` VARCHAR(128) CHARACTER SET utf8 COLLATE utf8_unicode_ci DEFAULT '[*]' COMMENT '暂时预留,未启用',
  `revoked` TINYINT(1) DEFAULT 0 COMMENT '是否撤销',
  `client_ip` VARCHAR(128) DEFAULT NULL COMMENT 'ipv6最长为128位',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `expires_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `admin_oauth_access_tokens_user_id_index` (`admin_user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;