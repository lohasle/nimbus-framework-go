package router

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/modules/system"
	"github.com/lohasle/nimbus-framework-go/internal/platform/config"
)

func TestVisibleAdminRouteContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := New(system.NewHandler(system.NewService(nil, config.Config{})), nil)
	routes := make(map[string]struct{}, len(engine.Routes()))
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}

	want := []string{
		"POST /admin-api/system/auth/login",
		"GET /admin-api/system/auth/get-permission-info",
		"GET /admin-api/system/user/page",
		"GET /admin-api/system/user/list",
		"GET /admin-api/system/user/get",
		"POST /admin-api/system/user/create",
		"PUT /admin-api/system/user/update",
		"DELETE /admin-api/system/user/delete",
		"GET /admin-api/system/user/export-excel",
		"GET /admin-api/system/user/get-import-template",
		"POST /admin-api/system/user/import",
		"GET /admin-api/system/role/simple-list",
		"GET /admin-api/system/permission/list-user-roles",
		"POST /admin-api/system/permission/assign-user-role",
		"GET /admin-api/system/area/tree",
		"GET /admin-api/infra/config/page",
		"GET /admin-api/infra/config/export-excel",
		"GET /admin-api/infra/file-config/page",
		"GET /admin-api/infra/api-access-log/page",
		"GET /admin-api/infra/api-access-log/export-excel",
		"GET /admin-api/member/user/page",
		"GET /admin-api/member/user/get",
		"PUT /admin-api/member/user/update",
		"PUT /admin-api/member/user/update-level",
		"PUT /admin-api/member/user/update-point",
		"GET /admin-api/member/level/list",
		"GET /admin-api/member/group/page",
		"GET /admin-api/member/tag/page",
		"GET /admin-api/member/point/record/page",
		"GET /admin-api/member/experience-record/page",
		"GET /admin-api/member/sign-in/record/page",
		"GET /admin-api/member/address/list",
		"GET /admin-api/pay/app/page",
		"GET /admin-api/pay/channel/page",
		"GET /admin-api/pay/channel/get",
		"GET /admin-api/pay/channel/list",
		"GET /admin-api/pay/channel/get-enable-code-list",
		"GET /admin-api/pay/order/page",
		"GET /admin-api/pay/refund/page",
		"GET /admin-api/pay/wallet/page",
		"PUT /admin-api/pay/wallet/update-balance",
		"GET /admin-api/pay/wallet-transaction/page",
	}

	for _, route := range want {
		if _, ok := routes[route]; !ok {
			t.Errorf("missing visible admin route %s", route)
		}
	}
}
