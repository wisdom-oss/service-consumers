package router

import (
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/thanhpk/randstr"
	"github.com/wisdom-oss/common-go/v3/middleware/gin/recoverer"
	"github.com/wisdom-oss/common-go/v3/types"

	errorHandler "github.com/wisdom-oss/common-go/v3/middleware/gin/error-handler"
)

// requestIDLength determines how long the generated request id will be.
const requestIDLength = 64

// ErrMethodNotAllowed is used if a request uses a unsupported or disabled
// HTTP method.
var ErrMethodNotAllowed = types.ServiceError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110.html#section-15.5.6",
	Status: http.StatusMethodNotAllowed,
	Title:  "Method Not Allowed",
	Detail: "The used HTTP method is not allowed on this route. Please check the documentation and your request",
}

// ErrRouteNotFound is used if a request matches no configured route in the
// service.
var ErrRouteNotFound = types.ServiceError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110.html#section-15.5.5",
	Status: http.StatusNotFound,
	Title:  "Not Found",
	Detail: "The requested path does not exist. Please check the documentation and your request",
}

func prepareRouter() *gin.Engine {
	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.UseH2C = true
	r.RedirectFixedPath = true

	r.Use(errorHandler.Handler)
	r.Use(gin.CustomRecovery(recoverer.RecoveryHandler))
	r.Use(requestid.New(
		requestid.WithGenerator(func() string {
			return randstr.Base62(requestIDLength)
		}),
	))

	r.NoMethod(func(c *gin.Context) {
		ErrMethodNotAllowed.Emit(c)
	})

	r.NoRoute(func(c *gin.Context) {
		ErrRouteNotFound.Emit(c)
	})

	return r
}
