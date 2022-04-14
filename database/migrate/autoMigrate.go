// autoMigrate.go needs to be executed only when it is required

package main

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/waldo237/gin-api-wm/config"
	"github.com/waldo237/gin-api-wm/database"
	"github.com/waldo237/gin-api-wm/database/model"
)

// Load all the models
type auth model.Auth
type user model.User
type post model.Post
type hobby model.Hobby
type userHobby model.UserHobby

var db *gorm.DB
var errorState int

func main() {
	configureDB := config.Database().RDBMS
	driver := configureDB.Env.Driver
	errorState = 0

	db = database.InitDB()

	dropAllTables()

	migrateTables()

	if driver != "sqlite3" {
		setPkFk()
	}

	if errorState == 0 {
		fmt.Println("Auto migration is completed!")
	} else {
		fmt.Println("Auto migration failed!")
	}
}

func dropAllTables() {
	// Careful! It will drop all the tables!
	if err := db.Migrator().DropTable(&userHobby{}, &hobby{}, &post{}, &user{}, &auth{}); err != nil {
		errorState = 1
		fmt.Println(err)
	} else {
		fmt.Println("Old tables are deleted!")
	}
}

func migrateTables() {
	configureDB := config.Database().RDBMS
	driver := configureDB.Env.Driver

	if driver == "mysql" {
		// db.Set() --> add table suffix during auto migration
		if err := db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&auth{},
			&user{}, &post{}, &hobby{}); err != nil {
			errorState = 1
			fmt.Println(err)
		} else {
			fmt.Println("New tables are  migrated successfully!")
		}
	} else {
		if err := db.AutoMigrate(&auth{},
			&user{}, &post{}, &hobby{}); err != nil {
			errorState = 1
			fmt.Println(err)
		} else {
			fmt.Println("New tables are  migrated successfully!")
		}
	}
}

func setPkFk() {
	// Manually set foreign key for MySQL and PostgreSQL
	if err := db.Migrator().CreateConstraint(&auth{}, "User"); err != nil {
		errorState = 1
		fmt.Println(err)
	}

	if err := db.Migrator().CreateConstraint(&user{}, "Posts"); err != nil {
		errorState = 1
		fmt.Println(err)
	}
}
