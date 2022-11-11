package models

import (
	"fmt"
	"time"
)

type Condo struct {
	Value float64
	Date  time.Time
	Payer string
	Apartment
}

func (c *Condo) ToString() string {
	return fmt.Sprintf("Condomínio referente ao mês %v de %v custando R$%v pago por %v", c.Date.Format("01"), c.Date.Format("2006"), c.Value, c.Payer)
}
