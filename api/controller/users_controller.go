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

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (server *Server) CreateUser(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		//clear previous error if any
		errList = map[string]string{}

		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			errList["Invalid_body"] = "Unable to get request"
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": http.StatusUnprocessableEntity,
				"error":  errList,
			})
			return
		}

		user := models.User{}

		err = json.Unmarshal(body, &user)
		if err != nil {
			errList["Unmarshal_error"] = "Cannot unmarshal body"
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": http.StatusUnprocessableEntity,
				"error":  errList,
			})
			return
		}

		user.Prepare()
		errMsg := user.Validate("")
		if len(errMsg) > 0 {
			errList = errMsg
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": http.StatusUnprocessableEntity,
				"error":  errList,
			})
			return
		}

		data, err := user.SaveUser(server.DB)
		if err != nil {
			formattedError := formaterror.FormatError(err.Error())
			errList = formattedError
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": http.StatusInternalServerError,
				"error":  errList,
			})
			return
		}

		//add user role to casbin table
		enforcer.AddGroupingPolicy(fmt.Sprintf(user.UserId), user.Role)

		c.JSON(http.StatusCreated, gin.H{
			"status": http.StatusCreated,
			"response": map[string]interface{}{
				"user_id":  data.UserId,
				"username": data.Username,
				"email":    data.Email,
				"role":     user.Role,
			},
		})
	}
}

func (server *Server) GetUsers(c *gin.Context) {
	//clear previous error if any
	errList = map[string]string{}

	user := models.User{}

	data, err := user.GetUsers(server.DB)
	if err != nil {
		errList["No_user"] = "No User Found"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"response": data,
	})
}

func (server *Server) GetUser(c *gin.Context) {
	errList = map[string]string{}

	uID := c.Param("id")

	user := models.User{}
	data, err := user.GetUser(uID, server.DB)
	if err != nil {
		errList["No_user"] = "No user found"
		c.JSON(http.StatusNotFound, gin.H{
			"status": http.StatusNotFound,
			"error":  errList,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"response": data,
	})
}

func (server *Server) UpdateUser(c *gin.Context) {
	errList = map[string]string{}

	uID := c.Param("id")
	// get user id from the token for valid tokens
	tokenID, err := auth.ExtractTokenUID(c.Request)
	if err != nil {
		errList["Invalid_request"] = "Invalid request"
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"error":  errList,
		})
		return
	}
	//if the id is not the authenticated user id
	if tokenID != "" && tokenID != uID {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}
	//start proccessing the request
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errList["Invalid_body"] = "Unable to get request"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}
	reqBody := map[string]string{}
	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		errList["Unmarshal_error"] = "Cannot unmarshal body"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}
	//check for previous details
	formerUser := models.User{}
	err = server.DB.Debug().Model(&models.User{}).Where("user_id = ?", uID).Take(&formerUser).Error
	if err != nil {
		errList["User_invalid"] = "The user is does not exist"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	newUser := models.User{}
	//when current password has content
	if reqBody["current_password"] != "" && reqBody["new_password"] == "" {
		errList["Empty_current"] = "Please Provide current password"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}
	if reqBody["current_password"] != "" && reqBody["new_password"] == "" {
		errList["Empty_new"] = "Please Provide new password"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}
	if reqBody["current_password"] != "" && reqBody["new_password"] == "" {
		//alse check if the new password
		if len(reqBody["new_password"]) < 6 {
			errList["Invalid_password"] = "Password should be atleast 6 characters"
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": http.StatusUnprocessableEntity,
				"error":  errList,
			})
			return
		}
		//if check former password with new password
		err := security.VerifyPassword(formerUser.Password, reqBody["current_password"])
		if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
			errList["Password_mismatch"] = "The password not correct"
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": http.StatusUnprocessableEntity,
				"error":  errList,
			})
			return
		}
		//update both the password and the email
		newUser.Username = reqBody["username"]
		newUser.Email = reqBody["email"]
		newUser.Password = reqBody["new_password"]
	}
	//if password field not entered, so update only the email and username
	newUser.Username = reqBody["username"]
	newUser.Email = reqBody["email"]

	newUser.Prepare()
	errMsg := newUser.Validate("update")
	if len(errMsg) > 0 {
		errList = errMsg
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	data, err := newUser.UpdateUser(uID, server.DB)
	if err != nil {
		errList := formaterror.FormatError(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"response": data,
	})
}

func (server *Server) DeleteUser(c *gin.Context) {
	errList = map[string]string{}
	uID := c.Param("id")
	//get user id from the token for valid tokens
	tokenID, err := auth.ExtractTokenUID(c.Request)
	if err != nil {
		errList["Unauthorize"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}
	// if the id is not the authenticated user id
	if tokenID != "" && tokenID != uID {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	user := models.User{}
	_, err = user.DeleteUser(uID, server.DB)
	if err != nil {
		errList["Other_error"] = "Please try again later"
		c.JSON(http.StatusNotFound, gin.H{
			"status": http.StatusNotFound,
			"error":  errList,
		})
		return
	}

	//also delete meeting, join, report
	meeting := models.Meeting{}
	join := models.Join{}
	report := models.Report{}

	_, err = meeting.DeleteUserMeeting(server.DB, uID)
	if err != nil {
		errList["Other_error"] = "Please try again later"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  err,
		})
		return
	}

	_, err = join.DeleteUserJoin(server.DB, uID)
	if err != nil {
		errList["Other_error"] = "Please try again later"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  err,
		})
		return
	}

	_, err = report.DeleteReport(server.DB, uID)
	if err != nil {
		errList["Other_error"] = "Please try again later"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"response": "User deleted",
	})
}
