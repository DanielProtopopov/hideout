package middleware

type (
	Impersonate struct {
		UserID      uint   `json:"UserID"`
		Auth0UserID string `json:"Auth0UserID"`
	}
	// CustomClaims contains custom data we want from the token.
	CustomClaims struct {
		Scope string `json:"scope"`
	}
)
