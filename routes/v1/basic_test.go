package v1_test

import (
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/joho/godotenv/autoload"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"

	validator "openapi.tanna.dev/go/validator/openapi3"

	_ "microservice/internal/db"
	"microservice/router"
)

var contract *openapi3.T

func TestMain(m *testing.M) {
	err := os.Chdir("../../")
	if err != nil {
		panic(err)
	}
	contract, err = openapi3.NewLoader().LoadFromFile("./openapi.yaml")
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestBasicHandler(t *testing.T) {
	r, err := router.Configure()
	assert.NoError(t, err)

	request := httptest.NewRequest("GET", "/v1/", nil)
	responseRecorder := httptest.NewRecorder()
	_ = validator.NewValidator(contract).ForTest(t, responseRecorder, request)
	r.ServeHTTP(responseRecorder, request)
}
