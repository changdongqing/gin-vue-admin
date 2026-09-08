package rulego

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/rulego/rulego/server/model"
	"github.com/rulego/rulego/server/services"
)

// gvaAuthenticator 实现 rulego services.Authenticator（v0.37.2 SPI）：
// Authenticate(authorization string) 的入参是 Authorization 头原始值（"Bearer <token>"；
// ?token= 兜底路径由 rulego 自动补 Bearer 前缀），拿不到 *http.Request，因此
// x-token → Authorization 的头映射在更外层的 mapXToken 中间件完成。
// 解析直接复用 GVA utils/jwt.go（HS256 CustomClaims），通过后组装 rulego UserContext，
// 角色 = GVA BaseClaims.AuthorityId 的字符串形式（Casbin 的 sub 口径）。
type gvaAuthenticator struct{}

var _ services.Authenticator = (*gvaAuthenticator)(nil)

func (gvaAuthenticator) Authenticate(authorization string) (*model.UserContext, error) {
	tokenStr := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
	if tokenStr == "" {
		return nil, fmt.Errorf("empty token")
	}
	claims, err := utils.NewJWT().ParseToken(tokenStr)
	if err != nil {
		return nil, err
	}
	return &model.UserContext{
		Username: claims.Username,
		Roles:    []string{strconv.FormatUint(uint64(claims.AuthorityId), 10)},
	}, nil
}

// gvaAuthorizer 实现 rulego services.Authorizer（v0.37.2 SPI）：
// Authorize(user, resource, action) 传入的是语义标签而非 HTTP path/method，
// 经 mapper.go 固定映射为虚拟路径后，走与 middleware/casbin_rbac.go 相同的
// Casbin 入口判权（"API 管理"录入 /rulego/** 虚拟路径即可按角色授权）。
type gvaAuthorizer struct{}

var _ services.Authorizer = (*gvaAuthorizer)(nil)

// superAdminAuthorityID GVA 内置超级管理员角色。其 Casbin 策略按实际 API 路径维护，
// /rulego 虚拟路径默认未录入，这里显式放行以保持 GVA"超管全通过"惯例；普通角色严格走策略。
const superAdminAuthorityID = "888"

func (gvaAuthorizer) Authorize(user *model.UserContext, resource, action string) error {
	if user == nil || len(user.Roles) == 0 {
		return fmt.Errorf("no role")
	}
	sub := user.Roles[0]
	if sub == superAdminAuthorityID {
		return nil
	}
	path, method, ok := mapAction(resource, action)
	if !ok {
		return fmt.Errorf("no mapping for resource=%q action=%q", resource, action)
	}
	// GetCasbin 在 GVA_DB 未就绪时会返回 nil（sync.Once 已消耗），必须判空防 panic。
	e := utils.GetCasbin()
	if e == nil {
		return fmt.Errorf("casbin not ready")
	}
	if allowed, _ := e.Enforce(sub, path, method); !allowed {
		return fmt.Errorf("forbidden: %s %s", method, path)
	}
	return nil
}
