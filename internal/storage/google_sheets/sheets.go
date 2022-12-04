package google_sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/gustavolopess/hoteleiro/internal/format"
	"github.com/gustavolopess/hoteleiro/internal/models"
	"github.com/gustavolopess/hoteleiro/internal/storage/s3_client"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const rentCell = "A3"
const rentDatesCells = "A3:E"

const billCell = "F3"
const readBillCells = "F3:H"

const condoCell = "I3"
const readCondosCells = "I3:K"

const cleaningCell = "L3"
const readCleaningCells = "L3:N"

const miscellaneousExpenseCell = "O3"
const readMiscellaneousExpenseCells = "O3:R"

const amortizationCell = "S3"
const readAmortizationCells = "S3:U"

const financingInstallmentCell = "V3"
const readFinancialInstallmentCells = "V3:X"

const dateLayout = "02/01/2006"

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	s3Client := s3_client.GetS3Client()
	tok := s3Client.GetGoogleSheetsAuthToken()

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

func NewSheetsClient(ctx context.Context, sheetsId string, credentialsJson []byte) *SheetsClient {
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(credentialsJson, "https://www.googleapis.com/auth/spreadsheets")
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
	existingCleanings, err := s.GetPayedCleanings(c.Apartment)
	if err != nil {
		return err
	}

	existingCleanings = append(existingCleanings, c)
	sort.Slice(existingCleanings, func(i, j int) bool {
		return existingCleanings[i].Date.Before(existingCleanings[j].Date)
	})

	var dataToWrite [][]interface{}
	for _, cleaning := range existingCleanings {
		dataToWrite = append(dataToWrite, []interface{}{cleaning.Date.Format(dateLayout), cleaning.Value, cleaning.Payer})
	}

	return s.upsertDataInRange(c.Apartment, cleaningCell, dataToWrite)
}

// AddCondo adds a new condo payment to the Condo table in the apartment Sheet
func (s *SheetsClient) AddCondo(c *models.Condo) error {
	existingCondos, err := s.GetPayedCondos(c.Apartment)
	if err != nil {
		return err
	}

	existingCondos = append(existingCondos, c)
	sort.Slice(existingCondos, func(i, j int) bool {
		return existingCondos[i].Date.Before(existingCondos[j].Date)
	})

	var dataToWrite [][]interface{}
	for _, condo := range existingCondos {
		dataToWrite = append(dataToWrite, []interface{}{condo.Date.Format(dateLayout), condo.Value, condo.Payer})
	}

	return s.upsertDataInRange(c.Apartment, condoCell, dataToWrite)
}

// AddApartment adds a new sheet on spreadsheet, which represents an apartment
func (s *SheetsClient) AddApartment(a *models.Apartment) error {
	return nil
}

// AddBill appends data to the Bill table in the apartment sheet
func (s *SheetsClient) AddBill(e *models.EnergyBill) error {
	existingBills, err := s.GetPayedBills(e.Apartment)
	if err != nil {
		return err
	}

	existingBills = append(existingBills, e)
	sort.Slice(existingBills, func(i, j int) bool {
		return existingBills[i].Date.Before(existingBills[j].Date)
	})

	var dataToWrite [][]interface{}
	for _, b := range existingBills {
		dataToWrite = append(dataToWrite, []interface{}{b.Date.Format(dateLayout), b.Value, b.Payer})
	}

	return s.upsertDataInRange(e.Apartment, billCell, dataToWrite)
}

// AddRent appends data to the rent table in the apartment sheet
func (s *SheetsClient) AddRent(r *models.Rent) error {
	existingRents, err := s.GetExistingRents(r.Apartment)
	if err != nil {
		return err
	}

	existingRents = append(existingRents, r)
	sort.Slice(existingRents, func(i, j int) bool {
		return existingRents[i].DateBegin.Before(existingRents[j].DateBegin)
	})

	var dataToWrite [][]interface{}
	for _, rent := range existingRents {
		dataToWrite = append(dataToWrite, []interface{}{
			rent.DateBegin.Format(dateLayout), rent.DateEnd.Format(dateLayout), rent.Value, rent.Renter, rent.Receiver,
		})
	}

	return s.upsertDataInRange(r.Apartment, rentCell, dataToWrite)
}

func (s *SheetsClient) AddMiscellaneousExpense(m *models.MiscellaneousExpense) error {
	existingExpenses, err := s.GetMiscellaneousExpenses(m.Apartment)
	if err != nil {
		return err
	}

	existingExpenses = append(existingExpenses, m)
	sort.Slice(existingExpenses, func(i, j int) bool {
		return existingExpenses[i].Date.Before(existingExpenses[j].Date)
	})

	var dataToWrite [][]interface{}
	for _, e := range existingExpenses {
		dataToWrite = append(dataToWrite, []interface{}{e.Date.Format(dateLayout), e.Value, e.Description, e.Payer})
	}

	return s.upsertDataInRange(m.Apartment, miscellaneousExpenseCell, dataToWrite)
}

func (s *SheetsClient) GetMiscellaneousExpenses(apartment models.Apartment) ([]*models.MiscellaneousExpense, error) {
	data, err := s.readDataFromRange(apartment, readMiscellaneousExpenseCells)
	if err != nil {
		return nil, err
	}

	expenses := make([]*models.MiscellaneousExpense, 0)
	for _, row := range data {
		date, err := time.Parse(dateLayout, row[0].(string))
		if err != nil {
			log.Println("failed to parse date", err.Error(), row)
			return nil, err
		}

		value, err := format.BrlToFloat64(row[1].(string))
		if err != nil {
			log.Println("failed to parse value of expense", row)
			return nil, err
		}

		expenses = append(expenses, &models.MiscellaneousExpense{
			Date:        date,
			Value:       value,
			Description: row[2].(string),
			Payer:       row[3].(string),
			Apartment:   apartment,
		})
	}

	return expenses, nil
}

func (s *SheetsClient) AddAmortization(a *models.Amortization) error {
	payedAmortizations, err := s.GetPayedAmortizations(a.Apartment)
	if err != nil {
		return err
	}

	payedAmortizations = append(payedAmortizations, a)
	sort.Slice(payedAmortizations, func(i, j int) bool {
		return payedAmortizations[i].Date.Before(payedAmortizations[j].Date)
	})

	var dataToWrite [][]interface{}
	for _, pa := range payedAmortizations {
		dataToWrite = append(dataToWrite, []interface{}{pa.Date.Format(dateLayout), pa.Value, pa.Payer})
	}

	return s.upsertDataInRange(a.Apartment, amortizationCell, dataToWrite)
}

func (s *SheetsClient) GetPayedAmortizations(apartment models.Apartment) ([]*models.Amortization, error) {
	payedAmortizationsData, err := s.readDataFromRange(apartment, readAmortizationCells)
	if err != nil {
		return nil, err
	}

	payedAmortizations := make([]*models.Amortization, 0)
	for _, am := range payedAmortizationsData {
		date, err := format.DDMMYYYYstringToTimeObj(am[0].(string))
		if err != nil {
			log.Println("failed to parse date", err.Error(), am)
			return nil, err
		}

		value, err := format.BrlToFloat64(am[1].(string))
		if err != nil {
			log.Println("failed to parse value of financial installment", am)
			return nil, err
		}

		payedAmortizations = append(payedAmortizations, &models.Amortization{
			Date:      date,
			Value:     value,
			Payer:     am[2].(string),
			Apartment: apartment,
		})
	}

	return payedAmortizations, nil
}

func (s *SheetsClient) AddFinancingInstallment(f *models.FinancingInstallment) error {
	payedFinancialInstallments, err := s.GetPayedFinancialInstallments(f.Apartment)
	if err != nil {
		return err
	}

	payedFinancialInstallments = append(payedFinancialInstallments, f)
	sort.Slice(payedFinancialInstallments, func(i, j int) bool {
		return payedFinancialInstallments[i].Date.Before(payedFinancialInstallments[j].Date)
	})

	var dataToWrite [][]interface{}
	for _, pfi := range payedFinancialInstallments {
		dataToWrite = append(dataToWrite, []interface{}{pfi.Date.Format(dateLayout), pfi.Value, pfi.Payer})
	}

	return s.upsertDataInRange(f.Apartment, financingInstallmentCell, dataToWrite)
}

func (s *SheetsClient) GetPayedFinancialInstallments(apartment models.Apartment) ([]*models.FinancingInstallment, error) {
	payedFinancialInstallmentsData, err := s.readDataFromRange(apartment, readFinancialInstallmentCells)
	if err != nil {
		return nil, err
	}

	payedFinancialInstallments := make([]*models.FinancingInstallment, 0)
	for _, fi := range payedFinancialInstallmentsData {
		date, err := format.DDMMYYYYstringToTimeObj(fi[0].(string))
		if err != nil {
			log.Println("failed to parse date", err.Error(), fi)
			return nil, err
		}

		value, err := format.BrlToFloat64(fi[1].(string))
		if err != nil {
			log.Println("failed to parse value of financial installment", fi)
			return nil, err
		}

		payedFinancialInstallments = append(payedFinancialInstallments, &models.FinancingInstallment{
			Date:      date,
			Value:     value,
			Payer:     fi[2].(string),
			Apartment: apartment,
		})
	}

	return payedFinancialInstallments, nil
}

func (s *SheetsClient) GetExistingRents(apartment models.Apartment) ([]*models.Rent, error) {
	existingRentsData, err := s.readDataFromRange(apartment, rentDatesCells)
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

		value, err := format.BrlToFloat64(rent[2].(string))
		if err != nil {
			log.Println("failed to parse value of rent", rent)
			return nil, err
		}

		existingRents = append(existingRents, &models.Rent{
			DateBegin: dateBegin,
			DateEnd:   dateEnd,
			Value:     value,
			Renter:    rent[3].(string),
			Receiver:  rent[4].(string),
			Apartment: apartment,
		})
	}

	return existingRents, nil
}

func (s *SheetsClient) GetPayedCondos(apartment models.Apartment) ([]*models.Condo, error) {
	payedCondosData, err := s.readDataFromRange(apartment, readCondosCells)
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

		value, err := format.BrlToFloat64(condo[1].(string))
		if err != nil {
			log.Println("failed to parse value of condo", err.Error(), condo)
		}

		existingCondos = append(existingCondos, &models.Condo{
			Value:     value,
			Date:      date,
			Payer:     condo[2].(string),
			Apartment: apartment,
		})
	}

	return existingCondos, nil
}

func (s *SheetsClient) GetPayedBills(apartment models.Apartment) ([]*models.EnergyBill, error) {
	payedBillsData, err := s.readDataFromRange(apartment, readBillCells)
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

		value, err := format.BrlToFloat64(bill[1].(string))
		if err != nil {
			log.Println("failed to parse value of bill", err.Error(), bill)
		}

		existingBills = append(existingBills, &models.EnergyBill{
			Value:     value,
			Date:      date,
			Payer:     bill[2].(string),
			Apartment: apartment,
		})
	}

	return existingBills, nil
}

func (s *SheetsClient) GetPayedCleanings(apartment models.Apartment) ([]*models.Cleaning, error) {
	payedCleaningsData, err := s.readDataFromRange(apartment, readCleaningCells)
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

		value, err := format.BrlToFloat64(cleaning[1].(string))
		if err != nil {
			log.Println("failed to parse value of cleaning", err.Error(), cleaning)
		}

		existingCleanings = append(existingCleanings, &models.Cleaning{
			Value:     value,
			Date:      date,
			Payer:     cleaning[2].(string),
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

func (s *SheetsClient) upsertDataInRange(apartment models.Apartment, upsertRange string, data [][]interface{}) error {
	resp, err := s.Spreadsheets.Values.Update(s.sheetsId, upsertRange, &sheets.ValueRange{
		Range:  upsertRange,
		Values: data,
	}).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return err
	}

	log.Printf("values added to range %s", resp.UpdatedRange)
	return nil
}

func (s *SheetsClient) readDataFromRange(apartment models.Apartment, readRange string) ([][]interface{}, error) {
	cells, err := s.Spreadsheets.Values.Get(s.sheetsId, readRange).ValueRenderOption("FORMATTED_VALUE").Do()
	if err != nil {
		return nil, err
	}

	return cells.Values, nil
}
