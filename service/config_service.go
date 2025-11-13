package service

import (
	"bookadmin/global"
	"bookadmin/model"
	"strconv"
	"sync"

	"go.uber.org/zap"
)

type ConfigService struct {
	configCache sync.Map // 配置缓存
}

// GetConfig 获取配置值
func (s *ConfigService) GetConfig(key string, defaultValue string) string {
	// 先从缓存中获取
	if value, ok := s.configCache.Load(key); ok {
		return value.(string)
	}

	// 从数据库获取
	var config model.SystemConfig
	if err := global.GVA_DB.Where("config_key = ?", key).First(&config).Error; err != nil {
		global.GVA_LOG.Warn("获取配置失败，使用默认值", zap.String("key", key), zap.Error(err))
		return defaultValue
	}

	// 存入缓存
	s.configCache.Store(key, config.ConfigValue)
	return config.ConfigValue
}

// GetIntConfig 获取整型配置
func (s *ConfigService) GetIntConfig(key string, defaultValue int) int {
	value := s.GetConfig(key, strconv.Itoa(defaultValue))
	intValue, err := strconv.Atoi(value)
	if err != nil {
		global.GVA_LOG.Warn("配置值转换失败，使用默认值", zap.String("key", key), zap.Error(err))
		return defaultValue
	}
	return intValue
}

// GetFloatConfig 获取浮点型配置
func (s *ConfigService) GetFloatConfig(key string, defaultValue float64) float64 {
	value := s.GetConfig(key, strconv.FormatFloat(defaultValue, 'f', -1, 64))
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		global.GVA_LOG.Warn("配置值转换失败，使用默认值", zap.String("key", key), zap.Error(err))
		return defaultValue
	}
	return floatValue
}

// GetBoolConfig 获取布尔型配置
func (s *ConfigService) GetBoolConfig(key string, defaultValue bool) bool {
	value := s.GetConfig(key, strconv.FormatBool(defaultValue))
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		global.GVA_LOG.Warn("配置值转换失败，使用默认值", zap.String("key", key), zap.Error(err))
		return defaultValue
	}
	return boolValue
}

// SetConfig 设置配置值
func (s *ConfigService) SetConfig(key, value string) error {
	var config model.SystemConfig
	result := global.GVA_DB.Where("config_key = ?", key).First(&config)

	if result.Error != nil {
		// 不存在，创建新配置
		config = model.SystemConfig{
			ConfigKey:   key,
			ConfigValue: value,
		}
		if err := global.GVA_DB.Create(&config).Error; err != nil {
			return err
		}
	} else {
		// 存在，更新配置
		if err := global.GVA_DB.Model(&config).Update("config_value", value).Error; err != nil {
			return err
		}
	}

	// 更新缓存
	s.configCache.Store(key, value)
	return nil
}

// RefreshCache 刷新配置缓存
func (s *ConfigService) RefreshCache() error {
	var configs []model.SystemConfig
	if err := global.GVA_DB.Find(&configs).Error; err != nil {
		return err
	}

	// 清空并重新加载缓存
	s.configCache = sync.Map{}
	for _, config := range configs {
		s.configCache.Store(config.ConfigKey, config.ConfigValue)
	}

	global.GVA_LOG.Info("配置缓存已刷新", zap.Int("count", len(configs)))
	return nil
}

// 全局配置服务实例
var GlobalConfigService = &ConfigService{}
