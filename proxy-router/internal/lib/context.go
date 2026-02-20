package lib

import "context"

type contextKey string

const requestIDKey contextKey = "request_id"

// Plain string key so gin.Context.Value() finds it via gin's Get() path
const requestIDStringKey = "request_id"

func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	// Try plain string key first — works with gin.Context which routes
	// string keys through its internal map (Set/Get)
	if v, ok := ctx.Value(requestIDStringKey).(string); ok {
		return v
	}
	// Fall back to typed key — works with standard context.WithValue
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}
