package lib

import (
	"testing"

	"go.uber.org/zap"
)

func TestLogger(t *testing.T) {
	l, err := NewLogger("debug", false, false, false, "")
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	l2 := l.With(
		zap.String("hello", "world"),
		zap.String("failure", "oh no"),
		zap.Stack("stack"),
		zap.Int("count", 42),
	)

	l2.Log(zap.DebugLevel, "Test", zap.String("ONE MORE FIREL", "FIREl"))

	// z:=zap.NewExample()
	// z2 := z.With(
	// 	zap.String("hello", "world"),
	// 	zap.String("failure", "oh no"),
	// 	zap.Stack("stack"),
	// 	zap.Int("count", 42),
	// )

	// z2.Debug("Test", zap.String("ONE MORE FIREL", "FIREl"))

	// s := z2.Sugar()
	// s.Debugw("Test", "ONE MORE FIREL", "FIREl")
}
