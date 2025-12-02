package httputils

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/yandzee/go-svc/data/jsoner"
)

const MaxSizeDefault int = 1024 * 1024 // 1 MB

type Jsoner struct {
	jsoner               jsoner.Jsoner
	DefaultDecodeOptions JSONDecodeOptions
}

type JSONDecodeOptions struct {
	MaxSize              int
	UnknownFieldsAllowed bool
}

type JSONDecodeResult struct {
	jsoner.JSONDecodeResult

	IsWrongContentType bool
	MaxBytesError      *http.MaxBytesError
}

func (jdr *JSONDecodeResult) AsHTTPStatus() (int, string) {
	switch {
	case jdr.IsWrongContentType:
		return http.StatusUnsupportedMediaType, "Content-Type is not application/json"
	case jdr.MaxBytesError != nil:
		return http.StatusRequestEntityTooLarge, fmt.Sprintf(
			"Request body must not be larger than %d bytes",
			jdr.MaxBytesError.Limit,
		)
	}

	msg := jdr.JSONDecodeResult.Error()
	if msg != jsoner.NoError {
		return http.StatusBadRequest, msg
	}

	return http.StatusOK, ""
}

func (jdr *JSONDecodeResult) Err() error {
	st, msg := jdr.AsHTTPStatus()

	if len(msg) == 0 || st == http.StatusOK {
		return nil
	}

	return errors.New(msg)
}

func (j *Jsoner) EncodeResponse(
	w http.ResponseWriter,
	d any,
	isManualErrHandling ...bool,
) error {
	w.Header().Set("Content-Type", "application/json")
	err := j.jsoner.Encode(w, d)

	if err != nil && (len(isManualErrHandling) == 0 || !isManualErrHandling[0]) {
		http.Error(w, "EncodeResponse: "+err.Error(), http.StatusInternalServerError)
		return err
	}

	return err
}

func (j *Jsoner) DecodeRequest(
	w http.ResponseWriter,
	r *http.Request,
	dst any,
	opts ...JSONDecodeOptions,
) *JSONDecodeResult {
	result := &JSONDecodeResult{}

	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))

		if mediaType != "application/json" {
			result.IsWrongContentType = true
			return result
		}
	}

	opt := j.DefaultDecodeOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt.MaxSize >= 0 {
		maxSize := opt.MaxSize
		if maxSize == 0 {
			maxSize = MaxSizeDefault
		}

		r.Body = http.MaxBytesReader(w, r.Body, int64(maxSize))
	}

	result.JSONDecodeResult = *j.jsoner.Decode(r.Body, dst, jsoner.JSONDecodeOptions{
		UnknownFieldsAllowed: opt.UnknownFieldsAllowed,
	})

	return result
}
