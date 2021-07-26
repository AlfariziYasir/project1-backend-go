package controller

import (
	"fmt"
	"project1/api/middleware"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
)

func (s *Server) InitializeRoutes() {
	//initialize casbin adapter
	adapter, err := gormadapter.NewAdapterByDB(s.DB)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize casbin adapter: %v", err))
	}

	//load model configuration file and policy store adapter
	enforcer, err := casbin.NewEnforcer("config/rbac_model.conf", adapter)
	if err != nil {
		panic(fmt.Sprintf("failed to create casbin enforcer: %v", err))
	}

	//add policy
	if hasPolicy := enforcer.HasPolicy("admin", "report", "read"); !hasPolicy {
		enforcer.AddPolicy("admin", "report", "read")
	}
	if hasPolicy := enforcer.HasPolicy("admin", "report", "write"); !hasPolicy {
		enforcer.AddPolicy("admin", "report", "write")
	}
	if hasPolicy := enforcer.HasPolicy("student", "report", "read"); !hasPolicy {
		enforcer.AddPolicy("student", "report", "read")
	}

	v1 := s.Router.Group("/api")
	{
		//login route
		v1.POST("/login", s.Login)
		v1.POST("/register", s.CreateUser(enforcer))
	}

	v2 := v1.Group("/users", middleware.TokenMiddleware())
	{
		v2.GET("/", middleware.Authorize("report", "write", enforcer), s.GetUsers)
		v2.GET("/:id", middleware.Authorize("report", "write", enforcer), s.GetUser)
		v2.PUT("/:id", middleware.Authorize("report", "write", enforcer), s.UpdateUser)
		v2.DELETE("/:id", middleware.Authorize("report", "write", enforcer), s.DeleteUser)
	}

	v3 := v1.Group("/meetings", middleware.TokenMiddleware())
	{
		v3.POST("/", middleware.Authorize("report", "read", enforcer), s.CreateMeeting)
		v3.GET("/", middleware.Authorize("report", "read", enforcer), s.GetMeetings)
		v3.GET("/:id", middleware.Authorize("report", "read", enforcer), s.GetMeeting)
		v3.PUT("/:id", middleware.Authorize("report", "write", enforcer), s.UpdateMeeting)
		v3.DELETE("/:id", middleware.Authorize("report", "write", enforcer), s.DeleteMeeting)
	}

	v4 := v1.Group("/joins", middleware.TokenMiddleware())
	{
		v4.GET("/:id", middleware.Authorize("report", "read", enforcer), s.GetJoin)
		v4.POST("/:id", middleware.Authorize("report", "read", enforcer), s.JoinMeeting)
		v4.DELETE("/:id", middleware.Authorize("report", "read", enforcer), s.DeleteJoin)
	}

	v5 := v1.Group("/reports", middleware.TokenMiddleware())
	{
		v5.POST("/:id", middleware.Authorize("report", "write", enforcer), s.SaveReport)
		v5.GET("s/:id", middleware.Authorize("report", "write", enforcer), s.GetReports)
		v5.PUT("/:id", middleware.Authorize("report", "write", enforcer), s.UpdateReport)
		v5.DELETE("/:id", middleware.Authorize("report", "write", enforcer), s.DeleteReport)
	}
}
