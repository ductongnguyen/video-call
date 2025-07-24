package http

import (
	"errors"
	"net/http"

	"video-call/internal/chat"
	"video-call/internal/models"
	"video-call/pkg/logger"
	"video-call/pkg/response"
	"video-call/pkg/utils"
	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for chat features
type Handler struct {
	chatUC chat.UseCase
	logger logger.Logger
}

// NewHandler creates a new chat HTTP handler
func NewHandler(chatUC chat.UseCase, logger logger.Logger) *Handler {
	return &Handler{
		chatUC: chatUC,
		logger: logger,
	}
}

// Use CreateConversationRequest from presenter.go

// getUserIDFromContext gets the user ID from the request context
func (h *Handler) getUserIDFromContext(c *gin.Context) (string, error) {
	user, err := utils.GetUserFromCtx(c.Request.Context())
	if err != nil {
		h.logger.Errorf(c.Request.Context(), "Failed to get user from context: %v", err)
		return "", err
	}
	return user.ID, nil
}

// validateConversationAccess checks if the user has access to the conversation
func (h *Handler) validateConversationAccess(c *gin.Context, userID, conversationID string) (bool, error) {
	if conversationID == "" {
		return false, errors.New("conversation ID is required")
	}

	isParticipant, err := h.chatUC.IsUserInConversation(c.Request.Context(), userID, conversationID)
	if err != nil {
		h.logger.Errorf(c.Request.Context(), "Failed to check conversation participation: %v", err)
		return false, err
	}

	if !isParticipant {
		c.AbortWithStatus(http.StatusForbidden)
		return false, nil
	}

	return true, nil
}

// CreateConversation creates a new conversation
func (h *Handler) CreateConversation(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		h.logger.Errorf(c.Request.Context(), "Failed to get user ID from context: %v", err)
		response.WithError(c, err)
		return
	}

	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf(c.Request.Context(), "Failed to bind request body: %v", err)
		response.WithError(c, err)
		return
	}

	// MemberIDs are already strings in the request

	// Create conversation
	name := req.Name // Create a copy to take address of
	conversation := &models.Conversation{
		IsGroup: req.IsGroup,
		Name:    &name,
	}

	// Set the creator as the first participant
	conversation.CreatedBy = userID

	// Create conversation
	if err := h.chatUC.CreateConversation(c.Request.Context(), conversation); err != nil {
		h.logger.Errorf(c.Request.Context(), "Failed to create conversation: %v", err)
		response.WithError(c, err)
		return
	}

	// Add members to conversation
	if len(req.MemberIDs) > 0 {
		h.logger.Warnf(c.Request.Context(), "Adding participants is not yet implemented")
		// TODO: Implement AddParticipants in use case and repository
		/*
		if err := h.chatUC.AddParticipants(c.Request.Context(), conversation.ID, req.MemberIDs); err != nil {
			h.logger.Errorf(c.Request.Context(), "Failed to add participants: %v", err)
		}
		*/
	}

	response.WithCode(c, http.StatusCreated, toConversationResponse(conversation))
}

// GetConversations gets all conversations for the current user
func (h *Handler) GetConversations(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		response.WithError(c, err)
		return
	}

	// Get conversations for the user
	conversations, err := h.chatUC.GetConversationsByUserID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Errorf(c.Request.Context(), "Failed to get user conversations: %v", err)
		response.WithError(c, err)
		return
	}

	// Convert to response models
	responseConvs := make([]ConversationResponse, len(conversations))
	for i, conv := range conversations {
		responseConvs[i] = toConversationResponse(conv)
	}

	response.WithData(c, http.StatusOK, responseConvs)
}

// GetConversation gets a conversation by ID
func (h *Handler) GetConversation(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		response.WithError(c, err)
		return
	}

	conversationID := c.Param("id")
	if ok, err := h.validateConversationAccess(c, userID, conversationID); !ok {
		if err != nil {
			response.WithError(c, err)
		}
		return
	}

	// Get conversation by ID
	conversation, err := h.chatUC.GetConversationByID(c.Request.Context(), conversationID)
	if err != nil {
		h.logger.Errorf(c.Request.Context(), "Failed to get conversation: %v", err)
		response.WithError(c, err)
		return
	}

	response.WithData(c, http.StatusOK, toConversationResponse(conversation))
}

// UpdateConversation updates a conversation
func (h *Handler) UpdateConversation(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		response.WithError(c, err)
		return
	}

	conversationID := c.Param("id")
	if ok, err := h.validateConversationAccess(c, userID, conversationID); !ok {
		if err != nil {
			response.WithError(c, err)
		}
		return
	}

	var req struct {
		Name *string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf(c.Request.Context(), "Failed to bind request body: %v", err)
		response.WithError(c, err)
		return
	}

	// Get existing conversation
	conversation, err := h.chatUC.GetConversationByID(c.Request.Context(), conversationID)
	if err != nil {
		h.logger.Errorf(c.Request.Context(), "Failed to get conversation: %v", err)
		response.WithError(c, err)
		return
	}

	// Update fields if provided
	if req.Name != nil {
		conversation.Name = req.Name
	}

	if err := h.chatUC.UpdateConversation(c.Request.Context(), conversation); err != nil {
		response.WithError(c, err)
		return
	}

	response.WithData(c, http.StatusOK, toConversationResponse(conversation))
}

// DeleteConversation deletes a conversation
func (h *Handler) DeleteConversation(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		response.WithError(c, err)
		return
	}

	conversationID := c.Param("id")
	if ok, err := h.validateConversationAccess(c, userID, conversationID); !ok {
		if err != nil {
			response.WithError(c, err)
		}
		return
	}

	if err := h.chatUC.DeleteConversation(c.Request.Context(), conversationID); err != nil {
		h.logger.Errorf(c.Request.Context(), "Failed to delete conversation: %v", err)
		response.WithError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// SendMessageRequest represents the request body for sending a message
// Use SendMessageRequest from presenter.go

// SendMessage sends a message to a conversation
func (h *Handler) SendMessage(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		h.logger.Errorf(c.Request.Context(), "Failed to get user ID from context: %v", err)
		response.WithError(c, err)
		return
	}

	conversationID := c.Param("id")
	if conversationID == "" {
		err := errors.New("conversation ID is required")
		h.logger.Errorf(c.Request.Context(), "%v", err)
		response.WithError(c, err)
		return
	}

	// Check if user is a participant in the conversation
	isParticipant, err := h.chatUC.IsUserInConversation(c.Request.Context(), userID, conversationID)
	if err != nil || !isParticipant {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf(c.Request.Context(), "Failed to bind request body: %v", err)
		response.WithError(c, err)
		return
	}

	// Create message
	message := &models.Message{
		ConversationID: conversationID,
		SenderID:       userID,
		Content:        req.Content,
	}

	if err := h.chatUC.CreateMessage(c.Request.Context(), *message); err != nil {
		h.logger.Errorf(c.Request.Context(), "Failed to send message: %v", err)
		response.WithError(c, err)
		return
	}

	response.WithCode(c, http.StatusCreated, toMessageResponse(message))
}

// GetMessagesRequest represents the query parameters for getting messages
type GetMessagesRequest struct {
	Limit  int `form:"limit,default=20"` // Number of messages to return (default: 20, max: 100)
	Offset int `form:"offset,default=0"` // Offset for pagination (default: 0)
}

// GetMessages gets messages for a conversation with pagination
func (h *Handler) GetMessages(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		response.WithError(c, err)
		return
	}

	conversationID := c.Param("id")
	if ok, err := h.validateConversationAccess(c, userID, conversationID); !ok {
		if err != nil {
			response.WithError(c, err)
		}
		return
	}

	// Parse query parameters
	var req GetMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.WithError(c, err)
		return
	}

	// Validate limit (max 100 messages per request)
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Get messages from use case
	messages, err := h.chatUC.GetMessages(c.Request.Context(), conversationID, req.Limit, req.Offset)
	if err != nil {
		h.logger.Errorf(c.Request.Context(), "Failed to get messages: %v", err)
		response.WithError(c, err)
		return
	}

	response.WithData(c, http.StatusOK, messages)
}
