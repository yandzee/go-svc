package router

import (
	"fmt"
	"log/slog"
)

type corsLogger struct {
	Log *slog.Logger
}

func (cl *corsLogger) Printf(f string, args ...interface{}) {
	msg := fmt.Sprintf(f, args...)
	cl.Log.Debug(msg)
}
