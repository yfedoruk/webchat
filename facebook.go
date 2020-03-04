package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

const fbCallback string = "/oauth2callback"

var (
	fbConf  *oauth2.Config
	fbState = "this_should_be_random"
)

func init() {
	c := config{}
	c.set("facebook")
	fbConf = &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  "https://" + host() + fbCallback,
		Scopes:       []string{"public_profile"},
		Endpoint:     facebook.Endpoint,
	}
}

type fbUser struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Picture struct {
		Data struct {
			Url string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
}

func handleFacebookLogin(w http.ResponseWriter, r *http.Request) {
	Url, err := url.Parse(fbConf.Endpoint.AuthURL)
	if err != nil {
		log.Fatal("Parse: ", err)
	}
	parameters := url.Values{}
	parameters.Add("client_id", fbConf.ClientID)
	parameters.Add("scope", strings.Join(fbConf.Scopes, " "))
	parameters.Add("redirect_uri", fbConf.RedirectURL)
	parameters.Add("response_type", "code")
	parameters.Add("state", fbState)
	Url.RawQuery = parameters.Encode()
	authUrl := fbConf.AuthCodeURL(fbState)
	http.Redirect(w, r, authUrl, http.StatusTemporaryRedirect)
}

func handleFacebookCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != fbState {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", fbState, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")

	token, err := fbConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	resp, err := http.Get("https://graph.facebook.com/me?access_token=" + url.QueryEscape(token.AccessToken) + "&fields=id,name,picture{url}")
	if err != nil {
		fmt.Printf("Get: %s\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ReadAll: %s\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	fbUser := fbUser{}
	_ = json.Unmarshal(response, &fbUser)

	cookie{
		fbUser.Name,
		fbUser.Picture.Data.Url,
	}.set(w)

	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)
	// http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
