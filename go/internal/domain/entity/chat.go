package entity

import (
	"errors"
	"github.com/google/uuid"
)

type ChatConfig struct {
	Model            *Model
	Temperature      float32
	TopP             float32
	N                int
	Stop             []string
	MaxTokens        int
	PresencePenalty  float32
	FrequencyPenalty float32
}

type Chat struct {
	ID                   string
	UserID               string
	InitialSystemMessage *Message
	Messages             []*Message
	ErasedMessages       []*Message
	Status               string
	TokenUsage           int
	Config               *ChatConfig
}

func NewChat(userID string, initialSystemMessage *Message, config *ChatConfig) (*Chat, error) {
	chat := &Chat{
		ID:                   uuid.NewString(),
		UserID:               userID,
		InitialSystemMessage: initialSystemMessage,
		Status:               "active",
		Config:               config,
		TokenUsage:           0,
	}
	err := chat.AddMessage(initialSystemMessage)
	err = chat.validate()
	if err != nil {
		return nil, err
	}
	return chat, nil
}

func (this *Chat) validate() error {
	if this.UserID == "" {
		return errors.New("user_id is empty")
	}
	if this.InitialSystemMessage == nil {
		return errors.New("initial_system_message is empty")
	}
	if this.Status != "active" && this.Status != "closed" {
		return errors.New("invalid status")
	}
	if this.Config.Temperature < 0.0 || this.Config.Temperature > 2.0 {
		return errors.New("invalid temperature")
	}
	if this.Config.TopP < 0.0 || this.Config.TopP > 1.0 {
		return errors.New("invalid top_p")
	}

	return nil
}

func (this *Chat) AddMessage(message *Message) error {
	if this.Status == "closed" {
		return errors.New("chat is closed, no more messages allowed")
	}
	for {
		if this.Config.Model.GetMaxTokens() >= message.GetCountTokens()+this.TokenUsage {
			// not full flow
			this.Messages = append(this.Messages, message)
			this.refreshTokenUsage()
			break
		}
		// full flow (remove the oldest message while not enough space)
		this.ErasedMessages = append(this.ErasedMessages, this.Messages[0])
		this.Messages = this.Messages[1:]
		this.refreshTokenUsage()
	}
	return nil
}

func (this *Chat) GetMessages() []*Message {
	return this.Messages
}

func (this *Chat) CountMessages() int {
	return len(this.Messages)
}

func (this *Chat) Close() {
	this.Status = "closed"
}

func (this *Chat) refreshTokenUsage() {
	this.TokenUsage = 0
	for _, message := range this.Messages {
		this.TokenUsage += message.GetCountTokens()
	}
}
