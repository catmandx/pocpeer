package models

import (
	"gorm.io/gorm"
)

type Application struct {
	Db      *gorm.DB
	Sinks   []Sink
	Sources []Source
}

type Sink interface {
	SendMessage(news News)
}

type Source interface {
	Init() (err error)
	Run(application Application) 
}