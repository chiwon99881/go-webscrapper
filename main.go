package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJobs struct {
	id       string
	title    string
	location string
	salary   string
	summary  string
}

var baseURL string = "https://kr.indeed.com/jobs?q=python&limit=50"

func cleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func getPage(pageNumber int) []extractedJobs {
	jobs := []extractedJobs{}
	queryString := pageNumber * 50
	url := baseURL + "&start=" + strconv.Itoa(queryString)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	searchCards := doc.Find(".jobsearch-SerpJobCard")
	searchCards.Each(func(i int, card *goquery.Selection) {
		id, _ := card.Attr("data-jk")
		title := cleanString(card.Find(".title>a").Text())
		location := cleanString(card.Find(".sjcl>span").Text())
		salary := cleanString(card.Find(".salaryText").Text())
		summary := cleanString(card.Find(".summary").Text())
		job := extractedJobs{
			id:       id,
			title:    title,
			location: location,
			salary:   salary,
			summary:  summary,
		}
		jobs = append(jobs, job)
	})
	return jobs
}

func getPages() int {
	pages := 0
	resp, err := http.Get(baseURL)
	if err != nil {
		log.Fatalln(err)
	}
	if resp.StatusCode != 200 {
		log.Fatalln("Status code is not 200 but we got this -> ", resp.StatusCode)
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})
	return pages
}

func writeJobsToCSV(jobs []extractedJobs) {
	file, err := os.Create("jobs.csv")
	if err != nil {
		log.Fatalln(err)
	}
	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"ID", "TITLE", "LOCATION", "SALARY", "SUMMARY"}
	wErr := w.Write(headers)
	if wErr != nil {
		log.Fatalln(wErr)
	}

	for _, job := range jobs {
		eachJob := []string{"https://kr.indeed.com/viewjob?jk=" + job.id, job.title, job.location, job.salary, job.summary}
		err := w.Write(eachJob)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func main() {
	finalJobs := []extractedJobs{}
	totalPages := getPages()

	for i := 0; i < totalPages; i++ {
		extractedJobs := getPage(i)
		finalJobs = append(finalJobs, extractedJobs...)
	}
	writeJobsToCSV(finalJobs)
	fmt.Println("CSV file done with this count: ", len(finalJobs))
}
