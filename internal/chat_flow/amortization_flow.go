package chat_flow

import (
	"fmt"

	"github.com/gustavolopess/hoteleiro/internal/models"
)

func (f *flow[T]) amortizationFlow(answer string) (string, interface{}) {
	switch f.step {
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
		return "Quem fez essa amortizaçao?", f.assembleKeyboardMenuWithPayers()
	case stepGetPayerAmortization:
		f.value.(*models.Amortization).Payer = answer
		err := f.store.AddAmortization(f.value.(*models.Amortization))
		if err != nil {
			return fmt.Sprintf("Falha ao adicionar amortizaçao %v - %v", f.value.(*models.Amortization).ToString(), err.Error()), nil
		}
		f.step = stepEnd
		return fmt.Sprintf("Amortizaçao registrada: %v", f.value.(*models.Amortization).ToString()), nil
	}
	return "", nil
}
