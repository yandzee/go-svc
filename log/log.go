package log

import (
	"fmt"
	"io"
	"log/slog"
)

func Discard() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func OrDiscard(l *slog.Logger) *slog.Logger {
	if l != nil {
		return l
	}

	return Discard()
}

func Strings[T any](key string, values []T) slog.Attr {
	arr := make([]any, 0, len(values))

	for _, v := range values {
		arr = append(arr, fmt.Sprintf("%v", v))
	}

	return slog.Any(key, arr)
}

func Error(key string, err error) slog.Attr {
	if err == nil {
		return slog.Any(key, nil)
	}

	return slog.String(key, err.Error())
}
