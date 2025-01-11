package userService

import (
	"encoding/json"
	"fmt"
	userEntity "idempotent-project/entity"
	"time"

	"github.com/go-redis/redis"
	"golang.org/x/exp/rand"
)

type UserService interface {
	Save(user_data *userEntity.User, rc *redis.Client) error
	GetUser(userId uint64, rc *redis.Client) (*userEntity.User, error)
	Operate(user_id uint64, val int32, rc *redis.Client) error
	GetKey(rc *redis.Client) (uint64, error)
}

type Service struct {
}

func (service *Service) Save(user *userEntity.User, rc *redis.Client) error {
	userData, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %v", err)
	}
	userKey := fmt.Sprintf("user:%d", user.Id)

	// Save user data to Redis with an expiration time (optional)
	err = rc.Set(userKey, userData, time.Hour*24).Err() // Expire after 24 hours
	if err != nil {
		return fmt.Errorf("failed to save user data to Redis: %v", err)
	}

	fmt.Println("User saved successfully in Redis")
	return nil
}

func (service *Service) GetUser(userId uint64, rc *redis.Client) (*userEntity.User, error) {
	userKey := fmt.Sprintf("user:%d", userId)

	// Get user data from Redis
	userData, err := rc.Get(userKey).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("user not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get user data from Redis: %v", err)
	}

	// Deserialize JSON into a User struct
	var user userEntity.User
	err = json.Unmarshal([]byte(userData), &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %v", err)
	}

	return &user, nil
}

func (service *Service) Operate(userId uint64, val int32, rc *redis.Client) error {
	userKey := fmt.Sprintf("user:%d", userId)

	// Get user data from Redis
	userData, err := rc.Get(userKey).Result()
	if err == redis.Nil {
		return fmt.Errorf("user not found")
	} else if err != nil {
		return fmt.Errorf("failed to get user data from Redis: %v", err)
	}

	// Deserialize JSON into a User struct
	var user userEntity.User
	err = json.Unmarshal([]byte(userData), &user)
	user.Balance -= val
	if err != nil {
		return fmt.Errorf("failed to unmarshal user data: %v", err)
	}
	updatedUserData, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal updated user data: %v", err)
	}
	// Save the updated user data back to Redis
	err = rc.Set(userKey, updatedUserData, 0).Err() // No expiration
	if err != nil {
		return fmt.Errorf("failed to save updated user data to Redis: %v", err)
	}

	fmt.Println("User updated successfully in Redis")
	return nil
}

func (service *Service) GetKey(rc *redis.Client) (uint64, error) {

	// Generate a random uint64 number
	var randomNum uint64
	for {
		randomNum = rand.Uint64()
		key := fmt.Sprintf("idempotent:%d", randomNum)
		val, _ := rc.Exists(key).Result()
		if val == 1 {
		} else {
			break
		}
	}

	return randomNum, nil
}
