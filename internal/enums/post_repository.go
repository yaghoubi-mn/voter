package enums

type SortBy string

const (
	SortByScore SortBy = "score desc"
	SortByDate  SortBy = "created_at desc"
	Trending    SortBy = "trending"
	DefaultSort SortBy = ""
)
