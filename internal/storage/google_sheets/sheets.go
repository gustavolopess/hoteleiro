package google_sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gustavolopess/hoteleiro/internal/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const rentCell = "A3"
const rentDatesCells = "A3:D"
const condoCell = "I3"
const readCondosCells = "I3:J"
const cleaningCell = "L3"
const readCleaningCells = "L3:N"
const billCell = "F3"
const readBillCells = "F3:G"

const dateLayout = "02/01/2006"

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

type SheetsClient struct {
	*sheets.Service
	sheetsId string
}

func NewSheetsClient(ctx context.Context, credentialsFile, sheetsId string) *SheetsClient {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	return &SheetsClient{srv, sheetsId}
}

// AddCleaning adds a new cleaning fee to the Cleaning table in the apartment sheet
func (s *SheetsClient) AddCleaning(c *models.Cleaning) error {
	return s.appendData(c.Apartment, cleaningCell, [][]interface{}{
		{c.Date.Format(dateLayout), c.Value, c.Cleaner},
	})
}

// AddCondo adds a new condo payment to the Condo table in the apartment Sheet
func (s *SheetsClient) AddCondo(c *models.Condo) error {
	return s.appendData(c.Apartment, condoCell, [][]interface{}{
		{c.Date.Format(dateLayout), c.Value},
	})
}

// AddApartment adds a new sheet on spreadsheet, which represents an apartment
func (s *SheetsClient) AddApartment(a *models.Apartment) error {
	return nil
}

// AddBill appends data to the Bill table in the apartment sheet
func (s *SheetsClient) AddBill(e *models.EnergyBill) error {
	return s.appendData(e.Apartment, billCell, [][]interface{}{
		{e.Date.Format(dateLayout), e.Value},
	})
}

// AddRent appends data to the rent table in the apartment sheet
func (s *SheetsClient) AddRent(r *models.Rent) error {
	return s.appendData(r.Apartment, rentCell, [][]interface{}{
		{r.DateBegin.Format(dateLayout), r.DateEnd.Format(dateLayout), r.Value, r.Renter},
	})
}

func (s *SheetsClient) GetExistingRents(apartment models.Apartment) ([]*models.Rent, error) {
	existingRentsData, err := s.readData(apartment, rentDatesCells)
	if err != nil {
		return nil, err
	}

	existingRents := make([]*models.Rent, 0)

	for _, rent := range existingRentsData {
		dateBegin, err := time.Parse(dateLayout, rent[0].(string))
		if err != nil {
			log.Println("failed to parse dateBegin", err.Error(), rent)
			return nil, err
		}

		dateEnd, err := time.Parse(dateLayout, rent[1].(string))
		if err != nil {
			log.Println("failed to parse dateEnd", err.Error(), rent)
			return nil, err
		}

		value, err := strconv.ParseFloat(rent[2].(string), 32)
		if err != nil {
			log.Println("failed to parse value of rent", rent)
		}

		existingRents = append(existingRents, &models.Rent{
			DateBegin: dateBegin,
			DateEnd:   dateEnd,
			Value:     value,
			Renter:    rent[3].(string),
			Apartment: apartment,
		})
	}

	return existingRents, nil
}

func (s *SheetsClient) GetPayedCondos(apartment models.Apartment) ([]*models.Condo, error) {
	payedCondosData, err := s.readData(apartment, readCondosCells)
	if err != nil {
		return nil, err
	}

	existingCondos := make([]*models.Condo, 0)
	for _, condo := range payedCondosData {
		date, err := time.Parse(dateLayout, condo[0].(string))
		if err != nil {
			log.Println("failed to parse date of condo bill", err.Error(), condo)
			return nil, err
		}

		value, err := strconv.ParseFloat(condo[1].(string), 32)
		if err != nil {
			log.Println("failed to parse value of condo", err.Error(), condo)
		}

		existingCondos = append(existingCondos, &models.Condo{
			Value:     value,
			Date:      date,
			Apartment: apartment,
		})
	}

	return existingCondos, nil
}

func (s *SheetsClient) GetPayedBills(apartment models.Apartment) ([]*models.EnergyBill, error) {
	payedBillsData, err := s.readData(apartment, readBillCells)
	if err != nil {
		return nil, err
	}

	existingBills := make([]*models.EnergyBill, 0)
	for _, bill := range payedBillsData {
		date, err := time.Parse(dateLayout, bill[0].(string))
		if err != nil {
			log.Println("failed to parse date of bill", err.Error(), bill)
			return nil, err
		}

		value, err := strconv.ParseFloat(bill[1].(string), 32)
		if err != nil {
			log.Println("failed to parse value of bill", err.Error(), bill)
		}

		existingBills = append(existingBills, &models.EnergyBill{
			Value:     value,
			Date:      date,
			Apartment: apartment,
		})
	}

	return existingBills, nil
}

func (s *SheetsClient) GetPayedCleanings(apartment models.Apartment) ([]*models.Cleaning, error) {
	payedCleaningsData, err := s.readData(apartment, readCleaningCells)
	if err != nil {
		return nil, err
	}

	existingCleanings := make([]*models.Cleaning, 0)
	for _, cleaning := range payedCleaningsData {
		date, err := time.Parse(dateLayout, cleaning[0].(string))
		if err != nil {
			log.Println("failed to parse date of cleaning", err.Error(), cleaning)
			return nil, err
		}

		value, err := strconv.ParseFloat(cleaning[1].(string), 32)
		if err != nil {
			log.Println("failed to parse value of cleaning", err.Error(), cleaning)
		}

		existingCleanings = append(existingCleanings, &models.Cleaning{
			Value:     value,
			Date:      date,
			Cleaner:   cleaning[2].(string),
			Apartment: apartment,
		})
	}

	return existingCleanings, nil
}

// GetAvailableApartments query the existing sheets and return its titles in an array
func (s *SheetsClient) GetAvailableApartments() ([]string, error) {
	sheetData, err := s.Spreadsheets.Get(s.sheetsId).Do()
	if err != nil {
		return nil, err
	}

	var apartmentNames []string
	for _, sheet := range sheetData.Sheets {
		apartmentNames = append(apartmentNames, sheet.Properties.Title)
	}

	return apartmentNames, nil
}

func (s *SheetsClient) appendData(apartment models.Apartment, appendRange string, data [][]interface{}) error {
	resp, err := s.Spreadsheets.Values.Append(s.sheetsId, appendRange, &sheets.ValueRange{
		Range:  appendRange,
		Values: data,
	}).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return err
	}

	log.Printf("values added to range %s", resp.Updates.UpdatedRange)
	return nil
}

func (s *SheetsClient) readData(apartment models.Apartment, readRange string) ([][]interface{}, error) {
	cells, err := s.Spreadsheets.Values.Get(s.sheetsId, readRange).ValueRenderOption("FORMATTED_VALUE").Do()
	if err != nil {
		return nil, err
	}

	return cells.Values, nil
}
