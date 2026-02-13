package configfile

import "context"

type jobIDContextKey struct{}
type queueNameContextKey struct{}
type priorityContextKey struct{}
type submitTypeContextKey struct{}

func WithJobID(ctx context.Context, jobID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, jobIDContextKey{}, jobID)
}

func jobIDFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	val := ctx.Value(jobIDContextKey{})
	jobID, ok := val.(string)
	if !ok || jobID == "" {
		return "", false
	}
	return jobID, true
}

func WithQueueName(ctx context.Context, queueName string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, queueNameContextKey{}, queueName)
}

func queueNameFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	val := ctx.Value(queueNameContextKey{})
	queueName, ok := val.(string)
	if !ok || queueName == "" {
		return "", false
	}
	return queueName, true
}

func WithPriority(ctx context.Context, priority int32) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, priorityContextKey{}, priority)
}

func priorityFromContext(ctx context.Context) (int32, bool) {
	if ctx == nil {
		return 0, false
	}
	val := ctx.Value(priorityContextKey{})
	priority, ok := val.(int32)
	if !ok {
		return 0, false
	}
	return priority, true
}

func WithSubmitType(ctx context.Context, submitType string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, submitTypeContextKey{}, submitType)
}

func submitTypeFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	val := ctx.Value(submitTypeContextKey{})
	submitType, ok := val.(string)
	if !ok || submitType == "" {
		return "", false
	}
	return submitType, true
}
