-- 注意：此文件用于 Docker MySQL 容器初始化
-- Docker Compose 会在容器首次启动时自动执行此文件
-- 数据库 bookadmin 已通过 docker-compose.yml 的 MYSQL_DATABASE 环境变量自动创建

-- 使用数据库
USE bookadmin;

-- 注意：表结构将由 GORM 自动创建，无需手动创建
-- 如果需要手动创建，可以使用以下 SQL：

-- CREATE TABLE IF NOT EXISTS `books` (
--   `id` bigint unsigned NOT NULL AUTO_INCREMENT,
--   `created_at` datetime(3) DEFAULT NULL,
--   `updated_at` datetime(3) DEFAULT NULL,
--   `deleted_at` datetime(3) DEFAULT NULL,
--   `title` varchar(191) NOT NULL COMMENT '书名',
--   `author` varchar(191) NOT NULL COMMENT '作者',
--   `publisher` varchar(191) DEFAULT NULL COMMENT '出版社',
--   `publish_date` varchar(191) DEFAULT NULL COMMENT '出版日期',
--   `isbn` varchar(191) DEFAULT NULL COMMENT 'ISBN',
--   `price` double DEFAULT NULL COMMENT '价格',
--   `description` text COMMENT '描述',
--   PRIMARY KEY (`id`),
--   UNIQUE KEY `idx_books_isbn` (`isbn`),
--   KEY `idx_books_deleted_at` (`deleted_at`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

