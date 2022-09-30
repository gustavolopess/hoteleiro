package models

import (
	"fmt"
	"time"
)

type EnergyBill struct {
	Date  time.Time
	Value float64
}

func (e *EnergyBill) ToString() string {
	return fmt.Sprintf("mês %v, do ano %v com valor de R$%v", e.Date.Month(), e.Date.Year(), e.Value)
}
