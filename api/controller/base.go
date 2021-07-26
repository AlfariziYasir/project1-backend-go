package controller

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"project1/api/middleware"
	"project1/api/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Server struct {
	DB     *gorm.DB
	Router *gin.Engine
}

var errList = make(map[string]string)

func (server *Server) Initialize(Dbdriver, DbUser, DbPassword, DbPort, DbHost, DbName string) {
	var err error

	DBURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", DbUser, DbPassword, DbHost, DbPort, DbName)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // Disable color
		},
	)

	server.DB, err = gorm.Open(mysql.Open(DBURL), &gorm.Config{
		Logger:                                   newLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		fmt.Printf("Cannot connect to %s database\n", Dbdriver)
		log.Fatal("This is the error:", err)
	} else {
		fmt.Printf("We are connected to the %s database", Dbdriver)
	}

	server.DB.Debug().Migrator().DropTable(
		&models.User{},
		&models.Meeting{},
		&models.ResetPassword{},
		&models.Join{},
		&models.Report{},
	)

	server.DB.Debug().AutoMigrate(
		&models.User{},
		&models.Meeting{},
		&models.ResetPassword{},
		&models.Join{},
		&models.Report{},
	)

	server.Router = gin.Default()
	server.Router.Use(middleware.CORSMiddleware())

	server.InitializeRoutes()
}

func (server *Server) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, server.Router))
}
