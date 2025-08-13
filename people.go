package donately

// Person represents a donor or user in the Donately system
// with their contact information, sign-in details, and associated accounts.
type Person struct {
	ID                          string    `json:"id"`
	Email                       string    `json:"email"`
	FirstName                   string    `json:"first_name"`
	LastName                    string    `json:"last_name"`
	PhoneNumber                 string    `json:"phone_number"`
	StreetAddress               string    `json:"street_address"`
	StreetAddress2              string    `json:"street_address_2"`
	City                        string    `json:"city"`
	State                       string    `json:"state"`
	ZipCode                     string    `json:"zip_code"`
	Country                     string    `json:"country"`
	Created                     int64     `json:"created"`
	Updated                     int64     `json:"updated"`
	LastSignIn                  IPAddress `json:"last_sign_in"`
	ConnectedToMultipleAccounts bool      `json:"connected_to_multiple_accounts"`
	HasAdminRoles               bool      `json:"has_admin_roles"`
	MetaData                    any       `json:"meta_data"`
	Accounts                    []Account `json:"accounts"`
}

// IPAddress contains geolocation and timing information
// for a person's last sign-in event.
type IPAddress struct {
	IPAddress  *string `json:"ip_address"`
	Object     string  `json:"object"`
	City       *string `json:"city"`
	State      *string `json:"state"`
	Country    *string `json:"country"`
	PostalCode *string `json:"postal_code"`
	SignInTime int64   `json:"sign_in_time"`
}
