//go:build !release

package router

import (
	"github.com/gin-gonic/gin"
)

// GenerateRouter returns a new [*gin.Engine] which has been configured
// for running in development environments.
func GenerateRouter() (*gin.Engine, error) {
	r := prepareRouter()
	return r, nil
}
