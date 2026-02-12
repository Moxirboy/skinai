package usecase

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
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
			Timeout: 15 * time.Second,
		},
	}
}

// PubMed E-utilities response types

type esearchResult struct {
	ESearchResult struct {
		Count  string   `json:"count"`
		IDList []string `json:"idlist"`
	} `json:"esearchresult"`
}

type pubmedArticleSet struct {
	XMLName  xml.Name        `xml:"PubmedArticleSet"`
	Articles []pubmedArticle `xml:"PubmedArticle"`
}

type pubmedArticle struct {
	Citation struct {
		PMID struct {
			Value string `xml:",chardata"`
		} `xml:"PMID"`
		Article struct {
			Title    string `xml:"ArticleTitle"`
			Abstract struct {
				Texts []string `xml:"AbstractText"`
			} `xml:"Abstract"`
			Journal struct {
				Title string `xml:"Title"`
				Date  struct {
					Year  string `xml:"Year"`
					Month string `xml:"Month"`
					Day   string `xml:"Day"`
				} `xml:"JournalIssue>PubDate"`
			} `xml:"Journal"`
			AuthorList struct {
				Authors []struct {
					LastName string `xml:"LastName"`
					ForeName string `xml:"ForeName"`
				} `xml:"Author"`
			} `xml:"AuthorList"`
		} `xml:"Article"`
	} `xml:"MedlineCitation"`
}

const (
	pubmedSearchURL  = "https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esearch.fcgi"
	pubmedFetchURL   = "https://eutils.ncbi.nlm.nih.gov/entrez/eutils/efetch.fcgi"
	pubmedArticleURL = "https://pubmed.ncbi.nlm.nih.gov/"
	pageSize         = 10
	searchTerm       = "dermatology OR skin disease OR skincare OR dermatitis"
)

func (u *newsUseCase) GetAll(ctx context.Context, query utils.PaginationQuery) (*domain.NewsList, error) {
	page := query.GetPage()
	if page < 1 {
		page = 1
	}
	retStart := (page - 1) * pageSize

	searchURL := fmt.Sprintf(
		"%s?db=pubmed&term=%s&retmode=json&retmax=%d&retstart=%d&sort=date",
		pubmedSearchURL, strings.ReplaceAll(searchTerm, " ", "+"),
		pageSize, retStart,
	)

	resp, err := u.client.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("pubmed search failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading search response: %w", err)
	}

	var search esearchResult
	if err := json.Unmarshal(body, &search); err != nil {
		return nil, fmt.Errorf("parsing search response: %w", err)
	}

	ids := search.ESearchResult.IDList
	if len(ids) == 0 {
		return &domain.NewsList{
			TotalCount: 0,
			TotalPages: 0,
			Page:       page,
			Size:       pageSize,
			HasMore:    false,
			News:       make([]*domain.NewWithSinglePhoto, 0),
		}, nil
	}

	totalCount := 0
	fmt.Sscanf(search.ESearchResult.Count, "%d", &totalCount)

	articles, err := u.fetchArticles(ids)
	if err != nil {
		return nil, err
	}

	totalPages := totalCount / pageSize
	if totalCount%pageSize > 0 {
		totalPages++
	}
	if totalPages > 100 {
		totalPages = 100
	}

	return &domain.NewsList{
		TotalCount: totalCount,
		TotalPages: totalPages,
		Page:       page,
		Size:       pageSize,
		HasMore:    page < totalPages,
		News:       articles,
	}, nil
}

func (u *newsUseCase) GetOneById(ctx context.Context, id string) (*domain.NewWithSinglePhoto, error) {
	articles, err := u.fetchArticles([]string{id})
	if err != nil {
		return nil, err
	}
	if len(articles) == 0 {
		return nil, nil
	}
	return articles[0], nil
}

func (u *newsUseCase) fetchArticles(ids []string) ([]*domain.NewWithSinglePhoto, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	fetchURL := fmt.Sprintf(
		"%s?db=pubmed&id=%s&retmode=xml",
		pubmedFetchURL, strings.Join(ids, ","),
	)

	resp, err := u.client.Get(fetchURL)
	if err != nil {
		return nil, fmt.Errorf("pubmed fetch failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading fetch response: %w", err)
	}

	var articleSet pubmedArticleSet
	if err := xml.Unmarshal(body, &articleSet); err != nil {
		return nil, fmt.Errorf("parsing article XML: %w", err)
	}

	results := make([]*domain.NewWithSinglePhoto, 0, len(articleSet.Articles))
	for _, a := range articleSet.Articles {
		pmid := a.Citation.PMID.Value
		article := a.Citation.Article

		abstractText := strings.Join(article.Abstract.Texts, " ")
		if len(abstractText) > 500 {
			abstractText = abstractText[:497] + "..."
		}

		var authors []string
		for _, auth := range article.AuthorList.Authors {
			if auth.LastName != "" {
				authors = append(authors, auth.ForeName+" "+auth.LastName)
			}
		}
		owner := strings.Join(authors, ", ")
		if len(owner) > 200 {
			owner = owner[:197] + "..."
		}
		if owner == "" {
			owner = article.Journal.Title
		}

		pubDate := article.Journal.Date
		dateStr := pubDate.Year
		if pubDate.Month != "" {
			dateStr = pubDate.Month + " " + dateStr
		}
		if pubDate.Day != "" {
			dateStr = pubDate.Day + " " + dateStr
		}

		results = append(results, &domain.NewWithSinglePhoto{
			ID:        pmid,
			Title:     article.Title,
			Body:      abstractText,
			Owner:     owner,
			Photo:     "",
			CreatedAt: dateStr,
			Source:    pubmedArticleURL + pmid,
		})
	}

	return results, nil
}
