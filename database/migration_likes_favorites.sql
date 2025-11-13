-- ================================================
-- 图书点赞和收藏功能数据库迁移脚本
-- 创建时间: 2025-11-10
-- 功能: 添加点赞表、收藏表，扩展books表字段
-- ================================================

USE bookadmin;

-- ================================================
-- 1. 创建点赞表 book_likes
-- ================================================
CREATE TABLE IF NOT EXISTS `book_likes` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `book_id` BIGINT UNSIGNED NOT NULL COMMENT '图书ID',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '点赞时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_book` (`user_id`, `book_id`) COMMENT '用户-图书唯一索引，防止重复点赞',
  KEY `idx_user_id` (`user_id`) COMMENT '用户ID索引',
  KEY `idx_book_id` (`book_id`) COMMENT '图书ID索引',
  KEY `idx_created_at` (`created_at`) COMMENT '创建时间索引，用于榜单统计'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图书点赞表';

-- ================================================
-- 2. 创建收藏表 book_favorites
-- ================================================
CREATE TABLE IF NOT EXISTS `book_favorites` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `book_id` BIGINT UNSIGNED NOT NULL COMMENT '图书ID',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '收藏时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_book` (`user_id`, `book_id`) COMMENT '用户-图书唯一索引，防止重复收藏',
  KEY `idx_user_id` (`user_id`) COMMENT '用户ID索引',
  KEY `idx_book_id` (`book_id`) COMMENT '图书ID索引',
  KEY `idx_created_at` (`created_at`) COMMENT '创建时间索引，用于榜单统计'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图书收藏表';

-- ================================================
-- 3. 扩展 books 表，添加统计字段
-- ================================================
-- 检查并添加 like_count 字段
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
               WHERE TABLE_SCHEMA = 'bookadmin' 
               AND TABLE_NAME = 'books' 
               AND COLUMN_NAME = 'like_count');

SET @sql := IF(@exist > 0, 
    'SELECT "like_count字段已存在" AS message',
    'ALTER TABLE `books` ADD COLUMN `like_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT ''点赞总数'' AFTER `description`'
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 检查并添加 favorite_count 字段
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
               WHERE TABLE_SCHEMA = 'bookadmin' 
               AND TABLE_NAME = 'books' 
               AND COLUMN_NAME = 'favorite_count');

SET @sql := IF(@exist > 0, 
    'SELECT "favorite_count字段已存在" AS message',
    'ALTER TABLE `books` ADD COLUMN `favorite_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT ''收藏总数'' AFTER `like_count`'
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 添加统计字段索引（如果不存在）
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
               WHERE TABLE_SCHEMA = 'bookadmin' 
               AND TABLE_NAME = 'books' 
               AND INDEX_NAME = 'idx_like_count');

SET @sql := IF(@exist > 0, 
    'SELECT "idx_like_count索引已存在" AS message',
    'ALTER TABLE `books` ADD INDEX `idx_like_count` (`like_count`)'
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
               WHERE TABLE_SCHEMA = 'bookadmin' 
               AND TABLE_NAME = 'books' 
               AND INDEX_NAME = 'idx_favorite_count');

SET @sql := IF(@exist > 0, 
    'SELECT "idx_favorite_count索引已存在" AS message',
    'ALTER TABLE `books` ADD INDEX `idx_favorite_count` (`favorite_count`)'
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ================================================
-- 4. 验证表结构
-- ================================================
-- 查看点赞表结构
SHOW CREATE TABLE book_likes;

-- 查看收藏表结构
SHOW CREATE TABLE book_favorites;

-- 查看books表新增字段
SHOW COLUMNS FROM books WHERE Field IN ('like_count', 'favorite_count');

-- ================================================
-- 说明：
-- 1. 使用 BIGINT UNSIGNED 存储ID，支持大数据量
-- 2. 唯一索引 uk_user_book 防止重复点赞/收藏
-- 3. created_at 索引用于榜单统计查询优化
-- 4. books表的冗余字段用于快速查询，避免JOIN
-- 5. 使用动态SQL检查字段是否存在，避免重复执行报错
-- ================================================

