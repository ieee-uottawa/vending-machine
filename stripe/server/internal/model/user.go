package model

type User struct {
	MembershipNumber string `json:"membership_number"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
}
