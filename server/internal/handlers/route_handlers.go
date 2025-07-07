package handlers

type ReturnType struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
