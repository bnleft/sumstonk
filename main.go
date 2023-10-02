package main

import (
	"bytes"
	language "cloud.google.com/go/language/apiv2"
	"cloud.google.com/go/language/apiv2/languagepb"
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/metaphorsystems/metaphor-go"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
	"regexp"
	"strconv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Metaphor Client
	metaphorApiKey := os.Getenv("METAPHOR_API_KEY")

	client, err := metaphor.NewClient(metaphorApiKey)
	if err != nil {
		log.Fatal(err)
		return
	}

	// User Input
	var stonk string
	fmt.Print("Enter a stonk (ex: AAPL): ")
	_, err = fmt.Scanln(&stonk)

	if err != nil {
		log.Fatal(err)
		return
	}

	// Metaphor Search API
	searchQuery := "news on " + stonk
	ctx := context.Background()
	numResults := 5
	domains := []string{"cnbc.com", "fool.com", "cnn.com", "foxbusiness.com", "reddit.com"}
	startDate := time.Now().Format(time.RFC3339)

	searchResults, err := client.Search(
		ctx,
		searchQuery,
		metaphor.WithAutoprompt(true),
		metaphor.WithNumResults(numResults),
		metaphor.WithIncludeDomains(domains),
		metaphor.WithStartPublishedDate(startDate),
	)

	if err != nil {
		log.Fatal(err)
		return
	}

	resultContents, err := searchResults.GetContents(ctx, client)
	if err != nil {
		log.Fatal(err)
		return
	}

	// We only need the urls
	urls := getURLs(resultContents)

	// SMMRY API
	summaries := getSummaries(urls)

	// Google Cloud Natural Language API
	sentiment := 0
	for i, _ := range summaries {
		buf := new(bytes.Buffer)

		analyzeSentiment(buf, summaries[i])
		value, sentimentString := getSentimentValue(buf.String())

		sentiment += value

		fmt.Printf("\nSource %d: %s", (i + 1), sentimentString)
	}

	score := float64(sentiment) / float64(len(summaries))
	sentimentString := sentimentToString(score)
	fmt.Printf("\n\nToday's sentiment: %s\n", sentimentString)
}

func sentimentToString(sentimentScore float64) string {
	// TODO: temporary thresholds
	if sentimentScore < -0.33 {
		return "negative"
	} else if sentimentScore < 0.33 {
		return "neutral"
	}

	return "positive"
}

func getURLs(response *metaphor.ContentsResponse) []string {
	var urls []string

	for _, result := range response.Contents {
		urls = append(urls, result.URL)
	}

	return urls
}

type SummaryResponseData struct {
	CharCount      string `json:"sm_api_character_count"`
	ContentReduced string `json:"sm_api_content_reduced"`
	Title          string `json:"sm_api_title"`
	Content        string `json:"sm_api_content"`
	Limitation     string `json:"sm_api_limitation"`
}

func getSummaries(urls []string) []string {
	var summaries []string

	smmryUrl := "https://api.smmry.com"
	smmryApiKey := os.Getenv("SMMRY_API_KEY")
	smmryCount := "3"

	for _, url := range urls {
		params := map[string]string{
			"SM_API_KEY": smmryApiKey,
			"SM_URL":     url,
			"SM_LENGTH":  smmryCount,
		}

		url := addURLParams(smmryUrl, params)

		// Call API
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
			break
		}
		defer resp.Body.Close()

		// Read response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
			break
		}

		// Parse body JSON
		var respData SummaryResponseData
		err = json.Unmarshal(body, &respData)
		if err != nil {
			log.Fatal(err)
			break
		}

		summary := respData.Content

		summaries = append(summaries, summary)
	}

	return summaries
}

func addURLParams(baseURL string, params map[string]string) string {
	u, _ := url.Parse(baseURL)
	q := u.Query()

	for key, value := range params {
		q.Add(key, value)
	}

	u.RawQuery = q.Encode()

	return u.String()
}

func analyzeSentiment(w io.Writer, text string) error {
	ctx := context.Background()

	client, err := language.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	resp, err := client.AnalyzeSentiment(ctx, &languagepb.AnalyzeSentimentRequest{
		Document: &languagepb.Document{
			Source: &languagepb.Document_Content{
				Content: text,
			},
			Type: languagepb.Document_PLAIN_TEXT,
		},
		EncodingType: languagepb.EncodingType_UTF8,
	})

	if err != nil {
		return fmt.Errorf("AnalyzeSentiment: %w", err)
	}
	fmt.Fprintf(w, "%q", resp)

	return nil
}

type Sentiment struct {
	Magnitude float64 `json:"magnitude"`
	Score     float64 `json:"score"`
}

func getSentimentValue(data string) (int, string) {
	// Define regular expressions to match magnitude and score
	magnitudeRegex := regexp.MustCompile(`magnitude:([0-9.-]+)`)
	scoreRegex := regexp.MustCompile(`score:([0-9.-]+)`)

	// Find the first magnitude and score using regular expressions
	magnitudeMatch := magnitudeRegex.FindStringSubmatch(data)
	scoreMatch := scoreRegex.FindStringSubmatch(data)

	if len(magnitudeMatch) < 2 && len(scoreMatch) < 2{
		log.Fatal("Magnitude and score not found in the string.")
		return 0, ""
	}

	magnitude, err := strconv.ParseFloat(magnitudeMatch[1], 64)
	if err != nil {
		log.Fatal(err)
		return 0, ""
	}

	score, err := strconv.ParseFloat(scoreMatch[1], 64)
	if err != nil {
		log.Fatal(err)
		return 0, ""
	}

	// TODO: determine threshold
	if magnitude > 1 {
		if score < -0.2 {
			return -1, "negative"
		} else if score > 0.2 {
			return 1, "positive"
		}
	}

	return 0, "neutral"
}
