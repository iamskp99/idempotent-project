package main

import (
	userController "idempotent-project/controller"
	userService "idempotent-project/service"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

func main() {
	r := gin.Default()

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis address
	})

	// Initialize the UserService
	userService := &userService.Service{}

	// Create the UserController
	userController := userController.NewUserController(userService, redisClient)

	// Define routes
	r.POST("/users", userController.SaveUser)
	r.GET("/users/:id", userController.GetUser)
	r.POST("/users/:id/operate", userController.OperateUser)
	r.GET("/users/getKey", userController.GetKey)

	// Run the server
	r.Run(":8080") // listen and serve on localhost:8080
}
