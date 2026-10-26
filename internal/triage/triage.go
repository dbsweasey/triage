package triage

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"triage/internal/store"
)

//go:embed schema.json
var outputSchema json.RawMessage

type requestBody struct {
	Model        string     `json:"model"`
	Instructions string     `json:"instructions"`
	Input        string     `json:"input"`
	Text         textConfig `json:"text"`
}

type textConfig struct {
	Format formatConfig `json:"format"`
}

type formatConfig struct {
	Type   string          `json:"type"`
	Name   string          `json:"name"`
	Strict bool            `json:"strict"`
	Schema json.RawMessage `json:"schema"`
}

type responseBody struct {
	Status string           `json:"status"`
	Output []responseOutput `json:"output"`
}

type responseOutput struct {
	Type    string          `json:"type"`
	Content []contentOutput `json:"content"`
}

type contentOutput struct {
	Type    string `json:"type"`
	Text    string `json:"text"`
	Refusal string `json:"refusal"`
}

type Result struct {
	Priority   string   `json:"priority"`
	Category   string   `json:"category"`
	Summary    string   `json:"summary"`
	Steps      []string `json:"steps"`
	ArticleIDs []int64  `json:"articleIds"`
}

type OpenAIClient struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

type Triager interface {
	GetResult(ctx context.Context, t store.Ticket, as []store.Article) (Result, error)
}

func buildPrompt(t store.Ticket, as []store.Article) (string, string) {
	systemPrompt := `
# Role
You are an ITSM triage agent. Given details from a support ticket and knowledge
base articles, you must determine the incident priority, its category,
a summary of how to resolve, and specific resolution steps.

# Input
You will receive a ticket, which contains a title and description. This
is how you determine the issue to be resolved and its priority.
You will also receive a list of articles. These have already been retrieved.
Use these as reference to create resolution steps and summary.

# Goal
Develop an output that will allow the ticket to be routed to the correct
department for resolution with a strong resolution summary, clear steps,
an accurate priority.
	`
	userPrompt := fmt.Sprintf(
		`Help me resolve the following ticket.
Title: %v
Description: %v`, t.Title, t.Description)
	if len(as) > 0 {
		// Create string of articles for user prompt
		var builder strings.Builder
		for _, a := range as {
			fmt.Fprintf(&builder, "\nArticle #%d - %v\n\t%v\n\tTags: %v\n", a.ID, a.Title, a.Body, a.Tags)
		}
		userPrompt += fmt.Sprintf(`

Here are the relevant knowledge articles:
%v`, builder.String())
	}

	return systemPrompt, userPrompt
}

func NewOpenAIClient(key string, model string) *OpenAIClient {
	client := http.Client{Timeout: 30 * time.Second}
	return &OpenAIClient{apiKey: key, model: model, baseURL: "https://api.openai.com/v1", client: &client}
}

func (c *OpenAIClient) GetResult(ctx context.Context, t store.Ticket, as []store.Article) (Result, error) {

	sys, user := buildPrompt(t, as)

	b := requestBody{Model: c.model, Instructions: sys, Input: user, Text: textConfig{formatConfig{"json_schema", "triage_result", true, outputSchema}}}
	body, err := json.Marshal(b)
	if err != nil {
		return Result{}, fmt.Errorf("parse body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/responses", bytes.NewReader(body))
	if err != nil {
		return Result{}, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("%w: do request: %w", ErrUpstream, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errMsg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return Result{}, fmt.Errorf("%w: response: %d - %s", ErrUpstream, resp.StatusCode, errMsg)
	}

	var response responseBody
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return Result{}, fmt.Errorf("%w: decode: %v", ErrBadResponse, err)
	}
	var result Result
	if response.Status != "completed" {
		return Result{}, fmt.Errorf("%w: status %q", ErrIncomplete, response.Status)
	}
	for _, o := range response.Output {
		if o.Type == "message" {
			for _, content := range o.Content {
				if content.Refusal != "" {
					return Result{}, fmt.Errorf("%w: %s", ErrRefused, content.Refusal)
				}
				if content.Type == "output_text" {
					if err := json.Unmarshal([]byte(content.Text), &result); err != nil {
						return Result{}, fmt.Errorf("%w: unmarshal: %w", ErrBadResponse, err)
					}
					return result, nil
				}
			}

		}
	}

	return Result{}, fmt.Errorf("%w: no result found", ErrBadResponse)
}

func ValidateResult(r Result) bool {
	// TOOD: implement
	// check if priority and category match enum
	// ensure no field is empty, steps must have at least 1 in the least, etc.
	return false
}

var (
	ErrRefused     = errors.New("model refused")
	ErrIncomplete  = errors.New("incomplete response")
	ErrBadResponse = errors.New("bad response")
	ErrUpstream    = errors.New("upstream error")
)
