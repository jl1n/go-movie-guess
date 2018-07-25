package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Results is the movie results
type Results struct {
	Page    int      `json:"page"`
	Results []Result `json:"results"`
}

// Result is a specific movie
type Result struct {
	Title      string `json:"title"`
	PosterPath string `json:"poster_path"`
	Overview   string `json:"overview"`
}

// GuessResult is the result of the user's guess
type GuessResult struct {
	Correct    string
	PosterPath string
}

var currMovie Result
var guessResult GuessResult

func main() {
	rand.Seed(time.Now().UnixNano())

	templates := template.Must(template.ParseFiles("index.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if err := templates.ExecuteTemplate(w, "index.html", guessResult); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
		var result Result
		var err error

		if result, err = getRandomMovieInfo(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Write([]byte(result.Overview))

	})

	// when user presses guess button
	http.HandleFunc("/guess", func(w http.ResponseWriter, r *http.Request) {

		// ignore punctuation of guess
		reg, err := regexp.Compile("[^a-zA-Z0-9]+")
		if err != nil {
			fmt.Println("regexp error", err)
		}
		processedAnswer := reg.ReplaceAllString(currMovie.Title, "")
		processedResponse := reg.ReplaceAllString(r.FormValue("guess"), "")

		// check if guess matches, case insensitive
		if strings.EqualFold(processedAnswer, processedResponse) {
			guessResult.Correct = "Correct"

		} else {
			guessResult.Correct = "Incorrect"

		}
		// send url of movie poster
		guessResult.PosterPath = currMovie.PosterPath

		// convert to json to send to frontend
		js, err := json.Marshal(guessResult)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write(js)
	})

	fmt.Println(http.ListenAndServe(":8080", nil))
}

func getRandomMovieInfo() (Result, error) {
	var response *http.Response
	var err error
	base := "https://api.themoviedb.org/3"
	endpoint := "/movie/popular"
	apikey := ""

	// mac keychain workaround
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// api call
	if response, err = client.Get(base + endpoint + "?api_key=" + apikey); err != nil {
		return Result{}, err
	}

	defer response.Body.Close()
	var body []byte
	if body, err = ioutil.ReadAll(response.Body); err != nil {
		return Result{}, err
	}

	var data Results

	// read json data
	err = json.Unmarshal(body, &data)

	if err != nil {
		fmt.Println("Cannot format JSON", err)
	}

	// pick random movie from list
	randMovie := rand.Intn(len(data.Results) - 1)

	// save movie info
	currMovie = data.Results[randMovie]

	return data.Results[randMovie], err

}
