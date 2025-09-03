package page

import "math"

type Page[T any] struct {
	Data []T            `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type PaginationMeta struct {
	// Index of requested page
	CurrentPage int `json:"currentPage"`

	// Actual number of elements in requested page
	CurrentPageSize int `json:"currentPageSize"`

	// Upper bound for the size of the requested page
	PageSizeLimit int `json:"pageSizeLimit"`

	// Number of all existing pages corresponding to data query
	TotalPages int `json:"totalPages"`

	// Number of all existing data entries corresponding to data query
	TotalEntries int `json:"totalEntries"`
}

func (p *Page[T]) SetSingle(data []T) {
	p.Data = data

	sel := Entire()
	p.FillMeta(&sel, len(data))
}

func (p *Page[T]) FillMeta(sel *Selector, total int) {
	p.Meta.Fill(sel, total)
	p.Meta.CurrentPageSize = len(p.Data)
}

func (pm *PaginationMeta) Fill(sel *Selector, total int) {
	pm.PageSizeLimit = sel.LimitOr(0)
	pm.TotalEntries = total

	pageSize := float64(pm.PageSizeLimit)
	off := sel.OffsetOr(0)

	switch {
	case total == 0:
		pm.TotalPages = 0
	case pm.PageSizeLimit == 0:
		pm.TotalPages = 1
	default:
		pm.TotalPages = int(math.Ceil(float64(total) / pageSize))
	}

	switch {
	case off == 0:
		pm.CurrentPage = 0
	case pm.PageSizeLimit == 0:
		pm.CurrentPage = 0
	default:
		pm.CurrentPage = int(math.Floor(float64(off) / pageSize))
	}
}
