package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gustavolopess/hoteleiro/internal/chat_flow"
	"github.com/gustavolopess/hoteleiro/internal/config"
	"github.com/gustavolopess/hoteleiro/internal/models"
	"github.com/gustavolopess/hoteleiro/internal/storage"
)

type MenuOption string

const (
	startCommand string     = "start"
	addRent      MenuOption = "Adicionar aluguel"
	addCleaning  MenuOption = "Adicionar faxina"
	addBill      MenuOption = "Adicionar conta de luz"
	addCondo     MenuOption = "Adicionar condomínio"
	addApartment MenuOption = "Adicionar imóvel"
)

func isMessageAMenuOption(msg string) bool {
	return (msg == string(addRent) ||
		msg == string(addBill) ||
		msg == string(addCleaning) ||
		msg == string(addCondo) ||
		msg == string(addApartment))
}

var numericKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(string(addRent)),
		tgbotapi.NewKeyboardButton(string(addCleaning)),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(string(addBill)),
		tgbotapi.NewKeyboardButton(string(addCondo)),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(string(addApartment)),
	),
)

var chatSessions = make(map[int64]chat_flow.ChatSession)

func main() {
	bot, err := tgbotapi.NewBotAPI("5627586393:AAHHTc0W5Fjy-dC1CLejshG3ZbJK4va5--E")
	if err != nil {
		log.Panic(err)
	}

	cfg := config.LoadConfig()

	store := storage.NewStore(cfg)

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			var replyText string

			if update.Message.IsCommand() && update.Message.Command() == "start" {
				msg.ReplyMarkup = numericKeyboard
			} else if isMessageAMenuOption(update.Message.Text) {
				replyText = initChatSession(update.Message.Chat.ID, update.Message.Text, store)
			} else {
				if _, ok := chatSessions[update.Message.Chat.ID]; ok {
					replyText = chatSessions[update.Message.Chat.ID].Next(update.Message.Text)
				}
			}

			if len(replyText) > 0 {
				msg.Text = replyText
			}
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}

		}
	}
}

func initChatSession(chatId int64, msgText string, store storage.Store) string {
	var chatSession chat_flow.ChatSession

	switch MenuOption(msgText) {
	case addBill:
		chatSession = chat_flow.NewChatSession[models.EnergyBill](chatId, store)
	case addRent:
		chatSession = chat_flow.NewChatSession[models.Rent](chatId, store)
	case addCleaning:
		chatSession = chat_flow.NewChatSession[models.Cleaning](chatId, store)
	case addCondo:
		chatSession = chat_flow.NewChatSession[models.Condo](chatId, store)
	case addApartment:
		chatSession = chat_flow.NewChatSession[models.Apartment](chatId, store)
	}

	if chatSession != nil {
		chatSessions[chatId] = chatSession
		return chatSession.Next(msgText)
	}

	return ""
}
