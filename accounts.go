package donately

// Account represents a Donately account (organization) with all its configuration,
// billing information, payment processors, and associated metadata.
type Account struct {
	ID                      string            `json:"id"`
	Title                   string            `json:"title"`
	Subdomain               string            `json:"subdomain"`
	DonatelyHomepageURL     string            `json:"donately_homepage_url"`
	Status                  string            `json:"status"`
	Currency                string            `json:"currency"`
	Created                 int64             `json:"created"`
	Updated                 int64             `json:"updated"`
	TaxID                   *string           `json:"tax_id"`
	TaxExemptStatus         *string           `json:"tax_exempt_status"`
	DBAName                 *string           `json:"dba_name"`
	HomeLinkURL             *string           `json:"home_link_url"`
	Description             *string           `json:"description"`
	Images                  AccountImages     `json:"images"`
	MailReplyTo             *string           `json:"mail_reply_to"`
	EmailFooter             *string           `json:"email_footer"`
	DontSendReceiptEmails   *bool             `json:"dont_send_receipt_emails"`
	Livemode                *bool             `json:"livemode"`
	StripeConnectStatus     *string           `json:"stripe_connect_status"`
	DonationFeePercent      float64           `json:"donation_fee_percent"`
	PublishableMerchantKeys PublishableKeys   `json:"publishable_merchant_keys"`
	FormID                  string            `json:"form_id"`
	GoogleAnalyticsID       *string           `json:"google_analytics_id"`
	Type                    *string           `json:"type"`
	BusinessType            *string           `json:"business_type"`
	City                    *string           `json:"city"`
	State                   *string           `json:"state"`
	ZipCode                 *string           `json:"zip_code"`
	Country                 *string           `json:"country"`
	Phone                   *string           `json:"phone"`
	Billing                 AccountBilling    `json:"billing"`
	Processors              AccountProcessors `json:"processors"`
	MetaData                map[string]any    `json:"meta_data"`
	ScriptTags              map[string]any    `json:"script_tags"`
	HasDonations            bool              `json:"has_donations"`
}

// AccountImages contains the logo and header images for an account
// in various sizes and formats.
type AccountImages struct {
	Logo   AccountImageSizes `json:"logo"`
	Header HeaderImage       `json:"header"`
}

// AccountImageSizes provides URLs for an image in different sizes
// from original down to mini thumbnails.
type AccountImageSizes struct {
	Original *string `json:"original"`
	Large    *string `json:"large"`
	Medium   *string `json:"medium"`
	Small    *string `json:"small"`
	Thumb    *string `json:"thumb"`
	Mini     *string `json:"mini"`
}

// HeaderImage contains the header image URL for an account.
type HeaderImage struct {
	Original *string `json:"original"`
}

// PublishableKeys contains the Stripe publishable keys for both
// live and test modes.
type PublishableKeys struct {
	StripePublishableKey     *string `json:"stripe_publishable_key"`
	StripeTestPublishableKey *string `json:"stripe_test_publishable_key"`
}

// AccountBilling contains billing and subscription information
// for a Donately account.
type AccountBilling struct {
	SubscriptionPlan          string  `json:"subscription_plan"`
	SubscriptionInterval      *string `json:"subscription_interval"`
	SubscriptionStartDate     *string `json:"subscription_start_date"`
	SubscriptionEndDate       *string `json:"subscription_end_date"`
	SubscriptionAmountInCents int64   `json:"subscription_amount_in_cents"`
	BillingMode               *string `json:"billing_mode"`
	BillingDayOfMonth         int     `json:"billing_day_of_month"`
	FailedChargeAt            *string `json:"failed_charge_at"`
}

// AccountProcessors contains payment processor configuration
// including default settings and processor-specific details.
type AccountProcessors struct {
	DefaultProcessor *string         `json:"default_processor"`
	DefaultCurrency  string          `json:"default_currency"`
	Livemode         *bool           `json:"livemode"`
	Stripe           StripeProcessor `json:"stripe"`
	Paypal           *any            `json:"paypal"`
}

// StripeProcessor contains Stripe-specific configuration and keys
// for payment processing.
type StripeProcessor struct {
	StripeConnectStatus      *string `json:"stripe_connect_status"`
	StripeConnectEmail       *string `json:"stripe_connect_email"`
	StripePublishableKey     *string `json:"stripe_publishable_key"`
	StripeTestPublishableKey *string `json:"stripe_test_publishable_key"`
}

// Donations contains aggregated donation statistics including
// count, total amount, and last donation timestamp.
type Donations struct {
	Count         int   `json:"count"`
	AmountInCents int64 `json:"amount_in_cents"`
	LastDonation  int64 `json:"last_donation"`
}

// Fundraisers contains aggregated fundraiser statistics including
// count and total amount raised.
type Fundraisers struct {
	Count         int   `json:"count"`
	AmountInCents int64 `json:"amount_in_cents"`
}

// Notifications contains the count of notifications for an account.
type Notifications struct {
	Count int `json:"count"`
}
