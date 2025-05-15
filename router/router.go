package router

import (
	"github.com/gin-gonic/gin"

	internal "microservice/internal/router"
	v1Routes "microservice/routes/v1"
)

// Configure generates a new router and adds routes to the router
//
// The router can also be imported during tests, as long as the tests are in a
// separate package.
// If the tests are in the same package (e.g. routes defined in `v3` and tests
// also defined in `v3`) an import cycle exists.
func Configure() (*gin.Engine, error) {
	r, err := internal.GenerateRouter()
	if err != nil {
		return nil, err
	}

	// TODO: Add your routes in this group
	v1 := r.Group("/v1")
	{
		v1.GET("/", v1Routes.BasicHandler)
	}

	return r, nil
}
