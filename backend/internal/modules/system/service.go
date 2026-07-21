package system

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lohasle/nimbus-framework-go/internal/platform/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	db  *gorm.DB
	cfg config.Config
}

func NewService(db *gorm.DB, cfg config.Config) *Service { return &Service{db: db, cfg: cfg} }

func (s *Service) TenantID(name string) (uint64, error) {
	var tenant Tenant
	err := s.db.Where("name = ? AND status = 0", name).First(&tenant).Error
	return tenant.ID, err
}

func (s *Service) Login(tenantID uint64, req LoginRequest) (TokenResponse, error) {
	var user AdminUser
	if err := s.db.Where("tenant_id = ? AND username = ? AND status = 0", tenantID, req.Username).First(&user).Error; err != nil {
		return TokenResponse{}, errors.New("用户名或密码错误")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return TokenResponse{}, errors.New("用户名或密码错误")
	}
	expires := time.Now().Add(s.cfg.TokenTTL)
	claims := jwt.MapClaims{"sub": strconv.FormatUint(user.ID, 10), "tenant_id": user.TenantID, "exp": expires.Unix(), "iat": time.Now().Unix()}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return TokenResponse{}, err
	}
	return TokenResponse{UserID: user.ID, UserType: 2, AccessToken: token, RefreshToken: token, ExpiresTime: expires.UnixMilli()}, nil
}

func (s *Service) ParseToken(raw string) (uint64, uint64, error) {
	token, err := jwt.Parse(raw, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return 0, 0, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, 0, errors.New("invalid claims")
	}
	userID, err := claims.GetSubject()
	if err != nil {
		return 0, 0, err
	}
	var uid uint64
	if _, err = fmt.Sscan(userID, &uid); err != nil {
		return 0, 0, err
	}
	tenant, ok := claims["tenant_id"].(float64)
	if !ok {
		return 0, 0, errors.New("invalid tenant claim")
	}
	return uid, uint64(tenant), nil
}

func (s *Service) User(id uint64) (AdminUser, error) {
	var user AdminUser
	err := s.db.First(&user, id).Error
	return user, err
}

func DefaultMenus() []Menu {
	userComponent := "system/user/index"
	infraConfigComponent := "infra/config/index"
	infraFileConfigComponent := "infra/fileConfig/index"
	infraAccessLogComponent := "infra/apiAccessLog/index"
	memberUserComponent := "member/user/index"
	memberLevelComponent := "member/level/index"
	memberGroupComponent := "member/group/index"
	memberTagComponent := "member/tag/index"
	payAppComponent := "pay/app/index"
	payOrderComponent := "pay/order/index"
	payRefundComponent := "pay/refund/index"
	return []Menu{
		{
			ID: 1, Name: "系统管理", Path: "/system", Icon: "ep:tools",
			Visible: true, KeepAlive: true, AlwaysShow: true,
			Children: []Menu{{
				ID: 100, ParentID: 1, Name: "用户管理", Path: "user",
				Component: &userComponent, Icon: "ep:avatar", Visible: true, KeepAlive: true,
			}},
		},
		{
			ID: 2, Name: "基础设施", Path: "/infra", Icon: "ep:setting",
			Visible: true, KeepAlive: true, AlwaysShow: true,
			Children: []Menu{
				{ID: 201, ParentID: 2, Name: "参数配置", Path: "config", Component: &infraConfigComponent, Icon: "ep:operation", Visible: true, KeepAlive: true},
				{ID: 202, ParentID: 2, Name: "文件配置", Path: "file-config", Component: &infraFileConfigComponent, Icon: "ep:folder-opened", Visible: true, KeepAlive: true},
				{ID: 203, ParentID: 2, Name: "访问日志", Path: "api-access-log", Component: &infraAccessLogComponent, Icon: "ep:document", Visible: true, KeepAlive: true},
			},
		},
		{
			ID: 3, Name: "会员中心", Path: "/member", Icon: "ep:user-filled",
			Visible: true, KeepAlive: true, AlwaysShow: true,
			Children: []Menu{
				{ID: 301, ParentID: 3, Name: "会员管理", Path: "user", Component: &memberUserComponent, Icon: "ep:user", Visible: true, KeepAlive: true},
				{ID: 302, ParentID: 3, Name: "会员等级", Path: "level", Component: &memberLevelComponent, Icon: "ep:medal", Visible: true, KeepAlive: true},
				{ID: 303, ParentID: 3, Name: "会员分组", Path: "group", Component: &memberGroupComponent, Icon: "ep:collection-tag", Visible: true, KeepAlive: true},
				{ID: 304, ParentID: 3, Name: "会员标签", Path: "tag", Component: &memberTagComponent, Icon: "ep:price-tag", Visible: true, KeepAlive: true},
			},
		},
		{
			ID: 4, Name: "支付中心", Path: "/pay", Icon: "ep:wallet-filled",
			Visible: true, KeepAlive: true, AlwaysShow: true,
			Children: []Menu{
				{ID: 401, ParentID: 4, Name: "支付应用", Path: "app", Component: &payAppComponent, Icon: "ep:grid", Visible: true, KeepAlive: true},
				{ID: 402, ParentID: 4, Name: "支付订单", Path: "order", Component: &payOrderComponent, Icon: "ep:tickets", Visible: true, KeepAlive: true},
				{ID: 403, ParentID: 4, Name: "退款管理", Path: "refund", Component: &payRefundComponent, Icon: "ep:refresh-left", Visible: true, KeepAlive: true},
			},
		},
	}
}
