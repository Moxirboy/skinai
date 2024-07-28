package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Define the structs to represent the JSON structure
type Response struct {
	Data []Item `json:"data"`
}

type Item struct {
	ID             int    `json:"id"`
	CategoryID     int    `json:"category_id"`
	Date           string `json:"date"`
	Title          string `json:"title"`
	Anons          string `json:"anons"`
	Views          int    `json:"views"`
	AnonsImage     string `json:"anons_image"`
	CategoryTitle  string `json:"category_title"`
	CategoryCode   string `json:"category_code"`
	ActivityTitle  string `json:"activity_title"`
	ActivityCode   string `json:"activity_code"`
	UrlToWeb   string `json:"url_to_web"`
}

func main() {
	// URL of the API endpoint that provides the JSON data
	url := "https://api-portal.gov.uz/news/category?code_name=news&page=1"

	// Make the HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making GET request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	// Print the response body for debugging
	fmt.Println("Response body:", string(body))

	// Unmarshal the JSON data into the Response struct
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	// Print the details of each item
	for _, item := range response.Data {
		fmt.Printf("ID: %d\n", item.ID)
		fmt.Printf("Category ID: %d\n", item.CategoryID)
		fmt.Printf("Date: %s\n", item.Date)
		fmt.Printf("Title: %s\n", item.Title)
		fmt.Printf("Anons: %s\n", item.Anons)
		fmt.Printf("Views: %d\n", item.Views)
		fmt.Printf("Anons Image: %s\n", item.AnonsImage)
		fmt.Printf("Category Title: %s\n", item.CategoryTitle)
		fmt.Printf("Category Code: %s\n", item.CategoryCode)
		fmt.Printf("Activity Title: %s\n", item.ActivityTitle)
		fmt.Printf("Activity Code: %s\n", item.ActivityCode)
		fmt.Printf("https://gov.uz/news/view/%d/\n",item.ID)
		fmt.Println()
	}
}
