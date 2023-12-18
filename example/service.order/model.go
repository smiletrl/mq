package order

type createRequest struct {
	Product string `json:"product"`
}

type Response struct {
	Data string `json:"data"`
}
