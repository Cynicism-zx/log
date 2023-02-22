package log

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// SetContext sets the context to the logger.
func SetContext(ctx context.Context, kvs ...interface{}) {
	logger.prefix = append(logger.prefix, kvs...)
}

// Info logs a message at info level.
func Info(ctx context.Context, kvs ...interface{}) {
	logger.log.Info("", getFields(ctx, kvs...)...)
}

// Warn logs a message at warn level.
func Warn(ctx context.Context, kvs ...interface{}) {
	logger.log.Warn("", getFields(ctx, kvs...)...)
}

// getFields returns zap fields.
func getFields(ctx context.Context, kvs ...interface{}) []zap.Field {
	if len(kvs) == 0 || len(kvs)%2 != 0 {
		logger.log.Warn(fmt.Sprint("Keyvalues must appear in pairs: ", kvs))
		return nil
	}

	fields := make([]zap.Field, 0, len(kvs)+len(logger.prefix))
	var (
		traceId, spanId string
		newKvs          []interface{}
	)
	if span := trace.SpanContextFromContext(ctx); span.HasTraceID() {
		traceId = span.TraceID().String()
	}

	if span := trace.SpanContextFromContext(ctx); span.HasSpanID() {
		spanId = span.SpanID().String()
	}

	fields = append(fields, zap.String("trace_id", traceId))
	fields = append(fields, zap.String("span_id", spanId))
	fields = append(fields, zap.Any("caller", caller(3)))

	for i := 0; i < len(kvs); i += 2 {
		for j := 0; j < len(logger.prefix); j += 2 {
			if kvs[i] == logger.prefix[j] {
				continue
			}
			newKvs = append(newKvs, logger.prefix[j], logger.prefix[j+1])
		}
		fields = append(fields, zap.Any(fmt.Sprint(kvs[i]), fmt.Sprint(kvs[i+1])))
	}

	for i := 0; i < len(newKvs); i += 2 {
		fields = append(fields, zap.Any(fmt.Sprint(newKvs[i]), fmt.Sprint(newKvs[i+1])))
	}

	return fields
}

func caller(depth int) interface{} {
	_, file, line, _ := runtime.Caller(depth)
	idx := strings.LastIndexByte(file, '/')
	if idx == -1 {
		return file[idx+1:] + ":" + strconv.Itoa(line)
	}
	idx = strings.LastIndexByte(file[:idx], '/')
	return file[idx+1:] + ":" + strconv.Itoa(line)
}
