package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"project1/api/auth"
	"project1/api/models"
	"project1/api/security"
	"project1/api/utils/formaterror"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (server *Server) Login(c *gin.Context) {
	errList = map[string]string{}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status":      http.StatusUnprocessableEntity,
			"first error": "Unable to get request",
		})
		return
	}

	user := models.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  "Cannot unmarshal body",
		})
		return
	}

	user.Prepare()
	errMsg := user.Validate("login")
	if len(errMsg) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errMsg,
		})
		return
	}

	data, err := server.SignIn(user.Email, user.Password)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  formattedError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"response": data,
	})
}

func (server *Server) SignIn(email, password string) (map[string]interface{}, error) {
	userData := make(map[string]interface{})

	user := models.User{}
	err := server.DB.Debug().Model(&models.User{}).Where("email = ?", email).Take(&user).Error
	if err != nil {
		fmt.Println("this is the error getting the user: ", err)
		return nil, err
	}
	err = security.VerifyPassword(user.Password, password)
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		fmt.Println("this is the error hashing the password: ", err)
		return nil, err
	}
	token, err := auth.CreateToken(user.UserId)
	if err != nil {
		fmt.Println("this the error creating the token: ", err)
		return nil, err
	}
	userData["token"] = token
	userData["user_id"] = user.UserId
	userData["email"] = user.Email
	userData["username"] = user.Username

	return userData, nil
}
