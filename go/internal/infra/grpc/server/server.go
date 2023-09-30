package server

import (
	"github.com/leo-the-nardo/chatservice/internal/application/usecase/chatcompletionstream"
	"github.com/leo-the-nardo/chatservice/internal/infra/grpc/pb"
	"github.com/leo-the-nardo/chatservice/internal/infra/grpc/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
)

type GRPCServer struct {
	ChatCompletionStreamUseCase chatcompletionstream.UseCase
	ChatConfigStream            chatcompletionstream.ConfigInputDTO
	ChatService                 service.ChatService
	Port                        string
	AuthToken                   string
	StreamChannel               chan chatcompletionstream.OutputDTO
}

func NewGRPCServer(
	useCase chatcompletionstream.UseCase,
	config chatcompletionstream.ConfigInputDTO,
	port string,
	authToken string,
	streamChannel chan chatcompletionstream.OutputDTO,
) *GRPCServer {
	chatService := service.NewChatService(useCase, config, streamChannel)
	return &GRPCServer{
		ChatCompletionStreamUseCase: useCase,
		ChatConfigStream:            config,
		ChatService:                 *chatService,
		Port:                        port,
		AuthToken:                   authToken,
		StreamChannel:               streamChannel,
	}
}

func (this *GRPCServer) Start() {
	opts := []grpc.ServerOption{
		grpc.StreamInterceptor(this.AuthInterceptor),
	}
	server := grpc.NewServer(opts...)
	pb.RegisterChatServiceServer(server, &this.ChatService)

	lis, err := net.Listen("tcp", ":"+this.Port)
	if err != nil {
		panic(err.Error())
	}

	if err := server.Serve(lis); err != nil {
		panic(err.Error())
	}

}

func (this *GRPCServer) AuthInterceptor(
	service any,
	serverStream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	ctx := serverStream.Context()
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "metadata is not provided")
	}
	token := meta.Get("authorization")
	if len(token) == 0 {
		return status.Error(codes.Unauthenticated, "authorization token is not provided")
	}
	if token[0] != this.AuthToken {
		return status.Error(codes.Unauthenticated, "invalid authorization token")
	}
	return handler(service, serverStream)
}
