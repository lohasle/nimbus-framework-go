package infra

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
	"gorm.io/gorm"
)

func (h *Handler) fileQuery(c *gin.Context) *gorm.DB {
	query := h.db.Model(&FileRecord{}).Where("tenant_id = ?", tenantID(c))
	if name := strings.TrimSpace(c.Query("name")); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if path := strings.TrimSpace(c.Query("path")); path != "" {
		query = query.Where("path LIKE ?", "%"+path+"%")
	}
	if typ := strings.TrimSpace(c.Query("type")); typ != "" {
		query = query.Where("content_type LIKE ?", "%"+typ+"%")
	}
	return query
}

// FilePage godoc
// @Summary Page uploaded files
// @Tags Infra File
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/file/page [get]
func (h *Handler) FilePage(c *gin.Context) {
	query := h.fileQuery(c)
	var total int64
	query.Count(&total)
	pn, ps := page(c)
	var rows []FileRecord
	query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

func (h *Handler) masterFileConfig(tenant uint64) (FileConfig, string, error) {
	var config FileConfig
	err := h.db.Where("tenant_id = ? AND master = ?", tenant, true).First(&config).Error
	if err != nil {
		return config, "", err
	}
	var settings struct {
		BasePath string `json:"basePath"`
	}
	_ = json.Unmarshal([]byte(config.Config), &settings)
	if settings.BasePath == "" {
		settings.BasePath = "./data/uploads"
	}
	absolute, err := filepath.Abs(settings.BasePath)
	return config, absolute, err
}
func safeUploadPath(directory, name string) (string, error) {
	directory = strings.Trim(strings.ReplaceAll(directory, "\\", "/"), "/")
	if directory == "." || strings.Contains(directory, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return "", fmt.Errorf("文件路径不合法")
	}
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(filepath.Base(name), ext)
	return filepath.Join(directory, time.Now().Format("20060102"), base+"-"+uuid.NewString()+ext), nil
}
func absoluteFileURL(c *gin.Context, id uint64) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if forwarded := c.GetHeader("X-Forwarded-Proto"); forwarded != "" {
		scheme = forwarded
	}
	return fmt.Sprintf("%s://%s/admin-api/infra/file/content/%d", scheme, c.Request.Host, id)
}

// FileUpload godoc
// @Summary Upload a file to the primary storage
// @Tags Infra File
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "File"
// @Param path formData string false "Original path"
// @Success 200 {object} httpx.Response
// @Router /infra/file/upload [post]
func (h *Handler) FileUpload(c *gin.Context) {
	header, err := c.FormFile("file")
	if err != nil {
		httpx.Fail(c, 400, 400, "请选择上传文件")
		return
	}
	config, base, err := h.masterFileConfig(tenantID(c))
	if err != nil {
		httpx.Fail(c, 500, 500, "未配置主文件存储")
		return
	}
	relative, err := safeUploadPath(c.PostForm("directory"), header.Filename)
	if err != nil {
		httpx.Fail(c, 400, 400, err.Error())
		return
	}
	target := filepath.Join(base, relative)
	if !strings.HasPrefix(target, base+string(os.PathSeparator)) {
		httpx.Fail(c, 400, 400, "文件路径不合法")
		return
	}
	if err = os.MkdirAll(filepath.Dir(target), 0750); err != nil {
		httpx.Fail(c, 500, 500, "创建上传目录失败")
		return
	}
	source, err := header.Open()
	if err != nil {
		httpx.Fail(c, 400, 400, "读取上传文件失败")
		return
	}
	defer source.Close()
	destination, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0640)
	if err != nil {
		httpx.Fail(c, 500, 500, "保存文件失败")
		return
	}
	size, copyErr := io.Copy(destination, source)
	closeErr := destination.Close()
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(target)
		httpx.Fail(c, 500, 500, "保存文件失败")
		return
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(header.Filename))
	}
	row := FileRecord{TenantID: tenantID(c), ConfigID: config.ID, Name: header.Filename, Path: filepath.ToSlash(relative), ContentType: contentType, Size: size}
	if err = h.db.Create(&row).Error; err != nil {
		_ = os.Remove(target)
		httpx.Fail(c, 500, 500, "记录文件失败")
		return
	}
	row.URL = absoluteFileURL(c, row.ID)
	h.db.Model(&row).Update("url", row.URL)
	httpx.OK(c, row.URL)
}

// FileContent godoc
// @Summary Download uploaded file content
// @Tags Infra File
// @Produce application/octet-stream
// @Success 200 {file} file
// @Router /infra/file/content/{id} [get]
func (h *Handler) FileContent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var row FileRecord
	if h.db.First(&row, id).Error != nil {
		c.Status(http.StatusNotFound)
		return
	}
	var config FileConfig
	if h.db.First(&config, row.ConfigID).Error != nil {
		c.Status(http.StatusNotFound)
		return
	}
	var settings struct {
		BasePath string `json:"basePath"`
	}
	_ = json.Unmarshal([]byte(config.Config), &settings)
	if settings.BasePath == "" {
		settings.BasePath = "./data/uploads"
	}
	base, _ := filepath.Abs(settings.BasePath)
	target := filepath.Join(base, filepath.FromSlash(row.Path))
	if !strings.HasPrefix(target, base+string(os.PathSeparator)) {
		c.Status(http.StatusForbidden)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", row.Name))
	c.Header("Content-Type", row.ContentType)
	c.File(target)
}

type FileCreateRequest struct {
	ConfigID uint64 `json:"configId"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	URL      string `json:"url"`
	Type     string `json:"type"`
	Size     int64  `json:"size"`
}

// FileCreate godoc
// @Summary Register a file uploaded through a presigned URL
// @Tags Infra File
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body FileCreateRequest true "File metadata"
// @Success 200 {object} httpx.Response
// @Router /infra/file/create [post]
func (h *Handler) FileCreate(c *gin.Context) {
	var req FileCreateRequest
	if c.ShouldBindJSON(&req) != nil || req.Path == "" {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	if req.ConfigID == 0 {
		config, _, err := h.masterFileConfig(tenantID(c))
		if err != nil {
			httpx.Fail(c, 500, 500, "未配置主文件存储")
			return
		}
		req.ConfigID = config.ID
	}
	row := FileRecord{TenantID: tenantID(c), ConfigID: req.ConfigID, Name: req.Name, Path: req.Path, URL: req.URL, ContentType: req.Type, Size: req.Size}
	if h.db.Create(&row).Error != nil {
		httpx.Fail(c, 500, 500, "创建文件记录失败")
		return
	}
	httpx.OK(c, row.ID)
}

// FilePresignedURL godoc
// @Summary Create a local-storage upload URL and final file path
// @Tags Infra File
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/file/presigned-url [get]
func (h *Handler) FilePresignedURL(c *gin.Context) {
	config, _, err := h.masterFileConfig(tenantID(c))
	if err != nil {
		httpx.Fail(c, 500, 500, "未配置主文件存储")
		return
	}
	path, err := safeUploadPath(c.Query("directory"), c.Query("name"))
	if err != nil {
		httpx.Fail(c, 400, 400, err.Error())
		return
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	uploadURL := fmt.Sprintf("%s://%s/admin-api/infra/file/upload", scheme, c.Request.Host)
	httpx.OK(c, gin.H{"configId": config.ID, "uploadUrl": uploadURL, "url": "", "path": filepath.ToSlash(path)})
}

// FileDelete godoc
// @Summary Delete a file and its stored content
// @Tags Infra File
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/file/delete [delete]
func (h *Handler) FileDelete(c *gin.Context) { h.deleteFiles(c, []uint64{queryID(c)}) }

// FileDeleteList godoc
// @Summary Delete files in batch
// @Tags Infra File
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/file/delete-list [delete]
func (h *Handler) FileDeleteList(c *gin.Context) { h.deleteFiles(c, parseIDs(c.Query("ids"))) }
func (h *Handler) deleteFiles(c *gin.Context, ids []uint64) {
	var rows []FileRecord
	h.db.Where("tenant_id = ? AND id IN ?", tenantID(c), ids).Find(&rows)
	for _, row := range rows {
		var config FileConfig
		if h.db.First(&config, row.ConfigID).Error == nil {
			var settings struct {
				BasePath string `json:"basePath"`
			}
			_ = json.Unmarshal([]byte(config.Config), &settings)
			if settings.BasePath != "" {
				base, _ := filepath.Abs(settings.BasePath)
				target := filepath.Join(base, filepath.FromSlash(row.Path))
				if strings.HasPrefix(target, base+string(os.PathSeparator)) {
					_ = os.Remove(target)
				}
			}
		}
	}
	h.db.Where("tenant_id = ? AND id IN ?", tenantID(c), ids).Delete(&FileRecord{})
	httpx.OK(c, true)
}

func parseIDs(raw string) []uint64 {
	result := []uint64{}
	for _, item := range strings.Split(raw, ",") {
		id, err := strconv.ParseUint(strings.TrimSpace(item), 10, 64)
		if err == nil && id > 0 {
			result = append(result, id)
		}
	}
	return result
}
