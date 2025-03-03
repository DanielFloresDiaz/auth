package models

type RateLimitValue struct {
	Tier  string `json:"tier" db:"tier"`
	Limit int    `json:"limit" db:"limit"`
}

type RateLimit struct {
	Seconds int `json:"seconds" db:"seconds"`

	RateLimits []RateLimitValue `json:"rate_limits" db:"rate_limits"`
}
