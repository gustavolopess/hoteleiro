package models

import (
	"fmt"
	"time"
)

type Cleaning struct {
	Date  time.Time
	Value float64
	Payer string
	Apartment
}

func (c *Cleaning) ToString() string {
	return fmt.Sprintf("faxina do dia %v, paga por %v, ao custo de R$%v", c.Date.Format("02/01/2006"), c.Payer, c.Value)
}
