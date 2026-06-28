package request

type CreateTokenRequest struct {
	UserID     int64  `json:"user_id"`
	Permission string `json:"permission"`
}
