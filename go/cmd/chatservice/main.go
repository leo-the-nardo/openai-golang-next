package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/leo-the-nardo/chatservice/configs"
	"github.com/leo-the-nardo/chatservice/internal/application/usecase/chatcompletion"
	"github.com/leo-the-nardo/chatservice/internal/infra/repository"
	"github.com/leo-the-nardo/chatservice/internal/infra/web"
	"github.com/leo-the-nardo/chatservice/internal/infra/webserver"
	"github.com/sashabaranov/go-openai"
)

func main() {
	config := configs.LoadConfig(".")
	dbConn, err := sql.Open(
		config.DBDriver,
		fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
			config.DBUser,
			config.DBPassword,
			config.DBHost,
			config.DBPort,
			config.DBName,
		),
	)
	if err != nil {
		panic(err)
	}
	defer dbConn.Close()

	repo := repository.NewChatRepository(dbConn)
	client := openai.NewClient(config.OpenAIApiKey)

	chatConfig := chatcompletion.ConfigInputDTO{
		Model:                config.Model,
		ModelMaxTokens:       config.ModelMaxTokens,
		Temperature:          float32(config.Temperature),
		TopP:                 float32(config.TopP),
		N:                    config.N,
		Stop:                 config.Stop,
		MaxTokens:            config.MaxTokens,
		InitialSystemMessage: config.InitialChatMessage,
	}

	useCase := chatcompletion.NewChatCompletionUseCase(repo, client)

	app := webserver.NewWebServer(":" + config.WebServerPort)
	chatGPTHandler := web.NewWebChatGPTHandler(useCase, chatConfig, config.AuthToken)
	app.AddHandler("/chat", chatGPTHandler.Handle)

	fmt.Println("server running on port " + config.WebServerPort)
	app.Start()
}
