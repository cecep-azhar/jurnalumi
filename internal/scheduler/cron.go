package scheduler

import (
	"log"
	"time"

	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/models"
	"github.com/cecep-azhar/jurnalumi/internal/services"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

var c *cron.Cron

func Start() {
	c = cron.New(cron.WithLocation(time.Local))

	// QA-P1-14: Fetch gold prices daily at 00:05
	_, err := c.AddFunc("5 0 * * *", fetchGoldPrices)
	if err != nil {
		log.Printf("[Cron] Failed to schedule gold snapshot: %v", err)
	}

	// QA-P1-21: Downgrade expired premium plans hourly
	_, err = c.AddFunc("0 * * * *", downgradeExpiredPlans)
	if err != nil {
		log.Printf("[Cron] Failed to schedule plan downgrade: %v", err)
	}

	c.Start()
	log.Println("[Cron] Scheduler started.")
}

func Stop() {
	if c != nil {
		c.Stop()
	}
}

func fetchGoldPrices() {
	log.Println("[Cron] Starting daily gold price snapshot...")
	
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var locked bool
		// Advisory lock scope is transaction since we use xact_lock
		tx.Raw("SELECT pg_try_advisory_xact_lock(12345)").Scan(&locked)
		if !locked {
			log.Println("[Cron] Could not acquire lock, skipping.")
			return nil
		}

		price, err := services.GetGoldPricePerGram()
		if err != nil {
			log.Printf("[Cron] Failed fetching gold price: %v", err)
			return err // rollback
		}

		// Insert snapshots for antam, perak, dinar
		
		snapshots := []models.PriceSnapshot{
			{CommodityType: "antam", PricePerGram: price, Source: "api", SnapshotDate: time.Now()},
			{CommodityType: "perak", PricePerGram: services.GetSilverPricePerGramFallback(), Source: "manual", SnapshotDate: time.Now()},
			{CommodityType: "dinar", PricePerGram: int64(float64(price) * 4.25), Source: "api", SnapshotDate: time.Now()},
		}

		for _, s := range snapshots {
			var existing models.PriceSnapshot
			res := tx.Where("DATE(snapshot_date) = CURRENT_DATE AND commodity_type = ?", s.CommodityType).First(&existing)
			if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
				return res.Error
			}
			if res.Error == gorm.ErrRecordNotFound {
				if err := tx.Create(&s).Error; err != nil {
					return err
				}
			} else {
				existing.PricePerGram = s.PricePerGram
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
			}
		}
		
		log.Println("[Cron] Successfully updated price snapshots.")
		return nil
	})

	if err != nil {
		log.Printf("[Cron] Snapshot transaction failed: %v", err)
	}
}

func downgradeExpiredPlans() {
	log.Println("[Cron] Starting to downgrade expired plans...")
	
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var locked bool
		// Advisory lock for downgrade task
		tx.Raw("SELECT pg_try_advisory_xact_lock(12346)").Scan(&locked)
		if !locked {
			log.Println("[Cron] Could not acquire lock for plan downgrade, skipping.")
			return nil
		}

		res := tx.Model(&models.Tenant{}).
			Where("plan = ? AND plan_expires_at < ?", "premium", time.Now()).
			Updates(map[string]interface{}{
				"plan":            "free",
				"plan_expires_at": nil,
			})

		if res.Error != nil {
			return res.Error
		}

		if res.RowsAffected > 0 {
			log.Printf("[Cron] Successfully downgraded %d expired plan(s) to free.", res.RowsAffected)
		} else {
			log.Println("[Cron] No expired plans found.")
		}
		
		return nil
	})

	if err != nil {
		log.Printf("[Cron] Plan downgrade transaction failed: %v", err)
	}
}
