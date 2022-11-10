package models

import (
	"fmt"
	"time"
)

type FinancingInstallment struct {
	Value float64
	Date  time.Time
	Payer string
	Apartment
}

func (f *FinancingInstallment) ToString() string {
	return fmt.Sprintf("Parcela de %v paga em %v por %v", f.Value, f.Date.Format("02/01/2006"), f.Payer)
}
