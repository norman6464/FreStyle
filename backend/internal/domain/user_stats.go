package domain

// UserStats はセッション・スコアから計算する集計済みユーザー統計（テーブル直結ではない）。
type UserStats struct {
	UserID        uint64  `json:"userId"`
	TotalSessions int     `json:"totalSessions"`
	AverageScore  float64 `json:"averageScore"`
	StreakDays    int     `json:"streakDays"`
	LongestStreak int     `json:"longestStreak"`
}
