package chat_flow

import (
	"fmt"

	"github.com/gustavolopess/hoteleiro/internal/models"
)

func (f *flow[T]) cleaningFlow(answer string) (string, interface{}) {
	switch f.step {
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
		f.step = stepGetCleaningPayer
		return "Quem pagou pela faxina?", f.assembleKeyboardMenuWithPayers()
	case stepGetCleaningPayer:
		f.value.(*models.Cleaning).Payer = answer
		f.store.AddCleaning(f.value.(*models.Cleaning))
		f.step = stepEnd
		return fmt.Sprintf("Faxina registrada: %v", f.value.(*models.Cleaning).ToString()), nil
	}
	return "", nil
}
