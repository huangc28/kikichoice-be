package webhooks

import "time"

// ClerkWebhookEvent represents the main webhook event structure from Clerk
type ClerkWebhookEvent struct {
	Data       ClerkUser `json:"data"`
	Object     string    `json:"object"`
	Type       string    `json:"type"`
	Timestamp  int64     `json:"timestamp"`
	InstanceID string    `json:"instance_id"`
}

// ClerkUser represents the user data structure from Clerk
type ClerkUser struct {
	ID             string              `json:"id"`
	FirstName      *string             `json:"first_name"`
	LastName       *string             `json:"last_name"`
	EmailAddresses []ClerkEmailAddress `json:"email_addresses"`
	ImageURL       *string             `json:"image_url"`
	CreatedAt      int64               `json:"created_at"`
	UpdatedAt      int64               `json:"updated_at"`
	ExternalID     *string             `json:"external_id"`
}

// ClerkEmailAddress represents an email address from Clerk
type ClerkEmailAddress struct {
	EmailAddress string                 `json:"email_address"`
	ID           string                 `json:"id"`
	Verification ClerkEmailVerification `json:"verification"`
}

// ClerkEmailVerification represents the verification status of an email
type ClerkEmailVerification struct {
	Status   string `json:"status"`
	Strategy string `json:"strategy"`
}

// GetPrimaryEmail returns the first verified email address, or the first email if none are verified
func (u ClerkUser) GetPrimaryEmail() *string {
	if len(u.EmailAddresses) == 0 {
		return nil
	}

	// First, look for a verified email
	for _, email := range u.EmailAddresses {
		if email.Verification.Status == "verified" {
			return &email.EmailAddress
		}
	}

	// If no verified email, return the first one
	return &u.EmailAddresses[0].EmailAddress
}

// GetFullName returns the combined first and last name
func (u ClerkUser) GetFullName() string {
	var name string
	if u.FirstName != nil {
		name = *u.FirstName
	}
	if u.LastName != nil {
		if name != "" {
			name += " "
		}
		name += *u.LastName
	}

	// If no name provided, use a default
	if name == "" {
		name = "User"
	}

	return name
}

// GetCreatedAtTime converts the timestamp to time.Time
func (u ClerkUser) GetCreatedAtTime() time.Time {
	return time.UnixMilli(u.CreatedAt)
}

// GetUpdatedAtTime converts the timestamp to time.Time
func (u ClerkUser) GetUpdatedAtTime() time.Time {
	return time.UnixMilli(u.UpdatedAt)
}
