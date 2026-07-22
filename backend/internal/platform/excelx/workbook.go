package excelx

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
	"github.com/xuri/excelize/v2"
)

func Write(c *gin.Context, book *excelize.File, filename string) {
	buffer, err := book.WriteToBuffer()
	if err != nil {
		httpx.Fail(c, http.StatusInternalServerError, 500, "生成 Excel 失败")
		return
	}
	encoded := url.PathEscape(filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", encoded))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buffer.Bytes())
}
