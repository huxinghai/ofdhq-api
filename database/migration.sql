
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

-- ------------------- 抖音本地生活 SPI 对接（核心字段建列，非核心字段存 extra JSON） -------------------

-- 预售订单（SPI: travel_spot.order.create_presale_order，biz_type 3011）
CREATE TABLE IF NOT EXISTS `douyin_presale_orders` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id` VARCHAR(64) NOT NULL COMMENT '抖音侧预售订单号',
  `order_out_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '我方订单号(雪花ID,创单成功必返)',
  `biz_type` INT NOT NULL DEFAULT 3011 COMMENT '业务类型:3011旅行社预售券',
  `create_order_time_unix` INT NOT NULL DEFAULT 0 COMMENT '下单时间(秒级时间戳)',
  `total_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '总原始金额(分)',
  `total_coupon_count` INT NOT NULL DEFAULT 0 COMMENT '券总张数(预留,一单一券)',
  `each_coupon_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '单张券原始金额(分)',
  `pay_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '用户实付(分)',
  `actual_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '实收=实付+平台补贴(分)',
  `discount_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '优惠总金额(分)',
  `merchant_discount_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '商家优惠金额(分)',
  `buyer_phone` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '买家手机号(平台加密串)',
  `pay_time_unix` INT NOT NULL DEFAULT 0 COMMENT '支付时间(秒),支付后创单模式必传',
  `order_source` TINYINT NOT NULL DEFAULT 0 COMMENT '订单来源:1抖音2抖省省3豆包',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态:1已创建2已取消3已退款',
  `extra` JSON NULL COMMENT '非核心字段原始报文(pay_info/commerce_info/item_list/product_snap_shot/buyer_info/departure_info等)',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_dpo_order_id` (`order_id`),
  KEY `idx_dpo_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='抖音预售订单';

-- 预约订单（SPI: travel_spot.order.create_order，biz_type 3012）
CREATE TABLE IF NOT EXISTS `douyin_book_orders` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id` VARCHAR(64) NOT NULL COMMENT '抖音侧预约订单号',
  `order_out_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '我方订单号(雪花ID,创单成功必返)',
  `biz_type` INT NOT NULL DEFAULT 3012 COMMENT '业务类型:3012旅行社预约单',
  `source_order_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '归属的预售订单ID',
  `presale_coupon_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '预售券ID',
  `create_order_time_unix` INT NOT NULL DEFAULT 0 COMMENT '创单时间(秒级时间戳)',
  `pay_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '用户实付(分)',
  `actual_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '实收(分)',
  `original_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '订单原始金额(分)',
  `discount_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '优惠总金额(分)',
  `merchant_discount_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '商家优惠金额(分)',
  `pay_time_unix` INT NOT NULL DEFAULT 0 COMMENT '支付时间(秒),支付后创单模式必传',
  `order_source` TINYINT NOT NULL DEFAULT 0 COMMENT '订单来源:1抖音2抖省省3豆包',
  `book_start_date` DATE DEFAULT NULL COMMENT '预约开始日期(yyyy-MM-dd)',
  `book_end_date` DATE DEFAULT NULL COMMENT '预约结束日期(yyyy-MM-dd)',
  `occupant_count` INT NOT NULL DEFAULT 0 COMMENT '入住人数(book_info.occupancies数量)',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态:1已创建2已取消3已退款',
  `extra` JSON NULL COMMENT '非核心字段原始报文(book_info详情/buyer_info/item_list/remark_from_guest/total_booking_count等)',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_dbo_order_id` (`order_id`),
  KEY `idx_dbo_source_order_id` (`source_order_id`),
  KEY `idx_dbo_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='抖音预约订单';

-- 取消/退款通知流水（SPI: travel_spot.order.cancel_apply / travel_spot.order.refund_notify）
CREATE TABLE IF NOT EXISTS `douyin_order_notices` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `dedup_key` VARCHAR(128) NOT NULL COMMENT '幂等键:cancel:{order_id}:{cancel_order_time_unix} / refund:{order_id}:{after_sale_id}:{refund_time_unix}',
  `notice_type` TINYINT NOT NULL COMMENT '通知类型:1取消通知2退款通知',
  `order_id` VARCHAR(64) NOT NULL COMMENT '抖音侧订单号',
  `order_out_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '第三方订单ID',
  `biz_type` INT NOT NULL DEFAULT 0 COMMENT '取消通知携带:3011预售券3012预约单',
  `sub_type` TINYINT NOT NULL DEFAULT 0 COMMENT '取消:1支付前2支付后3外部原因;退款:1订单退款2补差价退款',
  `notify_time_unix` INT NOT NULL DEFAULT 0 COMMENT '取消:cancel_order_time_unix;退款:refund_time_unix(秒)',
  `pay_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '订单实付金额(分,退款通知)',
  `refund_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '实际退款金额(分,退款通知)',
  `after_sale_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '售后单ID(退款通知)',
  `extra` JSON NULL COMMENT '非核心字段原始报文(cancel_reason/refund_count/refund_item_list/user_refund_amount等)',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_don_dedup_key` (`dedup_key`),
  KEY `idx_don_order_id` (`order_id`),
  KEY `idx_don_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='抖音订单取消/退款通知流水';