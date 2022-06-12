package models

import "time"

type Company struct {
	ID        string     `bson:"id"`
	Name      string     `bson:"name"`
	Code      string     `bson:"code"`
	Country   string     `bson:"country"`
	Website   string     `bson:"website"`
	Phone     string     `bson:"phone"`
	CreatedAt time.Time  `bson:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at,omit_empty"`
}
