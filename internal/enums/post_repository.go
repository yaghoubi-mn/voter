package enums

type SortBy string

const (
	SortByScore SortBy = "DESC score"
	SortByDate  SortBy = "DESC modified_at"
	DefaultSort SortBy = ""
)
