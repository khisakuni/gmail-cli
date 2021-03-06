package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"google.golang.org/api/googleapi"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

type messages struct {
	currentPageToken string
	nextPageToken    string
	prevPageTokens   []string
	gmailService     *gmail.Service
}

type message struct {
	subject string
	date    int64
	body    []byte
	sender  string
}

type byDate []message

func (d byDate) Len() int           { return len(d) }
func (d byDate) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d byDate) Less(i, j int) bool { return d[i].date < d[j].date }

type option struct {
	key, value string
}

type messagesResponse struct {
	messageIds    []string
	nextPageToken string
}

func (o option) Get() (key, value string) {
	return o.key, o.value
}

func newMessages() (*messages, error) {
	m := &messages{}
	srv, err := newGmailService()
	if err != nil {
		return nil, err
	}
	m.gmailService = srv

	return m, nil
}

func (m *messages) getNext() ([]string, error) {
	query := []googleapi.CallOption{
		option{key: "maxResults", value: "20"},
		option{key: "pageToken", value: m.nextPageToken},
	}

	res, err := getMessageIds(m, query)
	m.prevPageTokens = append([]string{m.currentPageToken}, m.prevPageTokens...)
	m.currentPageToken = m.nextPageToken
	m.nextPageToken = res.nextPageToken
	return res.messageIds, err
}

func (m *messages) getPrev() ([]string, error) {
	if len(m.prevPageTokens) < 2 {
		return []string{}, nil
	}
	m.nextPageToken = m.currentPageToken
	m.currentPageToken = m.prevPageTokens[0]
	m.prevPageTokens = m.prevPageTokens[1:]

	query := []googleapi.CallOption{
		option{key: "maxResults", value: "20"},
		option{key: "pageToken", value: m.currentPageToken},
	}
	res, err := getMessageIds(m, query)
	return res.messageIds, err
}

func getMessages(ids []string) []message {
	ch := make(chan message)
	getSubjects(ids, ch)
	result := make([]message, 0, len(ids))

	for c := 0; c < len(ids); {
		select {
		case subject := <-ch:
			result = append(result, subject)
			c++
		}
	}
	close(ch)

	return result
}

func getMessageIds(m *messages, query []googleapi.CallOption) (messagesResponse, error) {
	user := "me"
	r, err := m.gmailService.Users.Messages.List(user).Do(query...)
	res := messagesResponse{}
	if err != nil {
		log.Fatalf("Unable to retrieve labels. %v", err)
		return res, err
	}

	for _, message := range r.Messages {
		res.messageIds = append(res.messageIds, message.Id)
	}
	res.nextPageToken = r.NextPageToken

	return res, nil
}

func getSubjects(messageIds []string, ch chan<- message) {
	srv, _ := newGmailService()

	for _, id := range messageIds {
		go func(messageId string) {
			ch <- getMessage(srv, messageId)
		}(id)
	}
}

func getMessage(client *gmail.Service, id string) message {
	res, err := client.Users.Messages.Get("me", id).Do()
	if err != nil {
		// TODO: handle error
		log.Fatalf("oh no %v\n", err)
	}
	mes := message{date: res.InternalDate}
	// mes.body = res.Snippet
	// res.Payload.Body.Data
	body, _ := base64.URLEncoding.DecodeString(res.Payload.Body.Data)
	if len(body) > 0 {
		mes.body = body
	}

	for _, part := range res.Payload.Parts {

		if part.MimeType == "text/html" {
			body, _ := base64.URLEncoding.DecodeString(part.Body.Data)

			// fmt.Println(string(body))
			mes.body = body
			// fmt.Println(mes.body)
		} else if part.MimeType == "text/plain" {
			body, _ := base64.URLEncoding.DecodeString(part.Body.Data)
			mes.body = body
		}
	}

	// body, _ := base64.URLEncoding.DecodeString(res.Payload.Body.Data)
	// mes.body = string(body)

	for _, header := range res.Payload.Headers {
		if header.Name == "Subject" {
			mes.subject = header.Value
		}
		if header.Name == "From" {
			mes.sender = header.Value
		}
	}
	// mes.body = mes.subject + mes.sender
	return mes
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
