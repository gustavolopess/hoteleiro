package storage

import (
	"errors"
	"time"

	"github.com/gustavolopess/hoteleiro/models"
)

type Store interface {
	AddCleaning(c *models.Cleaning) error
	AddCondo(c *models.Condo) error
	AddBill(e *models.EnergyBill) error
	AddRent(r *models.Rent) error
}

type rentRegistry struct {
	*models.Rent
	isLastDay bool
}

type store struct {
	cleanings map[string]*models.Cleaning
	bills     map[string]*models.EnergyBill
	condos    map[string]*models.Condo
	rents     map[time.Time]*rentRegistry
}

func NewStore() Store {
	return &store{
		cleanings: make(map[string]*models.Cleaning),
		bills:     make(map[string]*models.EnergyBill),
		condos:    make(map[string]*models.Condo),
		rents:     make(map[time.Time]*rentRegistry),
	}
}

func (s *store) AddCleaning(c *models.Cleaning) error {
	s.cleanings[c.Date.Format("02/01/2006")] = c
	return nil
}

func (s *store) AddCondo(c *models.Condo) error {
	s.condos[c.Date.Format("02/01/2006")] = c
	return nil
}

func (s *store) AddBill(e *models.EnergyBill) error {
	s.bills[e.Date.Format("02/01/2006")] = e
	return nil
}

func (s *store) AddRent(r *models.Rent) error {
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
