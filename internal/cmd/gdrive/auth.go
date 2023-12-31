package gdrive

// https://developers.google.com/drive/api/quickstart/go

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

var GDriveAuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Google Drive",
	Long:  `Authenticate with Google Drive`,
	Run:   GDriveAuthCommand,
}

func GDriveAuthCommand(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	b, err := os.ReadFile("resources/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved "token.json".
	config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	// Test creating a new GDrive client
	_, err = drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}
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

// Request a token from the web, then returns the retrieved token.
const butlerGDriveAuthTokenWebServerAddr = "butler-gdrive-auth-token-web-server-addr"

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	var tok *oauth2.Token
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// get auth code from query string
		authCode := r.URL.Query().Get("code")

		tok, err = config.Exchange(context.TODO(), authCode)
		if err != nil {
			fmt.Fprintf(w, "Unable to retrieve token from web %v", err)
			log.Fatalf("Unable to retrieve token from web %v", err)
		}

		// shutdown the server after request is handled
		cancel()

		fmt.Fprintf(w, "Login successful!")
	})

	server := &http.Server{
		Addr:    ":8081",
		Handler: router,
		BaseContext: func(listener net.Listener) context.Context {
			ctx := context.WithValue(ctx, butlerGDriveAuthTokenWebServerAddr, listener.Addr().String())
			return ctx
		},
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Unable to start a HTTP server: %v", err)
		}
	}()

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser: \n%v\n", authURL)

	<-ctx.Done()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Unable to shutdown the HTTP server: %v", err)
	}

	return tok
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

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	filename := "resources/token.json"
	tok, err := tokenFromFile(filename)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(filename, tok)
	}

	return config.Client(context.Background(), tok)
}
