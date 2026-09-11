package scheduler

import (
	"log"
	"time"

	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/models"
	"github.com/cecep-azhar/jurnalumi/internal/services"
	"github.com/robfig/cron/v3"
)

// InitScheduler starts the cron scheduler
func InitScheduler() *cron.Cron {
	c := cron.New(cron.WithLocation(time.Local))

	// QA-P1-14: Fetch gold price daily at 00:00
	_, err := c.AddFunc("0 0 * * *", fetchAndStoreGoldPrice)
	if err != nil {
		log.Printf("Failed to schedule fetchAndStoreGoldPrice: %v", err)
	}

	c.Start()
	log.Println("Scheduler started.")
	
	// Also run once on startup so we have a snapshot immediately
	go fetchAndStoreGoldPrice()

	return c
}

func fetchAndStoreGoldPrice() {
	log.Println("CRON: Fetching and storing gold price snapshot...")
	price := services.FetchLiveGoldPrice()
	
	snapshot := models.PriceSnapshot{
		CommodityType: "emas",
		PricePerGram:  price,
		Source:        "logammulia.com/fallback",
		SnapshotDate:  time.Now(),
	}
	
	if err := db.DB.Create(&snapshot).Error; err != nil {
		log.Printf("CRON: Failed to save price snapshot: %v", err)
	} else {
		log.Printf("CRON: Saved price snapshot: %d IDR/gram", price)
	}
}
