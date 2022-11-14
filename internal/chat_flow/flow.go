package chat_flow

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/gustavolopess/hoteleiro/internal/format"
	"github.com/gustavolopess/hoteleiro/internal/models"
	"github.com/gustavolopess/hoteleiro/internal/storage"
)

type Flow[T models.Models] interface {
	Next(string) (string, interface{})
}

type Step int64

const (
	stepBeginEnergyBill Step = iota
	stepGetValueEnergyBill
	stepGetDateEnergyBill
	stepGetPayerEnergyBill

	stepBeginRent
	stepGetValueRent
	stepGetDateBeginRent
	stepGetDateEndRent
	stepGetRenter
	stepGetRentReceiver

	stepBeginCleaning
	stepGetValueCleaning
	stepGetDateCleaning
	stepGetCleaningPayer

	stepBeginCondo
	stepGetValueCondo
	stepGetDateCondo
	stepGetPayerCondo

	stepBeginApartment
	stepGetNameApartment
	stepGetAddressApartment

	stepBeginMiscellaneousExpense
	stepGetDescriptionMiscellaneousExpense
	stepGetValueMiscellaneousExpense
	stepGetDateMiscellaneousExpense
	stepGetPayerMiscellaneousExpense

	stepBeginAmortization
	stepGetPayerAmortization
	stepGetValueAmortization
	stepGetDateAmortization

	stepBeginFinancingInstallment
	stepGetFinancialInstallmentDate
	stepGetFinancialInstallmentValue
	stepGetFinancialInstallmentPayer

	stepEnd
)

type flow[T models.Models] struct {
	store               storage.Store
	step                Step
	askedApartment      bool
	apartmentName       string
	availableApartments []string
	value               any
	currentFlow         func(string) (string, interface{})
}

func NewFlow[T models.Models](store storage.Store) Flow[T] {
	f := &flow[T]{
		store: store,
	}

	var b T
	switch any(b).(type) {
	case models.EnergyBill:
		f.step = stepBeginEnergyBill
		f.currentFlow = f.energyBillFlow
	case models.Rent:
		f.step = stepBeginRent
		f.currentFlow = f.rentFlow
	case models.Cleaning:
		f.step = stepBeginCleaning
		f.currentFlow = f.cleaningFlow
	case models.Condo:
		f.step = stepBeginCondo
		f.currentFlow = f.condoFlow
	case models.Apartment:
		f.step = stepBeginApartment
		f.currentFlow = f.apartmentFlow
	case models.MiscellaneousExpense:
		f.step = stepBeginMiscellaneousExpense
		f.currentFlow = f.miscellaneousExpenseFlow
	case models.Amortization:
		f.step = stepBeginAmortization
		f.currentFlow = f.amortizationFlow
	case models.FinancingInstallment:
		f.step = stepBeginFinancingInstallment
		f.currentFlow = f.financingInstallmentFlow
	}

	return f
}

func (f *flow[T]) isApartmentValid(apartment string) bool {
	for _, apt := range f.availableApartments {
		if apt == apartment {
			return true
		}
	}
	return false
}

func (f *flow[T]) assembleKeyboardMenuWithApartments() tgbotapi.InlineKeyboardMarkup {
	var currRow []tgbotapi.InlineKeyboardButton

	keyboard := tgbotapi.NewInlineKeyboardMarkup()
	for i, apt := range f.availableApartments {
		if i > 0 && i%3 == 0 {
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, currRow)
			currRow = []tgbotapi.InlineKeyboardButton{}
		} else {
			currRow = append(currRow, tgbotapi.NewInlineKeyboardButtonData(apt, apt))
		}
	}
	if len(currRow) > 0 {
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, currRow)
	}

	return keyboard
}

func (f *flow[T]) assembleKeyboardMenuWithPayers() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Gustavo", "Gustavo"),
			tgbotapi.NewInlineKeyboardButtonData("Emerson", "Emerson"),
		),
	)
}

func (f *flow[T]) Next(answer string) (string, interface{}) {
	replyText, markup := f.next(answer)
	if len(replyText) > 0 && len(f.apartmentName) > 0 {
		replyText = fmt.Sprintf("[%s] %s", f.apartmentName, replyText)
	}
	return replyText, markup
}

func (f *flow[T]) next(answer string) (string, interface{}) {
	if !f.askedApartment {
		f.askedApartment = true
		availableApartments, err := f.store.GetAvailableApartments()
		if err != nil {
			f.step = stepEnd
			log.Printf("error while getting available apartments: %v", err.Error())
			return "Ocorreu um erro inesperado, tenete novamente :(", nil
		}
		f.availableApartments = availableApartments
		return "Selecione o apartamento", f.assembleKeyboardMenuWithApartments()
	}

	if f.apartmentName == "" {
		if !f.isApartmentValid(answer) {
			return "Imóvel nao existe. De qual imóvel estamos falando?", nil
		}
		f.apartmentName = answer
	}

	if f.step == stepEnd {
		return "", nil
	}

	return f.currentFlow(answer)
}

func parseDateFromFullDate(dateStr string) (time.Time, error) {
	t, err := format.DDMMYYYYstringToTimeObj(dateStr)
	if err != nil {
		return t, fmt.Errorf("%v nao é uma data válida, informe uma data no formato dd/mm/aaaa", dateStr)
	}
	return t, nil
}

func parsePriceFromStr(priceStr string) (float64, error) {
	p, err := format.StrPriceToFloat64(priceStr)
	if err != nil {
		return p, fmt.Errorf("%v nao é um número válido, informe novamente o valor", priceStr)
	}
	return p, nil
}
