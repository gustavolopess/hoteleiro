package chat_flow

import (
	"fmt"
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

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

	stepBeginRent
	stepGetValueRent
	stepGetDateBeginRent
	stepGetDateEndRent
	stepGetRenter

	stepBeginCleaning
	stepGetValueCleaning
	stepGetDateCleaning
	stepGetCleaner

	stepBeginCondo
	stepGetValueCondo
	stepGetDateCondo

	stepBeginApartment
	stepGetNameApartment
	stepGetAddressApartment

	stepBeginMiscellaneousExpense
	stepGetDescriptionMiscellaneousExpense
	stepGetValueMiscellaneousExpense
	stepGetDateMiscellaneousExpense

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
}

func NewFlow[T models.Models](store storage.Store) Flow[T] {
	f := &flow[T]{
		store: store,
	}

	var b T
	switch any(b).(type) {
	case models.EnergyBill:
		f.step = stepBeginEnergyBill
	case models.Rent:
		f.step = stepBeginRent
	case models.Cleaning:
		f.step = stepBeginCleaning
	case models.Condo:
		f.step = stepBeginCondo
	case models.Apartment:
		f.step = stepBeginApartment
	case models.MiscellaneousExpense:
		f.step = stepBeginMiscellaneousExpense
	case models.Amortization:
		f.step = stepBeginAmortization
	case models.FinancingInstallment:
		f.step = stepBeginFinancingInstallment
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

	switch f.step {
	case stepEnd:
		return "", nil

	// Energy bill flow
	case stepBeginEnergyBill:
		f.step = stepGetValueEnergyBill
		return "Qual o valor da conta de energia?", nil
	case stepGetValueEnergyBill:
		value, err := parsePriceFromStr(answer)
		if err != nil {
			return err.Error(), nil
		}
		if f.value == nil {
			f.value = &models.EnergyBill{
				Apartment: models.Apartment{Name: f.apartmentName},
			}
		}
		f.value.(*models.EnergyBill).Value = value
		f.step = stepGetDateEnergyBill
		return "Informe o mês e ano da conta de energia. Escreva no formato mm/aaaa", nil
	case stepGetDateEnergyBill:
		t, err := parseDateFromMonthAndYear(answer)
		if err != nil {
			return err.Error(), nil
		}
		f.value.(*models.EnergyBill).Date = t
		f.step = stepEnd
		err = f.store.AddBill(f.value.(*models.EnergyBill))
		if err != nil {
			return fmt.Sprintf("Falha ao registrar conta de energia %v - %v", f.value.(*models.EnergyBill).ToString(), err.Error()), nil
		}
		return fmt.Sprintf("Conta de energia adicionada - %v", f.value.(*models.EnergyBill).ToString()), nil

	// Rent flow
	case stepBeginRent:
		if f.value == nil {
			f.value = &models.Rent{
				Apartment: models.Apartment{Name: f.apartmentName},
			}
		}
		f.step = stepGetValueRent
		return "Qual o valor do aluguel?", nil
	case stepGetValueRent:
		value, err := parsePriceFromStr(answer)
		if err != nil {
			return err.Error(), nil
		}
		f.value.(*models.Rent).Value = value
		f.step = stepGetDateBeginRent
		return "Qual a data de início da locação? informe a data no formato dd/mm/aaaa", nil
	case stepGetDateBeginRent:
		t, err := parseDateFromFullDate(answer)
		if err != nil {
			return err.Error(), nil
		}
		f.value.(*models.Rent).DateBegin = t
		f.step = stepGetDateEndRent
		return "Qual a data final da locação? informe a data no formato dd/mm/aaaa", nil
	case stepGetDateEndRent:
		t, err := parseDateFromFullDate(answer)
		if err != nil {
			return err.Error(), nil
		}
		f.value.(*models.Rent).DateEnd = t
		f.step = stepGetRenter
		return "Qual o nome do inquilino?", nil
	case stepGetRenter:
		f.value.(*models.Rent).Renter = answer
		err := f.store.AddRent(f.value.(*models.Rent))
		if err != nil {
			return fmt.Sprintf("Falha ao adicionar o aluguel %v - %v", f.value.(*models.Rent).ToString(), err.Error()), nil
		}
		f.step = stepEnd
		return fmt.Sprintf("Aluguel adicionado! %v", f.value.(*models.Rent).ToString()), nil

	// Cleaning flow
	case stepBeginCleaning:
		f.value = &models.Cleaning{
			Apartment: models.Apartment{
				Name: f.apartmentName,
			},
		}
		f.step = stepGetValueCleaning
		return "Qual o valor pago na faxina?", nil
	case stepGetValueCleaning:
		value, err := parsePriceFromStr(answer)
		if err != nil {
			return err.Error(), nil
		}
		f.step = stepGetDateCleaning
		f.value.(*models.Cleaning).Value = value
		return "Em qual data a faxina foi realizada? informe uma data no formato dd/mm/aaaa", nil
	case stepGetDateCleaning:
		t, err := parseDateFromFullDate(answer)
		if err != nil {
			return err.Error(), nil
		}
		f.value.(*models.Cleaning).Date = t
		f.step = stepGetCleaner
		return "Quem foi o faxineiro(a)?", nil
	case stepGetCleaner:
		f.value.(*models.Cleaning).Cleaner = answer
		f.store.AddCleaning(f.value.(*models.Cleaning))
		f.step = stepEnd
		return fmt.Sprintf("Faxina registrada: %v", f.value.(*models.Cleaning).ToString()), nil

	// Condo flow
	case stepBeginCondo:
		f.step = stepGetValueCondo
		return "Qual o valor do condomínio?", nil
	case stepGetValueCondo:
		value, err := parsePriceFromStr(answer)
		if err != nil {
			return err.Error(), nil
		}
		f.step = stepGetDateCondo
		f.value = &models.Condo{
			Value: value,
		}
		return "Qual o mês e ano desta taxa de condomínio? informe uma data no formato mm/aaaa", nil
	case stepGetDateCondo:
		t, err := parseDateFromMonthAndYear(answer)
		if err != nil {
			return err.Error(), nil
		}
		f.value.(*models.Condo).Date = t
		_ = f.store.AddCondo(f.value.(*models.Condo))
		f.step = stepEnd
		return fmt.Sprintf("Taxa de condomínio registrada: %v", f.value.(*models.Condo).ToString()), nil

	// Apartment flow
	case stepBeginApartment:
		f.step = stepGetNameApartment
		return "Qual o nome do imóvel?", nil
	case stepGetNameApartment:
		f.step = stepGetAddressApartment
		f.value = &models.Apartment{
			Name: answer,
		}
		return "Qual o endereço do imóvel?", nil
	case stepGetAddressApartment:
		f.value.(*models.Apartment).Address = answer
		_ = f.store.AddApartment(f.value.(*models.Apartment))
		f.step = stepEnd

	// Miscellaneous expense flow
	case stepBeginMiscellaneousExpense:
		f.step = stepGetValueMiscellaneousExpense
		return "Informe um identificador para essa despesa (exemplo: compra de sofá, etc)", nil
	case stepGetDescriptionMiscellaneousExpense:
		f.step = stepGetValueMiscellaneousExpense
		f.value = &models.MiscellaneousExpense{
			Description: answer,
			Apartment:   models.Apartment{Name: f.apartmentName},
		}
		return "Qual o valor da despesa?", nil
	case stepGetValueMiscellaneousExpense:
		value, err := parsePriceFromStr(answer)
		if err != nil {
			return err.Error(), nil
		}
		f.step = stepGetDateMiscellaneousExpense
		f.value.(*models.MiscellaneousExpense).Value = value
	case stepGetDateMiscellaneousExpense:
		t, err := parseDateFromMonthAndYear(answer)
		if err != nil {
			return err.Error(), nil
		}
		f.value.(*models.MiscellaneousExpense).Date = t
		err = f.store.AddMiscellaneousExpense(f.value.(*models.MiscellaneousExpense))
		if err != nil {
			return fmt.Sprintf("Falha ao adicionar a despesa %v - %v", f.value.(*models.MiscellaneousExpense).ToString(), err.Error()), nil
		}
		f.step = stepEnd
		return fmt.Sprintf("Despesa registrada: %v", f.value.(*models.MiscellaneousExpense).ToString()), nil

	// Amortization flow
	case stepBeginAmortization:
		f.step = stepGetValueAmortization
		return "Qual foi o valor amortizado?", nil
	case stepGetValueAmortization:
		v, err := parsePriceFromStr(answer)
		if err != nil {
			return err.Error(), nil
		}
		f.value = &models.Amortization{
			Apartment: models.Apartment{Name: f.apartmentName},
			Value:     v,
		}
		f.step = stepGetDateAmortization
		return "Qual a data da amortizaçao?", nil
	case stepGetDateAmortization:
		t, err := parseDateFromFullDate(answer)
		if err != nil {
			return err.Error(), nil
		}
		f.value.(*models.Amortization).Date = t
		f.step = stepGetPayerAmortization
		return "Quem fez essa amortizaçao?", nil
	case stepGetPayerAmortization:
		f.value.(*models.Amortization).Payer = answer
		err := f.store.AddAmortization(f.value.(*models.Amortization))
		if err != nil {
			return fmt.Sprintf("Falha ao adicionar amortizaçao %v - %v", f.value.(*models.Amortization).ToString(), err.Error()), nil
		}
		f.step = stepEnd
		return fmt.Sprintf("Amortizaçao registrada: %v", f.value.(*models.Amortization).ToString()), nil

	// Financing installment flow
	case stepBeginFinancingInstallment:
		f.step = stepGetFinancialInstallmentValue
		return "Qual o valor pago na parcela?", nil
	case stepGetFinancialInstallmentValue:
		v, err := parsePriceFromStr(answer)
		if err != nil {
			return err.Error(), nil
		}
		f.value = &models.FinancingInstallment{
			Value:     v,
			Apartment: models.Apartment{Name: f.apartmentName},
		}
		f.step = stepGetFinancialInstallmentDate
		return "Qual foi a data em que essa parcela foi paga? informe no formato dd/mm/aaaa", nil
	case stepGetFinancialInstallmentDate:
		t, err := parseDateFromFullDate(answer)
		if err != nil {
			return err.Error(), nil
		}
		f.value.(*models.FinancingInstallment).Date = t
		f.step = stepGetFinancialInstallmentPayer
		return "Quem pagou esta parcela?", nil
	case stepGetFinancialInstallmentPayer:
		f.value.(*models.FinancingInstallment).Payer = answer
		err := f.store.AddFinancingInstallment(f.value.(*models.FinancingInstallment))
		if err != nil {
			return fmt.Sprintf("Falha ao registrar pagamento de parcela - %v", err.Error()), nil
		}
		f.step = stepEnd
		return fmt.Sprintf("Pagamento de parcela registrado: %v", f.value.(*models.FinancingInstallment).ToString()), nil
	}

	return "", nil
}

func parseDateFromMonthAndYear(dateStr string) (time.Time, error) {
	var t time.Time
	t, err := time.Parse("01/2006", dateStr)
	if err != nil {
		return t, fmt.Errorf("%v nao é uma data válida, informe uma data no formato mm/aaaa", dateStr)
	}
	return t, nil
}

func parseDateFromFullDate(dateStr string) (time.Time, error) {
	var t time.Time
	t, err := time.Parse("02/01/2006", dateStr)
	if err != nil {
		return t, fmt.Errorf("%v nao é uma data válida, informe uma data no formato dd/mm/aaaa", dateStr)
	}
	return t, nil
}

func parsePriceFromStr(priceStr string) (float64, error) {
	var value float64
	value, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return value, fmt.Errorf("%v nao é um número válido, informe novamente o valor", priceStr)
	}
	return value, nil
}
