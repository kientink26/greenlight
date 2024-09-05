package data

import (
	"math"
	"strings"

	"github.com/kientink26/greenlight/internal/validator"
)

type Filter struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

func ValidateFilter(v *validator.Validator, filter Filter) {
	v.Check(filter.Page > 0, "page", "must be positive")
	v.Check(filter.Page <= 1_000_000, "page", "must not be greater than 1 million")

	v.Check(filter.PageSize > 0, "page_size", "must be positive")
	v.Check(filter.PageSize <= 100, "page", "must not be greater than 100")

	v.Check(validator.Has(filter.SortSafeList, filter.Sort), "sort", "invalid sort value")
}

func (f Filter) sortColumn() string {
	for _, col := range f.SortSafeList {
		if col == f.Sort {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("invalid sort value " + f.Sort)
}

func (f Filter) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func (f Filter) limit() int {
	return f.PageSize
}

func (f Filter) offset() int {
	return (f.Page - 1) * f.PageSize
}

type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

func calculateMetadata(totalRecords int, f Filter) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}
	return Metadata{
		CurrentPage:  f.Page,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(f.PageSize))),
		PageSize:     f.PageSize,
		TotalRecords: totalRecords,
	}
}
