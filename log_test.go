package log

import (
	"context"
	"testing"
)

func TestInit(t *testing.T) {
	InitLogger()
	Warn(context.Background(), "test", "pass")
}
