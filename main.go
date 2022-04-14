package main

import (
	"fmt"
	// "io"
	// "os"
	// "time"

	"github.com/gin-gonic/gin"
	"github.com/waldo237/gin-api-wm/config"
	"github.com/waldo237/gin-api-wm/controller"
	"github.com/waldo237/gin-api-wm/database"
	"github.com/waldo237/gin-api-wm/lib/middleware"
)

var configure = config.Config()

func main() {
	if configure.Database.RDBMS.Activate == "yes" {
		if err := database.InitDB().Error; err != nil {
			fmt.Println(err)
			return
		}
	}

	// JWT
	middleware.AccessKey = []byte(configure.Security.JWT.AccessKey)
	middleware.AccessKeyTTL = configure.Security.JWT.AccessKeyTTL
	middleware.RefreshKey = []byte(configure.Security.JWT.RefreshKey)
	middleware.RefreshKeyTTL = configure.Security.JWT.RefreshKeyTTL



	router, err := SetupRouter()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = router.Run(":" + configure.Server.ServerPort)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// SetupRouter ...
func SetupRouter() (*gin.Engine, error) {
	if configure.Server.ServerEnv == "production" {
		gin.SetMode(gin.ReleaseMode) 
	}


	// Create a log file with start time
	// dt := time.Now()
	// t := dt.Format(time.RFC3339)
	// file, err := os.Create("./logs/start:" + t + ".log")
	// if err != nil {
	// 	return nil, err
	// }
	// gin.DefaultWriter = io.MultiWriter(file)

	// If it is required to write the logs to the file and the console
	// at the same time
	// gin.DefaultWriter = io.MultiWriter(file, os.Stdout)

	// Creates a router without any middleware by default
	// router := gin.New()

	// Logger middleware: gin.DefaultWriter = os.Stdout
	// router.Use(gin.Logger())

	// Recovery middleware recovers from any panics and writes a 500
	// if there is one
	// router.Use(gin.Recovery())

	// gin.Default() = gin.New() + gin.Logger() + gin.Recovery()
	router := gin.Default()

	// Which proxy to trust
	if configure.Security.TrustedIP == "nil" {
		err := router.SetTrustedProxies(nil)
		if err != nil {
			return router, err
		}
	} else {
		if configure.Security.TrustedIP != "" {
			err := router.SetTrustedProxies([]string{configure.Security.TrustedIP})
			if err != nil {
				return router, err
			}
		}
	}

	router.Use(middleware.CORS())
	router.Use(middleware.SentryCapture(configure.Logger.SentryDsn))
	router.Use(middleware.Firewall(
		configure.Security.Firewall.ListType,
		configure.Security.Firewall.IP,
	))

	// Render HTML
	router.Use(middleware.Pongo2())

	// API:v1.0
	v1 := router.Group("/api/v1/")
	{
		// RDBMS
		if configure.Database.RDBMS.Activate == "yes" {
			// Register - no JWT required
			v1.POST("register", controller.CreateUserAuth)

			// Login - app issues JWT
			v1.POST("login", controller.Login)

			// Refresh - app issues new JWT
			rJWT := v1.Group("refresh")
			rJWT.Use(middleware.RefreshJWT())
			rJWT.POST("", controller.Refresh)

			// User
			rUsers := v1.Group("users")
			rUsers.GET("", controller.GetUsers)    // Non-protected
			rUsers.GET("/:id", controller.GetUser) // Non-protected
			rUsers.Use(middleware.JWT())
			rUsers.POST("", controller.CreateUser)      // Protected
			rUsers.PUT("", controller.UpdateUser)       // Protected
			rUsers.PUT("/hobbies", controller.AddHobby) // Protected

			// Post
			rPosts := v1.Group("posts")
			rPosts.GET("", controller.GetPosts)    // Non-protected
			rPosts.GET("/:id", controller.GetPost) // Non-protected
			rPosts.Use(middleware.JWT())
			rPosts.POST("", controller.CreatePost)       // Protected
			rPosts.PUT("/:id", controller.UpdatePost)    // Protected
			rPosts.DELETE("/:id", controller.DeletePost) // Protected

			// Hobby
			rHobbies := v1.Group("hobbies")
			rHobbies.GET("", controller.GetHobbies) // Non-protected
		}

		// Basic Auth demo
		user := configure.Security.BasicAuth.Username
		pass := configure.Security.BasicAuth.Password
		rBasicAuth := v1.Group("access_resources")
		rBasicAuth.Use(gin.BasicAuth(gin.Accounts{
			user: pass,
		}))
		rBasicAuth.GET("", controller.AccessResource) // Protected
	}

	return router, nil
}
