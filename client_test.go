package donately

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDonatelyClient(t *testing.T) {
	tests := []struct {
		name          string
		options       []ClientOption
		expectedError bool
		errorMessage  string
	}{
		{
			name:          "missing API key",
			options:       []ClientOption{},
			expectedError: true,
			errorMessage:  "missing API key!",
		},
		{
			name: "valid API key",
			options: []ClientOption{
				WithAPIKey("test-api-key"),
			},
			expectedError: false,
		},
		{
			name: "valid API key with custom base URL",
			options: []ClientOption{
				WithAPIKey("test-api-key"),
				WithBaseURL("https://custom.api.com/v1"),
			},
			expectedError: false,
		},
		{
			name: "valid API key with retry enabled",
			options: []ClientOption{
				WithAPIKey("test-api-key"),
				WithRetry(),
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewDonatelyClient(tt.options...)

			if tt.expectedError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestClientOptions(t *testing.T) {
	t.Run("WithAPIKey sets API key", func(t *testing.T) {
		opts := clientOption{}
		WithAPIKey("test-key")(&opts)
		assert.Equal(t, "test-key", opts.apiKey)
	})

	t.Run("WithBaseURL sets base URL", func(t *testing.T) {
		opts := clientOption{}
		WithBaseURL("https://test.com")(&opts)
		assert.Equal(t, "https://test.com", opts.baseURL)
	})

	t.Run("WithRetry enables retry", func(t *testing.T) {
		opts := clientOption{}
		WithRetry()(&opts)
		assert.True(t, opts.doRetry)
	})
}

func setupTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, Client) {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := NewDonatelyClient(
		WithAPIKey("test-api-key"),
		WithBaseURL(server.URL),
	)
	require.NoError(t, err)

	return server, client
}

func TestFindAccount(t *testing.T) {
	expectedAccount := Account{
		ID:    "acc_123",
		Title: "Test Account",
	}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.True(t, strings.HasSuffix(r.URL.Path, "/accounts/acc_123"))

		resp := APIResponse{
			Data: mustMarshal(t, expectedAccount),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	account, err := client.FindAccount(context.Background(), "acc_123")
	require.NoError(t, err)

	assert.Equal(t, expectedAccount.ID, account.ID)
	assert.Equal(t, expectedAccount.Title, account.Title)
}

func TestListPeople(t *testing.T) {
	expectedPeople := []Person{
		{ID: "person_1", Email: "test1@example.com"},
		{ID: "person_2", Email: "test2@example.com"},
	}
	account := Account{ID: "acc_123"}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.True(t, strings.HasPrefix(r.URL.Path, "/people"))

		// Check query parameters
		params := r.URL.Query()
		assert.Equal(t, "acc_123", params.Get("account_id"))
		assert.Equal(t, "10", params.Get("offset"))
		assert.Equal(t, "20", params.Get("limit"))

		resp := APIResponse{
			Data: mustMarshal(t, expectedPeople),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	people, err := client.ListPeople(context.Background(), account, 10, 20)
	require.NoError(t, err)

	assert.Len(t, people, len(expectedPeople))
}

func TestFindPerson(t *testing.T) {
	expectedPerson := Person{
		ID:    "person_123",
		Email: "test@example.com",
	}
	account := Account{ID: "acc_123"}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.True(t, strings.Contains(r.URL.Path, "/people/person_123"))

		params := r.URL.Query()
		assert.Equal(t, "acc_123", params.Get("account_id"))

		resp := APIResponse{
			Data: mustMarshal(t, expectedPerson),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	person, err := client.FindPerson(context.Background(), "person_123", account)
	require.NoError(t, err)

	assert.Equal(t, expectedPerson.ID, person.ID)
}

func TestMe(t *testing.T) {
	expectedPerson := Person{
		ID:    "person_me",
		Email: "me@example.com",
	}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/me", r.URL.Path)

		resp := APIResponse{
			Data: mustMarshal(t, expectedPerson),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	person, err := client.Me(context.Background())
	require.NoError(t, err)

	assert.Equal(t, expectedPerson.ID, person.ID)
}

func TestSavePerson(t *testing.T) {
	inputPerson := Person{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Accounts:  []Account{{ID: "acc_123"}},
	}

	expectedPerson := Person{
		ID:        "person_new",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/people", r.URL.Path)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		// Parse form data
		err := r.ParseForm()
		require.NoError(t, err)

		assert.Equal(t, "acc_123", r.Form.Get("account_id"))
		assert.Equal(t, "John", r.Form.Get("first_name"))

		resp := APIResponse{
			Data: mustMarshal(t, expectedPerson),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	person, err := client.SavePerson(context.Background(), inputPerson)
	require.NoError(t, err)

	assert.Equal(t, expectedPerson.ID, person.ID)
}

func TestSavePersonMissingAccount(t *testing.T) {
	inputPerson := Person{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not make request when account is missing")
	})
	defer server.Close()

	_, err := client.SavePerson(context.Background(), inputPerson)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing account information")
}

func TestListDonations(t *testing.T) {
	expectedDonations := []Donation{
		{ID: "don_1", AmountInCents: 1000},
		{ID: "don_2", AmountInCents: 2000},
	}
	account := Account{ID: "acc_123"}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.True(t, strings.HasPrefix(r.URL.Path, "/donations"))

		params := r.URL.Query()
		assert.Equal(t, "acc_123", params.Get("account_id"))

		resp := APIResponse{
			Data: mustMarshal(t, expectedDonations),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	donations, err := client.ListDonations(context.Background(), account, 0, 0)
	require.NoError(t, err)

	assert.Len(t, donations, len(expectedDonations))
}

func TestListMyDonations(t *testing.T) {
	expectedDonations := []Donation{
		{ID: "don_1", AmountInCents: 1000},
	}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/me/donations", r.URL.Path)

		resp := APIResponse{
			Data: mustMarshal(t, expectedDonations),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	donations, err := client.ListMyDonations(context.Background())
	require.NoError(t, err)

	assert.Len(t, donations, len(expectedDonations))
}

func TestFindDonation(t *testing.T) {
	expectedDonation := Donation{
		ID:            "don_123",
		AmountInCents: 5000,
	}
	account := Account{ID: "acc_123"}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.True(t, strings.Contains(r.URL.Path, "/donations/don_123"))

		params := r.URL.Query()
		assert.Equal(t, "acc_123", params.Get("account_id"))

		resp := APIResponse{
			Data: mustMarshal(t, expectedDonation),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	donation, err := client.FindDonation(context.Background(), "don_123", account)
	require.NoError(t, err)

	assert.Equal(t, expectedDonation.ID, donation.ID)
}

func TestSaveDonation(t *testing.T) {
	inputDonation := Donation{
		AmountInCents: 2500,
		DonationType:  "one_time",
		Account:       Account{ID: "acc_123"},
		Person:        Person{Email: "donor@example.com"},
		Comment:       "Test donation",
		Anonymous:     true,
	}

	expectedDonation := Donation{
		ID:            "don_new",
		AmountInCents: 2500,
		DonationType:  "one_time",
	}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.True(t, strings.HasPrefix(r.URL.Path, "/donations"))

		// Parse query parameters
		params := r.URL.Query()
		assert.Equal(t, "acc_123", params.Get("account_id"))
		assert.Equal(t, "2500", params.Get("amount_in_cents"))
		assert.Equal(t, "true", params.Get("anonymous"))

		resp := APIResponse{
			Data: mustMarshal(t, expectedDonation),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	donation, err := client.SaveDonation(context.Background(), inputDonation)
	require.NoError(t, err)

	assert.Equal(t, expectedDonation.ID, donation.ID)
}

func TestRefundDonation(t *testing.T) {
	inputDonation := Donation{
		ID:      "don_123",
		Account: Account{ID: "acc_123"},
	}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.True(t, strings.Contains(r.URL.Path, "/donations/don_123/refund"))

		resp := APIResponse{
			Data: json.RawMessage(`{}`),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	err := client.RefundDonation(context.Background(), inputDonation, "Customer request")
	require.NoError(t, err)
}

func TestSendDonationReceipt(t *testing.T) {
	inputDonation := Donation{
		ID: "don_123",
	}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.True(t, strings.Contains(r.URL.Path, "/donations/don_123/receipt"))

		resp := APIResponse{
			Data: json.RawMessage(`{}`),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	err := client.SendDonationReceipt(context.Background(), inputDonation)
	require.NoError(t, err)
}

func TestListSubscriptions(t *testing.T) {
	expectedSubscriptions := []Subscription{
		{ID: "sub_1", AmountInCents: 1000},
		{ID: "sub_2", AmountInCents: 2000},
	}
	account := Account{ID: "acc_123"}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.True(t, strings.HasPrefix(r.URL.Path, "/subscriptions"))

		params := r.URL.Query()
		assert.Equal(t, "acc_123", params.Get("account_id"))

		resp := APIResponse{
			Data: mustMarshal(t, expectedSubscriptions),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	subscriptions, err := client.ListSubscriptions(context.Background(), account)
	require.NoError(t, err)

	assert.Len(t, subscriptions, len(expectedSubscriptions))
}

func TestListMySubscriptions(t *testing.T) {
	expectedSubscriptions := []Subscription{
		{ID: "sub_1", AmountInCents: 1000},
	}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/me/subscriptions", r.URL.Path)

		resp := APIResponse{
			Data: mustMarshal(t, expectedSubscriptions),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	subscriptions, err := client.ListMySubscriptions(context.Background())
	require.NoError(t, err)

	assert.Len(t, subscriptions, len(expectedSubscriptions))
}

func TestFindSubscription(t *testing.T) {
	expectedSubscription := Subscription{
		ID:            "sub_123",
		AmountInCents: 3000,
	}
	account := Account{ID: "acc_123"}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.True(t, strings.Contains(r.URL.Path, "/subscriptions/sub_123"))

		params := r.URL.Query()
		assert.Equal(t, "acc_123", params.Get("account_id"))

		resp := APIResponse{
			Data: mustMarshal(t, expectedSubscription),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	subscription, err := client.FindSubscription(context.Background(), "sub_123", account)
	require.NoError(t, err)

	assert.Equal(t, expectedSubscription.ID, subscription.ID)
}

func TestSaveSubscription(t *testing.T) {
	inputSubscription := Subscription{
		AmountInCents:      1500,
		RecurringFrequency: "monthly",
	}

	expectedSubscription := Subscription{
		ID:                 "sub_new",
		AmountInCents:      1500,
		RecurringFrequency: "monthly",
	}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/subscriptions", r.URL.Path)

		resp := APIResponse{
			Data: mustMarshal(t, expectedSubscription),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	subscription, err := client.SaveSubscription(context.Background(), inputSubscription)
	require.NoError(t, err)

	assert.Equal(t, expectedSubscription.ID, subscription.ID)
}

func TestListCampaigns(t *testing.T) {
	expectedCampaigns := []Campaign{
		{ID: "camp_1", Title: "Campaign 1"},
		{ID: "camp_2", Title: "Campaign 2"},
	}
	account := Account{ID: "acc_123"}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.True(t, strings.HasPrefix(r.URL.Path, "/campaigns"))

		params := r.URL.Query()
		assert.Equal(t, "acc_123", params.Get("account_id"))

		resp := APIResponse{
			Data: mustMarshal(t, expectedCampaigns),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	campaigns, err := client.ListCampaigns(context.Background(), account)
	require.NoError(t, err)

	assert.Len(t, campaigns, len(expectedCampaigns))
}

func TestFindCampaign(t *testing.T) {
	expectedCampaign := Campaign{
		ID:    "camp_123",
		Title: "Test Campaign",
	}
	account := Account{ID: "acc_123"}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.True(t, strings.Contains(r.URL.Path, "/campaigns/camp_123"))

		params := r.URL.Query()
		assert.Equal(t, "acc_123", params.Get("account_id"))

		resp := APIResponse{
			Data: mustMarshal(t, expectedCampaign),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	campaign, err := client.FindCampaign(context.Background(), "camp_123", account)
	require.NoError(t, err)

	assert.Equal(t, expectedCampaign.ID, campaign.ID)
}

func TestSaveCampaign(t *testing.T) {
	inputCampaign := Campaign{
		Title:       "New Campaign",
		Description: stringPtr("A test campaign"),
	}

	expectedCampaign := Campaign{
		ID:          "camp_new",
		Title:       "New Campaign",
		Description: stringPtr("A test campaign"),
	}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/campaigns", r.URL.Path)

		resp := APIResponse{
			Data: mustMarshal(t, expectedCampaign),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	campaign, err := client.SaveCampaign(context.Background(), inputCampaign)
	require.NoError(t, err)

	assert.Equal(t, expectedCampaign.ID, campaign.ID)
}

func TestDeleteCampaign(t *testing.T) {
	inputCampaign := Campaign{
		ID: "camp_123",
	}

	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.True(t, strings.Contains(r.URL.Path, "/campaigns/camp_123"))

		resp := APIResponse{
			Data: json.RawMessage(`{}`),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	err := client.DeleteCampaign(context.Background(), inputCampaign)
	require.NoError(t, err)
}

func TestAPIErrorHandling(t *testing.T) {
	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{
			Type:    "error",
			Message: "Invalid API key",
			Code:    "invalid_api_key",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.FindAccount(context.Background(), "acc_123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid_api_key")
}

func TestHTTPErrorHandling(t *testing.T) {
	server, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		resp := APIResponse{
			Data: json.RawMessage(`{}`),
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.FindAccount(context.Background(), "acc_123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestRetryOnRetryableError(t *testing.T) {
	attempts := 0
	server, _ := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			// First attempt returns "retry later"
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("retry later"))
			return
		}
		// Second attempt succeeds
		resp := APIResponse{
			Data: mustMarshal(t, Donation{ID: "don_123"}),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	// Create client with retry enabled
	client, err := NewDonatelyClient(
		WithAPIKey("test-api-key"),
		WithBaseURL(server.URL),
		WithRetry(),
	)
	require.NoError(t, err)

	donation := Donation{
		Account:       Account{ID: "acc_123"},
		AmountInCents: 1000,
	}

	_, err = client.SaveDonation(context.Background(), donation)
	require.NoError(t, err)

	assert.Equal(t, 2, attempts)
}

func mustMarshal(t *testing.T, v any) json.RawMessage {
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return json.RawMessage(data)
}

func stringPtr(s string) *string {
	return &s
}
