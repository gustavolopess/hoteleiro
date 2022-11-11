package chat_flow

import (
	"fmt"

	"github.com/gustavolopess/hoteleiro/internal/models"
)

func (f *flow[T]) energyBillFlow(answer string) (string, interface{}) {
	switch f.step {
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
		f.step = stepGetPayerEnergyBill
		return "Quem pagou essa conta de energia?", f.assembleKeyboardMenuWithPayers()
	case stepGetPayerEnergyBill:
		f.value.(*models.EnergyBill).Payer = answer
		if err := f.store.AddBill(f.value.(*models.EnergyBill)); err != nil {
			return fmt.Sprintf("Falha ao registrar conta de energia %v - %v", f.value.(*models.EnergyBill).ToString(), err.Error()), nil
		}
		return fmt.Sprintf("Conta de energia adicionada - %v", f.value.(*models.EnergyBill).ToString()), nil
	}
	return "", nil
}
