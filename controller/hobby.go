package controller

import (
	"net/http"

	"github.com/waldo237/gin-api-wm/database"
	"github.com/waldo237/gin-api-wm/database/model"
	"github.com/waldo237/gin-api-wm/lib/renderer"

	"github.com/gin-gonic/gin"
)

// GetHobbies - GET /hobbies
func GetHobbies(c *gin.Context) {
	db := database.GetDB()
	hobbies := []model.Hobby{}

	if err := db.Find(&hobbies).Error; err != nil {
		renderer.Render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
	} else {
		renderer.Render(c, hobbies, http.StatusOK)
	}
}
