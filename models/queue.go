package models

import "gorm.io/gorm"

type Queue struct {
	gorm.Model
	NameOfPax     string `json:"name_of_pax"`
	QueuePosition int    `json:"queue_position"`
	Countdown     int    `json:"countdown"`
}
