package request

import "github.com/google/uuid"

type CreateTokenRequest struct {
	UserID     uuid.UUID `json:"user_id"`
	Permission string    `json:"permission"`
}
