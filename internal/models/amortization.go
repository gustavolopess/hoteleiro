package models

import (
	"fmt"
	"time"
)

type Amortization struct {
	Payer string
	Value float64
	Date  time.Time
	Apartment
}

func (a *Amortization) ToString() string {
	return fmt.Sprintf("%v amortizou R$%v no dia %v", a.Payer, a.Value, a.Date.Format("02/01/2006"))
}
