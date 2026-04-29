package domain

// Health はバックエンドの稼働状態を表すドメインモデル。
// Spring Boot 側の /actuator/health 相当。
type Health struct {
	Status   string `json:"status"`
	DBStatus string `json:"db"`
}

const (
	StatusUp   = "UP"
	StatusDown = "DOWN"
)
