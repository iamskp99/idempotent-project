package userController

import (
	"fmt"
	userEntity "idempotent-project/entity"
	userService "idempotent-project/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

type UserController struct {
	Service     userService.UserService
	RedisClient *redis.Client
}

func NewUserController(service userService.UserService, redisClient *redis.Client) *UserController {
	return &UserController{
		Service:     service,
		RedisClient: redisClient,
	}
}

func (uc *UserController) SaveUser(c *gin.Context) {
	var user userEntity.User

	// Bind incoming JSON to the User struct
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the service to save the user
	err := uc.Service.Save(&user, uc.RedisClient)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User saved successfully"})
}

func (uc *UserController) GetUser(c *gin.Context) {
	// Get user ID from path parameter
	userId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Call the service to retrieve the user
	user, err := uc.Service.GetUser(userId, uc.RedisClient)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (uc *UserController) OperateUser(c *gin.Context) {
	// Get user ID from path parameter
	userId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
	}

	idemKey := c.GetHeader("Idempotency-Key")

	keyToCheck := fmt.Sprintf("idempotent:%d", idemKey)

	val, _ := uc.RedisClient.Exists(keyToCheck).Result()
	if val == 1 {
		c.JSON(http.StatusOK, gin.H{"message": "The Request has already been processed"})
		return
	} else {
		err := uc.RedisClient.Set(keyToCheck, keyToCheck, time.Hour*24).Err() // Expire after 24 hours
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Get value to operate from request body (could be positive or negative)
	var requestBody struct {
		Value int32 `json:"value"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the service to update the user's balance
	err = uc.Service.Operate(userId, requestBody.Value, uc.RedisClient)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User balance updated"})
}

func (uc *UserController) GetKey(c *gin.Context) {
	key, err := uc.Service.GetKey(uc.RedisClient)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"key": key})
}
