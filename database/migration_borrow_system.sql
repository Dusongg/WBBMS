-- 借阅系统完善 - 数据库迁移脚本

-- 1. 创建预约表
CREATE TABLE IF NOT EXISTS `reservations` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  `reader_id` BIGINT UNSIGNED NOT NULL COMMENT '读者ID',
  `book_id` BIGINT UNSIGNED NOT NULL COMMENT '图书ID',
  `status` ENUM('pending', 'available', 'fulfilled', 'cancelled', 'expired') NOT NULL DEFAULT 'pending' COMMENT '状态',
  `reserve_date` DATETIME(3) NULL COMMENT '预约日期',
  `available_date` DATETIME(3) NULL COMMENT '可取书日期',
  `pickup_deadline` DATETIME(3) NULL COMMENT '取书截止日期',
  `fulfilled_date` DATETIME(3) NULL COMMENT '完成日期',
  `queue_position` INT NULL COMMENT '队列位置',
  `remark` TEXT NULL COMMENT '备注',
  PRIMARY KEY (`id`),
  INDEX `idx_reservations_deleted_at` (`deleted_at`),
  INDEX `idx_reader_book` (`reader_id`, `book_id`),
  INDEX `idx_reservations_reader_id` (`reader_id`),
  INDEX `idx_reservations_book_id` (`book_id`),
  INDEX `idx_reservations_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图书预约表';

-- 2. 创建罚款记录表
CREATE TABLE IF NOT EXISTS `fine_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  `reader_id` BIGINT UNSIGNED NOT NULL COMMENT '读者ID',
  `borrow_record_id` BIGINT UNSIGNED NULL COMMENT '借阅记录ID',
  `fine_type` VARCHAR(50) NULL COMMENT '罚款类型',
  `amount` DECIMAL(10,2) NOT NULL COMMENT '罚款金额',
  `paid_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '已支付金额',
  `status` ENUM('unpaid', 'paid', 'waived') NOT NULL DEFAULT 'unpaid' COMMENT '状态',
  `overdue_days` INT NULL COMMENT '逾期天数',
  `fine_date` DATETIME(3) NULL COMMENT '罚款日期',
  `paid_date` DATETIME(3) NULL COMMENT '支付日期',
  `operator_id` BIGINT UNSIGNED NULL COMMENT '操作员ID',
  `remark` TEXT NULL COMMENT '备注',
  PRIMARY KEY (`id`),
  INDEX `idx_fine_records_deleted_at` (`deleted_at`),
  INDEX `idx_fine_records_reader_id` (`reader_id`),
  INDEX `idx_fine_records_borrow_record_id` (`borrow_record_id`),
  INDEX `idx_fine_records_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='罚款记录表';

-- 3. 创建黑名单表
CREATE TABLE IF NOT EXISTS `blacklists` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  `reader_id` BIGINT UNSIGNED NOT NULL COMMENT '读者ID',
  `reason` VARCHAR(50) NULL COMMENT '拉黑原因',
  `description` TEXT NULL COMMENT '详细描述',
  `status` ENUM('active', 'lifted', 'expired') NOT NULL DEFAULT 'active' COMMENT '状态',
  `start_date` DATETIME(3) NULL COMMENT '开始日期',
  `end_date` DATETIME(3) NULL COMMENT '结束日期',
  `lifted_date` DATETIME(3) NULL COMMENT '解除日期',
  `operator_id` BIGINT UNSIGNED NULL COMMENT '操作员ID',
  `remark` TEXT NULL COMMENT '备注',
  PRIMARY KEY (`id`),
  INDEX `idx_blacklists_deleted_at` (`deleted_at`),
  INDEX `idx_blacklists_reader_id` (`reader_id`),
  INDEX `idx_blacklists_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='黑名单表';

-- 4. 创建系统配置表
CREATE TABLE IF NOT EXISTS `system_configs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  `config_key` VARCHAR(100) NOT NULL COMMENT '配置键',
  `config_value` TEXT NULL COMMENT '配置值',
  `description` TEXT NULL COMMENT '配置描述',
  `config_type` VARCHAR(50) NULL COMMENT '配置类型',
  `is_system` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否系统配置',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_system_configs_config_key` (`config_key`),
  INDEX `idx_system_configs_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- 5. 修改 borrow_records 表 - 添加新字段
ALTER TABLE `borrow_records` 
ADD COLUMN `max_renew_count` INT NOT NULL DEFAULT 2 COMMENT '最大续借次数' AFTER `renew_count`,
ADD COLUMN `overdue_days` INT NOT NULL DEFAULT 0 COMMENT '逾期天数' AFTER `fine_amount`,
ADD COLUMN `reservation_id` BIGINT UNSIGNED NULL COMMENT '预约ID' AFTER `overdue_days`,
ADD INDEX `idx_borrow_records_reader_id` (`reader_id`),
ADD INDEX `idx_borrow_records_book_id` (`book_id`),
ADD INDEX `idx_borrow_records_due_date` (`due_date`),
ADD INDEX `idx_borrow_records_status` (`status`);

-- 6. 修改 readers 表 - 添加新字段
ALTER TABLE `readers`
ADD COLUMN `max_renew` INT NOT NULL DEFAULT 2 COMMENT '最大续借次数' AFTER `borrow_days`,
ADD COLUMN `renew_days` INT NOT NULL DEFAULT 15 COMMENT '每次续借延长天数' AFTER `max_renew`,
ADD COLUMN `max_reservations` INT NOT NULL DEFAULT 3 COMMENT '最大预约数量' AFTER `renew_days`,
ADD COLUMN `total_fine` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '累计罚款' AFTER `max_reservations`,
ADD COLUMN `unpaid_fine` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '未支付罚款' AFTER `total_fine`,
ADD COLUMN `is_blacklisted` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否在黑名单' AFTER `unpaid_fine`;

-- 7. 插入默认系统配置
INSERT INTO `system_configs` (`config_key`, `config_value`, `description`, `config_type`, `is_system`) VALUES
('max_borrow_books', '5', '每个用户最多可借图书数量', 'int', 1),
('borrow_days', '30', '图书借阅期限（天）', 'int', 1),
('max_renew_times', '2', '最大续借次数', 'int', 1),
('renew_days', '15', '每次续借延长天数', 'int', 1),
('max_reservations', '3', '每个用户最多预约数量', 'int', 1),
('reservation_pickup_days', '3', '预约取书有效期（天）', 'int', 1),
('overdue_fine_per_day', '0.5', '逾期罚款（元/天）', 'float', 1),
('max_fine_rate', '0.5', '罚款上限比例（图书价格的百分比）', 'float', 1),
('overdue_reminder_days', '3', '到期前提前提醒天数', 'int', 1),
('overdue_block_days', '7', '逾期多久后禁止借书（天）', 'int', 1),
('overdue_blacklist_days', '30', '逾期多久后自动拉黑（天）', 'int', 1)
ON DUPLICATE KEY UPDATE `config_value` = VALUES(`config_value`);

