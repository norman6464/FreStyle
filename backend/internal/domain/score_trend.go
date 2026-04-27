package domain

// ScoreTrendPoint は時系列でのスコア推移 1 ポイント。
type ScoreTrendPoint struct {
	Date         string  `json:"date"`
	OverallScore float64 `json:"overallScore"`
}

// ScoreTrend はユーザーのスコア推移。
type ScoreTrend struct {
	UserID uint64            `json:"userId"`
	Points []ScoreTrendPoint `json:"points"`
}
