package chat_flow

import (
	"fmt"

	"github.com/gustavolopess/hoteleiro/internal/models"
)

func (f *flow[T]) rentFlow(answer string) (string, interface{}) {
	switch f.step {
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
		f.step = stepGetRentReceiver
		return "Quem recebeu o dinheiro do aluguel?", f.assembleKeyboardMenuWithPayers()
	case stepGetRentReceiver:
		f.value.(*models.Rent).Receiver = answer
		err := f.store.AddRent(f.value.(*models.Rent))
		if err != nil {
			return fmt.Sprintf("Falha ao adicionar o aluguel %v - %v", f.value.(*models.Rent).ToString(), err.Error()), nil
		}
		f.step = stepEnd
		return fmt.Sprintf("Aluguel adicionado! %v", f.value.(*models.Rent).ToString()), nil
	}
	return "", nil
}
