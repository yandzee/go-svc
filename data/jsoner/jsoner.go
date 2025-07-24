package jsoner

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var NoError = ""

type Jsoner struct{}

type JSONDecodeOptions struct {
	UnknownFieldsAllowed bool
}

type JSONDecodeResult struct {
	IsUnexpectedEOF    bool
	IsEmptyInput       bool
	IsMultipleJSONs    bool
	SyntaxError        *json.SyntaxError
	UnmarshalTypeError *json.UnmarshalTypeError
	UnknownError       error
}

func (jdr *JSONDecodeResult) Error() string {
	switch {
	case jdr.IsUnexpectedEOF:
		return "Input contains badly-formed JSON"
	case jdr.IsEmptyInput:
		return "Input is empty"
	case jdr.SyntaxError != nil:
		return fmt.Sprintf(
			"Input contains badly-formed JSON (at position %d)",
			jdr.SyntaxError.Offset,
		)
	case jdr.UnmarshalTypeError != nil:
		return fmt.Sprintf(
			"Request body contains an invalid value for the %q field (at position %d)",
			jdr.UnmarshalTypeError.Field,
			jdr.UnmarshalTypeError.Offset,
		)
	case jdr.IsMultipleJSONs:
		return "Input must only contain a single JSON object"
	case jdr.UnknownError != nil:
		return fmt.Sprintf(
			"Unhandled error on JSON decoding: %s",
			jdr.UnknownError.Error(),
		)
	}

	return NoError
}

func (jdr *JSONDecodeResult) Err() error {
	msg := jdr.Error()

	if msg == NoError {
		return nil
	}

	return errors.New(msg)
}

func (j *Jsoner) Encode(w io.Writer, d any) error {
	return json.NewEncoder(w).Encode(d)
}

func (j *Jsoner) Decode(
	r io.Reader,
	dst any,
	opts ...JSONDecodeOptions,
) *JSONDecodeResult {
	result := &JSONDecodeResult{}
	dec := json.NewDecoder(r)

	if len(opts) > 0 && !opts[0].UnknownFieldsAllowed {
		dec.DisallowUnknownFields()
	}

	err := dec.Decode(&dst)
	switch {
	case err == nil:
		break
	case errors.Is(err, io.ErrUnexpectedEOF):
		result.IsUnexpectedEOF = true
	case errors.Is(err, io.EOF):
		result.IsEmptyInput = true
	case errors.As(err, &result.SyntaxError):
	case errors.As(err, &result.UnmarshalTypeError):
	default:
		result.UnknownError = err
	}

	if err != nil {
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
