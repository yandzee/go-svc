package page

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const (
	DefaultLimitKey  = "limit"
	DefaultOffsetKey = "offset"
	DefaultLastKey   = "last"
)

var (
	ErrLimitParse  = errors.New("limit parse error")
	ErrOffsetParse = errors.New("offset parse error")
)

type Pager struct {
	LimitKey  string
	OffsetKey string
	LastKey   string
}

func (sr *Pager) FromHTTPRequest(r *http.Request) (Selector, error) {
	return sr.FromURLValues(r.URL.Query())
}

func (sr *Pager) FromURLValues(q url.Values) (Selector, error) {
	sel := Selector{}

	limitKey := sr.LimitKey
	if len(limitKey) == 0 {
		limitKey = DefaultLimitKey
	}

	if s := q.Get(limitKey); len(s) > 0 {
		limit, err := strconv.ParseInt(s, 10, 0)
		if err != nil {
			return sel, errors.Join(
				ErrLimitParse,
				fmt.Errorf("failed to parse limit `%s`", s),
				err,
			)
		}

		sel.Limit = new(int)
		*sel.Limit = int(limit)
	}

	offsetKey := sr.OffsetKey
	if len(offsetKey) == 0 {
		offsetKey = DefaultOffsetKey
	}

	if s := q.Get(offsetKey); len(s) > 0 {
		offset, err := strconv.ParseInt(s, 10, 0)
		if err != nil {
			return sel, errors.Join(
				ErrOffsetParse,
				fmt.Errorf("failed to parse offset `%s`", s),
				err,
			)
		}

		sel.Offset = new(int)
		*sel.Offset = int(offset)
	}

	lastKey := sr.LastKey
	if len(lastKey) == 0 {
		lastKey = DefaultLastKey
	}

	if s := q.Get(lastKey); len(s) > 0 {
		sel.Last = &s
	}

	return sel, nil
}
