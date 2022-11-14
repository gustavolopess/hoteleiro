package storage

import (
	"context"
	"log"

	"github.com/gustavolopess/hoteleiro/internal/config"
	"github.com/gustavolopess/hoteleiro/internal/models"
	"github.com/gustavolopess/hoteleiro/internal/storage/errors"
	"github.com/gustavolopess/hoteleiro/internal/storage/google_sheets"
)

type Store interface {
	AddCleaning(c *models.Cleaning) error
	AddCondo(c *models.Condo) error
	AddApartment(a *models.Apartment) error
	AddBill(e *models.EnergyBill) error
	AddRent(r *models.Rent) error
	AddMiscellaneousExpense(e *models.MiscellaneousExpense) error
	AddAmortization(a *models.Amortization) error
	AddFinancingInstallment(f *models.FinancingInstallment) error
	GetAvailableApartments() ([]string, error)
	GetExistingRents(apartment models.Apartment) ([]*models.Rent, error)
	GetPayedCondos(apartment models.Apartment) ([]*models.Condo, error)
	GetPayedBills(apartment models.Apartment) ([]*models.EnergyBill, error)
	GetPayedCleanings(apartment models.Apartment) ([]*models.Cleaning, error)
	GetMiscellaneousExpenses(apartment models.Apartment) ([]*models.MiscellaneousExpense, error)
}

type store struct {
	client Store
}

func NewStore(cfg *config.Config) Store {
	sheetsClient := google_sheets.NewSheetsClient(context.Background(), "credentials.json", "1lfWxf_Wj5IjKjPlu6V2k519Y_RVJh1UU2pDL9VuFxCo")
	return &store{
		client: sheetsClient,
	}
}

func (s *store) AddCleaning(c *models.Cleaning) error {
	payedCleanings, err := s.GetPayedCleanings(c.Apartment)
	if err != nil {
		return err
	}

	if isAlreadyCleanedAtDay(c, payedCleanings) {
		return errors.ErrCleaningAlreadyHappened
	}

	return s.client.AddCleaning(c)
}

func (s *store) AddCondo(c *models.Condo) error {
	payedCondos, err := s.GetPayedCondos(c.Apartment)
	if err != nil {
		return err
	}

	if isCondoPayedAtMonth(c, payedCondos) {
		return errors.ErrCondoAlreadyPayed
	}

	return s.client.AddCondo(c)
}

func (s *store) AddApartment(a *models.Apartment) error {
	return s.client.AddApartment(a)
}

func (s *store) AddBill(e *models.EnergyBill) error {
	payedBills, err := s.GetPayedBills(e.Apartment)
	if err != nil {
		return err
	}

	if isBillAlreadyPayed(e, payedBills) {
		return errors.ErrBillAlreadyPayed
	}

	return s.client.AddBill(e)
}

func (s *store) AddRent(r *models.Rent) error {
	existingRents, err := s.GetExistingRents(r.Apartment)
	if err != nil {
		return err
	}

	if !isRentDatesAvailable(r, existingRents) {
		return errors.ErrRentDatesUsed
	}

	if r.DateBegin.After(r.DateEnd) || r.DateBegin.Equal(r.DateEnd) {
		return errors.ErrRentReversedDates
	}

	return s.client.AddRent(r)
}

func (s *store) AddMiscellaneousExpense(m *models.MiscellaneousExpense) error {
	return s.client.AddMiscellaneousExpense(m)
}

func (s *store) AddAmortization(a *models.Amortization) error {
	return s.client.AddAmortization(a)
}

func (s *store) AddFinancingInstallment(f *models.FinancingInstallment) error {
	return s.client.AddFinancingInstallment(f)
}

func (s *store) GetAvailableApartments() ([]string, error) {
	return s.client.GetAvailableApartments()
}

func (s *store) GetExistingRents(apartment models.Apartment) ([]*models.Rent, error) {
	return s.client.GetExistingRents(apartment)
}

func (s *store) GetPayedCondos(apartment models.Apartment) ([]*models.Condo, error) {
	return s.client.GetPayedCondos(apartment)
}

func (s *store) GetPayedBills(apartment models.Apartment) ([]*models.EnergyBill, error) {
	return s.client.GetPayedBills(apartment)
}

func (s *store) GetPayedCleanings(apartment models.Apartment) ([]*models.Cleaning, error) {
	return s.client.GetPayedCleanings(apartment)
}

func (s *store) GetMiscellaneousExpenses(apartment models.Apartment) ([]*models.MiscellaneousExpense, error) {
	return s.client.GetMiscellaneousExpenses(apartment)
}

func isRentDatesAvailable(r *models.Rent, existingRents []*models.Rent) bool {
	for _, er := range existingRents {
		dateBegin, dateEnd := er.DateBegin, er.DateEnd

		if r.DateBegin.Equal(dateBegin) || (r.DateBegin.After(dateBegin) && r.DateBegin.Before(dateEnd)) {
			log.Println("begin date is at an used date range")
			return false
		}

		if (r.DateEnd.After(dateBegin) && r.DateEnd.Before(dateEnd)) || r.DateEnd.Equal(dateEnd) {
			log.Println("end date is at an used date range")
			return false
		}
	}

	return true
}

func isCondoPayedAtMonth(c *models.Condo, condosPayed []*models.Condo) bool {
	for _, cp := range condosPayed {
		if cp.Date.Month() == c.Date.Month() && cp.Date.Year() == c.Date.Year() {
			return true
		}
	}

	return false
}

func isBillAlreadyPayed(b *models.EnergyBill, billsPayed []*models.EnergyBill) bool {
	for _, bp := range billsPayed {
		if bp.Date.Month() == b.Date.Month() && bp.Date.Year() == b.Date.Year() {
			return true
		}
	}

	return false
}

func isAlreadyCleanedAtDay(c *models.Cleaning, cleaningsPayed []*models.Cleaning) bool {
	for _, cp := range cleaningsPayed {
		if cp.Date.Day() == c.Date.Day() && cp.Date.Month() == c.Date.Month() && cp.Date.Year() == c.Date.Year() {
			return true
		}
	}

	return false
}
