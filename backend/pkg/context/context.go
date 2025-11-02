package context

import (
	"context"
	"fmt"
)

// ContextKey represents a context key type
type ContextKey string

// Context keys
const (
	UserIDKey     ContextKey = "user_id"
	RequestIDKey  ContextKey = "request_id"
	TenantIDKey   ContextKey = "tenant_id"
	RoleKey       ContextKey = "role"
	EmailKey      ContextKey = "email"
)

// WithUserID adds user_id to context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserID extracts user_id from context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// MustGetUserID extracts user_id or panics
func MustGetUserID(ctx context.Context) string {
	userID, ok := GetUserID(ctx)
	if !ok {
		panic("user_id not found in context")
	}
	return userID
}

// WithRequestID adds request_id to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID extracts request_id from context
func GetRequestID(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(RequestIDKey).(string)
	return requestID, ok
}

// MustGetRequestID extracts request_id or panics
func MustGetRequestID(ctx context.Context) string {
	requestID, ok := GetRequestID(ctx)
	if !ok {
		panic("request_id not found in context")
	}
	return requestID
}

// WithTenantID adds tenant_id to context
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// GetTenantID extracts tenant_id from context
func GetTenantID(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value(TenantIDKey).(string)
	return tenantID, ok
}

// WithRole adds role to context
func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, RoleKey, role)
}

// GetRole extracts role from context
func GetRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(RoleKey).(string)
	return role, ok
}

// WithEmail adds email to context
func WithEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, EmailKey, email)
}

// GetEmail extracts email from context
func GetEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(EmailKey).(string)
	return email, ok
}

// WithValues adds multiple values to context
func WithValues(ctx context.Context, values map[ContextKey]interface{}) context.Context {
	for key, value := range values {
		ctx = context.WithValue(ctx, key, value)
	}
	return ctx
}

// GetString extracts a string value from context
func GetString(ctx context.Context, key ContextKey) (string, bool) {
	value, ok := ctx.Value(key).(string)
	return value, ok
}

// GetInt extracts an int value from context
func GetInt(ctx context.Context, key ContextKey) (int, bool) {
	value, ok := ctx.Value(key).(int)
	return value, ok
}

// String returns string representation of the key
func (k ContextKey) String() string {
	return fmt.Sprintf("context.%s", string(k))
}
