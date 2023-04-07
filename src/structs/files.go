package structs

// EnvironmentConfiguration matches the json5 file used to specify the
// needed environment variables
type EnvironmentConfiguration struct {
	RequiredEnvironmentVariables []string `json:"required"`
	OptionalEnvironmentVariables []struct {
		EnvironmentKey string `json:"key"`
		DefaultValue   string `json:"default"`
	} `json:"optional"`
}

type RequestError struct {
	ErrorCode        string `json:"code"`
	ErrorTitle       string `json:"title"`
	ErrorDescription string `json:"description"`
	HttpCode         int    `json:"httpCode"`
}
