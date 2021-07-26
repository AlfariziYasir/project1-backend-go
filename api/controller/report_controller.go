package controller

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"project1/api/auth"
	"project1/api/models"
	"project1/api/utils/formaterror"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (server *Server) SaveReport(c *gin.Context) {
	errList = map[string]string{}

	id := c.Param("id")
	mID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		errList["Invalid_request"] = "invalid request"
		c.JSON(http.StatusBadRequest, gin.H{
			"status":   http.StatusBadRequest,
			"response": errList,
		})
		return
	}

	//check token
	uID, err := auth.ExtractTokenUID(c.Request)
	if err != nil {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	//check if the user exist
	user := models.User{}
	err = server.DB.Model(&models.User{}).Where("user_id = ?", uID).Take(&user).Error
	if err != nil {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	//check if the meeting exist
	meeting := models.Meeting{}
	err = server.DB.Model(&models.Meeting{}).Where("id = ?", mID).Take(&meeting).Error
	if err != nil {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	//binding data
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		if err != nil {
			errList["Invalid_body"] = "Unable to get request"
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": http.StatusUnprocessableEntity,
				"error":  errList,
			})
			return
		}
	}

	report := models.Report{}
	err = json.Unmarshal(body, &report)
	if err != nil {
		errList["Unmarshal_error"] = "Cannot unmarshal body"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	//enter userid and meetingid to report struct
	report.UserId = user.UserId
	report.MeetingId = meeting.ID

	report.Prepare()
	errMsg := report.Validate("")
	if len(errMsg) > 0 {
		errList = errMsg
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status":  http.StatusUnprocessableEntity,
			"reponse": errList,
		})
		return
	}

	data, err := report.SaveReport(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		errList = formattedError
		c.JSON(http.StatusNotFound, gin.H{
			"status": http.StatusNotFound,
			"error":  errList,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"reponse": data,
	})
}

func (server *Server) GetReports(c *gin.Context) {
	errList = map[string]string{}

	//is valid meeting id given to server
	id := c.Param("id")
	mID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		errList["Invalid_request"] = "Invalid Request"
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"error":  errList,
		})
		return
	}

	//check if the meeting exist
	meeting := models.Meeting{}
	err = server.DB.Model(&models.Meeting{}).Where("id = ?", mID).Take(&meeting).Error
	if err != nil {
		errList["No_post"] = "No post found"
		c.JSON(http.StatusNotFound, gin.H{
			"status": http.StatusNotFound,
			"error":  errList,
		})
		return
	}

	report := models.Report{}
	data, err := report.GetReports(server.DB, meeting.ID)
	if err != nil {
		errList["No_comments"] = "No comments found"
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

func (server *Server) UpdateReport(c *gin.Context) {
	errList = map[string]string{}

	id := c.Param("id")
	rID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		errList["Invalid_request"] = "Invalid Request"
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"error":  errList,
		})
		return
	}

	//check if the auth token is valid and get the user id from it
	uID, err := auth.ExtractTokenUID(c.Request)
	if err != nil {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	//check if report exist
	oriReport := models.Report{}
	err = server.DB.Model(&models.Report{}).Where("id = ?", rID).Take(&oriReport).Error
	if err != nil {
		errList["No_comment"] = "No Comment Found"
		c.JSON(http.StatusNotFound, gin.H{
			"status": http.StatusNotFound,
			"error":  errList,
		})
		return
	}

	if uID != oriReport.UserId {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	//binding data
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errList["Invalid_body"] = "Unable to get request"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}
	//start processing the request data
	report := models.Report{}
	err = json.Unmarshal(body, &report)
	if err != nil {
		errList["Unmarshal_error"] = "Cannot unmarshal body"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	report.Prepare()
	errMsg := report.Validate("")
	if len(errMsg) > 0 {
		errList = errMsg
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	report.ID = oriReport.ID
	report.UserId = oriReport.UserId
	report.MeetingId = oriReport.MeetingId

	data, err := report.SaveReport(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		errList = formattedError
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"response": data,
	})
}

func (server *Server) DeleteReport(c *gin.Context) {
	id := c.Param("id")
	rID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		errList["Invalid_request"] = "invalid request"
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"error":  errList,
		})
		return
	}

	//is this user authenticated
	uID, err := auth.ExtractTokenUID(c.Request)
	if err != nil {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	//check if report exist
	report := models.Report{}
	err = server.DB.Model(&models.Report{}).Where("id = ?", rID).Take(&report).Error
	if err != nil {
		errList["No_post"] = "No Post Found"
		c.JSON(http.StatusNotFound, gin.H{
			"status": http.StatusNotFound,
			"error":  errList,
		})
		return
	}

	//is the authenticated user, the owner of this post?
	if uID != report.UserId {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	//if all conditions are met, delete the report
	_, err = report.DeleteReport(server.DB, uID)
	if err != nil {
		errList["Other_error"] = "Please try again later"
		c.JSON(http.StatusNotFound, gin.H{
			"status": http.StatusNotFound,
			"error":  errList,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"response": "Report deleted",
	})
}
