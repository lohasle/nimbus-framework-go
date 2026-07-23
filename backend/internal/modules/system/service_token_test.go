package system

import (
	"testing"
	"time"

	"github.com/lohasle/nimbus-framework-go/internal/platform/config"
)

func TestTokenPairUsesSeparateTypedTokens(t *testing.T) {
	service := NewService(nil, config.Config{
		JWTSecret:       "test-secret",
		TokenTTL:        2 * time.Hour,
		RefreshTokenTTL: 30 * 24 * time.Hour,
	})
	user := AdminUser{ID: 42, TenantID: 7}

	pair, err := service.issueTokenPair(user)
	if err != nil {
		t.Fatalf("issue token pair: %v", err)
	}
	if pair.AccessToken == pair.RefreshToken {
		t.Fatal("access token and refresh token must differ")
	}
	if uid, tenantID, err := service.ParseToken(pair.AccessToken); err != nil || uid != user.ID || tenantID != user.TenantID {
		t.Fatalf("parse access token: uid=%d tenant=%d err=%v", uid, tenantID, err)
	}
	if _, _, err := service.ParseToken(pair.RefreshToken); err == nil {
		t.Fatal("refresh token must not be accepted as an access token")
	}
	if _, err := service.parseToken(pair.AccessToken, tokenTypeRefresh); err == nil {
		t.Fatal("access token must not be accepted as a refresh token")
	}
}
