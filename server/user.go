package server

import (
	"backend-assignment/requests"
	"backend-assignment/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	logger "github.com/rs/zerolog/log"
)

func RegisterUser(s *Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		user := requests.User{}
		requestId := requests.ID(c)
		log := logger.With().Str("requestID", requestId).Str("email", user.Email).Logger()
		if err := c.ShouldBindJSON(&user); err != nil {
			log.Error().Err(err).Msg("failed to parse request body")
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse request body"})
			return
		}
		exist, err := s.db.CheckExistingUser(ctx, user.Email)
		if err != nil {
			log.Error().Err(err).Str("email", user.Email).Msg("error checking user in db")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user, try again"})
			return
		}
		if exist {
			log.Error().Err(err).Str("email", user.Email).Msg("user already exists in db")
			c.JSON(http.StatusForbidden, gin.H{"error": "user already exist"})
			return
		}
		password := "qwerty"
		err = s.db.CreateUser(ctx, user.Name, user.Email, password)
		if err != nil {
			log.Error().Err(err).Str("email", user.Email).Msg("error creating user in db")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user, try again"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"user": user})
	}
}

func LoginUser(s *Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		user := requests.LoginUser{}

		if err := c.ShouldBindJSON(&user); err != nil {
			log.Error().Err(err).Msg("failed to parse request body")
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse request body"})
			return
		}
		requestId := requests.ID(c)
		log := logger.With().Str("requestID", requestId).Str("email", user.Email).Logger()

		exist, err := s.db.CheckExistingUser(ctx, user.Email)
		if err != nil {
			log.Error().Err(err).Str("email", user.Email).Msg("error checking user in db")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login user, try again"})
			return
		}
		if !exist {
			log.Error().Err(err).Str("email", user.Email).Msg("user does not exists in db")
			c.JSON(http.StatusNotFound, gin.H{"error": "user does not exist"})
			return
		}
		validLogin, err := s.db.CheckUserCredentials(ctx, user.Email, user.Password)
		if err != nil {
			log.Error().Err(err).Str("email", user.Email).Msg("error checking user in db")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login user, try again"})
			return
		}
		if !validLogin {
			log.Error().Err(err).Str("email", user.Email).Msg("entered credentials are not valid")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user with entered credentials does not exist"})
			return
		}
		accessToken, err := utils.GenerateJWTToken(user.Email)
		if err != nil {
			log.Error().Err(err).Str("email", user.Email).Msg("error generating access token")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token, try again"})
			return
		}

		// Send the token in the response
		c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
	}
}

func RateMovie(s *Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		movieData := requests.Movie{}
		if err := c.ShouldBindJSON(&movieData); err != nil {
			log.Error().Err(err).Msg("failed to parse request body")
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse request body"})
			return
		}
		requestId := requests.ID(c)
		log := logger.With().Str("requestID", requestId).Str("name", movieData.Name).Logger()
		var rating int8
		rating = movieData.Rating

		if rating < 1 || rating > 10 {
			log.Error().Err(errors.New("entered rating is out of valid range [1, 10]")).Msg("incorrect rating value")
			c.JSON(http.StatusBadRequest, gin.H{"error": "rating should be in the range [1, 10]"})
			return
		}
		movieName := utils.SearchMovie(movieData.Name)
		if movieName == "" {
			log.Error().Err(errors.New("searched movie does not exist in the database")).Msg("invalid movie name")
			c.JSON(http.StatusBadRequest, gin.H{"error": "searched movie does not exist"})
			return
		}
		avgRating, err := s.db.UpdateMovieRating(ctx, movieName, rating)
		if err != nil {
			log.Error().Err(err).Str("name", movieData.Name).Msg("error updating movie in db")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update movie rating, try again"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"average_rating": avgRating})
	}
}
func ListMovies(s *Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		requestId := requests.ID(c)
		log := logger.With().Str("requestID", requestId).Logger()
		// Validate through access-token
		movieList, err := s.db.GetMoviesData(ctx)
		if err != nil {
			log.Error().Err(err).Msg("error in retreiving movie list from db")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retreive movie list, try again"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"movie_list": movieList})
	}
}
func ListMovieRatings(s *Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		requestId := requests.ID(c)
		log := logger.With().Str("requestID", requestId).Logger()
		movieList, err := s.db.GetMovieRatings(ctx)
		if err != nil {
			log.Error().Err(err).Msg("error in retreiving movie list from db")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retreive movie list with ratings, try again"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"movie_list": movieList})
	}
}
