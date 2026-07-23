package system

import (
	"crypto/rand"
	"encoding/hex"
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

const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

type tokenClaims struct {
	TenantID uint64 `json:"tenant_id"`
	Type     string `json:"token_type"`
	jwt.RegisteredClaims
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
	return s.issueTokenPair(user)
}

func (s *Service) ParseToken(raw string) (uint64, uint64, error) {
	claims, err := s.parseToken(raw, tokenTypeAccess)
	if err != nil {
		return 0, 0, err
	}
	userID, err := claims.GetSubject()
	if err != nil {
		return 0, 0, err
	}
	var uid uint64
	if _, err = fmt.Sscan(userID, &uid); err != nil {
		return 0, 0, err
	}
	return uid, claims.TenantID, nil
}

func (s *Service) RefreshToken(raw string) (TokenResponse, error) {
	claims, err := s.parseToken(raw, tokenTypeRefresh)
	if err != nil {
		return TokenResponse{}, errors.New("无效的刷新令牌")
	}
	userID, err := claims.GetSubject()
	if err != nil {
		return TokenResponse{}, errors.New("无效的刷新令牌")
	}
	uid, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return TokenResponse{}, errors.New("无效的刷新令牌")
	}
	var user AdminUser
	if err := s.db.Where("id = ? AND tenant_id = ? AND status = 0", uid, claims.TenantID).First(&user).Error; err != nil {
		return TokenResponse{}, errors.New("无效的刷新令牌")
	}
	return s.issueTokenPair(user)
}

func (s *Service) issueTokenPair(user AdminUser) (TokenResponse, error) {
	now := time.Now()
	accessExpires := now.Add(s.cfg.TokenTTL)
	refreshExpires := now.Add(s.cfg.RefreshTokenTTL)
	accessToken, err := s.signToken(user, tokenTypeAccess, now, accessExpires)
	if err != nil {
		return TokenResponse{}, err
	}
	refreshToken, err := s.signToken(user, tokenTypeRefresh, now, refreshExpires)
	if err != nil {
		return TokenResponse{}, err
	}
	return TokenResponse{
		UserID: user.ID, UserType: 2, AccessToken: accessToken,
		RefreshToken: refreshToken, ExpiresTime: accessExpires.UnixMilli(),
	}, nil
}

func (s *Service) signToken(user AdminUser, tokenType string, issuedAt, expiresAt time.Time) (string, error) {
	tokenID, err := newTokenID()
	if err != nil {
		return "", err
	}
	claims := tokenClaims{
		TenantID: user.TenantID,
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Subject:   strconv.FormatUint(user.ID, 10),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
}

func newTokenID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func (s *Service) parseToken(raw, expectedType string) (*tokenClaims, error) {
	claims := &tokenClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid || claims.Type != expectedType || claims.TenantID == 0 {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (s *Service) User(id uint64) (AdminUser, error) {
	var user AdminUser
	err := s.db.First(&user, id).Error
	return user, err
}

func DefaultMenus() []Menu {
	userComponent := "system/user/index"
	userComponentName := "SystemUser"
	infraConfigComponent := "infra/config/index"
	infraConfigComponentName := "InfraConfig"
	infraFileConfigComponent := "infra/fileConfig/index"
	infraFileConfigComponentName := "InfraFileConfig"
	infraAccessLogComponent := "infra/apiAccessLog/index"
	infraAccessLogComponentName := "InfraApiAccessLog"
	return []Menu{
		{
			ID: 1, Name: "系统管理", Path: "/system", Icon: "ep:tools",
			Visible: true, KeepAlive: true, AlwaysShow: true,
			Children: []Menu{{
				ID: 100, ParentID: 1, Name: "用户管理", Path: "user",
				Component: &userComponent, ComponentName: &userComponentName, Icon: "ep:avatar", Visible: true, KeepAlive: true,
			}},
		},
		{
			ID: 2, Name: "基础设施", Path: "/infra", Icon: "ep:setting",
			Visible: true, KeepAlive: true, AlwaysShow: true,
			Children: []Menu{
				{ID: 201, ParentID: 2, Name: "参数配置", Path: "config", Component: &infraConfigComponent, ComponentName: &infraConfigComponentName, Icon: "ep:operation", Visible: true, KeepAlive: true},
				{ID: 202, ParentID: 2, Name: "文件配置", Path: "file-config", Component: &infraFileConfigComponent, ComponentName: &infraFileConfigComponentName, Icon: "ep:folder-opened", Visible: true, KeepAlive: true},
				{ID: 203, ParentID: 2, Name: "访问日志", Path: "api-access-log", Component: &infraAccessLogComponent, ComponentName: &infraAccessLogComponentName, Icon: "ep:document", Visible: true, KeepAlive: true},
			},
		},
	}
}
