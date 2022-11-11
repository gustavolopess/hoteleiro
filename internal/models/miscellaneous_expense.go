package models

import (
	"fmt"
	"time"
)

type MiscellaneousExpense struct {
	Value       float64
	Date        time.Time
	Description string
	Payer       string
	Apartment
}

func (m *MiscellaneousExpense) ToString() string {
	return fmt.Sprintf("despesa de %v, com identificaçao \"%v\", do dia %v, paga por %v", m.Value, m.Description, m.Date.Format("02/01/2006"), m.Payer)
}
