package utils

import (
	"gorm.io/driver/sqlite" // Sqlite driver based on GGO
	"gorm.io/gorm"
	"github.com/catmandx/pocpeer/models"
  )
  
func ConnectDb() (db *gorm.DB, err error) {
	db, err = gorm.Open(sqlite.Open("pocpeer.db"), &gorm.Config{})
	db.AutoMigrate(&models.News{})
	return db, err
}

func SaveOne(news models.News){

}

func SaveMany(news models.News){

}
