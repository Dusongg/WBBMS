package service

import (
	"bookadmin/global"
	"bookadmin/model"
	"errors"
	"math"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FineService struct{}

// CalculateOverdueFine 计算逾期罚款
func (s *FineService) CalculateOverdueFine(borrowRecord *model.BorrowRecord) (float64, int, error) {
	// 计算逾期天数
	now := time.Now()
	if borrowRecord.DueDate.After(now) {
		// 还未到期
		return 0, 0, nil
	}

	overdueDays := int(now.Sub(borrowRecord.DueDate).Hours() / 24)
	if overdueDays <= 0 {
		return 0, 0, nil
	}

	// 获取罚款配置
	finePerDay := GlobalConfigService.GetFloatConfig(model.ConfigOverdueFinePerDay, 0.5)
	maxFineRate := GlobalConfigService.GetFloatConfig(model.ConfigMaxFineRate, 0.5)

	// 计算基础罚款
	fineAmount := float64(overdueDays) * finePerDay

	// 获取图书价格，计算罚款上限
	var book model.Book
	if err := global.GVA_DB.First(&book, borrowRecord.BookID).Error; err == nil {
		maxFine := book.Price * maxFineRate
		if fineAmount > maxFine {
			fineAmount = maxFine
		}
	}

	// 四舍五入到两位小数
	fineAmount = math.Round(fineAmount*100) / 100

	return fineAmount, overdueDays, nil
}

// CreateFinetRecord 创建罚款记录
func (s *FineService) CreateFineRecord(readerID uint, borrowRecordID uint, fineType string, amount float64, overdueDays int, operatorID uint) error {
	fine := model.FineRecord{
		ReaderID:       readerID,
		BorrowRecordID: borrowRecordID,
		FineType:       fineType,
		Amount:         amount,
		PaidAmount:     0,
		Status:         model.FineStatusUnpaid,
		OverdueDays:    overdueDays,
		FineDate:       time.Now(),
		OperatorID:     operatorID,
	}

	if err := global.GVA_DB.Create(&fine).Error; err != nil {
		global.GVA_LOG.Error("创建罚款记录失败", zap.Error(err))
		return errors.New("创建罚款记录失败")
	}

	// 更新读者的罚款金额
	if err := global.GVA_DB.Model(&model.Reader{}).Where("id = ?", readerID).
		UpdateColumn("unpaid_fine", gorm.Expr("unpaid_fine + ?", amount)).
		UpdateColumn("total_fine", gorm.Expr("total_fine + ?", amount)).
		Error; err != nil {
		global.GVA_LOG.Error("更新读者罚款金额失败", zap.Error(err))
	}

	global.GVA_LOG.Info("创建罚款记录成功", zap.Uint("reader_id", readerID), zap.Float64("amount", amount))
	return nil
}

// PayFine 支付罚款
func (s *FineService) PayFine(fineID uint, paidAmount float64, operatorID uint) error {
	tx := global.GVA_DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查找罚款记录
	var fine model.FineRecord
	if err := tx.First(&fine, fineID).Error; err != nil {
		tx.Rollback()
		return errors.New("罚款记录不存在")
	}

	if fine.Status == model.FineStatusPaid {
		tx.Rollback()
		return errors.New("罚款已支付")
	}

	if fine.Status == model.FineStatusWaived {
		tx.Rollback()
		return errors.New("罚款已豁免")
	}

	// 更新罚款记录
	now := time.Now()
	fine.PaidAmount += paidAmount
	fine.OperatorID = operatorID

	if fine.PaidAmount >= fine.Amount {
		fine.Status = model.FineStatusPaid
		fine.PaidDate = &now
	}

	if err := tx.Save(&fine).Error; err != nil {
		tx.Rollback()
		return errors.New("更新罚款记录失败")
	}

	// 更新读者的未支付罚款
	if err := tx.Model(&model.Reader{}).Where("id = ?", fine.ReaderID).
		UpdateColumn("unpaid_fine", gorm.Expr("unpaid_fine - ?", paidAmount)).
		Error; err != nil {
		tx.Rollback()
		return errors.New("更新读者罚款金额失败")
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("支付罚款失败")
	}

	global.GVA_LOG.Info("支付罚款成功", zap.Uint("fine_id", fineID), zap.Float64("amount", paidAmount))
	return nil
}

// WaiveFine 豁免罚款
func (s *FineService) WaiveFine(fineID uint, operatorID uint, remark string) error {
	tx := global.GVA_DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查找罚款记录
	var fine model.FineRecord
	if err := tx.First(&fine, fineID).Error; err != nil {
		tx.Rollback()
		return errors.New("罚款记录不存在")
	}

	if fine.Status == model.FineStatusPaid {
		tx.Rollback()
		return errors.New("罚款已支付，无法豁免")
	}

	if fine.Status == model.FineStatusWaived {
		tx.Rollback()
		return errors.New("罚款已豁免")
	}

	// 计算未支付的金额
	unpaidAmount := fine.Amount - fine.PaidAmount

	// 更新罚款记录
	fine.Status = model.FineStatusWaived
	fine.OperatorID = operatorID
	fine.Remark = remark

	if err := tx.Save(&fine).Error; err != nil {
		tx.Rollback()
		return errors.New("更新罚款记录失败")
	}

	// 更新读者的未支付罚款
	if unpaidAmount > 0 {
		if err := tx.Model(&model.Reader{}).Where("id = ?", fine.ReaderID).
			UpdateColumn("unpaid_fine", gorm.Expr("unpaid_fine - ?", unpaidAmount)).
			Error; err != nil {
			tx.Rollback()
			return errors.New("更新读者罚款金额失败")
		}
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("豁免罚款失败")
	}

	global.GVA_LOG.Info("豁免罚款成功", zap.Uint("fine_id", fineID))
	return nil
}

// GetReaderUnpaidFines 获取读者未支付的罚款列表
func (s *FineService) GetReaderUnpaidFines(readerID uint) ([]model.FineRecord, error) {
	var fines []model.FineRecord
	if err := global.GVA_DB.Where("reader_id = ? AND status = ?", readerID, model.FineStatusUnpaid).
		Preload("BorrowRecord").
		Preload("BorrowRecord.Book").
		Order("fine_date DESC").
		Find(&fines).Error; err != nil {
		return nil, err
	}
	return fines, nil
}
