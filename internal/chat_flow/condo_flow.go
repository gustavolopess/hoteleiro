package chat_flow

import (
	"fmt"

	"github.com/gustavolopess/hoteleiro/internal/models"
)

func (f *flow[T]) condoFlow(answer string) (string, interface{}) {
	switch f.step {
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
		return "Em que data esta taxa de condomínio foi paga? informe uma data no formato dd/mm/aaaa", nil
	case stepGetDateCondo:
		t, err := parseDateFromFullDate(answer)
		if err != nil {
			return err.Error(), nil
		}
		f.value.(*models.Condo).Date = t
		f.step = stepGetPayerCondo
		return "Quem pagou essa taxa de condomínio?", f.assembleKeyboardMenuWithPayers()
	case stepGetPayerCondo:
		f.value.(*models.Condo).Payer = answer
		if err := f.store.AddCondo(f.value.(*models.Condo)); err != nil {
			return fmt.Sprintf("Falha ao adicionar taxa de condomínio: %v", err.Error()), nil
		}
		f.step = stepEnd
		return fmt.Sprintf("Taxa de condomínio registrada: %v", f.value.(*models.Condo).ToString()), nil
	}
	return "", nil
}
