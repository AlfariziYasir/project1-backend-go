package middleware

import (
	"project1/api/auth"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

func Authorize(obj, act string, enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		//get current user/subject
		tokenID, err := auth.ExtractTokenUID(c.Request)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{
				"msg": "User hasn't logged in yet",
			})
			return
		}

		//load policy from database
		err = enforcer.LoadPolicy()
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"msg": "Failed to load policy from DB"})
			return
		}

		//casbin enforces policy
		ok, err := enforcer.Enforce(tokenID, obj, act)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"msg": "Error occurred when authorizing user"})
			return
		}

		if !ok {
			c.AbortWithStatusJSON(403, gin.H{"msg": "You are not authorized"})
			return
		}
		c.Next()
	}
}
