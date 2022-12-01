package response

import "net/http"

func Success() Response {
	return Response{
		Message: http.StatusText(http.StatusOK),
		Status:  http.StatusOK,
	}
}

func BadRequest(message string) Response {
	return Response{
		Message: message,
		Status:  http.StatusBadRequest,
	}
}

func Unauthorized() Response {
	return Response{
		Message: http.StatusText(http.StatusUnauthorized),
		Status:  http.StatusUnauthorized,
	}
}

func Forbidden() Response {
	return Response{
		Message: http.StatusText(http.StatusForbidden),
		Status:  http.StatusForbidden,
	}
}

func NotFound(message string) Response {
	return Response{
		Message: message,
		Status:  http.StatusNotFound,
	}
}

func Conflict(message string) Response {
	return Response{
		Message: message,
		Status:  http.StatusConflict,
	}
}

func InternalServerError() Response {
	return Response{
		Message: http.StatusText(http.StatusInternalServerError),
		Status:  http.StatusInternalServerError,
	}
}
