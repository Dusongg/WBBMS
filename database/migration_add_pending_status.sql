-- 添加借阅审批功能的数据库迁移
-- 执行时间：请在执行前备份数据库

-- 1. 修改 borrow_records 表的 status 枚举，添加 pending 和 rejected 状态
ALTER TABLE `borrow_records` 
MODIFY COLUMN `status` 
ENUM('pending','borrowed','returned','overdue','renewed','reserved','rejected') 
DEFAULT 'pending' 
COMMENT '状态';

-- 注意：由于修改了默认值，现有的 borrowed 状态记录不会受影响
-- 新创建的记录默认状态为 pending（待批准）

