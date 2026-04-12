CREATE DATABASE IF NOT EXISTS lottery_system
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

CREATE DATABASE IF NOT EXISTS bitstorm
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE lottery_system;

CREATE TABLE IF NOT EXISTS `t_user_info` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `user_name` varchar(65) NOT NULL COMMENT '用户名字',
  `pwd` varchar(65) NOT NULL COMMENT '用户密码',
  `sex` int(11) NOT NULL COMMENT '性别',
  `age` int(11) NOT NULL COMMENT '年龄',
  `email` varchar(128) DEFAULT NULL COMMENT '邮箱',
  `contact` varchar(128) DEFAULT NULL COMMENT '联系地址',
  `mobile` varchar(64) NOT NULL COMMENT '手机号',
  `id_card` varchar(64) NOT NULL COMMENT '证件号',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `modify_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`user_name`),
  UNIQUE KEY `idx_mobile` (`mobile`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户信息表';

INSERT IGNORE INTO `t_user_info`
  (`id`, `user_name`, `pwd`, `age`, `sex`, `mobile`, `id_card`)
VALUES
  (1, 'admin', '123321', 18, 1, '18676662555', '5106811222223333');

USE bitstorm;

CREATE TABLE IF NOT EXISTS `t_seckill_stock` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `goods_id` bigint(20) DEFAULT NULL COMMENT '商品ID',
  `stock` int(11) NOT NULL COMMENT '库存大小',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `modify_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_goodsid` (`goods_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='秒杀库存表';

CREATE TABLE IF NOT EXISTS `t_seckill_record` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `goods_id` bigint(20) NOT NULL COMMENT '商品ID',
  `sec_num` varchar(128) DEFAULT NULL COMMENT '秒杀号',
  `order_num` varchar(128) DEFAULT NULL COMMENT '订单号',
  `price` int(11) NOT NULL COMMENT '金额',
  `status` int(11) NOT NULL COMMENT '状态',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `modify_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_secnum` (`sec_num`),
  UNIQUE KEY `idx_ordernum` (`order_num`),
  KEY `idx_userid` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='秒杀记录表';

CREATE TABLE IF NOT EXISTS `t_goods` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `goods_num` varchar(128) DEFAULT NULL COMMENT '商品编号',
  `goods_name` varchar(128) DEFAULT NULL COMMENT '商品名字',
  `price` float NOT NULL COMMENT '价格',
  `pic_url` varchar(128) DEFAULT NULL COMMENT '商品图片',
  `seller` bigint(20) NOT NULL COMMENT '卖家ID',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `modify_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_goodsnum` (`goods_num`),
  KEY `idx_seller` (`seller`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品表';

CREATE TABLE IF NOT EXISTS `t_order` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `seller` bigint(20) NOT NULL COMMENT '买方ID',
  `buyer` bigint(20) NOT NULL COMMENT '卖房ID',
  `goods_id` bigint(20) NOT NULL COMMENT '商品ID',
  `goods_num` varchar(128) DEFAULT NULL COMMENT '商品编号',
  `order_num` varchar(128) DEFAULT NULL COMMENT '订单号',
  `price` int(11) NOT NULL COMMENT '金额',
  `status` int(11) NOT NULL COMMENT '状态',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `modify_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ordernum` (`order_num`),
  KEY `idx_goodsid` (`goods_id`),
  KEY `idx_seller` (`seller`),
  KEY `idx_buyer` (`buyer`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单表';

CREATE TABLE IF NOT EXISTS `t_quota` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `goods_id` bigint(20) DEFAULT NULL COMMENT '商品ID',
  `num` int(11) NOT NULL COMMENT '限额',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `modify_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_goodsid` (`goods_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='限额表';

CREATE TABLE IF NOT EXISTS `t_user_quota` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `goods_id` bigint(20) DEFAULT NULL COMMENT '商品ID',
  `num` int(11) NOT NULL COMMENT '限额',
  `killed_num` int(11) NOT NULL COMMENT '已经消耗的额度',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `modify_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  KEY `idx_goodsid` (`goods_id`),
  KEY `idx_usergoodsid` (`user_id`, `goods_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户限额表';

INSERT IGNORE INTO `t_goods`
  (`id`, `goods_name`, `goods_num`, `price`, `pic_url`, `seller`)
VALUES
  (1, 'redhat', 'abc123', 18, 'http://', 135);

INSERT IGNORE INTO `t_seckill_stock`
  (`id`, `goods_id`, `stock`)
VALUES
  (1, 1, 3);

CREATE TABLE IF NOT EXISTS `t_seckill_async_result` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `sec_num` varchar(128) NOT NULL COMMENT '秒杀号',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `goods_id` bigint(20) NOT NULL COMMENT '商品ID',
  `goods_num` varchar(128) DEFAULT NULL COMMENT '商品编号',
  `order_num` varchar(128) DEFAULT NULL COMMENT '订单号',
  `status` int(11) NOT NULL COMMENT '状态: 1=处理中, 2=成功, 6=失败',
  `reason` varchar(512) DEFAULT NULL COMMENT '失败原因',
  `attempt` int(11) NOT NULL DEFAULT 0 COMMENT '重试次数',
  `last_error` varchar(512) DEFAULT NULL COMMENT '最后错误信息',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `modify_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_secnum` (`sec_num`),
  KEY `idx_userid` (`user_id`),
  KEY `idx_goodsid` (`goods_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='秒杀异步结果表';
