package system

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/excelx"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
)

// UserExport godoc
// @Summary Export operations-console users
// @Tags System User
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /system/user/export-excel [get]
func (h *Handler) UserExport(c *gin.Context) {
	var users []AdminUser
	h.service.db.Where("tenant_id = ?", tenantIDFromContext(c)).Order("id").Find(&users)
	book := excelize.NewFile()
	sheet := "用户数据"
	book.SetSheetName("Sheet1", sheet)
	headers := []string{"用户编号", "用户名称", "用户昵称", "部门编号", "手机号码", "邮箱", "性别", "状态", "创建时间"}
	for column, value := range headers {
		cell, _ := excelize.CoordinatesToCellName(column+1, 1)
		_ = book.SetCellValue(sheet, cell, value)
	}
	for index, user := range users {
		values := []any{user.ID, user.Username, user.Nickname, user.DeptID, user.Mobile, user.Email, user.Sex, user.Status, user.CreatedAt.Format(time.DateTime)}
		for column, value := range values {
			cell, _ := excelize.CoordinatesToCellName(column+1, index+2)
			_ = book.SetCellValue(sheet, cell, value)
		}
	}
	excelx.Write(c, book, "用户数据.xlsx")
}

// UserImportTemplate godoc
// @Summary Download user import template
// @Tags System User
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /system/user/get-import-template [get]
func (h *Handler) UserImportTemplate(c *gin.Context) {
	book := excelize.NewFile()
	sheet := "用户导入模板"
	book.SetSheetName("Sheet1", sheet)
	headers := []string{"用户名称*", "用户昵称*", "部门编号", "手机号码", "邮箱", "性别(0未知/1男/2女)", "状态(0启用/1停用)", "初始密码"}
	example := []any{"demo", "演示用户", 1, "13800000001", "demo@nimbus.local", 0, 0, "123456"}
	for column, value := range headers {
		cell, _ := excelize.CoordinatesToCellName(column+1, 1)
		_ = book.SetCellValue(sheet, cell, value)
	}
	for column, value := range example {
		cell, _ := excelize.CoordinatesToCellName(column+1, 2)
		_ = book.SetCellValue(sheet, cell, value)
	}
	excelx.Write(c, book, "用户导入模板.xlsx")
}

type UserImportResult struct {
	CreateUsernames  []string          `json:"createUsernames"`
	UpdateUsernames  []string          `json:"updateUsernames"`
	FailureUsernames map[string]string `json:"failureUsernames"`
}

// UserImport godoc
// @Summary Import operations-console users from XLSX
// @Tags System User
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "XLSX file"
// @Param updateSupport query bool false "Update existing users"
// @Success 200 {object} httpx.Response
// @Router /system/user/import [post]
func (h *Handler) UserImport(c *gin.Context) {
	result := UserImportResult{CreateUsernames: []string{}, UpdateUsernames: []string{}, FailureUsernames: map[string]string{}}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "请选择导入文件")
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "无法读取导入文件")
		return
	}
	defer file.Close()
	book, err := excelize.OpenReader(file)
	if err != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "仅支持有效的 XLSX 文件")
		return
	}
	defer book.Close()
	sheets := book.GetSheetList()
	if len(sheets) == 0 {
		httpx.Fail(c, http.StatusBadRequest, 400, "导入文件没有工作表")
		return
	}
	rows, err := book.GetRows(sheets[0])
	if err != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "读取工作表失败")
		return
	}
	updateSupport := c.Query("updateSupport") == "1" || strings.EqualFold(c.Query("updateSupport"), "true")
	tenantID := tenantIDFromContext(c)
	for index, row := range rows {
		if index == 0 || rowBlank(row) {
			continue
		}
		username, nickname := cell(row, 0), cell(row, 1)
		if username == "" || nickname == "" {
			result.FailureUsernames[displayUsername(username, index)] = "用户名称和用户昵称不能为空"
			continue
		}
		var user AdminUser
		found := h.service.db.Where("tenant_id = ? AND username = ?", tenantID, username).First(&user).Error == nil
		if found && !updateSupport {
			result.FailureUsernames[username] = "用户已存在"
			continue
		}
		deptID := parseCellUint(cell(row, 2), 1)
		user.TenantID, user.Username, user.Nickname, user.DeptID = tenantID, username, nickname, deptID
		user.Mobile, user.Email = cell(row, 3), cell(row, 4)
		user.Sex, user.Status = int(parseCellUint(cell(row, 5), 0)), int(parseCellUint(cell(row, 6), 0))
		if !found {
			password := cell(row, 7)
			if password == "" {
				password = "123456"
			}
			hash, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if hashErr != nil {
				result.FailureUsernames[username] = "密码处理失败"
				continue
			}
			user.PasswordHash, user.LoginDate = string(hash), time.Now()
			if err = h.service.db.Create(&user).Error; err != nil {
				result.FailureUsernames[username] = "创建失败"
				continue
			}
			result.CreateUsernames = append(result.CreateUsernames, username)
			continue
		}
		if err = h.service.db.Save(&user).Error; err != nil {
			result.FailureUsernames[username] = "更新失败"
			continue
		}
		result.UpdateUsernames = append(result.UpdateUsernames, username)
	}
	httpx.OK(c, result)
}

func cell(row []string, index int) string {
	if index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func rowBlank(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func parseCellUint(value string, fallback uint64) uint64 {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func displayUsername(username string, rowIndex int) string {
	if username != "" {
		return username
	}
	return fmt.Sprintf("第 %d 行", rowIndex+1)
}
