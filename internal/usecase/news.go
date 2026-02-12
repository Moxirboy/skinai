package usecase

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testDeployment/internal/domain"
	"testDeployment/pkg/Bot"
	"testDeployment/pkg/utils"
	"time"
)

type newsUseCase struct {
	bot    Bot.Bot
	client *http.Client
}

func NewNewsUseCase(_ interface{}, bot Bot.Bot) INewsUseCase {
	return &newsUseCase{
		bot: bot,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

// ──────────────────────────────────────────────
// Source 1: Europe PMC  (JSON, free, no key)
// ──────────────────────────────────────────────

type epmcResponse struct {
	HitCount int `json:"hitCount"`
	Results  []struct {
		ID            string `json:"id"`
		Title         string `json:"title"`
		Abstract      string `json:"abstractText"`
		AuthorString  string `json:"authorString"`
		JournalTitle  string `json:"journalTitle"`
		DateOfCreat   string `json:"firstPublicationDate"`
		PMID          string `json:"pmid"`
		DOI           string `json:"doi"`
	} `json:"resultList>result"`
}

// Nested structure for the actual JSON shape
type epmcResponseRaw struct {
	HitCount   int `json:"hitCount"`
	ResultList struct {
		Result []struct {
			ID           string `json:"id"`
			Title        string `json:"title"`
			Abstract     string `json:"abstractText"`
			AuthorString string `json:"authorString"`
			JournalTitle string `json:"journalTitle"`
			DateOfCreat  string `json:"firstPublicationDate"`
			PMID         string `json:"pmid"`
			DOI          string `json:"doi"`
		} `json:"result"`
	} `json:"resultList"`
}

func (u *newsUseCase) fetchEuropePMC(query string, pageSize, offset int) ([]*domain.NewWithSinglePhoto, int, error) {
	apiURL := fmt.Sprintf(
		"https://www.ebi.ac.uk/europepmc/webservices/rest/search?query=%s&format=json&pageSize=%d&cursorMark=*&sort=DATE_CREATED+desc",
		url.QueryEscape(query), pageSize,
	)

	resp, err := u.client.Get(apiURL)
	if err != nil {
		return nil, 0, fmt.Errorf("europe pmc request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("europe pmc read body: %w", err)
	}

	var raw epmcResponseRaw
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, 0, fmt.Errorf("europe pmc parse json: %w (body: %.200s)", err, string(body))
	}

	var articles []*domain.NewWithSinglePhoto
	for _, r := range raw.ResultList.Result {
		abstract := r.Abstract
		if len(abstract) > 500 {
			abstract = abstract[:497] + "..."
		}
		owner := r.AuthorString
		if len(owner) > 200 {
			owner = owner[:197] + "..."
		}
		if owner == "" {
			owner = r.JournalTitle
		}

		link := ""
		if r.DOI != "" {
			link = "https://doi.org/" + r.DOI
		} else if r.PMID != "" {
			link = "https://pubmed.ncbi.nlm.nih.gov/" + r.PMID
		}

		articles = append(articles, &domain.NewWithSinglePhoto{
			ID:        r.ID,
			Title:     r.Title,
			Body:      abstract,
			Owner:     owner,
			CreatedAt: r.DateOfCreat,
			Source:    link,
			Category: "Research",
		})
	}
	return articles, raw.HitCount, nil
}

// ──────────────────────────────────────────────
// Source 2: WHO Disease Outbreak News (RSS)
// ──────────────────────────────────────────────

type whoRSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Items []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
			PubDate     string `xml:"pubDate"`
			Category    string `xml:"category"`
		} `xml:"item"`
	} `xml:"channel"`
}

func (u *newsUseCase) fetchWHO() ([]*domain.NewWithSinglePhoto, error) {
	feeds := []string{
		"https://www.who.int/rss-feeds/news/en/",
		"https://www.who.int/rss-feeds/headlines/en/",
	}

	var allArticles []*domain.NewWithSinglePhoto
	for _, feedURL := range feeds {
		req, err := http.NewRequest("GET", feedURL, nil)
		if err != nil {
			log.Printf("[NEWS] WHO request build error for %s: %v", feedURL, err)
			continue
		}
		req.Header.Set("User-Agent", "SkinAI-Bot/1.0 (health news aggregator)")

		resp, err := u.client.Do(req)
		if err != nil {
			log.Printf("[NEWS] WHO fetch error for %s: %v", feedURL, err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("[NEWS] WHO read body error: %v", err)
			continue
		}

		var rss whoRSS
		if err := xml.Unmarshal(body, &rss); err != nil {
			log.Printf("[NEWS] WHO XML parse error: %v (body: %.200s)", err, string(body))
			continue
		}

		for i, item := range rss.Channel.Items {
			if i >= 10 {
				break
			}
			desc := item.Description
			// strip HTML tags from description
			desc = stripHTMLTags(desc)
			if len(desc) > 500 {
				desc = desc[:497] + "..."
			}
			cat := item.Category
			if cat == "" {
				cat = "WHO News"
			}
			allArticles = append(allArticles, &domain.NewWithSinglePhoto{
				ID:        fmt.Sprintf("who-%d", i),
				Title:     item.Title,
				Body:      desc,
				Owner:     "World Health Organization",
				CreatedAt: item.PubDate,
				Source:    item.Link,
				Category: cat,
			})
		}
	}
	return allArticles, nil
}

// ──────────────────────────────────────────────
// Source 3: MedlinePlus Health Topics (RSS)
// ──────────────────────────────────────────────

func (u *newsUseCase) fetchMedlinePlus() ([]*domain.NewWithSinglePhoto, error) {
	// MedlinePlus skin-related RSS
	feedURL := "https://medlineplus.gov/feeds/topic_685.xml" // Skin Conditions

	req, err := http.NewRequest("GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("medlineplus request build: %w", err)
	}
	req.Header.Set("User-Agent", "SkinAI-Bot/1.0 (health news aggregator)")

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("medlineplus fetch: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("medlineplus read body: %w", err)
	}

	// MedlinePlus uses Atom/RSS format
	type mlpFeed struct {
		XMLName xml.Name `xml:"feed"`
		Entries []struct {
			Title   string `xml:"title"`
			Summary string `xml:"summary"`
			Updated string `xml:"updated"`
			ID      string `xml:"id"`
			Link    struct {
				Href string `xml:"href,attr"`
			} `xml:"link"`
		} `xml:"entry"`
	}

	var feed mlpFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		// Try RSS format
		type rssFeed struct {
			XMLName xml.Name `xml:"rss"`
			Channel struct {
				Items []struct {
					Title       string `xml:"title"`
					Link        string `xml:"link"`
					Description string `xml:"description"`
					PubDate     string `xml:"pubDate"`
				} `xml:"item"`
			} `xml:"channel"`
		}
		var rss rssFeed
		if err2 := xml.Unmarshal(body, &rss); err2 != nil {
			return nil, fmt.Errorf("medlineplus parse error (atom: %v, rss: %v, body: %.200s)", err, err2, string(body))
		}
		var articles []*domain.NewWithSinglePhoto
		for i, item := range rss.Channel.Items {
			if i >= 10 {
				break
			}
			desc := stripHTMLTags(item.Description)
			if len(desc) > 500 {
				desc = desc[:497] + "..."
			}
			articles = append(articles, &domain.NewWithSinglePhoto{
				ID:        fmt.Sprintf("mlp-rss-%d", i),
				Title:     item.Title,
				Body:      desc,
				Owner:     "MedlinePlus",
				CreatedAt: item.PubDate,
				Source:    item.Link,
				Category: "Health Topics",
			})
		}
		return articles, nil
	}

	var articles []*domain.NewWithSinglePhoto
	for i, entry := range feed.Entries {
		if i >= 10 {
			break
		}
		summary := stripHTMLTags(entry.Summary)
		if len(summary) > 500 {
			summary = summary[:497] + "..."
		}
		articles = append(articles, &domain.NewWithSinglePhoto{
			ID:        fmt.Sprintf("mlp-%d", i),
			Title:     entry.Title,
			Body:      summary,
			Owner:     "MedlinePlus",
			CreatedAt: entry.Updated,
			Source:    entry.Link.Href,
			Category: "Health Topics",
		})
	}
	return articles, nil
}

// ──────────────────────────────────────────────
// Main methods
// ──────────────────────────────────────────────

// Search queries rotated across topics
var searchTopics = []string{
	"dermatology skin disease treatment",
	"skin cancer melanoma diagnosis",
	"eczema psoriasis dermatitis therapy",
	"cosmetic dermatology skincare",
	"AI artificial intelligence dermatology",
}

const perPage = 10

func (u *newsUseCase) GetAll(ctx context.Context, query utils.PaginationQuery) (*domain.NewsList, error) {
	page := query.GetPage()
	if page < 1 {
		page = 1
	}

	// Pick a rotating topic based on page
	topic := searchTopics[(page-1)%len(searchTopics)]

	type sourceResult struct {
		articles []*domain.NewWithSinglePhoto
		total    int
		source   string
		err      error
	}

	results := make(chan sourceResult, 3)
	var wg sync.WaitGroup

	// Source 1: Europe PMC
	wg.Add(1)
	go func() {
		defer wg.Done()
		articles, total, err := u.fetchEuropePMC(topic, perPage, (page-1)*perPage)
		if err != nil {
			log.Printf("[NEWS] Europe PMC error: %v", err)
		}
		results <- sourceResult{articles, total, "EuropePMC", err}
	}()

	// Source 2: WHO
	wg.Add(1)
	go func() {
		defer wg.Done()
		articles, err := u.fetchWHO()
		if err != nil {
			log.Printf("[NEWS] WHO error: %v", err)
		}
		results <- sourceResult{articles, len(articles), "WHO", err}
	}()

	// Source 3: MedlinePlus
	wg.Add(1)
	go func() {
		defer wg.Done()
		articles, err := u.fetchMedlinePlus()
		if err != nil {
			log.Printf("[NEWS] MedlinePlus error: %v", err)
		}
		results <- sourceResult{articles, len(articles), "MedlinePlus", err}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var allArticles []*domain.NewWithSinglePhoto
	totalHits := 0
	successSources := 0

	for r := range results {
		if r.err != nil {
			log.Printf("[NEWS] Source %s failed: %v", r.source, r.err)
			continue
		}
		log.Printf("[NEWS] Source %s returned %d articles", r.source, len(r.articles))
		allArticles = append(allArticles, r.articles...)
		totalHits += r.total
		successSources++
	}

	if len(allArticles) == 0 {
		log.Printf("[NEWS] WARNING: All sources returned 0 articles for query: %s", topic)
		return &domain.NewsList{
			TotalCount: 0,
			TotalPages: 0,
			Page:       page,
			Size:       perPage,
			HasMore:    false,
			News:       make([]*domain.NewWithSinglePhoto, 0),
		}, nil
	}

	// Limit to perPage total
	if len(allArticles) > perPage {
		allArticles = allArticles[:perPage]
	}

	totalPages := totalHits / perPage
	if totalHits%perPage > 0 {
		totalPages++
	}
	if totalPages > 100 {
		totalPages = 100
	}

	return &domain.NewsList{
		TotalCount: totalHits,
		TotalPages: totalPages,
		Page:       page,
		Size:       perPage,
		HasMore:    page < totalPages,
		News:       allArticles,
	}, nil
}

func (u *newsUseCase) GetOneById(ctx context.Context, id string) (*domain.NewWithSinglePhoto, error) {
	// Try Europe PMC by ID
	articles, _, err := u.fetchEuropePMC(id, 1, 0)
	if err == nil && len(articles) > 0 {
		return articles[0], nil
	}
	return nil, fmt.Errorf("article %s not found", id)
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

func stripHTMLTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}
