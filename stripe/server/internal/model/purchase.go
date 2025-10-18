package model

import "time"

type PaymentMethod int
type PaymentStatus int

const (
	CASH PaymentMethod = iota
	CARD_ONLINE
	CARD_IN_PERSON
)

const (
	COMPLETED PaymentStatus = iota
	CANCELED
	OPEN
)

type Purchase struct {
	Id                   int           `json:"id"`
	ItemId               int           `json:"item_id"`
	UserMembershipNumber string        `json:"user_membership_number"`
	Amount               float32       `json:"amount"`
	Date                 time.Time     `json:"date"`
	PaymentMethod        PaymentMethod `json:"payment_method"`
	PaymentStatus        PaymentStatus `json:"payment_status"`
}
