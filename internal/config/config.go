package config

import "time"

const (
	JWTRefreshExpireTime = 7 * 24 * time.Hour
	JWTAccessExpireTime  = 10 * time.Minute

	PageLimit = 20

	TimeFormat = "2006-01-02 15:04"
)
