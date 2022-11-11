package chat_flow

import "github.com/gustavolopess/hoteleiro/internal/models"

func (f *flow[T]) apartmentFlow(answer string) (string, interface{}) {
	switch f.step {
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
	}
	return "", nil
}
