package models

// TokenPair is the service-level token response used by login and refresh.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// LoginResult is the service-level login payload.
type LoginResult struct {
	Account AccountResponse `json:"account"`
	Tokens  TokenPair       `json:"tokens"`
}
