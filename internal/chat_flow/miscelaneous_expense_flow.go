package chat_flow

import (
	"fmt"

	"github.com/gustavolopess/hoteleiro/internal/models"
)

func (f *flow[T]) miscellaneousExpenseFlow(answer string) (string, interface{}) {
	switch f.step {
	case stepBeginMiscellaneousExpense:
		f.step = stepGetDescriptionMiscellaneousExpense
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
		return "Qual a data da despesa? dd/mm/aaaa", nil
	case stepGetDateMiscellaneousExpense:
		t, err := parseDateFromFullDate(answer)
		if err != nil {
			return err.Error(), nil
		}
		f.value.(*models.MiscellaneousExpense).Date = t
		f.step = stepGetPayerMiscellaneousExpense
		return "Quem pagou por essa despesa?", f.assembleKeyboardMenuWithPayers()
	case stepGetPayerMiscellaneousExpense:
		f.value.(*models.MiscellaneousExpense).Payer = answer
		if err := f.store.AddMiscellaneousExpense(f.value.(*models.MiscellaneousExpense)); err != nil {
			return fmt.Sprintf("Falha ao adicionar a despesa %v - %v", f.value.(*models.MiscellaneousExpense).ToString(), err.Error()), nil
		}
		f.step = stepEnd
		return fmt.Sprintf("Despesa registrada: %v", f.value.(*models.MiscellaneousExpense).ToString()), nil
	}
	return "", nil
}
