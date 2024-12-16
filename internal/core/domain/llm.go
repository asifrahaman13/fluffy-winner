package domain

type Query struct {
	Search string `json:"search" bson:"search"`
}