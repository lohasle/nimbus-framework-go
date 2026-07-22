package system

import (
	_ "embed"
	"encoding/csv"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
)

//go:embed area.csv
var areaCSV string

type AreaNode struct {
	ID       uint64      `json:"id"`
	Name     string      `json:"name"`
	Type     int         `json:"type"`
	ParentID uint64      `json:"parentId"`
	Children []*AreaNode `json:"children,omitempty"`
}

var (
	areaOnce sync.Once
	areaRoot *AreaNode
	areaErr  error
)

// AreaTree godoc
// @Summary Get the China administrative area tree
// @Tags System Area
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/area/tree [get]
func (h *Handler) AreaTree(c *gin.Context) {
	root, err := loadAreaTree()
	if err != nil {
		httpx.Fail(c, http.StatusInternalServerError, 500, "加载地区数据失败")
		return
	}
	httpx.OK(c, root.Children)
}

// AreaByIP godoc
// @Summary Resolve a local IP description
// @Description Returns 内网地址 for loopback/private IPs. Public IP geolocation is intentionally not sent to an external service.
// @Tags System Area
// @Produce json
// @Security BearerAuth
// @Param ip query string true "IP address"
// @Success 200 {object} httpx.Response
// @Router /system/area/get-by-ip [get]
func (h *Handler) AreaByIP(c *gin.Context) {
	ip := net.ParseIP(strings.TrimSpace(c.Query("ip")))
	if ip == nil {
		httpx.OK(c, nil)
		return
	}
	if ip.IsLoopback() || ip.IsPrivate() {
		httpx.OK(c, "内网地址")
		return
	}
	httpx.OK(c, nil)
}

func loadAreaTree() (*AreaNode, error) {
	areaOnce.Do(func() {
		rows, err := csv.NewReader(strings.NewReader(areaCSV)).ReadAll()
		if err != nil {
			areaErr = err
			return
		}
		nodes := make(map[uint64]*AreaNode, len(rows))
		for index, row := range rows {
			if index == 0 || len(row) < 4 {
				continue
			}
			id, idErr := strconv.ParseUint(row[0], 10, 64)
			typeValue, typeErr := strconv.Atoi(row[2])
			parentID, parentErr := strconv.ParseUint(row[3], 10, 64)
			if idErr != nil || typeErr != nil || parentErr != nil {
				continue
			}
			nodes[id] = &AreaNode{ID: id, Name: row[1], Type: typeValue, ParentID: parentID, Children: []*AreaNode{}}
		}
		for _, node := range nodes {
			if parent := nodes[node.ParentID]; parent != nil {
				parent.Children = append(parent.Children, node)
			}
		}
		areaRoot = nodes[1]
		if areaRoot == nil {
			areaErr = errors.New("china area root is missing")
		}
	})
	return areaRoot, areaErr
}
