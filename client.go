// Package donately provides a client for interacting with the Donately API.
// It implements a complete REST client with support for accounts, people,
// donations, subscriptions, and campaigns management.
package donately

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cenkalti/backoff/v5"
)

// Client defines the interface for interacting with the Donately API.
// It provides methods for managing accounts, people, donations, subscriptions, and campaigns.
type Client interface {
	// FindAccount retrieves a Donately account by its ID.
	FindAccount(context.Context, string) (Account, error)

	// ListPeople retrieves a paginated list of people for the given account.
	// The offset and limit parameters control pagination (0 values disable pagination).
	ListPeople(context.Context, Account, int, int) ([]Person, error)

	// FindPerson retrieves a specific person by ID for the given account.
	FindPerson(context.Context, string, Account) (Person, error)

	// Me retrieves the authenticated user's person record.
	Me(context.Context) (Person, error)

	// SavePerson creates or updates a person record. If the person has no ID, it will be created.
	SavePerson(context.Context, Person) (Person, error)

	// ListDonations retrieves a paginated list of donations for the given account.
	// The offset and limit parameters control pagination (0 values disable pagination).
	ListDonations(context.Context, Account, int, int) ([]Donation, error)

	// ListMyDonations retrieves donations for the authenticated user.
	ListMyDonations(context.Context) ([]Donation, error)

	// FindDonation retrieves a specific donation by ID for the given account.
	FindDonation(context.Context, string, Account) (Donation, error)

	// SaveDonation creates or updates a donation record. If the donation has no ID, it will be created.
	SaveDonation(context.Context, Donation) (Donation, error)

	// RefundDonation processes a refund for the given donation with the specified reason.
	RefundDonation(context.Context, Donation, string) error

	// SendDonationReceipt sends a receipt email for the given donation.
	SendDonationReceipt(context.Context, Donation) error

	// ListSubscriptions retrieves all subscriptions for the given account.
	ListSubscriptions(context.Context, Account) ([]Subscription, error)

	// ListMySubscriptions retrieves subscriptions for the authenticated user.
	ListMySubscriptions(context.Context) ([]Subscription, error)

	// FindSubscription retrieves a specific subscription by ID for the given account.
	FindSubscription(context.Context, string, Account) (Subscription, error)

	// SaveSubscription creates or updates a subscription record. If the subscription has no ID, it will be created.
	SaveSubscription(context.Context, Subscription) (Subscription, error)

	// ListCampaigns retrieves all campaigns for the given account.
	ListCampaigns(context.Context, Account) ([]Campaign, error)

	// FindCampaign retrieves a specific campaign by ID for the given account.
	FindCampaign(context.Context, string, Account) (Campaign, error)

	// SaveCampaign creates or updates a campaign record. If the campaign has no ID, it will be created.
	SaveCampaign(context.Context, Campaign) (Campaign, error)

	// DeleteCampaign deletes the specified campaign.
	DeleteCampaign(context.Context, Campaign) error
}

type clientOption struct {
	apiKey             string
	baseURL            string
	donatelyAPIVersion string
	doRetry            bool
	debug              bool
}

type donatelyClient struct {
	opts   clientOption
	client *http.Client
}

// ClientOption defines a function type for configuring client options.
type ClientOption func(*clientOption)

// APIResponse represents the standard response format from the Donately API.
// It contains the response data and metadata including error information when applicable.
type APIResponse struct {
	Data      json.RawMessage `json:"data"`
	Type      string          `json:"type"`
	Message   string          `json:"message"`
	Code      string          `json:"code"`
	RequestID string          `json:"request_id"`
}

// WithAPIKey returns a ClientOption that sets the API key for authentication.
func WithAPIKey(key string) ClientOption {
	return func(opt *clientOption) {
		opt.apiKey = key
	}
}

// WithBaseURL returns a ClientOption that sets the base URL for the Donately API.
// If not provided, defaults to "https://api.com/v2".
func WithBaseURL(url string) ClientOption {
	return func(opt *clientOption) {
		opt.baseURL = url
	}
}

// WithRetry returns a ClientOption that enables retries (when applicable) for the Donately API.
// If not provided, defaults to false.
func WithRetry() ClientOption {
	return func(opt *clientOption) {
		opt.doRetry = true
	}
}

// NewDonatelyClient creates a new Donately API client with the provided options.
// An API key must be provided using WithAPIKey, otherwise an error is returned.
// The client uses "https://api.com/v2" as the default base URL.
func NewDonatelyClient(options ...ClientOption) (Client, error) {
	clientOptions := clientOption{
		baseURL:            "https://api.donately.com/v2",
		donatelyAPIVersion: "2018-04-01",
	}

	for _, option := range options {
		option(&clientOptions)
	}

	if clientOptions.apiKey == "" {
		return &donatelyClient{}, errors.New("missing API key!")
	}

	if clientOptions.baseURL == "" {
		return &donatelyClient{}, errors.New("missing base URL!")
	}

	return &donatelyClient{
		opts:   clientOptions,
		client: &http.Client{},
	}, nil
}

type retryable interface {
	CanRetry() bool
}

type retryableError struct {
	Err      error
	canRetry bool
}

func (e retryableError) Error() string {
	return e.Err.Error()
}

func (e retryableError) Unwrap() error {
	return e.Err
}

func (e retryableError) CanRetry() bool {
	return e.canRetry
}

func (c *donatelyClient) makeRequest(ctx context.Context, method, endpoint string, body any) (*APIResponse, error) {
	return c.makeRequestWithContentType(ctx, method, endpoint, body, "application/json")
}

func (c *donatelyClient) makeRequestWithContentType(ctx context.Context, method, endpoint string, body any, contentType string) (*APIResponse, error) {
	var reqBody io.Reader
	if body != nil {
		switch contentType {
		case "application/x-www-form-urlencoded":
			if formData, ok := body.(url.Values); ok {
				reqBody = strings.NewReader(formData.Encode())
			} else {
				return nil, fmt.Errorf("body must be url.Values for form-encoded requests")
			}
		default:
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			reqBody = bytes.NewReader(jsonBody)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.opts.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Donately-Version", c.opts.donatelyAPIVersion)
	req.Header.Set("Authorization", "Bearer "+c.opts.apiKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")

	requestLine := fmt.Sprintf("%s %s %s", req.Method, req.URL.RequestURI(), req.Proto)

	if c.opts.debug {
		fmt.Println("Issuing request", requestLine)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		rawBody := string(respBody)

		errorReturned := fmt.Errorf("failed to unmarshal response: %w", err)

		if "retry later" == strings.ToLower(strings.TrimSpace(rawBody)) {
			return nil, retryableError{Err: errorReturned, canRetry: true}
		}

		return nil, errorReturned
	}

	if apiResp.Type != "" && apiResp.Message != "" && apiResp.Code != "" {
		return nil, fmt.Errorf("API error: %s - (%s) %s", apiResp.Code, apiResp.Type, apiResp.Message)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error: %d (Raw Response: %v)", resp.StatusCode, apiResp)
	}

	return &apiResp, nil
}

func (c *donatelyClient) FindAccount(ctx context.Context, id string) (Account, error) {
	endpoint := fmt.Sprintf("/accounts/%s", url.PathEscape(id))

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Account{}, err
	}

	var account Account
	if err := json.Unmarshal(resp.Data, &account); err != nil {
		return Account{}, fmt.Errorf("failed to unmarshal account: %w", err)
	}

	return account, nil
}

func (c *donatelyClient) ListPeople(ctx context.Context, account Account, offset, limit int) ([]Person, error) {
	params := url.Values{}
	params.Set("account_id", account.ID)

	if offset > 0 {
		params.Set("offset", strconv.Itoa(offset))
	}

	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	resp, err := c.makeRequest(ctx, http.MethodGet, "/people?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	var people []Person
	if err := json.Unmarshal(resp.Data, &people); err != nil {
		return nil, fmt.Errorf("failed to unmarshal people: %w", err)
	}

	return people, nil
}

func (c *donatelyClient) FindPerson(ctx context.Context, id string, account Account) (Person, error) {
	endpoint := fmt.Sprintf("/people/%s", url.PathEscape(id))

	params := url.Values{}
	params.Set("account_id", account.ID)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return Person{}, err
	}

	var person Person
	if err := json.Unmarshal(resp.Data, &person); err != nil {
		return Person{}, fmt.Errorf("failed to unmarshal person: %w", err)
	}

	return person, nil
}

func (c *donatelyClient) Me(ctx context.Context) (Person, error) {
	resp, err := c.makeRequest(ctx, http.MethodGet, "/me", nil)
	if err != nil {
		return Person{}, err
	}

	var person Person
	if err := json.Unmarshal(resp.Data, &person); err != nil {
		return Person{}, fmt.Errorf("failed to unmarshal person: %w", err)
	}

	return person, nil
}

func (c *donatelyClient) SavePerson(ctx context.Context, person Person) (Person, error) {
	var endpoint string

	if person.ID == "" {
		endpoint = "/people"
	} else {
		endpoint = fmt.Sprintf("/people/%s", url.PathEscape(person.ID))
	}

	if len(person.Accounts) == 0 || person.Accounts[0].ID == "" {
		return Person{}, errors.New("missing account information")
	}

	accountId := person.Accounts[0].ID

	formData := url.Values{}

	formData.Set("account_id", accountId)

	if person.FirstName != "" {
		formData.Set("first_name", person.FirstName)
	}
	if person.LastName != "" {
		formData.Set("last_name", person.LastName)
	}
	if person.Email != "" {
		formData.Set("email", person.Email)
	}
	if person.PhoneNumber != "" {
		formData.Set("phone_number", person.PhoneNumber)
	}
	if person.StreetAddress != "" {
		formData.Set("street_address", person.StreetAddress)
	}
	if person.StreetAddress2 != "" {
		formData.Set("street_address_2", person.StreetAddress2)
	}
	if person.City != "" {
		formData.Set("city", person.City)
	}
	if person.State != "" {
		formData.Set("state", person.State)
	}
	if person.ZipCode != "" {
		formData.Set("zip_code", person.ZipCode)
	}
	if person.Country != "" {
		formData.Set("country", person.Country)
	}

	resp, err := c.makeRequestWithContentType(ctx, http.MethodPost, endpoint, formData, "application/x-www-form-urlencoded")
	if err != nil {
		return Person{}, err
	}

	var savedPerson Person
	if err := json.Unmarshal(resp.Data, &savedPerson); err != nil {
		return Person{}, fmt.Errorf("failed to unmarshal saved person: %w", err)
	}

	return savedPerson, nil
}

func (c *donatelyClient) ListDonations(ctx context.Context, account Account, offset, limit int) ([]Donation, error) {
	params := url.Values{}
	params.Set("account_id", account.ID)

	if offset > 0 {
		params.Set("offset", strconv.Itoa(offset))
	}

	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	resp, err := c.makeRequest(ctx, http.MethodGet, "/donations?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	var donations []Donation
	if err := json.Unmarshal(resp.Data, &donations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal donations: %w", err)
	}

	return donations, nil
}

func (c *donatelyClient) ListMyDonations(ctx context.Context) ([]Donation, error) {
	resp, err := c.makeRequest(ctx, http.MethodGet, "/me/donations", nil)
	if err != nil {
		return nil, err
	}

	var donations []Donation
	if err := json.Unmarshal(resp.Data, &donations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal my donations: %w", err)
	}

	return donations, nil
}

func (c *donatelyClient) FindDonation(ctx context.Context, id string, account Account) (Donation, error) {
	params := url.Values{}
	params.Set("account_id", account.ID)

	endpoint := fmt.Sprintf("/donations/%s", url.PathEscape(id))
	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return Donation{}, err
	}

	var donation Donation
	if err := json.Unmarshal(resp.Data, &donation); err != nil {
		return Donation{}, fmt.Errorf("failed to unmarshal donation: %w", err)
	}

	return donation, nil
}

func (c *donatelyClient) SaveDonation(ctx context.Context, donation Donation) (Donation, error) {
	var endpoint string

	if donation.ID == "" {
		endpoint = "/donations"
	} else {
		endpoint = fmt.Sprintf("/donations/%s", url.PathEscape(donation.ID))
	}

	if donation.Account.ID == "" {
		return Donation{}, errors.New("missing account information")
	}

	params := url.Values{}
	params.Set("account_id", donation.Account.ID)

	if donation.AmountInCents > 0 {
		params.Set("amount_in_cents", fmt.Sprintf("%d", donation.AmountInCents))
	}
	if donation.DonationType != "" {
		params.Set("donation_type", donation.DonationType)
	}
	if donation.Campaign.ID != "" {
		params.Set("campaign_id", donation.Campaign.ID)
	}
	if donation.Person.Email != "" {
		params.Set("email", donation.Person.Email)
	}
	if donation.Comment != "" {
		params.Set("comment", donation.Comment)
	}
	if donation.Anonymous {
		params.Set("anonymous", "true")
	}
	if donation.OnBehalfOf != "" {
		params.Set("on_behalf_of", donation.OnBehalfOf)
	}
	if donation.Status != "" {
		params.Set("status", donation.Status)
	}

	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	resp, err := c.makeRequest(ctx, http.MethodPost, endpoint, nil)

	if c.opts.doRetry {
		re, ok := err.(retryable)
		if ok && re.CanRetry() {
			operation := func() (*APIResponse, error) {
				return c.makeRequest(ctx, http.MethodPost, endpoint, nil)
			}
			resp, err = backoff.Retry(ctx, operation, backoff.WithBackOff(backoff.NewExponentialBackOff()))
			if err != nil {
				return Donation{}, err

			}
		} else {
			return Donation{}, err
		}
	}

	var savedDonation Donation
	if err := json.Unmarshal(resp.Data, &savedDonation); err != nil {
		return Donation{}, fmt.Errorf("failed to unmarshal saved donation: %w", err)
	}

	return savedDonation, nil
}

func (c *donatelyClient) RefundDonation(ctx context.Context, donation Donation, reason string) error {
	endpoint := fmt.Sprintf("/donations/%s/refund", url.PathEscape(donation.ID))

	if donation.Account.ID == "" {
		return errors.New("missing account information")
	}

	formData := url.Values{}
	formData.Set("account_id", donation.Account.ID)
	formData.Set("refund_reason", reason)

	_, err := c.makeRequest(ctx, http.MethodPost, endpoint, formData)
	return err
}

func (c *donatelyClient) SendDonationReceipt(ctx context.Context, donation Donation) error {
	endpoint := fmt.Sprintf("/donations/%s/receipt", url.PathEscape(donation.ID))
	_, err := c.makeRequest(ctx, http.MethodPost, endpoint, nil)
	return err
}

// Subscriptions operations
func (c *donatelyClient) ListSubscriptions(ctx context.Context, account Account) ([]Subscription, error) {
	params := url.Values{}
	params.Set("account_id", account.ID)

	resp, err := c.makeRequest(ctx, http.MethodGet, "/subscriptions?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	var subscriptions []Subscription
	if err := json.Unmarshal(resp.Data, &subscriptions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subscriptions: %w", err)
	}

	return subscriptions, nil
}

func (c *donatelyClient) ListMySubscriptions(ctx context.Context) ([]Subscription, error) {
	resp, err := c.makeRequest(ctx, http.MethodGet, "/me/subscriptions", nil)
	if err != nil {
		return nil, err
	}

	var subscriptions []Subscription
	if err := json.Unmarshal(resp.Data, &subscriptions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal my subscriptions: %w", err)
	}

	return subscriptions, nil
}

func (c *donatelyClient) FindSubscription(ctx context.Context, id string, account Account) (Subscription, error) {
	endpoint := fmt.Sprintf("/subscriptions/%s", url.PathEscape(id))

	params := url.Values{}
	params.Set("account_id", account.ID)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return Subscription{}, err
	}

	var subscription Subscription
	if err := json.Unmarshal(resp.Data, &subscription); err != nil {
		return Subscription{}, fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	return subscription, nil
}

func (c *donatelyClient) SaveSubscription(ctx context.Context, subscription Subscription) (Subscription, error) {
	var endpoint string

	if subscription.ID == "" {
		endpoint = "/subscriptions"
	} else {
		endpoint = fmt.Sprintf("/subscriptions/%s", url.PathEscape(subscription.ID))
	}

	resp, err := c.makeRequest(ctx, http.MethodPost, endpoint, subscription)
	if err != nil {
		return Subscription{}, err
	}

	var savedSubscription Subscription
	if err := json.Unmarshal(resp.Data, &savedSubscription); err != nil {
		return Subscription{}, fmt.Errorf("failed to unmarshal saved subscription: %w", err)
	}

	return savedSubscription, nil
}

func (c *donatelyClient) ListCampaigns(ctx context.Context, account Account) ([]Campaign, error) {
	params := url.Values{}
	params.Set("account_id", account.ID)

	resp, err := c.makeRequest(ctx, http.MethodGet, "/campaigns?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	var campaigns []Campaign
	if err := json.Unmarshal(resp.Data, &campaigns); err != nil {
		return nil, fmt.Errorf("failed to unmarshal campaigns: %w", err)
	}

	return campaigns, nil
}

func (c *donatelyClient) FindCampaign(ctx context.Context, id string, account Account) (Campaign, error) {
	endpoint := fmt.Sprintf("/campaigns/%s", url.PathEscape(id))

	params := url.Values{}
	params.Add("account_id", account.ID)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return Campaign{}, err
	}

	var campaign Campaign
	if err := json.Unmarshal(resp.Data, &campaign); err != nil {
		return Campaign{}, fmt.Errorf("failed to unmarshal campaign: %w", err)
	}

	return campaign, nil
}

func (c *donatelyClient) SaveCampaign(ctx context.Context, campaign Campaign) (Campaign, error) {
	var endpoint string

	if campaign.ID == "" {
		endpoint = "/campaigns"
	} else {
		endpoint = fmt.Sprintf("/campaigns/%s", url.PathEscape(campaign.ID))
	}

	resp, err := c.makeRequest(ctx, http.MethodPost, endpoint, campaign)
	if err != nil {
		return Campaign{}, err
	}

	var savedCampaign Campaign
	if err := json.Unmarshal(resp.Data, &savedCampaign); err != nil {
		return Campaign{}, fmt.Errorf("failed to unmarshal saved campaign: %w", err)
	}

	return savedCampaign, nil
}

func (c *donatelyClient) DeleteCampaign(ctx context.Context, campaign Campaign) error {
	endpoint := fmt.Sprintf("/campaigns/%s", url.PathEscape(campaign.ID))
	_, err := c.makeRequest(ctx, http.MethodDelete, endpoint, nil)
	return err
}
