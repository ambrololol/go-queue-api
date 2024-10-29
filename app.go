package main

import (
	"golang_training/controller"
	"golang_training/module"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// init db
	module.InitDB()

	// goroutines countdown
	go controller.ProcessCountdowns()

	r := gin.Default()

	// Routes
	r.POST("/register", controller.RegisterHandler)
	r.POST("/login", controller.LoginHandler)
	r.POST("/enqueue", controller.EnqueueHandler)

	protected := r.Group("/queue")
	protected.Use(module.JWTAuthMiddleware())
	{
		protected.GET("/dequeue", controller.DequeueHandler)
		protected.GET("/list", controller.ListQueueHandler)
		protected.GET("/status/:name_of_pax", controller.CheckStatusHandler)
	}

	// Start the server
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
