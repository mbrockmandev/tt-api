package models

type Library struct {
	ID            int    `db:"id" json:"id"`
	Name          string `db:"name" json:"name"`
	City          string `db:"city" json:"city"`
	StreetAddress string `db:"street_address" json:"street_address"`
	PostalCode    string `db:"postal_code" json:"postal_code"`
	Country       string `db:"country" json:"country"`
	Phone         string `db:"phone" json:"phone"`
}
