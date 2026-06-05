package main

import (
	"log"
	"strings"

	"github.com/kemit/trip-planner/internal/config"
	"github.com/kemit/trip-planner/internal/database"
	"github.com/kemit/trip-planner/internal/models"
)

func main() {
	cfg := config.Load()
	db := database.Connect(cfg)

	urls := []string{
		// ... add your Cloudinary URLs here ...
	}

	for _, u := range urls {
		parts := strings.Split(u, "/")
		filename := parts[len(parts)-1]
		
		namePart := strings.TrimSuffix(filename, ".jpg")
		
		nameParts := strings.Split(namePart, "_")
		if len(nameParts) > 1 {
			placeNameWords := nameParts[:len(nameParts)-1]
			
			placeName := strings.Join(placeNameWords, " ")
			placeName = strings.ReplaceAll(placeName, " s ", "'s ")
			
			// Exact overrides based on typical data differences
			if placeName == "Hanging Church Cairo" {
				placeName = "Hanging Church (Cairo)"
			}
			if placeName == "Mohamed Ali Mosque Citadel" {
				placeName = "Mohamed Ali Mosque (Citadel)"
			}
			
			var place models.Place
			res := db.Where("name = ?", placeName).First(&place)
			if res.Error == nil {
				place.ThumbnailURL = u
				db.Save(&place)
				log.Printf("✅ Updated %s with URL", place.Name)
			} else {
				// try with fuzzy LIKE
				likeQuery := "%" + strings.Join(placeNameWords, "%") + "%"
				res2 := db.Where("name LIKE ?", likeQuery).First(&place)
				if res2.Error == nil {
					place.ThumbnailURL = u
					db.Save(&place)
					log.Printf("✅ Updated (fuzzy match) %s with URL", place.Name)
				} else {
					log.Printf("❌ Could not find place %s (tried fuzzy: %s)", placeName, likeQuery)
				}
			}
		}
	}
}
