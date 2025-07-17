package middleware

type (
	UserInfo struct {
		UserID    uint  `json:"UserID"`
		ExpiresAt int64 `json:"ExpiresAt"`
	}
)
