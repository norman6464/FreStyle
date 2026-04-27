package domain

// RankingEntry はランキング 1 行。
type RankingEntry struct {
	UserID       uint64  `json:"userId"`
	DisplayName  string  `json:"displayName"`
	AverageScore float64 `json:"averageScore"`
	Rank         int     `json:"rank"`
}
