package domain

// Health はバックエンドの稼働状態を表す。
type Health struct {
	Status   string `json:"status"`
	DBStatus string `json:"db"`
}

const (
	StatusUp   = "UP"
	StatusDown = "DOWN"
)
