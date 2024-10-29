package controller

import (
	"fmt"
	"golang_training/models"
	"golang_training/module"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var countdownChannel = make(chan int) // Channel for countdown signals

// Handler to enqueue a new item
func EnqueueHandler(c *gin.Context) {
	var newItem models.Queue

	if err := c.BindJSON(&newItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if newItem.NameOfPax == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name_of_pax is required"})
		return
	}

	// Menambahkan position untuk new entry
	var lastItem models.Queue
	if result := module.DB.Order("queue_position desc").First(&lastItem); result.Error == nil {
		newItem.QueuePosition = lastItem.QueuePosition + 1
	} else {
		newItem.QueuePosition = 1
	}

	// Kalau posisi ke-1, langsung start hitung mundur
	if newItem.QueuePosition == 1 {
		newItem.Countdown = 300
		fmt.Println("Start queue for item: ", newItem.NameOfPax)
		go startCountdown(&newItem)
	} else {
		newItem.Countdown = 300
	}

	// Tambah entry ke db
	if result := module.DB.Create(&newItem); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Item enqueued", "item": newItem})
}

func startCountdown(item *models.Queue) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if item.Countdown > 0 {
				fmt.Println("countdown: ", item.ID, item.Countdown)
				item.Countdown -= 60 // update mengurangi 60 detik
				module.DB.Model(item).Update("countdown", item.Countdown)
			}
			// Cek ketika countdown sudah di 0, assign value ke channel
			if item.Countdown <= 0 {
				fmt.Println("countdown telah selesai, lanjut ke queue berikutnya")
				countdownChannel <- item.QueuePosition
				return
			}
		}
	}
}

func ProcessCountdowns() {
	for pos := range countdownChannel {
		DequeuePosition(pos)
	}
}

// Handler to dequeue and return the first item
func DequeueHandler(c *gin.Context) {
	var firstItem models.Queue

	if result := module.DB.Order("queue_position asc").First(&firstItem); result.Error != nil {
		c.JSON(http.StatusNoContent, gin.H{"message": "Queue is empty"})
		return
	}

	module.DB.Delete(&firstItem)

	var newFirstItem models.Queue
	if result := module.DB.Order("queue_position asc").First(&newFirstItem); result.Error == nil {
		newFirstItem.Countdown = 300
		module.DB.Model(&newFirstItem).Update("countdown", newFirstItem.Countdown)
		go startCountdown(&newFirstItem) // Start countdown for the new first item
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dequeued item", "item": firstItem})
}

// queue list
func ListQueueHandler(c *gin.Context) {
	var queue []models.Queue

	// select * from queue order by queue_position asc
	if result := module.DB.Order("queue_position asc").Find(&queue); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"queue": queue})
}

func DequeuePosition(position int) {
	var item models.Queue
	if result := module.DB.Where("queue_position = ?", position).First(&item); result.Error == nil {
		module.DB.Delete(&item)

		// reorder semua position
		var items []models.Queue
		module.DB.Where("queue_position > ?", position).Order("queue_position asc").Find(&items)

		for _, queueItem := range items {
			queueItem.QueuePosition -= 1
			module.DB.Save(&queueItem)
		}

		var newFirstItem models.Queue
		if result := module.DB.Order("queue_position asc").First(&newFirstItem); result.Error == nil {
			newFirstItem.Countdown = 300
			module.DB.Model(&newFirstItem).Update("countdown", newFirstItem.Countdown)
			go startCountdown(&newFirstItem)
		}
	}
}

func CheckStatusHandler(c *gin.Context) {
	nameOfPax := c.Param("name_of_pax")

	var item models.Queue
	if result := module.DB.Where("name_of_pax = ?", nameOfPax).First(&item); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found in queue"})
		return
	}

	// Return the remaining countdown time
	c.JSON(http.StatusOK, gin.H{
		"message":        "User found",
		"name_of_pax":    item.NameOfPax,
		"queue_position": item.QueuePosition,
		"time_remaining": item.Countdown, // Remaining countdown time in seconds
	})
}
