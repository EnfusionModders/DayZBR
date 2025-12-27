package server

type GenericResponse struct {
	Success int         `json:"success"`
	Error   string      `json:"error"`
	Data    interface{} `json:"data"`
}

func buildSuccessResponse(data interface{}) *GenericResponse {
	return &GenericResponse{
		Success: 1,
		Error:   "",
		Data:    data,
	}
}

func buildErrorResponse(message string) *GenericResponse {
	return &GenericResponse{
		Success: 0,
		Error:   message,
		Data:    nil,
	}
}
