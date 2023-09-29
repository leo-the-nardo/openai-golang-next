package chatcompletion

import (
	"context"
	"errors"
	"github.com/leo-the-nardo/chatservice/internal/domain/entity"
	"github.com/leo-the-nardo/chatservice/internal/domain/gateway"
	openai "github.com/sashabaranov/go-openai"
)

type ConfigInputDTO struct {
	Model                string   `json:"model"`
	ModelMaxTokens       int      `json:"model_max_tokens"`
	Temperature          float32  `json:"temperature"`
	TopP                 float32  `json:"top_p"`
	N                    int      `json:"n"`
	Stop                 []string `json:"stop"`
	MaxTokens            int      `json:"max_tokens"`
	PresencePenalty      float32  `json:"presence_penalty"`
	FrequencyPenalty     float32  `json:"frequency_penalty"`
	InitialSystemMessage string   `json:"initial_system_message"`
}

type InputDTO struct {
	ChatID      string `json:"chat_id"`
	UserID      string `json:"user_id"`
	UserMessage string `json:"user_message"`
	Config      ConfigInputDTO
}

type OutputDTO struct {
	ChatID  string `json:"chat_id"`
	UserID  string `json:"user_id"`
	Content string `json:"content"`
}

type UseCase struct {
	chatGateway  gateway.ChatGateway
	openAiClient *openai.Client
	stream       chan OutputDTO
}

func NewChatCompletionUseCase(
	chatGateway gateway.ChatGateway,
	openAiClient *openai.Client,
) *UseCase {
	useCase := &UseCase{
		chatGateway:  chatGateway,
		openAiClient: openAiClient,
	}
	return useCase
}

func (this *UseCase) Execute(
	input InputDTO,
	ctx context.Context,
) (*OutputDTO, error) {
	chat, err := this.getOrCreateChat(input, ctx)
	if err != nil {
		return nil, err
	}
	userMessage, err := entity.NewMessage("user", input.UserMessage, chat.Config.Model)
	if err != nil {
		return nil, errors.New("failed to add user message:" + err.Error())
	}
	err = chat.AddMessage(userMessage)
	if err != nil {
		return nil, errors.New("failed to create user message:" + err.Error())
	}

	var messages []openai.ChatCompletionMessage
	for _, msg := range chat.Messages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	resp, err := this.openAiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:            chat.Config.Model.GetName(),
			Messages:         messages,
			MaxTokens:        chat.Config.MaxTokens,
			Temperature:      chat.Config.Temperature,
			TopP:             chat.Config.TopP,
			N:                chat.Config.N,
			Stop:             chat.Config.Stop,
			PresencePenalty:  chat.Config.PresencePenalty,
			FrequencyPenalty: chat.Config.FrequencyPenalty,
		},
	)
	if err != nil {
		return nil, errors.New("failed to create chat completion stream:" + err.Error())
	}
	msgContent := resp.Choices[0].Message.Content

	assistant, err := entity.NewMessage("assistant", msgContent, chat.Config.Model)
	if err != nil {
		return nil, errors.New("failed to create assistant message:" + err.Error())
	}

	err = chat.AddMessage(assistant)
	if err != nil {
		return nil, errors.New("failed to add assistant message:" + err.Error())
	}

	err = this.chatGateway.Save(ctx, chat)
	if err != nil {
		return nil, errors.New("failed to save chat:" + err.Error())
	}

	return &OutputDTO{
		ChatID:  chat.ID,
		UserID:  input.UserID,
		Content: msgContent,
	}, nil
}

func (this *UseCase) getOrCreateChat(
	input InputDTO,
	ctx context.Context,
) (*entity.Chat, error) {
	chat, err := this.chatGateway.FindById(ctx, input.ChatID)
	if err != nil {
		return nil, errors.New("failed to get chat by user id:" + err.Error())
	}
	if chat == nil {
		chat, err = createNewChat(input)
		if err != nil {
			return nil, errors.New("failed to create new chat:" + err.Error())
		}
		err = this.chatGateway.Create(ctx, chat)
		if err != nil {
			return nil, errors.New("failed to persist chat:" + err.Error())
		}
	}
	return chat, nil

}

func createNewChat(input InputDTO) (*entity.Chat, error) {
	model := entity.NewModel(input.Config.Model, input.Config.ModelMaxTokens)
	initialMessage, err := entity.NewMessage("system", input.Config.InitialSystemMessage, model)
	if err != nil {
		return nil, errors.New("failed to create initial message:" + err.Error())
	}
	config := &entity.ChatConfig{
		Model:            model,
		Temperature:      input.Config.Temperature,
		TopP:             input.Config.TopP,
		N:                input.Config.N,
		Stop:             input.Config.Stop,
		MaxTokens:        input.Config.MaxTokens,
		PresencePenalty:  input.Config.PresencePenalty,
		FrequencyPenalty: input.Config.FrequencyPenalty,
	}
	chat, err := entity.NewChat(input.UserID, initialMessage, config)
	if err != nil {
		return nil, errors.New("failed to create new chat:" + err.Error())
	}
	return chat, nil
}
