package database

import (
	"fmt"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/assert"

	"github.com/mbrockmandev/tometracker/internal/models"
)

func TestCreateUser(t *testing.T) {
	repo := setupTestDB()
	defer teardownTestDB(repo)

	user := models.User{
		Email:     fmt.Sprintf("testuser+%v@example.com", time.Now().Unix()),
		Password:  "password",
		FirstName: "first",
		LastName:  "last",
		Role:      "admin",
	}

	_, err := repo.CreateUser(&user)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateUser(t *testing.T) {
	repo := setupTestDB()
	defer teardownTestDB(repo)

	email := fmt.Sprintf("testuser+%v@example.com", time.Now().Unix())
	user := models.User{
		Email:     email,
		Password:  "password2",
		FirstName: "first2",
		LastName:  "last2",
		Role:      "admin",
		UpdatedAt: time.Now(),
	}

	_, err := repo.CreateUser(&user)
	if err != nil {
		t.Log(err)
	}

	addedUser, err := repo.GetUserByEmail(email)
	if err != nil {
		t.Log(err)
	}

	user.LastName = "newlast2"
	err = repo.UpdateUser(addedUser.ID, &user)
	if err != nil {
		t.Log(err)
	}

	updatedUser, err := repo.GetUserByEmail(email)
	if err != nil {
		t.Log(err)
	}

	assert.Equal(t, user.LastName, updatedUser.LastName)

	_ = repo.DeleteUser(addedUser.ID)
}

func TestDeleteUser(t *testing.T) {
	repo := setupTestDB()
	defer teardownTestDB(repo)

	id, err := repo.CreateUser(&models.User{
		Email:     "testuser@example.com",
		Password:  "password",
		FirstName: "first",
		LastName:  "last",
		Role:      "admin",
	})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err = repo.GetUserById(id)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}

	err = repo.DeleteUser(id)
	if err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	user, _ := repo.GetUserById(id)
	if user != nil {
		t.Fatalf("user should have been deleted")
	}
}

func TestGetAllUsers(t *testing.T) {
	repo := setupTestDB()
	defer teardownTestDB(repo)

	users, err := repo.GetAllUsers()
	if err != nil {
		fmt.Println(err)
	}
	if users == nil {
		t.Fatal("users should not be nil")
	}

	if len(users) == 0 {
		t.Fatal("expected at least one user, got none")
	}

	// for _, user := range users {
	// fmt.Println(user)
	// }
}
