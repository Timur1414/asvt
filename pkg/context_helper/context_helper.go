package context_helper

import "context"

type contextKey string

const (
	ContextKeyRequestId contextKey = "request_id"
	ContextKeyUser      contextKey = "user"
)

func GetRequestIdFromContext(ctx context.Context) string {
	requestId, ok := ctx.Value(ContextKeyRequestId).(string)
	if !ok {
		return ""
	}
	return requestId
}

func GetUserIdFromContext(ctx context.Context) int {
	userId, ok := ctx.Value(ContextKeyUser).(int)
	if !ok {
		return -1
	}
	return userId
}
