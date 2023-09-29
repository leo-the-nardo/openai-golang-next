package web

import (
	"encoding/json"
	"github.com/leo-the-nardo/chatservice/internal/application/usecase/chatcompletion"
	"io"
	"net/http"
)

type ChatGPTHandler struct {
	CompletionUseCase *chatcompletion.UseCase
	Config            chatcompletion.ConfigInputDTO
	AuthToken         string
}

func NewWebChatGPTHandler(useCase *chatcompletion.UseCase, config chatcompletion.ConfigInputDTO, authToken string) *ChatGPTHandler {
	return &ChatGPTHandler{
		CompletionUseCase: useCase,
		Config:            config,
		AuthToken:         authToken,
	}
}

func (this *ChatGPTHandler) Handle(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if req.Header.Get("Authorization") != this.AuthToken {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if !json.Valid(body) {
		http.Error(res, "invalid json", http.StatusBadRequest)
		return
	}
	var inputDTO chatcompletion.InputDTO
	err = json.Unmarshal(body, &inputDTO)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	inputDTO.Config = this.Config
	result, err := this.CompletionUseCase.Execute(inputDTO, req.Context())
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(result)
}
