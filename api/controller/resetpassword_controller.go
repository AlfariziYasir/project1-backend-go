package controller

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"project1/api/mailer"
	"project1/api/models"
	"project1/api/security"
	"project1/api/utils/formaterror"

	"github.com/gin-gonic/gin"
)

func (server *Server) ForgotPassword(c *gin.Context) {
	errList = map[string]string{}

	//start processing the request
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

	err = server.DB.Model(&models.User{}).Where("email = ?", user.Email).Take(&user).Error
	if err != nil {
		errList["No_email"] = "Sorry, we do not recognize this email"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	rp := models.ResetPassword{}
	rp.Prepare()

	//generate token
	token := security.TokenHash(user.Email)
	rp.Email = user.Email
	rp.Token = token

	data, err := rp.SaveData(server.DB)
	if err != nil {
		errList = formaterror.FormatError(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	response, err := mailer.SendMail.SendResetPassword(data.Email, os.Getenv("SENDGRID_FROM"), rp.Token, os.Getenv("SENDGRID_API_KEY"), os.Getenv("APP_ENV"))
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"response": response.RespBody,
	})
}
