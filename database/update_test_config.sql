-- 更新测试配置：借阅期限改为1分钟，到期前提醒改为30秒
-- 注意：配置值为0时，代码会自动使用测试模式（1分钟和30秒）

-- 更新借阅期限配置（0表示测试模式1分钟）
UPDATE system_configs 
SET config_value = '0', 
    description = '图书借阅期限（天，0表示测试模式1分钟）'
WHERE config_key = 'borrow_days';

-- 更新到期前提醒配置（0表示测试模式30秒）
UPDATE system_configs 
SET config_value = '0', 
    description = '到期前提前提醒天数（0表示测试模式30秒）'
WHERE config_key = 'overdue_reminder_days';

-- 如果配置不存在，插入配置
INSERT INTO system_configs (config_key, config_value, description, config_type, is_system, created_at, updated_at)
VALUES 
    ('borrow_days', '0', '图书借阅期限（天，0表示测试模式1分钟）', 'int', 1, NOW(), NOW()),
    ('overdue_reminder_days', '0', '到期前提前提醒天数（0表示测试模式30秒）', 'int', 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE 
    config_value = VALUES(config_value),
    description = VALUES(description),
    updated_at = NOW();

