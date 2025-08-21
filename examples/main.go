package main

import (
	"log"
	"strconv"
	"time"

	"github.com/Caknoooo/go-pagination"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Database connection
	dsn := "root:root@tcp(localhost:3306)/sports_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate all models
	err = db.AutoMigrate(&Province{}, &Sport{}, &Event{}, &Athlete{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Seed data
	seedData(db)

	r := gin.Default()

	// Province endpoints
	r.GET("/provinces", func(c *gin.Context) {
		filter := &ProvinceFilter{}
		response := pagination.PaginatedAPIResponseWithCustomFilter[Province](
			db, c, filter, "Provinces retrieved successfully",
		)
		c.JSON(response.Code, response)
	})

	// Sport endpoints
	r.GET("/sports", func(c *gin.Context) {
		filter := &SportFilter{}
		response := pagination.PaginatedAPIResponseWithCustomFilter[Sport](
			db, c, filter, "Sports retrieved successfully",
		)
		c.JSON(response.Code, response)
	})

	// Event endpoints
	r.GET("/events", func(c *gin.Context) {
		filter := &EventFilter{}
		response := pagination.PaginatedAPIResponseWithCustomFilter[Event](
			db, c, filter, "Events retrieved successfully",
		)
		c.JSON(response.Code, response)
	})

	// Athlete endpoints
	r.GET("/athletes", func(c *gin.Context) {
		filter := &AthleteFilter{}
		response := pagination.PaginatedAPIResponseWithCustomFilter[Athlete](
			db, c, filter, "Athletes retrieved successfully",
		)
		c.JSON(response.Code, response)
	})

	// Athletes with relationships
	r.GET("/athletes/detailed", func(c *gin.Context) {
		filter := &AthleteFilter{}
		filter.BindPagination(c)
		c.ShouldBindQuery(filter)

		// Load relationships
		filter.Includes = []string{"Province", "Sport", "Event"}

		athletes, total, err := pagination.PaginatedQuery[Athlete](
			db, filter, filter.GetPagination(), filter.GetIncludes(),
		)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		paginationResponse := pagination.CalculatePagination(filter.GetPagination(), total)
		response := pagination.NewPaginatedResponse(200, "Detailed athletes retrieved successfully", athletes, paginationResponse)

		c.JSON(200, response)
	})

	// Athletes by province
	r.GET("/provinces/:id/athletes", func(c *gin.Context) {
		provinceID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid province ID"})
			return
		}

		filter := &AthleteFilter{
			ProvinceID: provinceID,
		}

		response := pagination.PaginatedAPIResponseWithCustomFilter[Athlete](
			db, c, filter, "Athletes from province retrieved successfully",
		)
		c.JSON(response.Code, response)
	})

	// Athletes by sport
	r.GET("/sports/:id/athletes", func(c *gin.Context) {
		sportID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid sport ID"})
			return
		}

		filter := &AthleteFilter{
			SportID: sportID,
		}

		response := pagination.PaginatedAPIResponseWithCustomFilter[Athlete](
			db, c, filter, "Athletes from sport retrieved successfully",
		)
		c.JSON(response.Code, response)
	})

	// Athletes by event
	r.GET("/events/:id/athletes", func(c *gin.Context) {
		eventID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		filter := &AthleteFilter{
			EventID: eventID,
		}

		response := pagination.PaginatedAPIResponseWithCustomFilter[Athlete](
			db, c, filter, "Athletes from event retrieved successfully",
		)
		c.JSON(response.Code, response)
	})

	log.Println("Server starting on :8080")
	log.Println("Available endpoints:")
	log.Println("GET /provinces - Filter: ?id=1&name=jakarta&code=JKT&search=name&page=1&per_page=10")
	log.Println("GET /sports - Filter: ?id=1&name=sepak&category=team&search=name&page=1&per_page=10")
	log.Println("GET /events - Filter: ?id=1&name=pon&location=jakarta&start_year=2024&search=name&page=1&per_page=10")
	log.Println("GET /athletes - Filter: ?id=1&province_id=1&sport_id=1&event_id=1&min_age=18&max_age=30&search=name&page=1&per_page=10")
	log.Println("GET /athletes/detailed - Same as athletes but with relationships loaded")
	log.Println("GET /provinces/:id/athletes - Athletes from specific province")
	log.Println("GET /sports/:id/athletes - Athletes from specific sport")
	log.Println("GET /events/:id/athletes - Athletes from specific event")

	r.Run(":8080")
}

func seedData(db *gorm.DB) {
	// Check if data already exists
	var count int64
	db.Model(&Province{}).Count(&count)
	if count > 0 {
		return
	}

	log.Println("Seeding database...")

	// Seed provinces
	provinces := []Province{
		{Name: "DKI Jakarta", Code: "JKT"},
		{Name: "Jawa Barat", Code: "JBR"},
		{Name: "Jawa Tengah", Code: "JTG"},
		{Name: "Jawa Timur", Code: "JTM"},
		{Name: "Bali", Code: "BAL"},
		{Name: "Sumatera Utara", Code: "SUT"},
		{Name: "Sumatera Barat", Code: "SBR"},
	}

	for _, province := range provinces {
		db.Create(&province)
	}

	// Seed sports
	sports := []Sport{
		{Name: "Sepak Bola", Category: "Team Sport", Description: "Olahraga tim dengan bola"},
		{Name: "Basket", Category: "Team Sport", Description: "Olahraga tim dengan keranjang"},
		{Name: "Voli", Category: "Team Sport", Description: "Olahraga tim dengan net"},
		{Name: "Badminton", Category: "Individual Sport", Description: "Olahraga individu dengan raket"},
		{Name: "Renang", Category: "Individual Sport", Description: "Olahraga air individu"},
		{Name: "Tenis", Category: "Individual Sport", Description: "Olahraga raket individu"},
		{Name: "Atletik", Category: "Individual Sport", Description: "Lari, lempar, lompat"},
	}

	for _, sport := range sports {
		db.Create(&sport)
	}

	// Seed events
	events := []Event{
		{
			Name:        "PON XXI Papua 2024",
			Description: "Pekan Olahraga Nasional XXI",
			StartDate:   time.Date(2024, 10, 15, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 10, 30, 0, 0, 0, 0, time.UTC),
			Location:    "Papua",
		},
		{
			Name:        "SEA Games 2023",
			Description: "Southeast Asian Games 2023",
			StartDate:   time.Date(2023, 5, 12, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2023, 5, 23, 0, 0, 0, 0, time.UTC),
			Location:    "Cambodia",
		},
		{
			Name:        "Asian Games 2022",
			Description: "Asian Games Hangzhou 2022",
			StartDate:   time.Date(2022, 9, 10, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2022, 9, 25, 0, 0, 0, 0, time.UTC),
			Location:    "Hangzhou",
		},
		{
			Name:        "Pekan Olahraga Daerah 2024",
			Description: "Kompetisi olahraga tingkat daerah",
			StartDate:   time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC),
			Location:    "Jakarta",
		},
	}

	for _, event := range events {
		db.Create(&event)
	}

	// Seed athletes
	athletes := []Athlete{
		{Name: "Budi Santoso", ProvinceID: 1, SportID: 1, EventID: 1, Age: 25},
		{Name: "Siti Nurhaliza", ProvinceID: 1, SportID: 2, EventID: 1, Age: 23},
		{Name: "Ahmad Subandrio", ProvinceID: 2, SportID: 1, EventID: 2, Age: 27},
		{Name: "Dewi Sartika", ProvinceID: 2, SportID: 3, EventID: 2, Age: 24},
		{Name: "Rudi Tabuti", ProvinceID: 3, SportID: 4, EventID: 3, Age: 26},
		{Name: "Maya Sari", ProvinceID: 3, SportID: 5, EventID: 3, Age: 22},
		{Name: "Andi Lala", ProvinceID: 4, SportID: 1, EventID: 4, Age: 28},
		{Name: "Rina Marlina", ProvinceID: 4, SportID: 2, EventID: 4, Age: 21},
		{Name: "Agus Salim", ProvinceID: 5, SportID: 3, EventID: 1, Age: 29},
		{Name: "Putri Indah", ProvinceID: 5, SportID: 4, EventID: 2, Age: 20},
		{Name: "Joko Widodo", ProvinceID: 6, SportID: 5, EventID: 3, Age: 30},
		{Name: "Sari Dewi", ProvinceID: 6, SportID: 6, EventID: 4, Age: 19},
		{Name: "Bambang Pamungkas", ProvinceID: 7, SportID: 1, EventID: 1, Age: 32},
		{Name: "Taufik Hidayat", ProvinceID: 7, SportID: 4, EventID: 2, Age: 33},
		{Name: "Liliyana Natsir", ProvinceID: 1, SportID: 4, EventID: 3, Age: 31},
	}

	for _, athlete := range athletes {
		db.Create(&athlete)
	}

	log.Println("Database seeded successfully!")
}
