package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/leo-the-nardo/chatservice/internal/domain/entity"
	"github.com/leo-the-nardo/chatservice/internal/infra/db"
	"time"
)

type ChatRepository struct {
	DB      *sql.DB
	Queries *db.Queries
}

func NewChatRepository(database *sql.DB) *ChatRepository {
	return &ChatRepository{
		DB:      database,
		Queries: db.New(database), //SQLC boilerplate
	}
}

func (this *ChatRepository) Create(ctx context.Context, chat *entity.Chat) error {
	err := this.Queries.CreateChat(ctx, db.CreateChatParams{
		ID:               chat.ID,
		UserID:           chat.UserID,
		InitialMessageID: chat.InitialSystemMessage.ID,
		Status:           chat.Status,
		TokenUsage:       int32(chat.TokenUsage),
		Model:            chat.Config.Model.GetName(),
		ModelMaxTokens:   int32(chat.Config.Model.GetMaxTokens()),
		Temperature:      float64(chat.Config.Temperature),
		TopP:             float64(chat.Config.TopP),
		N:                int32(chat.Config.N),
		Stop:             chat.Config.Stop[0],
		MaxTokens:        int32(chat.Config.MaxTokens),
		PresencePenalty:  float64(chat.Config.PresencePenalty),
		FrequencyPenalty: float64(chat.Config.FrequencyPenalty),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	})
	if err != nil {
		return err
	}
	err = this.Queries.AddMessage(ctx, db.AddMessageParams{
		ID:        chat.InitialSystemMessage.ID,
		ChatID:    chat.ID,
		Content:   chat.InitialSystemMessage.Content,
		Role:      chat.InitialSystemMessage.Role,
		Tokens:    int32(chat.InitialSystemMessage.Tokens),
		CreatedAt: time.Now(),
	})
	return err
}

func (this *ChatRepository) FindById(ctx context.Context, id string) (*entity.Chat, error) {
	if id == "" {
		return nil, nil
	}
	dbChat, err := this.Queries.FindChatById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	dbMessages, err := this.Queries.FindMessagesByChatId(ctx, id)
	if err != nil {
		return nil, err
	}
	erasedDBMessages, err := this.Queries.FindErasedMessagesByChatId(ctx, id)
	if err != nil {
		return nil, err
	}
	chat := toEntity(dbChat, dbMessages, erasedDBMessages)

	return chat, nil
}

func (this *ChatRepository) Save(ctx context.Context, chat *entity.Chat) error {
	params := db.SaveChatParams{
		ID:               chat.ID,
		UserID:           chat.UserID,
		Status:           chat.Status,
		TokenUsage:       int32(chat.TokenUsage),
		Model:            chat.Config.Model.GetName(),
		ModelMaxTokens:   int32(chat.Config.Model.GetMaxTokens()),
		Temperature:      float64(chat.Config.Temperature),
		TopP:             float64(chat.Config.TopP),
		N:                int32(chat.Config.N),
		Stop:             chat.Config.Stop[0],
		MaxTokens:        int32(chat.Config.MaxTokens),
		PresencePenalty:  float64(chat.Config.PresencePenalty),
		FrequencyPenalty: float64(chat.Config.FrequencyPenalty),
		UpdatedAt:        time.Now(),
	}
	err := this.Queries.SaveChat(
		ctx,
		params,
	)
	if err != nil {
		return err
	}
	// delete messages
	err = this.Queries.DeleteChatMessages(ctx, chat.ID)
	if err != nil {
		return err
	}
	// delete erased messages
	err = this.Queries.DeleteErasedChatMessages(ctx, chat.ID)
	if err != nil {
		return err
	}
	// save messages
	i := 0
	for _, message := range chat.Messages {
		err = this.Queries.AddMessage(
			ctx,
			db.AddMessageParams{
				ID:        message.ID,
				ChatID:    chat.ID,
				Content:   message.Content,
				Role:      message.Role,
				Tokens:    int32(message.Tokens),
				Model:     chat.Config.Model.GetName(),
				CreatedAt: message.CreatedAt,
				OrderMsg:  int32(i),
				Erased:    false,
			},
		)
		if err != nil {
			return err
		}
		i++
	}
	// save erased messages
	i = 0
	for _, message := range chat.ErasedMessages {
		err = this.Queries.AddMessage(
			ctx,
			db.AddMessageParams{
				ID:        message.ID,
				ChatID:    chat.ID,
				Content:   message.Content,
				Role:      message.Role,
				Tokens:    int32(message.Tokens),
				Model:     chat.Config.Model.GetName(),
				CreatedAt: message.CreatedAt,
				OrderMsg:  int32(i),
				Erased:    true,
			},
		)
		if err != nil {
			return err
		}
		i++
	}
	return nil

}

func toEntity(dbChat db.Chat, dbMessages []db.Message, erasedDbMessages []db.Message) *entity.Chat {
	var messages []*entity.Message
	for _, dbMessage := range dbMessages {
		messages = append(messages, &entity.Message{
			ID:        dbMessage.ID,
			Content:   dbMessage.Content,
			Role:      dbMessage.Role,
			CreatedAt: dbMessage.CreatedAt,
			Model:     entity.NewModel(dbMessage.Model, int(dbChat.ModelMaxTokens)),
			Tokens:    int(dbMessage.Tokens)},
		)
	}

	var erasedMessages []*entity.Message
	for _, dbMessage := range erasedDbMessages {
		erasedMessages = append(erasedMessages, &entity.Message{
			ID:        dbMessage.ID,
			Content:   dbMessage.Content,
			Role:      dbMessage.Role,
			CreatedAt: dbMessage.CreatedAt,
			Model:     entity.NewModel(dbMessage.Model, int(dbChat.ModelMaxTokens)),
			Tokens:    int(dbMessage.Tokens)},
		)
	}

	chat := &entity.Chat{
		ID:             dbChat.ID,
		UserID:         dbChat.UserID,
		Status:         dbChat.Status,
		TokenUsage:     int(dbChat.TokenUsage),
		Messages:       messages,
		ErasedMessages: erasedMessages,
		Config: &entity.ChatConfig{
			Model:            entity.NewModel(dbChat.Model, int(dbChat.ModelMaxTokens)),
			Temperature:      float32(dbChat.Temperature),
			TopP:             float32(dbChat.TopP),
			N:                int(dbChat.N),
			Stop:             []string{dbChat.Stop},
			MaxTokens:        int(dbChat.MaxTokens),
			PresencePenalty:  float32(dbChat.PresencePenalty),
			FrequencyPenalty: float32(dbChat.FrequencyPenalty),
		},
	}
	return chat
}
