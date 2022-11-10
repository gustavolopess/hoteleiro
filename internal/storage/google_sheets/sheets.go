package google_sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gustavolopess/hoteleiro/internal/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

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
	return s.appendData(c.Apartment, "L3", [][]interface{}{
		{c.Date, c.Value, c.Cleaner},
	})
}

// AddCondo adds a new condo payment to the Condo table in the apartment Sheet
func (s *SheetsClient) AddCondo(c *models.Condo) error {
	return s.appendData(c.Apartment, "I3", [][]interface{}{
		{c.Date, c.Value},
	})
}

// AddApartment adds a new sheet on spreadsheet, which represents an apartment
func (s *SheetsClient) AddApartment(a *models.Apartment) error {
	return nil
}

// AddBill appends data to the Bill table in the apartment sheet
func (s *SheetsClient) AddBill(e *models.EnergyBill) error {
	return s.appendData(e.Apartment, "F3", [][]interface{}{
		{e.Date, e.Value},
	})
}

// AddRent appends data to the rent table in the apartment sheet
func (s *SheetsClient) AddRent(r *models.Rent) error {
	return s.appendData(r.Apartment, "A3", [][]interface{}{
		{r.DateBegin.Local().String(), r.DateEnd.Local().String(), r.Value, r.Renter},
	})
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
