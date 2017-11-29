package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

type option struct {
	key, value string
}

func (o option) Get() (key, value string) {
	return o.key, o.value
}

func getMessageIds() ([]string, error) {
	srv, _ := newGmailService()
	messageIds := make([]string, 0)
	user := "me"
	query := option{key: "maxResults", value: "20"}
	r, err := srv.Users.Messages.List(user).Do(query)
	if err != nil {
		log.Fatalf("Unable to retrieve labels. %v", err)
		return nil, err
	}

	for _, message := range r.Messages {
		messageIds = append(messageIds, message.Id)
	}

	return messageIds, nil
}

func getSubjects(messageIds []string, ch chan<- string) {
	srv, _ := newGmailService()

	fmt.Println("starting to get messages")
	for _, id := range messageIds {
		go func(messageId string) {
			ch <- getMessage(srv, messageId)
		}(id)
	}
}

func getMessage(client *gmail.Service, id string) string {
	res, err := client.Users.Messages.Get("me", id).Do()
	if err != nil {
		// TODO: handle error
		log.Fatalf("oh no %v\n", err)
	}
	for _, header := range res.Payload.Headers {
		if header.Name == "Subject" {
			return header.Value
		}
	}
	return ""
}

func newGmailService() (*gmail.Service, error) {
	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/gmail-go-quickstart.json
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	return gmail.New(client)
}

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("gmail-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// func main() {
// 	ctx := context.Background()

// 	b, err := ioutil.ReadFile("client_secret.json")
// 	if err != nil {
// 		log.Fatalf("Unable to read client secret file: %v", err)
// 	}

// 	// If modifying these scopes, delete your previously saved credentials
// 	// at ~/.credentials/gmail-go-quickstart.json
// 	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
// 	if err != nil {
// 		log.Fatalf("Unable to parse client secret file to config: %v", err)
// 	}
// 	client := getClient(ctx, config)

// 	srv, err := gmail.New(client)
// 	if err != nil {
// 		log.Fatalf("Unable to retrieve gmail Client %v", err)
// 	}

// 	messageIds := make([]string, 0)
// 	user := "me"
// 	r, err := srv.Users.Messages.List(user).Do()
// 	if err != nil {
// 		log.Fatalf("Unable to retrieve labels. %v", err)
// 	}

// 	for _, message := range r.Messages {
// 		messageIds = append(messageIds, message.Id)
// 	}

// 	fmt.Println(messageIds)
// 	fmt.Printf("%v messages\n", len(messageIds))

// 	ch := make(chan string)

// 	for _, id := range messageIds {
// 		go getMessage(srv, id, ch)
// 	}

// 	subjects := make([]string, 0)
// 	for e := range ch {
// 		fmt.Printf(">> %v\n", e)
// 		subjects = append(subjects, e)
// 		if len(subjects) >= len(messageIds) {
// 			close(ch)
// 		}
// 	}

// 	fmt.Println(subjects)
// }
