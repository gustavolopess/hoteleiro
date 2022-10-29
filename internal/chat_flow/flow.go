package chat_flow

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gustavolopess/hoteleiro/internal/models"
	"github.com/gustavolopess/hoteleiro/internal/storage"
)

type Flow[T models.Models] interface {
	Next(string) string
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

	stepEnd
)

type flow[T models.Models] struct {
	store          storage.Store
	step           Step
	askedApartment bool
	apartmentName  string
	value          any
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
	}

	return f
}

func (f *flow[T]) Next(answer string) string {
	if !f.askedApartment {
		f.askedApartment = true
		return "De qual imóvel estamos falando?"
	}

	if f.apartmentName == "" {
		f.apartmentName = answer
	}

	switch f.step {
	case stepEnd:
		return ""

	// Energy bill flow
	case stepBeginEnergyBill:
		f.step = stepGetValueEnergyBill
		return "Qual o valor da conta de energia?"
	case stepGetValueEnergyBill:
		value, err := parsePriceFromStr(answer)
		if err != nil {
			return err.Error()
		}
		if f.value == nil {
			f.value = &models.EnergyBill{
				Apartment: models.Apartment{Name: f.apartmentName},
			}
		}
		f.value.(*models.EnergyBill).Value = value
		f.step = stepGetDateEnergyBill
		return "Informe o mês e ano da conta de energia. Escreva no formato mm/aaaa"
	case stepGetDateEnergyBill:
		t, err := parseDateFromMonthAndYear(answer)
		if err != nil {
			return err.Error()
		}
		f.value.(*models.EnergyBill).Date = t
		f.step = stepEnd
		err = f.store.AddBill(f.value.(*models.EnergyBill))
		if err != nil {
			return fmt.Sprintf("Falha ao registrar conta de energia %v - %v", f.value.(*models.EnergyBill).ToString(), err.Error())
		}
		return fmt.Sprintf("Conta de energia adicionada - %v", f.value.(*models.EnergyBill).ToString())

	// Rent flow
	case stepBeginRent:
		if f.value == nil {
			f.value = &models.Rent{
				Apartment: models.Apartment{Name: f.apartmentName},
			}
		}
		f.step = stepGetValueRent
		return "Qual o valor do aluguel?"
	case stepGetValueRent:
		value, err := parsePriceFromStr(answer)
		if err != nil {
			return err.Error()
		}
		f.value.(*models.Rent).Value = value
		f.step = stepGetDateBeginRent
		return "Qual a data de início da locação? informe a data no formato dd/mm/aaaa"
	case stepGetDateBeginRent:
		t, err := parseDateFromFullDate(answer)
		if err != nil {
			return err.Error()
		}
		f.value.(*models.Rent).DateBegin = t
		f.step = stepGetDateEndRent
		return "Qual a data final da locação? informe a data no formato dd/mm/aaaa"
	case stepGetDateEndRent:
		t, err := parseDateFromFullDate(answer)
		if err != nil {
			return err.Error()
		}
		f.value.(*models.Rent).DateEnd = t
		f.step = stepGetRenter
		return "Qual o nome do inquilino?"
	case stepGetRenter:
		f.value.(*models.Rent).Renter = answer
		err := f.store.AddRent(f.value.(*models.Rent))
		if err != nil {
			return fmt.Sprintf("Falha ao adicionar o aluguel %v - %v", f.value.(*models.Rent).ToString(), err.Error())
		}
		f.step = stepEnd
		return fmt.Sprintf("Aluguel adicionado! %v", f.value.(*models.Rent).ToString())

	// Cleaning flow
	case stepBeginCleaning:
		f.value = &models.Cleaning{
			Apartment: models.Apartment{
				Name: f.apartmentName,
			},
		}
		f.step = stepGetValueCleaning
		return "Qual o valor pago na faxina?"
	case stepGetValueCleaning:
		value, err := parsePriceFromStr(answer)
		if err != nil {
			return err.Error()
		}
		f.step = stepGetDateCleaning
		f.value.(*models.Cleaning).Value = value
		return "Em qual data a faxina foi realizada? informe uma data no formato dd/mm/aaaa"
	case stepGetDateCleaning:
		t, err := parseDateFromFullDate(answer)
		if err != nil {
			return err.Error()
		}
		f.value.(*models.Cleaning).Date = t
		f.step = stepGetCleaner
		return "Quem foi o faxineiro(a)?"
	case stepGetCleaner:
		f.value.(*models.Cleaning).Cleaner = answer
		f.store.AddCleaning(f.value.(*models.Cleaning))
		f.step = stepEnd
		return fmt.Sprintf("Faxina registrada: %v", f.value.(*models.Cleaning).ToString())

	// Condo flow
	case stepBeginCondo:
		f.step = stepGetValueCondo
		return "Qual o valor do condomínio?"
	case stepGetValueCondo:
		value, err := parsePriceFromStr(answer)
		if err != nil {
			return err.Error()
		}
		f.step = stepGetDateCondo
		f.value = &models.Condo{
			Value: value,
		}
		return "Qual o mês e ano desta taxa de condomínio? informe uma data no formato mm/aaaa"
	case stepGetDateCondo:
		t, err := parseDateFromMonthAndYear(answer)
		if err != nil {
			return err.Error()
		}
		f.value.(*models.Condo).Date = t
		_ = f.store.AddCondo(f.value.(*models.Condo))
		f.step = stepEnd
		return fmt.Sprintf("Taxa de condomínio registrada: %v", f.value.(*models.Condo).ToString())

	// Apartment flow
	case stepBeginApartment:
		f.step = stepGetNameApartment
		return "Qual o nome do imóvel?"
	case stepGetNameApartment:
		f.step = stepGetAddressApartment
		f.value = &models.Apartment{
			Name: answer,
		}
		return "Qual o endereço do imóvel?"
	case stepGetAddressApartment:
		f.value.(*models.Apartment).Address = answer
		_ = f.store.AddApartment(f.value.(*models.Apartment))
		f.step = stepEnd
	}

	return ""
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
