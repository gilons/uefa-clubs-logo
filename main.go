package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gocolly/colly"
)

type Club struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Path string `json:"path"`
}

func main() {
	addr := ":7772"
	http.HandleFunc("/ping", pin)
	http.HandleFunc("/search", getAllImageLinks)
	fmt.Println("Listener started at ...")
	log.Fatal(http.ListenAndServe(addr, nil))
}

func pin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ping")
	w.Write([]byte("pong"))
}

func getAllImageLinks(w http.ResponseWriter, r *http.Request) {
	//Verify the param "URL" exists
	URL := r.URL.Query().Get("url")
	if URL == "" {
		log.Println("missing URL argument")
		return
	}
	log.Println("visiting", URL)

	//Create a new collector which will be in charge of collect the data from HTML
	main := colly.NewCollector()
	var response []Club
	main.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link != "" {
			pageClubs := collectForPage(link)
			response = append(response, pageClubs...)
		}
	})
	main.Visit(URL)
	b, err := json.MarshalIndent(response, "", "")
	if err != nil {
		log.Println("failed to serialize response:", err)
		return
	}
	ioutil.WriteFile("eufa_clubs.json", b, 0644)
	// Add some header and write the body for our endpoint
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}

func collectForPage(url string) []Club {

	c := colly.NewCollector()
	v := colly.NewCollector()
	//Slices to store the data
	var response []Club

	//onHTML function allows the collector to use a callback function when the specific HTML tag is reached
	//in this case whenever our collector finds an
	//anchor tag with href it will call the anonymous function
	// specified below which will get the info from the href and append it to our slice
	c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("src"))
		if link != "" {
			path, err := DownloadFile(link)
			if err != nil {
				log.Fatal(err)
			}
			response = append(response, Club{
				URL:  link,
				Name: "",
				Path: path,
			})
		}
	})

	//Command to visit the website
	c.Visit(url)
	index := 0
	v.OnHTML("h3", func(e *colly.HTMLElement) {
		text := e.Text
		fmt.Println("visiting this one: ")
		fmt.Println(text)
		response[index].Name = text
		index++

	})
	v.Visit(url)
	return response
}

/*DownloadFile funcion This function is used to download each image collected*/
func DownloadFile(url string) (string, error) {
	// Create the file
	urlArr := strings.Split(url, "/")
	filename := urlArr[len(urlArr)-1]
	path := "../uefa/" + filename
	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return filepath.Abs(path)
}
