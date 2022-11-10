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
	startCommand            string     = "start"
	addRent                 MenuOption = "Adicionar aluguel"
	addCleaning             MenuOption = "Adicionar faxina"
	addBill                 MenuOption = "Adicionar conta de luz"
	addCondo                MenuOption = "Adicionar conta de condomínio"
	addApartment            MenuOption = "Adicionar imóvel"
	addMiscellaneousExpense MenuOption = "Adicionar despesa diversa"
	addAmortization         MenuOption = "Adicionar amortizaçao"
	addFinancingInstallment MenuOption = "Adicionar pagamento de parcela do financiamento"
)

func isMessageAMenuOption(msg string) bool {
	return (msg == string(addRent) ||
		msg == string(addBill) ||
		msg == string(addCleaning) ||
		msg == string(addCondo) ||
		msg == string(addApartment) ||
		msg == string(addAmortization) ||
		msg == string(addMiscellaneousExpense) ||
		msg == string(addFinancingInstallment))
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
		tgbotapi.NewKeyboardButton(string(addAmortization)),
		tgbotapi.NewKeyboardButton(string(addMiscellaneousExpense)),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(string(addFinancingInstallment)),
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
		var msg tgbotapi.MessageConfig
		isMessage := update.Message != nil
		isCallback := update.CallbackQuery != nil
		var msgText string
		var chatId int64
		if isMessage { // If we got a message
			chatId, msgText = update.Message.Chat.ID, update.Message.Text
		} else if isCallback {
			chatId, msgText = update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data
		}

		msg = tgbotapi.NewMessage(chatId, msgText)

		if isMessage && update.Message.IsCommand() && update.Message.Command() == "start" {
			msg.ReplyMarkup = numericKeyboard
		} else if isMessageAMenuOption(msgText) {
			msg.Text, msg.ReplyMarkup = initChatSession(chatId, msgText, store)
		} else {
			if _, ok := chatSessions[chatId]; ok {
				msg.Text, msg.ReplyMarkup = chatSessions[chatId].Next(msgText)
			}
		}

		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

func initChatSession(chatId int64, msgText string, store storage.Store) (string, interface{}) {
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
	case addMiscellaneousExpense:
		chatSession = chat_flow.NewChatSession[models.MiscellaneousExpense](chatId, store)
	case addAmortization:
		chatSession = chat_flow.NewChatSession[models.Amortization](chatId, store)
	case addFinancingInstallment:
		chatSession = chat_flow.NewChatSession[models.FinancingInstallment](chatId, store)
	}

	if chatSession != nil {
		chatSessions[chatId] = chatSession
		return chatSession.Next(msgText)
	}

	return "", nil
}
