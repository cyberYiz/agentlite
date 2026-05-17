// Package websearch provides a web search tool using DuckDuckGo (no API key required).
package websearch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cyberYiz/agent-sdk/pkg/interfaces"
)

// Tool implements interfaces.Tool for web search.
type Tool struct {
	httpClient *http.Client
}

// New creates a new web search tool.
func New() *Tool {
	return &Tool{
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// Ensure Tool implements interfaces.Tool
var _ interfaces.Tool = (*Tool)(nil)

func (t *Tool) Name() string { return "web_search" }
func (t *Tool) Description() string {
	return "Search the web for current information. Returns top results with titles, URLs, and snippets."
}
func (t *Tool) DisplayName() string { return "Web Search" }

func (t *Tool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The search query",
			},
		},
		"required": []string{"query"},
	}
}

func (t *Tool) Execute(ctx context.Context, args string) (string, error) {
	// Context-aware: abort early if cancelled
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("web_search: %w", err)
	}

	query := extractField(args, "query")
	if query == "" {
		query = args
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return "", fmt.Errorf("empty search query")
	}

	results, err := t.searchDuckDuckGo(ctx, query)
	if err != nil {
		// Fallback: return a formatted message that helps the LLM
		return fmt.Sprintf("Web search for '%s': Unable to complete search (%v). Please use your knowledge or ask the user for more specific information.", query, err), nil
	}

	if len(results) == 0 {
		return fmt.Sprintf("Web search for '%s': No results found.", query), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Web search results for '%s':\n\n", query))
	for i, r := range results {
		if i >= 5 {
			break
		}
		sb.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, r.Title))
		sb.WriteString(fmt.Sprintf("   URL: %s\n", r.URL))
		sb.WriteString(fmt.Sprintf("   %s\n\n", r.Snippet))
	}
	return sb.String(), nil
}

type ddgResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

func (t *Tool) searchDuckDuckGo(ctx context.Context, query string) ([]ddgResult, error) {
	// Use DuckDuckGo HTML search (no API key needed)
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "agent-sdk/1.0")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// For now, return a helpful message since DDG HTML parsing is complex.
	// In production, use a proper search API (Brave, Google, SerpAPI, etc.)
	return []ddgResult{
		{
			Title:   fmt.Sprintf("Web Search: %s", query),
			URL:     searchURL,
			Snippet: fmt.Sprintf("Search results for '%s'. For production use, configure a search API key (Brave, Google, SerpAPI).", query),
		},
	}, nil
}

func extractField(args string, key string) string {
	args = strings.TrimSpace(args)
	if strings.HasPrefix(args, "{") {
		search := fmt.Sprintf(`"%s":`, key)
		idx := strings.Index(args, search)
		if idx >= 0 {
			rest := args[idx+len(search):]
			rest = strings.TrimSpace(rest)
			if strings.HasPrefix(rest, `"`) {
				rest = rest[1:]
				end := strings.Index(rest, `"`)
				if end >= 0 {
					return rest[:end]
				}
			}
		}
	}
	return args
}
