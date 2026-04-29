package domain

// UserStats は集計済みのユーザー統計。テーブル直結ではなく
// セッション・スコアからの計算結果として扱う。
type UserStats struct {
	UserID        uint64  `json:"userId"`
	TotalSessions int     `json:"totalSessions"`
	AverageScore  float64 `json:"averageScore"`
	StreakDays    int     `json:"streakDays"`
	LongestStreak int     `json:"longestStreak"`
}
