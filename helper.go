package log

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unsafe"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// SetContext sets the context field to the log-out.
// if ctx miss the key, then use the value in kvs.
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
	fields = append(fields, zap.Any("caller", caller(3)))
	fields = getTraceAndSpan(ctx, fields)

	for i := 0; i < len(kvs); i += 2 {
		fields = append(fields, zap.Any(fmt.Sprint(kvs[i]), kvs[i+1]))
	}

	for j := 0; j < len(logger.prefix); j += 2 {
		k := logger.prefix[j]
		v := ctx.Value(k)
		if v != nil {
			fields = append(fields, zap.Any(fmt.Sprint(logger.prefix[j]), v))
			continue
		}
		fields = append(fields, zap.Any(fmt.Sprint(logger.prefix[j]), logger.prefix[j+1]))
	}

	return fields
}

func getTraceAndSpan(ctx context.Context, fields []zap.Field) []zap.Field {
	var traceId, spanId string
	if span := trace.SpanContextFromContext(ctx); span.HasTraceID() {
		traceId = span.TraceID().String()
	}

	if span := trace.SpanContextFromContext(ctx); span.HasSpanID() {
		spanId = span.SpanID().String()
	}

	fields = append(fields, zap.String("trace_id", traceId))
	fields = append(fields, zap.String("span_id", spanId))
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

// contextInternals returns the context kvs.
func contextInternals(ctx context.Context) []zap.Field {
	contextValues := reflect.ValueOf(ctx).Elem()
	contextKeys := reflect.TypeOf(ctx).Elem()
	var fields []zap.Field

	var keys []string
	if contextKeys.Kind() == reflect.Struct {
		for i := 0; i < contextValues.NumField(); i++ {
			reflectValue := contextValues.Field(i)
			reflectValue = reflect.NewAt(reflectValue.Type(), unsafe.Pointer(reflectValue.UnsafeAddr())).Elem()

			reflectField := contextKeys.Field(i)

			if reflectField.Name == "Context" {
				fields = append(fields, contextInternals(reflectValue.Interface().(context.Context))...)
			} else {
				if reflectField.Name == "key" {
					keys = append(keys, reflectValue.Interface().(string))
				}
			}
		}
	}

	for _, key := range keys {
		fields = append(fields, zap.Any(key, ctx.Value(key)))
	}
	return fields
}
