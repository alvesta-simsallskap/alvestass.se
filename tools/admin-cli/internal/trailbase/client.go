package trailbase

import (
	"fmt"

	tb "github.com/trailbaseio/trailbase/client/go/trailbase"
)

// ClubInfo mirrors the club_info table in Trailbase.
type ClubInfo struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Tagline          string `json:"tagline"`
	FoundingYear     int    `json:"founding_year"`
	ShortDescription string `json:"short_description"`
	Address          string `json:"address"`
	City             string `json:"city"`
	PostalCode       string `json:"postal_code"`
	Phone            string `json:"phone"`
	Email            string `json:"email"`
}

// Tokens aliases the SDK type so callers need not import the SDK directly.
type Tokens = tb.Tokens

// Client wraps the official Trailbase SDK client.
type Client struct {
	sdk *tb.Client
}

// NewClient returns a Client for the given base URL (unauthenticated).
func NewClient(baseURL string) (*Client, error) {
	sdk, err := tb.NewClient(baseURL)
	if err != nil {
		return nil, err
	}
	return &Client{sdk: sdk}, nil
}

// NewClientWithTokens restores a session from previously saved tokens.
func NewClientWithTokens(baseURL string, tokens *Tokens) (*Client, error) {
	sdk, err := tb.NewClientWithTokens(baseURL, tokens)
	if err != nil {
		return nil, err
	}
	return &Client{sdk: sdk}, nil
}

// Login authenticates against Trailbase and stores session tokens in the client.
func (c *Client) Login(email, password string) error {
	if _, err := c.sdk.Login(email, password); err != nil {
		return fmt.Errorf("inloggning misslyckades: %w", err)
	}
	return nil
}

// Tokens returns the current session tokens for persistence.
func (c *Client) Tokens() *Tokens {
	return c.sdk.Tokens()
}

// GetClubInfo fetches the single club_info record via the SDK RecordApi.
func (c *Client) GetClubInfo() (*ClubInfo, error) {
	api := tb.NewRecordApi[ClubInfo](c.sdk, "club_info")
	record, err := api.Read(tb.IntRecordId(1))
	if err != nil {
		return nil, fmt.Errorf("kunde inte hämta klubbinfo: %w", err)
	}
	return record, nil
}

// UpdateClubInfo saves the full club_info record back to Trailbase.
func (c *Client) UpdateClubInfo(info ClubInfo) error {
	api := tb.NewRecordApi[ClubInfo](c.sdk, "club_info")
	if err := api.Update(tb.IntRecordId(1), info); err != nil {
		return fmt.Errorf("kunde inte spara klubbinfo: %w", err)
	}
	return nil
}
