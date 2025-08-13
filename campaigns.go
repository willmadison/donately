package donately

import "time"

// Campaign represents a fundraising campaign with its goals,
// progress, content, and associated metadata.
type Campaign struct {
	ID                  string         `json:"id"`
	Title               string         `json:"title"`
	Slug                string         `json:"slug"`
	Type                string         `json:"type"`
	URL                 string         `json:"url"`
	Status              string         `json:"status"`
	Permalink           string         `json:"permalink"`
	Description         *string        `json:"description"`
	Content             *string        `json:"content"`
	Created             int64          `json:"created"`
	Updated             int64          `json:"updated"`
	StartDate           *string        `json:"start_date"` // use *time.Time if ISO format is confirmed
	EndDate             *string        `json:"end_date"`
	GoalInCents         int64          `json:"goal_in_cents"`
	AmountRaisedInCents int64          `json:"amount_raised_in_cents"`
	PercentFunded       float64        `json:"percent_funded"`
	DonorsCount         int            `json:"donors_count"`
	Images              CampaignImages `json:"images"`
	Account             Account        `json:"account"`
	FormID              string         `json:"form_id"`
	MetaData            any            `json:"meta_data"`
	InternalID          int64          `json:"internal_id"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	Recurring           *bool          `json:"recurring"`
	FundraiserGoal      *int64         `json:"fundraiser_goal"`
	DonationAmount      *int64         `json:"donation_amount"`
}

// CampaignImages contains the photo and cover photo for a campaign
// in various sizes.
type CampaignImages struct {
	Photo      CampaignPhotoSizes      `json:"photo"`
	CoverPhoto CampaignCoverPhotoSizes `json:"cover_photo"`
}

// CampaignPhotoSizes provides campaign photo URLs in different sizes
// from original down to icon size.
type CampaignPhotoSizes struct {
	Original *string `json:"original"`
	Medium   *string `json:"medium"`
	MediumV2 *string `json:"medium_v2"`
	Average  *string `json:"average"`
	Small    *string `json:"small"`
	Thumb    *string `json:"thumb"`
	SquareV2 *string `json:"square_v2"`
	Icon     *string `json:"icon"`
}

// CampaignCoverPhotoSizes provides campaign cover photo URLs
// in original and large sizes.
type CampaignCoverPhotoSizes struct {
	Original *string `json:"original"`
	Large    *string `json:"large"`
}
