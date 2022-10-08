package models

type Apartment struct {
	Name    string
	Address string
}

func (a *Apartment) ToString() string {
	return a.Name + " - " + a.Address
}
