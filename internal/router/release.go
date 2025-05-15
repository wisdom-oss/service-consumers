//go:build release

package router

import (
	"microservice/internal/configuration"

	"github.com/gin-gonic/gin"
	"github.com/wisdom-oss/common-go/v3/middleware/gin/jwt"
)

// GenerateRouter returns a new [*gin.Engine] which has been configured
// to run in release scenarios.
// This enables security hardening and decreases the default logging level.
func GenerateRouter() (*gin.Engine, error) {
	r := prepareRouter()
	gin.SetMode(gin.ReleaseMode)

	/* Configure OpenID Connect */
	authority := configuration.Default.Viper().GetString(configuration.ConfigurationKey_OidcAuthority)

	jwtValidator := jwt.Validator{}
	err := jwtValidator.Discover(authority)
	if err != nil {
		return nil, err
	}

	if !configuration.Default.Viper().GetBool(configuration.ConfigurationKey_AuthorizationRequired) {
		jwtValidator.EnableOptional()
	}

	r.Use(jwtValidator.Handler)
	return r, nil

}
