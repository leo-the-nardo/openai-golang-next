package gateway

import (
	"context"
	"github.com/leo-the-nardo/chatservice/internal/domain/entity"
)

type ChatGateway interface {
	Create(ctx context.Context, chat *entity.Chat) error
	FindById(ctx context.Context, id string) (*entity.Chat, error)
	Save(ctx context.Context, chat *entity.Chat) error
}
