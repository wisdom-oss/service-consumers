package structs

// RequestBody contains the information about a consumer which should
// be stored in the database or updated in the database
type RequestBody struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"long"`
}

type GetConsumerQueryParameters struct {
	UsageAbove  []int    `schema:"usage_above"`
	ConsumerIds []string `schema:"id"`
	AreaKeys    []string `schema:"in"`
}

type UpdateConsumerQueryParameters struct {
	UpdateName     bool `schema:"updateName"`
	UpdateLocation bool `schema:"updateLocation"`
}
