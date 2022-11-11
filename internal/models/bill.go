package models

import (
	"fmt"
	"time"
)

type EnergyBill struct {
	Date  time.Time
	Value float64
	Payer string
	Apartment
}

func (e *EnergyBill) ToString() string {
	return fmt.Sprintf("mês %v, do ano %v com valor de R$%v paga por %v", e.Date.Month(), e.Date.Year(), e.Value, e.Payer)
}
