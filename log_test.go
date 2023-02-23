package log

import (
	"context"
	"testing"
)

func TestInit(t *testing.T) {
	InitLogger()
	ctx := context.WithValue(context.Background(), "sp", "2222")
	//SetContext(ctx, "trace_id", "1234567890")
	Warn(ctx, "test", "pass")
}
