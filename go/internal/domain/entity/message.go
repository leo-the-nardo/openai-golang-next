package entity

import (
	"errors"
	"github.com/google/uuid"
	"github.com/pkoukk/tiktoken-go"
	"time"
)

type Message struct {
	ID        string
	Role      string
	Content   string
	Tokens    int
	Model     *Model
	CreatedAt time.Time
}

func NewMessage(role string, content string, model *Model) (*Message, error) {
	tokens, err := countTokens(content, model)
	msg := &Message{
		ID:        uuid.NewString(),
		Role:      role,
		Content:   content,
		Tokens:    tokens,
		Model:     model,
		CreatedAt: time.Now(),
	}
	err = msg.validate()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (this *Message) validate() error {
	if this.Role != "user" && this.Role != "system" && this.Role != "assistant" {
		return errors.New("invalid role")
	}
	if this.Content == "" {
		return errors.New("content is empty")
	}
	if this.CreatedAt.IsZero() {
		return errors.New("created_at is empty")
	}
	return nil
}

func countTokens(content string, model *Model) (int, error) {
	tkm, err := tiktoken.EncodingForModel(model.GetName())
	if err != nil {
		return 0, err
	}
	return len(tkm.Encode(content, nil, nil)), nil
}

func (this *Message) GetCountTokens() int {
	return this.Tokens
}
