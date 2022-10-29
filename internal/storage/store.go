package storage

import (
	"context"
	"errors"
	"time"

	"github.com/gustavolopess/hoteleiro/internal/config"
	"github.com/gustavolopess/hoteleiro/internal/models"
	"github.com/gustavolopess/hoteleiro/internal/storage/google_sheets"
)

type Store interface {
	AddCleaning(c *models.Cleaning) error
	AddCondo(c *models.Condo) error
	AddApartment(a *models.Apartment) error
	AddBill(e *models.EnergyBill) error
	AddRent(r *models.Rent) error
}

type rentRegistry struct {
	*models.Rent
	isLastDay bool
}

type store struct {
	client     Store
	cleanings  map[string]*models.Cleaning
	bills      map[string]*models.EnergyBill
	condos     map[string]*models.Condo
	rents      map[time.Time]*rentRegistry
	apartments map[string]*models.Apartment
}

func NewStore(cfg *config.Config) Store {
	sheetsClient := google_sheets.NewSheetsClient(context.Background(), "credentials.json", "1lfWxf_Wj5IjKjPlu6V2k519Y_RVJh1UU2pDL9VuFxCo")
	return &store{
		client:    sheetsClient,
		cleanings: make(map[string]*models.Cleaning),
		bills:     make(map[string]*models.EnergyBill),
		condos:    make(map[string]*models.Condo),
		rents:     make(map[time.Time]*rentRegistry),
	}
}

func (s *store) AddCleaning(c *models.Cleaning) error {
	// TODO: check if there's some cleaning at same day for same apartment before add
	return s.client.AddCleaning(c)
}

func (s *store) AddCondo(c *models.Condo) error {
	// TODO: check if there's some condo at same month for same apartment before add
	return s.client.AddCondo(c)
}

func (s *store) AddApartment(a *models.Apartment) error {
	// TODO: check if there's some apartment with same name before add
	return s.client.AddApartment(a)
}

func (s *store) AddBill(e *models.EnergyBill) error {
	s.bills[e.Date.Format("02/01/2006")] = e
	return nil
}

func (s *store) AddRent(r *models.Rent) error {
	// TODO: get list of rents at window of 1month forward and backward to run the verification below
	dateIterator := r.DateBegin
	rentsCopy := s.rents

	for dateIterator.Before(r.DateEnd) || dateIterator.Equal(r.DateEnd) {
		if registry, ok := rentsCopy[dateIterator]; ok && registry.isLastDay && dateIterator.Equal(r.DateBegin) {
			rentsCopy[dateIterator] = &rentRegistry{
				isLastDay: false,
				Rent:      r,
			}
		} else if !ok && dateIterator.Equal(r.DateBegin) {
			rentsCopy[dateIterator] = &rentRegistry{
				isLastDay: false,
				Rent:      r,
			}
		} else if !ok && dateIterator.Equal(r.DateEnd) {
			rentsCopy[dateIterator] = &rentRegistry{
				isLastDay: true,
				Rent:      nil,
			}
		} else if !ok {
			rentsCopy[dateIterator] = &rentRegistry{
				isLastDay: false,
				Rent:      nil,
			}
		} else {
			return errors.New("aluguel inválido ou sobrepondo um aluguel existente")
		}

		dateIterator = dateIterator.Add(time.Hour * 24)
	}

	s.rents = rentsCopy
	return nil
}
