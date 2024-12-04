package response

type Response struct {
	Status string `json:"status"`
	Erorr  string `json:"erorr,omitempty"`
}

const (
	StatusOk    = "OK"
	StatusError = "Error"
)

func OK() Response {
	return Response{
		Status: StatusOk,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Erorr:  msg,
	}
}
