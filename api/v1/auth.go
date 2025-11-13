package v1

import (
	"bookadmin/global"
	"bookadmin/model"
	"bookadmin/model/common/response"
	"bookadmin/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthApi struct{}

// Login 用户登录
func (a *AuthApi) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		global.GVA_LOG.Error("参数绑定失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	// 查找用户
	var user model.User
	if err := global.GVA_DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		global.GVA_LOG.Error("用户不存在", zap.Error(err))
		c.JSON(200, response.FailWithMessage("用户名或密码错误"))
		return
	}

	// 验证密码（简化处理，实际应使用bcrypt）
	if user.Password != req.Password {
		c.JSON(200, response.FailWithMessage("用户名或密码错误"))
		return
	}

	// 检查用户状态
	if user.Status != "active" {
		c.JSON(200, response.FailWithMessage("用户已被禁用"))
		return
	}

	// 生成token
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		global.GVA_LOG.Error("生成token失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("登录失败"))
		return
	}

	c.JSON(200, response.OkWithData(gin.H{
		"token":    token,
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"real_name": user.RealName,
	}))
}

// Register 用户注册（读者注册）
func (a *AuthApi) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		RealName string `json:"real_name" binding:"required"`
		IDCard   string `json:"id_card" binding:"required"`
		Address  string `json:"address"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		global.GVA_LOG.Error("参数绑定失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("参数错误"))
		return
	}

	// 检查用户名是否已存在
	var existUser model.User
	if err := global.GVA_DB.Where("username = ?", req.Username).First(&existUser).Error; err == nil {
		c.JSON(200, response.FailWithMessage("用户名已存在"))
		return
	}

	// 创建用户
	user := model.User{
		Username: req.Username,
		Password: req.Password, // 简化处理，实际应加密
		Email:    req.Email,
		Phone:    req.Phone,
		Role:     model.RoleReader,
		Status:   "active",
		RealName: req.RealName,
	}

	if err := global.GVA_DB.Create(&user).Error; err != nil {
		global.GVA_LOG.Error("创建用户失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("注册失败"))
		return
	}

	// 生成读者编号
	readerNo := "R" + strconv.FormatUint(uint64(user.ID+10000), 10)

	// 创建读者信息
	reader := model.Reader{
		UserID:     user.ID,
		ReaderNo:   readerNo,
		IDCard:     req.IDCard,
		Address:    req.Address,
		Status:     model.ReaderStatusPending, // 待审核
		MaxBorrow:  5,
		BorrowDays: 30,
	}

	if err := global.GVA_DB.Create(&reader).Error; err != nil {
		global.GVA_LOG.Error("创建读者失败", zap.Error(err))
		// 删除已创建的用户
		global.GVA_DB.Delete(&user)
		c.JSON(200, response.FailWithMessage("注册失败"))
		return
	}

	c.JSON(200, response.OkWithMessage("注册成功，等待审核"))
}

// GetUserInfo 获取当前用户信息
func (a *AuthApi) GetUserInfo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(200, response.FailWithMessage("未登录"))
		return
	}

	var user model.User
	if err := global.GVA_DB.Preload("Reader").First(&user, userID).Error; err != nil {
		global.GVA_LOG.Error("获取用户信息失败", zap.Error(err))
		c.JSON(200, response.FailWithMessage("获取用户信息失败"))
		return
	}

	c.JSON(200, response.OkWithData(user))
}

