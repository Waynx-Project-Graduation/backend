package main

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/kemit/trip-planner/internal/config"
	"github.com/kemit/trip-planner/internal/database"
	"github.com/kemit/trip-planner/internal/models"
)

// Seed populates the database with places data from the CSV file
func main() {
	cfg := config.Load()
	db := database.Connect(cfg)
	database.AutoMigrate(db)

	// Open the CSV file
	csvPath := "databaseFiles/kem_places (1).csv"
	file, err := os.Open(csvPath)
	if err != nil {
		log.Printf("⚠️  Failed to open CSV file %s: %v. Skipping seeding.", csvPath, err)
		return
	}
	defer file.Close()

	// Handle BOM (UTF-8 BOM: 0xEF 0xBB 0xBF)
	// Read first 3 bytes and check for BOM
	bom := make([]byte, 3)
	n, _ := file.Read(bom)
	if n < 3 || bom[0] != 0xEF || bom[1] != 0xBB || bom[2] != 0xBF {
		// No BOM found, seek back to start
		file.Seek(0, 0)
	}

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to parse CSV: %v", err)
	}

	if len(records) < 2 {
		log.Fatal("CSV file has no data rows")
	}

	// Skip header row (index 0)
	header := records[0]
	log.Printf("CSV columns: %v", header)
	// Expected: place_id, place_name, city, category, budget_level, best_season,
	//           crowd_level, suitable_for, suitable_age, duration_needed, rating, description

	seeded := 0
	skipped := 0

	for i, row := range records[1:] {
		if len(row) < 12 {
			log.Printf("⚠️  Row %d has insufficient columns (%d), skipping", i+2, len(row))
			continue
		}

		// Parse place_id
		placeID, err := strconv.ParseUint(strings.TrimSpace(row[0]), 10, 64)
		if err != nil {
			log.Printf("⚠️  Row %d: invalid place_id '%s', skipping", i+2, row[0])
			continue
		}

		// Parse duration_needed
		duration, _ := strconv.Atoi(strings.TrimSpace(row[9]))

		// Parse rating
		rating, _ := strconv.ParseFloat(strings.TrimSpace(row[10]), 64)

		// Convert pipe-delimited fields to comma-separated
		suitableFor := strings.ReplaceAll(strings.TrimSpace(row[7]), "|", ",")
		suitableAge := strings.ReplaceAll(strings.TrimSpace(row[8]), "|", ",")

		// Clean up \r from description (Windows line endings)
		description := strings.TrimSpace(row[11])
		description = strings.TrimRight(description, "\r")

		place := models.Place{
			ID:             uint(placeID),
			Name:           strings.TrimSpace(row[1]),
			City:           strings.TrimSpace(row[2]),
			Category:       strings.TrimSpace(row[3]),
			BudgetLevel:    strings.TrimSpace(row[4]),
			BestSeason:     strings.TrimSpace(row[5]),
			CrowdLevel:     strings.TrimSpace(row[6]),
			SuitableFor:    suitableFor,
			SuitableAge:    suitableAge,
			DurationNeeded: duration,
			Rating:         rating,
			Description:    description,
		}

		result := db.Where("id = ?", place.ID).FirstOrCreate(&place)
		if result.Error != nil {
			log.Printf("❌ Failed to seed place %d (%s): %v", place.ID, place.Name, result.Error)
		} else if result.RowsAffected > 0 {
			log.Printf("✅ Seeded: %d — %s (%s)", place.ID, place.Name, place.City)
			seeded++
		} else {
			log.Printf("⏭️  Skipped (exists): %d — %s (%s)", place.ID, place.Name, place.City)
			skipped++
		}
	}

	log.Printf("🌱 Database seeding completed! Seeded: %d, Skipped: %d", seeded, skipped)
}
