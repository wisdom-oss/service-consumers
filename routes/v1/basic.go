package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BasicHandler contains just a response, that is used to show the templating.
func BasicHandler(c *gin.Context) {
	c.String(http.StatusOK, "hello there")
}
