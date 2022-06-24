package models

import ("time")
type News struct {
	ID			uint
	CveNum 		string
	Links 		string
	Text 		string
	Author 		string
	Platform 	string 
	PlatformId  string
	CreatedAt	time.Time
}