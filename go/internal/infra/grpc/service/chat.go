package service

import (
	"github.com/leo-the-nardo/chatservice/internal/application/usecase/chatcompletionstream"
	"github.com/leo-the-nardo/chatservice/internal/infra/grpc/pb"
)

type ChatService struct {
	pb.UnimplementedChatServiceServer //gRPC boilerplate
	ChatCompletionStreamUseCase       chatcompletionstream.UseCase
	ChatConfigStream                  chatcompletionstream.ConfigInputDTO
	StreamChannel                     chan chatcompletionstream.OutputDTO
}

func NewChatService(useCase chatcompletionstream.UseCase, config chatcompletionstream.ConfigInputDTO, streamChannel chan chatcompletionstream.OutputDTO) *ChatService {
	return &ChatService{
		ChatCompletionStreamUseCase: useCase,
		ChatConfigStream:            config,
		StreamChannel:               streamChannel,
	}
}

func (this *ChatService) ChatStream(req *pb.ChatRequest, stream pb.ChatService_ChatStreamServer) error {
	chatConfig := chatcompletionstream.ConfigInputDTO{
		Model:                this.ChatConfigStream.Model,
		ModelMaxTokens:       this.ChatConfigStream.ModelMaxTokens,
		Temperature:          this.ChatConfigStream.Temperature,
		TopP:                 this.ChatConfigStream.TopP,
		N:                    this.ChatConfigStream.N,
		Stop:                 this.ChatConfigStream.Stop,
		MaxTokens:            this.ChatConfigStream.MaxTokens,
		PresencePenalty:      this.ChatConfigStream.PresencePenalty,
		FrequencyPenalty:     this.ChatConfigStream.FrequencyPenalty,
		InitialSystemMessage: this.ChatConfigStream.InitialSystemMessage,
	}

	input := &chatcompletionstream.InputDTO{
		ChatID:      req.GetChatId(),
		UserID:      req.GetUserId(),
		UserMessage: req.GetUserMessage(),
		Config:      chatConfig,
	}

	ctx := stream.Context()

	go func() {
		for msg := range this.StreamChannel {
			stream.Send(&pb.ChatResponse{
				ChatId:  msg.ChatID,
				UserId:  msg.UserID,
				Content: msg.Content,
			})
		}
	}()

	_, err := this.ChatCompletionStreamUseCase.Execute(input, ctx)
	if err != nil {
		return err
	}
	return nil
}
