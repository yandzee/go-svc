package page

type Selector struct {
	Limit  *int
	Offset *int
	Last   *string
}

func Entire() Selector {
	return NewSelector(0, 0)
}

func NewSelector(limit, offset int, last ...string) Selector {
	var sel *Selector
	return sel.Or(limit, offset, last...)
}

func (s *Selector) Or(limit, offset int, last ...string) Selector {
	sel := Selector{
		Limit:  &limit,
		Offset: &offset,
	}

	if len(last) > 0 {
		sel.Last = &last[0]
	}

	if s != nil {
		if s.Limit != nil {
			sel.Limit = s.Limit
		}

		if s.Offset != nil {
			sel.Offset = s.Offset
		}

		if s.Last != nil {
			sel.Last = s.Last
		}
	}

	return sel
}

func (s *Selector) LimitOr(limit int) int {
	if s == nil || s.Limit == nil || *s.Limit == 0 {
		return limit
	}

	return *s.Limit
}

func (s *Selector) OffsetOr(off int) int {
	if s == nil || s.Offset == nil || *s.Offset == 0 {
		return off
	}

	return *s.Offset
}

func (s *Selector) LastOr(ls string) string {
	if s == nil || s.Last == nil || len(*s.Last) == 0 {
		return ls
	}

	return *s.Last
}
