package chat_flow

import (
	"fmt"

	"github.com/gustavolopess/hoteleiro/internal/models"
)

func (f *flow[T]) financingInstallmentFlow(answer string) (string, interface{}) {
	switch f.step {
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
		return "Quem pagou esta parcela?", f.assembleKeyboardMenuWithPayers()
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
