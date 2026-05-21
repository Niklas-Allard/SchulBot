package news

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://www.tagesschau.de/api2u"

// Regions maps German state names (lowercase) to API region IDs.
var Regions = map[string]string{
	"bw": "1", "baden-württemberg": "1", "badenwürttemberg": "1",
	"by": "2", "bayern": "2",
	"be": "3", "berlin": "3",
	"bb": "4", "brandenburg": "4",
	"hb": "5", "bremen": "5",
	"hh": "6", "hamburg": "6",
	"he": "7", "hessen": "7",
	"mv": "8", "mecklenburg": "8",
	"ni": "9", "niedersachsen": "9",
	"nw": "10", "nrw": "10", "nordrhein-westfalen": "10",
	"rp": "11", "rheinland-pfalz": "11",
	"sl": "12", "saarland": "12",
	"sn": "13", "sachsen": "13",
	"st": "14", "sachsen-anhalt": "14",
	"sh": "15", "schleswig-holstein": "15",
	"th": "16", "thüringen": "16",
}

// Ressorts maps user-friendly names to API values.
var Ressorts = map[string]string{
	"inland": "inland", "deutschland": "inland", "de": "inland",
	"ausland": "ausland", "international": "ausland", "welt": "ausland",
	"sport": "sport",
	"wirtschaft": "wirtschaft", "finanzen": "wirtschaft",
	"wissenschaft": "wissenschaft", "tech": "wissenschaft", "technologie": "wissenschaft",
	"gesundheit": "gesundheit",
	"faktenfinder": "faktenfinder", "fakten": "faktenfinder",
}

type newsItem struct {
	Title         string `json:"title"`
	Topline       string `json:"topline"`
	TeaserText    string `json:"teaserText"`
	FirstSentence string `json:"firstSentence"`
	Date          string `json:"date"`
	ShareURL      string `json:"shareURL"`
	Ressort       string `json:"ressort"`
}

type searchItem struct {
	Title      string `json:"title"`
	TeaserText string `json:"teaserText"`
	ShareURL   string `json:"shareURL"`
	Date       string `json:"date"`
}

type apiClient struct {
	client *http.Client
}

func newAPIClient() *apiClient {
	return &apiClient{client: &http.Client{Timeout: 10 * time.Second}}
}

func (c *apiClient) get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	u := baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "SchulBot/1.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, clip(string(body), 150))
	}
	return body, nil
}

func (c *apiClient) homepage(ctx context.Context) ([]newsItem, error) {
	data, err := c.get(ctx, "/homepage/", nil)
	if err != nil {
		return nil, err
	}
	var r struct {
		News         []newsItem `json:"news"`
		BreakingNews []newsItem `json:"breakingNews"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parse homepage: %w", err)
	}
	// Breaking news first if present
	items := append(r.BreakingNews, r.News...)
	return items, nil
}

func (c *apiClient) news(ctx context.Context, ressort, regions string) ([]newsItem, error) {
	params := url.Values{}
	if ressort != "" {
		params.Set("ressort", ressort)
	}
	if regions != "" {
		params.Set("regions", regions)
	}
	data, err := c.get(ctx, "/news/", params)
	if err != nil {
		return nil, err
	}
	var r struct {
		News []newsItem `json:"news"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parse news: %w", err)
	}
	return r.News, nil
}

func (c *apiClient) search(ctx context.Context, query string) ([]searchItem, error) {
	params := url.Values{
		"searchText": {query},
		"pageSize":   {"10"},
		"resultPage": {"1"},
	}
	data, err := c.get(ctx, "/search/", params)
	if err != nil {
		return nil, err
	}
	var r struct {
		SearchResults []searchItem `json:"searchResults"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parse search: %w", err)
	}
	return r.SearchResults, nil
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
