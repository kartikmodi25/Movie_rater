package server

import (
	"backend-assignment/database/postgres"
	"backend-assignment/requests"
	"backend-assignment/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func RegisterUser(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		user := requests.User{}
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse request body"})
			return
		}
		exist, err := postgres.CheckExistingUser(db, user.Email)
		if err != nil {
			log.Error().Err(err).Str("email", user.Email).Msg("error checking user in db")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user, try again"})
			return
		}
		if exist {
			log.Error().Err(err).Str("email", user.Email).Msg("user already exists in db")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user already exist"})
			return
		}
		password := "qwerty"
		err = postgres.CreateUser(db, user.Name, user.Email, password)
		c.JSON(http.StatusCreated, user)
	}
}

func LoginUser(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		user := requests.LoginUser{}
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse request body"})
			return
		}

		exist, err := postgres.CheckExistingUser(db, user.Email)
		if err != nil {
			log.Error().Err(err).Str("email", user.Email).Msg("error checking user in db")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login user, try again"})
			return
		}
		if !exist {
			log.Error().Err(err).Str("email", user.Email).Msg("user does not exists in db")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user does not exist"})
			return
		}
		validLogin, err := postgres.CheckUserCredentials(db, user.Email, user.Password)
		if err != nil {
			log.Error().Err(err).Str("email", user.Email).Msg("error checking user in db")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login user, try again"})
			return
		}
		if !validLogin {
			log.Error().Err(err).Str("email", user.Email).Msg("entered credentials are not valid")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user with entered credentials does not exist"})
			return
		}
		accessToken := "dummyAccessToken"
		err = postgres.GenerateUserToken(db, user.Email, accessToken)
		if err != nil {
			log.Error().Err(err).Str("email", user.Email).Msg("error generating access token")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token, try again"})
			return
		}
		c.JSON(http.StatusCreated, accessToken)
	}
}

func RateMovie(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		movieData := requests.Movie{}
		if err := c.ShouldBindJSON(&movieData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse request body"})
			return
		}
		var rating int8
		rating = movieData.Rating

		if rating < 1 || rating > 10 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rating should be in the range [1, 10]"})
			return
		}
		movieName := utils.SearchMovie(movieData.Name)
		if movieName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Searched movie does not exist"})
			return
		}
		avgRating, err := postgres.UpdateMovieRating(db, movieName, rating)
		if err != nil {
			log.Error().Err(err).Str("name", movieData.Name).Msg("error updating movie in db")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update movie rating, try again"})
			return
		}
		c.JSON(http.StatusOK, avgRating)
	}
}