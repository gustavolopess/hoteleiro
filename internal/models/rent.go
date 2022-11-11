package models

import (
	"fmt"
	"time"
)

type Rent struct {
	DateBegin time.Time
	DateEnd   time.Time
	Value     float64
	Renter    string
	Receiver  string
	Apartment
}

func (r *Rent) ToString() string {
	return fmt.Sprintf(`do dia %v ao dia %v pelo valor de R$%v para o inquilino %v - recebido por %v`,
		r.DateBegin.Format("02/01/2006"),
		r.DateEnd.Format("02/01/2006"),
		r.Value,
		r.Renter,
		r.Receiver,
	)
}
