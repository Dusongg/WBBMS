package utils

import (
	"fmt"
	"time"
)

// GetCurrentWeekPeriod 获取当前周期标识（2025-W45）
func GetCurrentWeekPeriod() string {
	now := time.Now()
	year, week := now.ISOWeek()
	return fmt.Sprintf("%d-W%02d", year, week)
}

// GetCurrentMonthPeriod 获取当前月份标识（2025-11）
func GetCurrentMonthPeriod() string {
	now := time.Now()
	return fmt.Sprintf("%d-%02d", now.Year(), now.Month())
}

// GetWeekStartEnd 获取指定周期的开始和结束时间
func GetWeekStartEnd(period string) (time.Time, time.Time, error) {
	var year, week int
	_, err := fmt.Sscanf(period, "%d-W%d", &year, &week)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	// 计算该周的第一天（周一）
	jan1 := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	weekday := int(jan1.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	
	// ISO周从周一开始
	daysToMonday := (8 - weekday) % 7
	firstMonday := jan1.AddDate(0, 0, daysToMonday)
	
	// 加上周数
	weekStart := firstMonday.AddDate(0, 0, (week-1)*7)
	weekEnd := weekStart.AddDate(0, 0, 7).Add(-time.Second)

	return weekStart, weekEnd, nil
}

// GetMonthStartEnd 获取指定月份的开始和结束时间
func GetMonthStartEnd(period string) (time.Time, time.Time, error) {
	var year, month int
	_, err := fmt.Sscanf(period, "%d-%d", &year, &month)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

	return monthStart, monthEnd, nil
}

// GetLastWeekPeriod 获取上周周期
func GetLastWeekPeriod() string {
	now := time.Now().AddDate(0, 0, -7)
	year, week := now.ISOWeek()
	return fmt.Sprintf("%d-W%02d", year, week)
}

// GetLastMonthPeriod 获取上月周期
func GetLastMonthPeriod() string {
	now := time.Now().AddDate(0, -1, 0)
	return fmt.Sprintf("%d-%02d", now.Year(), now.Month())
}

