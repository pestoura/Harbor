package controllers

import (
	"fmt"
	"github.com/goharbor/harbor/src/controller/config"
	"github.com/goharbor/harbor/src/lib/orm"
	"net/http"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/core/auth/authproxy"
	"github.com/goharbor/harbor/src/lib/log"
)

const (
	authproxyTokenKey = "token"
	postURIKey        = "postURI"
)

var helper = &authproxy.Auth{}

// AuthProxyController handles requests with token that can be reviewed by authproxy.
type AuthProxyController struct {
	api.BaseController
}

// Prepare checks the auth mode and fail early
func (apc *AuthProxyController) Prepare() {
	am, err := config.AuthMode(orm.Context())
	if err != nil {
		apc.SendInternalServerError(err)
		return
	}
	if am != common.HTTPAuth {
		apc.SendPreconditionFailedError(fmt.Errorf("the auth mode %s does not support this flow", am))
		return
	}
}

// HandleRedirect reviews the token and login the user based on the review status.
func (apc *AuthProxyController) HandleRedirect() {
	token := apc.Ctx.Request.URL.Query().Get(authproxyTokenKey)
	if token == "" {
		log.Errorf("No token found in request.")
		apc.Ctx.Redirect(http.StatusMovedPermanently, "/")
		return
	}
	u, err := helper.VerifyToken(token)
	if err != nil {
		log.Errorf("Failed to verify token, error: %v", err)
		apc.Ctx.Redirect(http.StatusMovedPermanently, "/")
		return
	}
	if err := helper.PostAuthenticate(u); err != nil {
		log.Errorf("Failed to onboard user, error: %v", err)
		apc.Ctx.Redirect(http.StatusMovedPermanently, "/")
		return
	}
	apc.PopulateUserSession(*u)
	uri := apc.Ctx.Request.URL.Query().Get(postURIKey)
	if uri == "" {
		uri = "/"
	}
	apc.Ctx.Redirect(http.StatusMovedPermanently, uri)
	return
}
