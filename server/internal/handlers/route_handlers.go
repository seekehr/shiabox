package handlers

type ReturnType struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type StreamReturnType struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Done    bool        `json:"done"` // so the client doesnt keep listening indefinitely
}
