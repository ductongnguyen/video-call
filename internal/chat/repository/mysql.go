package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"video-call/internal/chat"
	"video-call/internal/models"
	"gorm.io/gorm"
)

// repo implements the chat.Repository interface.
type repo struct {
	db *gorm.DB
}

// NewRepository is the constructor for repo.
func NewRepository(db *gorm.DB) chat.Repository {
	return &repo{db: db}
}

// CreateConversation implements chat.Repository.
func (r *repo) CreateConversation(ctx context.Context, conversation *models.Conversation) error {
	if conversation.ID == "" {
		conversation.ID = uuid.New().String()
	}
	if conversation.CreatedAt.IsZero() {
		conversation.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(conversation).Error
}

// GetConversationByID implements chat.Repository.
func (r *repo) GetConversationByID(ctx context.Context, conversationID string) (*models.Conversation, error) {
	var conversation models.Conversation
	if err := r.db.WithContext(ctx).First(&conversation, "id = ?", conversationID).Error; err != nil {
		return nil, err
	}
	return &conversation, nil
}

// GetConversationsByUserID implements chat.Repository.
func (r *repo) GetConversationsByUserID(ctx context.Context, userID string) ([]*models.Conversation, error) {
	var conversations []*models.Conversation
	if err := r.db.WithContext(ctx).
		Joins("JOIN conversation_participants cp ON cp.conversation_id = conversations.id").
		Where("cp.user_id = ?", userID).
		Order("conversations.updated_at DESC").
		Find(&conversations).Error; err != nil {
		return nil, err
	}
	return conversations, nil
}

// UpdateConversation implements chat.Repository.
func (r *repo) UpdateConversation(ctx context.Context, conversation *models.Conversation) error {
	return r.db.WithContext(ctx).Save(conversation).Error
}

// DeleteConversation implements chat.Repository.
func (r *repo) DeleteConversation(ctx context.Context, conversationID string) error {
	return r.db.WithContext(ctx).Delete(&models.Conversation{}, "id = ?", conversationID).Error
}

// IsUserInConversation implements chat.Repository.
func (r *repo) IsUserInConversation(ctx context.Context, userID, conversationID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ConversationParticipant{}).
		Where("user_id = ? AND conversation_id = ?", userID, conversationID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *repo) CreateMessage(ctx context.Context, message models.Message) error {
	if message.ID == "" {
		message.ID = uuid.New().String()
	}
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(&message).Error
}

// GetMessages implements chat.Repository.
func (r *repo) GetMessages(ctx context.Context, conversationID string, limit, offset int) ([]*models.Message, error) {
	var messages []*models.Message

	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	return messages, nil
}
