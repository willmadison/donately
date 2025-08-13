package donately

import "time"

// Subscription represents a recurring donation subscription
// with payment details, scheduling information, and associated metadata.
type Subscription struct {
	ID                       string         `json:"id"`
	DonationType             string         `json:"donation_type"`
	Status                   string         `json:"status"`
	Processor                string         `json:"processor"`
	Livemode                 bool           `json:"livemode"`
	AmountInCents            int64          `json:"amount_in_cents"`
	Currency                 string         `json:"currency"`
	Created                  int64          `json:"created"`
	Updated                  int64          `json:"updated"`
	RecurringStartDay        int64          `json:"recurring_start_day"`
	RecurringStopDay         int64          `json:"recurring_stop_day"`
	RecurringFrequency       string         `json:"recurring_frequency"`
	RecurringDayOfMonth      int            `json:"recurring_day_of_month"`
	CreditCardType           string         `json:"cc_type"`
	CreditCardLast4          string         `json:"cc_last4"`
	CreditCardExpMonth       string         `json:"cc_exp_month"`
	CreditCardExpYear        string         `json:"cc_exp_year"`
	Anonymous                bool           `json:"anonymous"`
	OnBehalfOf               *string        `json:"on_behalf_of"`
	Comment                  *string        `json:"comment"`
	TrackingCodes            string         `json:"tracking_codes"`
	MetaData                 map[string]any `json:"meta_data"`
	DonationParent           DonationLite   `json:"donation_parent"`
	Person                   Person         `json:"person"`
	Account                  Account        `json:"account"`
	Campaign                 any            `json:"campaign"`
	Fundraiser               any            `json:"fundraiser"`
	ChargeSource             ChargeSource   `json:"charge_source"`
	InternalID               int64          `json:"internal_id"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
	RestartRecurringSchedule *string        `json:"restart_recurring_schedule"`
	ReferrerID               *string        `json:"referrer_id"`
	Notes                    *string        `json:"notes"`
}

// DonationLite represents a simplified donation record used
// as a reference within subscription data.
type DonationLite struct {
	ID            string `json:"id"`
	Object        string `json:"object"`
	DonationType  string `json:"donation_type"`
	Processor     string `json:"processor"`
	Status        string `json:"status"`
	Livemode      bool   `json:"livemode"`
	DonationDate  int64  `json:"donation_date"`
	AmountInCents int64  `json:"amount_in_cents"`
	Currency      string `json:"currency"`
	Recurring     bool   `json:"recurring"`
	Refunded      *bool  `json:"refunded"`
}
