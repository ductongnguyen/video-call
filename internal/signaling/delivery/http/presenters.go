package http

import "video-call/internal/models"

type callRequest struct {
	CallerID string `json:"caller_id" binding:"required,uuid"`
	CalleeID string `json:"callee_id" binding:"required,uuid"`
}

type callResponse struct {
	Call *models.Call `json:"call"`
	Role string       `json:"role"`
}
