package chat_flow

import (
	"github.com/gustavolopess/hoteleiro/internal/models"
	"github.com/gustavolopess/hoteleiro/internal/storage"
)

type ChatSession interface {
	Next(string) string
}

type chatSession[T models.Models] struct {
	chatId   int64
	chatFlow Flow[T]
}

func NewChatSession[T models.Models](chatId int64, store storage.Store) ChatSession {
	return &chatSession[T]{
		chatId:   chatId,
		chatFlow: NewFlow[T](store),
	}
}

func (s *chatSession[T]) Next(answer string) string {
	return s.chatFlow.Next(answer)
}
