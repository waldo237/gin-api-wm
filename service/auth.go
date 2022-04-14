package service

import (
	"github.com/waldo237/gin-api-wm/database"
	"github.com/waldo237/gin-api-wm/database/model"
)

// GetUserByEmail ...
func GetUserByEmail(email string) (*model.Auth, error) {
	db := database.GetDB()

	var auth model.Auth

	if err := db.Where("email = ? ", email).First(&auth).Error; err != nil {
		return nil, err
	}

	return &auth, nil
}
