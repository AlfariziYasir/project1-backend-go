package controller

import (
	"net/http"
	"project1/api/auth"
	"project1/api/models"
	"project1/api/utils/formaterror"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (server *Server) JoinMeeting(c *gin.Context) {
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
		errList["Unauthorized"] = "unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"reponse": errList,
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

	join := models.Join{}
	join.UserId = user.UserId
	join.MeetingId = uint(meeting.ID)

	data, err := join.SaveJoin(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		errList = formattedError
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

func (server *Server) GetJoin(c *gin.Context) {
	errList = map[string]string{}

	id := c.Param("id")
	mID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		errList["Invalid_request"] = "invalid request"
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"reponse": errList,
		})
		return
	}

	//check if the meeting exist
	meeting := models.Meeting{}
	err = server.DB.Model(&models.Meeting{}).Where("id = ?", mID).Take(&meeting).Error
	if err != nil {
		errList["no_meeting"] = "no meeting"
		c.JSON(http.StatusNotFound, gin.H{
			"status":   http.StatusNotFound,
			"response": errList,
		})
		return
	}

	join := models.Join{}
	data, count, err := join.GetJoinInfo(server.DB, uint(mID))
	if err != nil {
		errList["No_join"] = "no join found"
		c.JSON(http.StatusNotFound, gin.H{
			"status":  http.StatusNotFound,
			"reponse": errList,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"total_user" : count,
		"response": data,
	})
}

func (server *Server) DeleteJoin(c *gin.Context) {
	id := c.Param("id")
	jID, err := strconv.ParseUint(id, 10, 64)
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

	//check if join exist
	join := models.Join{}
	err = server.DB.Model(&models.Join{}).Where("id = ?", jID).Take(&join).Error
	if err != nil {
		errList["No_like"] = "No Like Found"
		c.JSON(http.StatusNotFound, gin.H{
			"status": http.StatusNotFound,
			"error":  errList,
		})
		return
	}

	if uID != join.UserId {
		errList["Unauthorized"] = "Unauthorized"
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": http.StatusUnauthorized,
			"error":  errList,
		})
		return
	}

	_, err = join.DeleteJoin(server.DB)
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
		"response": "Like deleted",
	})
}
