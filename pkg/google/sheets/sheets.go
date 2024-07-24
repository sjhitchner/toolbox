package sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Sheet struct {
	client *http.Client
	srv    *sheets.Service
}

func New(tokenFile, credentialsFile string) (*Sheet, error) {
	ctx := context.Background()
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to read client secret file")
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to parse client secret file to config")
	}

	client, err := getClient(config)
	if err != nil {
		return nil, err
	}

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	return &Sheet{
		client: client,
		srv:    srv,
	}, nil
}

func (t *Sheet) GetRange(spreadSheetID, rangeName string) ([][]interface{}, error) {
	// Read data from Google Sheets
	readRange := fmt.Sprintf("%s!%s", sheetID, rangeName)
	resp, err := t.srv.Spreadsheets.Values.Get(spreadSheetID, readRange).Do()
	if err != nil {
		return errors.Wrapf(err, "Unable to retrieve data from sheet")
	}

	// Process the data
	if len(resp.Values) == 0 {
		return errors.New("No data found.")
	}

	for _, row := range resp.Values {
		fmt.Println(row)
	}

	return nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) (*http.Client, error) {
	token, err := getToken(config)
	if err != nil {
		return nil, err
	}

	return config.Client(context.Background(), token), nil
}

func getToken(config *oauth2.Config, tokenFile string) (*oauth2.Token, error) {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	token, err := tokenFromFile(tokenFile)
	if err == nil {
		return token, nil
	}

	token, err = getTokenFromWeb(config)
	if err != nil {
		return nil, err
	}

	if err := saveToken(tokenFile, token); err != nil {
		return nil, err
	}

	return token, nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, errors.Wrapf(err, "Unable to read authorization code")
	}

	token, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to retrieve token from web")
	}

	return token, nil
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}

	if err := json.NewDecoder(f).Decode(token); err != nil {
		return nil, err
	}

	return token, nil
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrapf(err, "Unable to cache oauth token")
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(token); err != nil {
		return err
	}

	return nil
}
