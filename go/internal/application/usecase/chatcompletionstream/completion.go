package chat_completion_stream

import (
	"context"
	"errors"
	"github.com/leo-the-nardo/chatservice/internal/domain/entity"
	"github.com/leo-the-nardo/chatservice/internal/domain/gateway"
	openai "github.com/sashabaranov/go-openai"
	"io"
	"strings"
)

type ChatCompletionConfigInputDTO struct {
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

type ChatCompletionInputDTO struct {
	ChatID      string `json:"chat_id"`
	UserID      string `json:"user_id"`
	UserMessage string `json:"user_message"`
	Config      *ChatCompletionConfigInputDTO
}

type ChatCompletionOutputDTO struct {
	ChatID  string `json:"chat_id"`
	UserID  string `json:"user_id"`
	Content string `json:"content"`
}

type ChatCompletionUseCase struct {
	chatGateway  gateway.ChatGateway
	openAiClient *openai.Client
	stream       chan ChatCompletionOutputDTO
}

func NewChatCompletionUseCase(
	chatGateway gateway.ChatGateway,
	openAiClient *openai.Client,
	stream chan ChatCompletionOutputDTO,
) *ChatCompletionUseCase {
	useCase := &ChatCompletionUseCase{
		chatGateway:  chatGateway,
		openAiClient: openAiClient,
		stream:       stream,
	}
	return useCase
}

func (this *ChatCompletionUseCase) Execute(
	input *ChatCompletionInputDTO,
	ctx context.Context,
) (*ChatCompletionOutputDTO, error) {
	chat, err := this.chatGateway.FindById(ctx, input.ChatID)
	if err != nil {
		if err.Error() == "chat not found" { // not exists flow
			chat, err = createNewChat(input)
			err = this.chatGateway.Create(ctx, chat)
			if err != nil {
				return nil, errors.New("failed to persist new chat:" + err.Error())
			}
		} // unknown error fetch flow
		return nil, errors.New("failed to fetch chat:" + err.Error())
	}
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
	resp, err := this.openAiClient.CreateChatCompletionStream(
		ctx,
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
			Stream:           true,
		},
	)
	if err != nil {
		return nil, errors.New("failed to create chat completion stream:" + err.Error())
	}

	var fullResponse strings.Builder
	for {
		response, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, errors.New("failed to receive streaming response:" + err.Error())
		}
		fullResponse.WriteString(response.Choices[0].Delta.Content)
		r := ChatCompletionOutputDTO{
			ChatID:  chat.ID,
			UserID:  chat.UserID,
			Content: fullResponse.String(),
		}
		this.stream <- r
	}
	assistant, err := entity.NewMessage("assistant", fullResponse.String(), chat.Config.Model)
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

	return &ChatCompletionOutputDTO{
		ChatID:  chat.ID,
		UserID:  chat.UserID,
		Content: fullResponse.String(),
	}, nil
}

func createNewChat(input *ChatCompletionInputDTO) (*entity.Chat, error) {
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
