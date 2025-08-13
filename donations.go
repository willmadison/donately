package donately

import (
	"time"
)

// Donation represents a donation record from the Donately API.
// It contains all the details about a donation including amount, donor information,
// associated campaign, payment processor details, and metadata.
type Donation struct {
	ID                  string       `json:"id"`
	DonationType        string       `json:"donation_type"`
	Processor           string       `json:"processor"`
	Status              string       `json:"status"`
	Livemode            bool         `json:"livemode"`
	DonationDate        int64        `json:"donation_date"`
	AmountInCents       int64        `json:"amount_in_cents"`
	Currency            string       `json:"currency"`
	Recurring           bool         `json:"recurring"`
	Refunded            *bool        `json:"refunded"`
	TransactionID       string       `json:"transaction_id"`
	Created             int64        `json:"created"`
	Updated             int64        `json:"updated"`
	AmountFormatted     string       `json:"amount_formatted"`
	Anonymous           bool         `json:"anonymous"`
	OnBehalfOf          string       `json:"on_behalf_of"`
	Comment             string       `json:"comment"`
	TrackingCodes       string       `json:"tracking_codes"`
	MetaData            MetaData     `json:"meta_data"`
	Person              Person       `json:"person"`
	Account             Account      `json:"account"`
	Campaign            Campaign     `json:"campaign"`
	Fundraiser          any          `json:"fundraiser"`
	Subscription        Subscription `json:"subscription"`
	Parent              any          `json:"parent"`
	Refunds             []any        `json:"refunds"`
	ChargeSource        ChargeSource `json:"charge_source"`
	ReferrerID          *string      `json:"referrer_id"`
	RemoteIP            string       `json:"remote_ip"`
	FeeInCents          int64        `json:"fee_in_cents"`
	InternalID          int64        `json:"internal_id"`
	CreatedAt           time.Time    `json:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"`
	FeeStripeChargeID   string       `json:"fee_stripe_charge_id"`
	StripeCustomerID    string       `json:"stripe_customer_id"`
	StripeConnectIDHash string       `json:"stripe_connect_id_hash"`
	AmountInCentsUSD    int64        `json:"amount_in_cents_usd"`
	Notes               *string      `json:"notes"`
}

// MetaData contains additional metadata associated with a donation,
// including base amount and donor fee payment information.
type MetaData struct {
	BaseAmount    int64 `json:"base-amount"`
	DonorPaysFees int64 `json:"donor-pays-fees"`
}

// ChargeSource represents the payment method used for a donation,
// typically containing credit card or payment processor information from Stripe.
type ChargeSource struct {
	ID                 string         `json:"id"`
	Object             string         `json:"object"`
	AddressCity        string         `json:"address_city"`
	AddressCountry     string         `json:"address_country"`
	AddressLine1       string         `json:"address_line1"`
	AddressLine1Check  string         `json:"address_line1_check"`
	AddressLine2       string         `json:"address_line2"`
	AddressState       string         `json:"address_state"`
	AddressZip         string         `json:"address_zip"`
	AddressZipCheck    string         `json:"address_zip_check"`
	Brand              string         `json:"brand"`
	Country            string         `json:"country"`
	Customer           string         `json:"customer"`
	CVCCheck           string         `json:"cvc_check"`
	DynamicLast4       *string        `json:"dynamic_last4"`
	ExpMonth           int            `json:"exp_month"`
	ExpYear            int            `json:"exp_year"`
	Fingerprint        string         `json:"fingerprint"`
	Funding            string         `json:"funding"`
	Last4              string         `json:"last4"`
	Metadata           map[string]any `json:"metadata"`
	Name               string         `json:"name"`
	TokenizationMethod *string        `json:"tokenization_method"`
}
