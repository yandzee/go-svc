package httputils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const MaxSizeDefault int = 1024 * 1024 // 1 MB

type Jsoner struct {
	MaxSize              int
	UnknownFieldsAllowed bool
}

type JSONDecodeResult struct {
	IsWrongContentType bool
	IsUnexpectedEOF    bool
	IsEmptyBody        bool
	IsMultipleJSONs    bool
	SyntaxError        *json.SyntaxError
	UnmarshalTypeError *json.UnmarshalTypeError
	MaxBytesError      *http.MaxBytesError
	UnknownError       error
}

func (jdr *JSONDecodeResult) AsHTTPStatus() (int, string) {
	switch {
	case jdr.IsWrongContentType:
		return http.StatusUnsupportedMediaType, "Content-Type is not application/json"
	case jdr.IsUnexpectedEOF:
		return http.StatusBadRequest, "Request body contains badly-formed JSON"
	case jdr.IsEmptyBody:
		return http.StatusBadRequest, "Request body must not be empty"
	case jdr.IsMultipleJSONs:
		return http.StatusBadRequest, "Request body must only contain a single JSON object"

	case jdr.SyntaxError != nil:
		return http.StatusBadRequest, fmt.Sprintf(
			"Request body contains badly-formed JSON (at position %d)",
			jdr.SyntaxError.Offset,
		)
	case jdr.UnmarshalTypeError != nil:
		return http.StatusBadRequest, fmt.Sprintf(
			"Request body contains an invalid value for the %q field (at position %d)",
			jdr.UnmarshalTypeError.Field,
			jdr.UnmarshalTypeError.Offset,
		)
	case jdr.MaxBytesError != nil:
		return http.StatusRequestEntityTooLarge, fmt.Sprintf(
			"Request body must not be larger than %d bytes",
			jdr.MaxBytesError.Limit,
		)
	case jdr.UnknownError != nil:
		return http.StatusInternalServerError, fmt.Sprintf(
			"Unhandled error on JSON decoding: %s",
			jdr.UnknownError.Error(),
		)
	}

	return http.StatusOK, ""
}

func (jdr *JSONDecodeResult) Err() error {
	st, msg := jdr.AsHTTPStatus()

	if len(msg) == 0 || st == http.StatusOK {
		return nil
	}

	return fmt.Errorf("%s", msg)
}

func (j *Jsoner) EncodeResponse(
	w http.ResponseWriter,
	d any,
	isManualErrHandling ...bool,
) error {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(d)

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

	if j.MaxSize >= 0 {
		maxSize := j.MaxSize
		if maxSize == 0 {
			maxSize = MaxSizeDefault
		}

		r.Body = http.MaxBytesReader(w, r.Body, int64(maxSize))
	}

	dec := json.NewDecoder(r.Body)

	if !j.UnknownFieldsAllowed {
		dec.DisallowUnknownFields()
	}

	err := dec.Decode(&dst)
	if err != nil {
		switch {
		case errors.Is(err, io.ErrUnexpectedEOF):
			result.IsUnexpectedEOF = true
		case errors.Is(err, io.EOF):
			result.IsEmptyBody = true
		case errors.As(err, &result.SyntaxError):
		case errors.As(err, &result.UnmarshalTypeError):
		case errors.As(err, &result.MaxBytesError):
		default:
			result.UnknownError = err
		}

		return result
	}

	err = dec.Decode(&struct{}{})

	switch {
	case errors.Is(err, io.EOF):
	default:
		result.IsMultipleJSONs = true
	}

	return result
}
