-- 更新生产环境配置：借阅期限改为30天，到期前提醒改为3天

-- 更新借阅期限配置（30天）
UPDATE system_configs 
SET config_value = '30', 
    description = '图书借阅期限（天）'
WHERE config_key = 'borrow_days';

-- 更新到期前提醒配置（3天）
UPDATE system_configs 
SET config_value = '3', 
    description = '到期前提前提醒天数'
WHERE config_key = 'overdue_reminder_days';

-- 如果配置不存在，插入配置
INSERT INTO system_configs (config_key, config_value, description, config_type, is_system, created_at, updated_at)
VALUES 
    ('borrow_days', '30', '图书借阅期限（天）', 'int', 1, NOW(), NOW()),
    ('overdue_reminder_days', '3', '到期前提前提醒天数', 'int', 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE 
    config_value = VALUES(config_value),
    description = VALUES(description),
    updated_at = NOW();

