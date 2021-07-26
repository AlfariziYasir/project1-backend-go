package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"project1/api/auth"
	"project1/api/models"
	"project1/api/utils/formaterror"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (server *Server) CreateMeeting(c *gin.Context) {
	errList = map[string]string{}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errList["Invalid_body"] = "Unable to get request"
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	meeting := models.Meeting{}
	err = json.Unmarshal(body, &meeting)
	if err != nil {
		errList["Unmarshal_error"] = "Cannot unmarshal body"
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	uID, err := auth.ExtractTokenUID(c.Request)
	if err != nil {
		errList["Unauthorized"] = "Unauthorized"
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	// check if user exist
	user := models.User{}
	err = server.DB.Debug().Model(&models.User{}).Where("user_id = ?", uID).Take(&user).Error
	if err != nil {
		errList["Unauthorized"] = "Unauthorized"
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	fmt.Println(uID)
	meeting.Prepare(uID)
	errMsg := meeting.Validate()
	if len(errMsg) > 0 {
		errList = errMsg
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
	}

	data, err := meeting.SaveMeeting(server.DB)
	if err != nil {
		errList := formaterror.FormatError(err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":   http.StatusCreated,
		"response": data,
	})
}

func (server *Server) GetMeetings(c *gin.Context) {
	meeting := models.Meeting{}

	data, err := meeting.GetMeetings(server.DB)
	if err != nil {
		errList["No_post"] = "No Post Found"
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

func (server *Server) GetMeeting(c *gin.Context) {
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

	meeting := models.Meeting{}
	data, err := meeting.GetMeeting(server.DB, uint(mID))
	if err != nil {
		errList["No_post"] = "No Post Found"
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

func (server *Server) UpdateMeeting(c *gin.Context) {
	errList = map[string]string{}

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

	uID, err := auth.ExtractTokenUID(c.Request)
	if err != nil {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	//check if the meeting exist
	oriMeeting := models.Meeting{}
	err = server.DB.Debug().Model(&models.Meeting{}).Where("id = ?", mID).Take(&oriMeeting).Error
	if err != nil {
		errList["No_post"] = "No Post Found"
		c.JSON(http.StatusNotFound, gin.H{
			"status": http.StatusNotFound,
			"error":  errList,
		})
		return
	}

	if uID != oriMeeting.UserId {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	// read the data meeting
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errList["Invalid_body"] = "Unable to get request"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	//start processing request data
	meeting := models.Meeting{}
	err = json.Unmarshal(body, &meeting)
	if err != nil {
		errList["Unmarshal_error"] = "Cannot unmarshal body"
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	meeting.ID = oriMeeting.ID
	meeting.UserId = oriMeeting.UserId

	meeting.Prepare(uID)
	errMsg := meeting.Validate()
	if len(errMsg) > 0 {
		errList = errMsg
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  errList,
		})
		return
	}

	data, err := meeting.UpdateMeeting(server.DB)
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

func (server *Server) DeleteMeeting(c *gin.Context) {
	id := c.Param("id")
	mID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		errList["Invalid_request"] = "invalid request"
		c.JSON(http.StatusBadRequest, gin.H{
			"status":   http.StatusBadRequest,
			"response": errList,
		})
	}

	// is this user authenticated
	uid, err := auth.ExtractTokenUID(c.Request)
	if err != nil {
		errList["Unauthorized"] = "unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":   http.StatusUnauthorized,
			"response": errList,
		})
	}

	// check if the meeting exist
	meeting := models.Meeting{}
	err = server.DB.Debug().Model(&models.Meeting{}).Where("id = ?", mID).Take(&meeting).Error
	if err != nil {
		errList["No_post"] = "no post found"
		c.JSON(http.StatusNotFound, gin.H{
			"status":   http.StatusNotFound,
			"response": errList,
		})
		return
	}

	if uid != meeting.UserId {
		errList["Unauthorized"] = "unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":   http.StatusUnauthorized,
			"response": errList,
		})
		return
	}

	_, err = meeting.DeleteMeeting(server.DB)
	if err != nil {
		errList["Other_error"] = "Please try again later"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	report := models.Report{}
	join := models.Join{}

	_, err = report.DeleteReport(server.DB, uid)
	if err != nil {
		errList["Other_error"] = "Please try again later"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	_, err = join.DeleteMeetingJoin(server.DB, uint(mID))
	if err != nil {
		errList["Other_error"] = "Please try again later"
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  errList,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"response": "meeting deleted",
	})
}

func (server *Server) GetUserMeetings(c *gin.Context) {
	id := c.Param("id")

	meeting := models.Meeting{}
	data, err := meeting.GetUserMeetings(server.DB, id)
	if err != nil {
		errList["Invalid_request"] = "Invalid Request"
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"error":  errList,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"response": data,
	})
}
