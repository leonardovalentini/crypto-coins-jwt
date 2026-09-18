package contextKey

import (
	"context"
	"net/http"

	"github.com/leonardovalentini/crypto-coins/lib/errs"
)

type ctxKey int

const (
	userIdKey ctxKey = iota
	jobIdKey
	requestIdKey
)

func WithJobId(ctx context.Context, jobID string) context.Context {
	return context.WithValue(ctx, jobIdKey, jobID)
}

func GetJobId(ctx context.Context) (string, bool) {
	jobID, ok := ctx.Value(jobIdKey).(string)
	return jobID, ok
}

func WithRequestId(ctx context.Context, requestId string) context.Context {
	return context.WithValue(ctx, requestIdKey, requestId)
}

func GetRequestId(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIdKey).(string)
	return id, ok
}

func SetUserId(r *http.Request, userId string) *context.Context {
	c := context.WithValue(r.Context(), userIdKey, userId)
	return &c
}

func GetUserId(ctx context.Context) (*string, error) {
	userId, ok := ctx.Value(userIdKey).(string)
	if !ok {
		return nil, errs.NewAuthenticationError("User id not found")
	}

	return &userId, nil
}
