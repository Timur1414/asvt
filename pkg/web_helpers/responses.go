package web_helpers

import "net/http"

type SimpleResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

type OkResponse SimpleResponse

func NewOkResponse() OkResponse {
	return OkResponse{
		Code:    http.StatusOK,
		Message: "Ok",
	}
}

type UuidResponse struct {
	SimpleResponse
	RequestId string `json:"request_id,omitempty"`
}

func NewUuidResponse(code int, message string, requestId string) UuidResponse {
	return UuidResponse{
		SimpleResponse: SimpleResponse{
			Code:    code,
			Message: message,
		},
		RequestId: requestId,
	}
}

type TooManyRequestsResponse UuidResponse

func NewTooManyRequestsResponse(requestId string) TooManyRequestsResponse {
	return TooManyRequestsResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusTooManyRequests,
			Message: "Слишком много запросов",
		},
		RequestId: requestId,
	}
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func NewValidationError(field string, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

type ValidationErrorResponse struct {
	UuidResponse
	Errors []ValidationError `json:"errors"`
}

func NewValidationErrorResponse(requestId string, errors []ValidationError) ValidationErrorResponse {
	return ValidationErrorResponse{
		UuidResponse: NewUuidResponse(http.StatusBadRequest, "Ошибка валидации", requestId),
		Errors:       errors,
	}
}

type ForbiddenResponse UuidResponse

func NewForbiddenResponse(requestId string) ForbiddenResponse {
	return ForbiddenResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusForbidden,
			Message: "Доступ запрещён",
		},
		RequestId: requestId,
	}
}

type UnauthorizedResponse UuidResponse

func NewUnauthorizedResponse(requestId string) UnauthorizedResponse {
	return UnauthorizedResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusUnauthorized,
			Message: "Не авторизирован",
		},
		RequestId: requestId,
	}
}

type ServerErrorResponse UuidResponse

func NewServerErrorResponse(requestId string) ServerErrorResponse {
	return ServerErrorResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusInternalServerError,
			Message: "Ошибка сервера",
		},
		RequestId: requestId,
	}
}
