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
	dsn := "root:Cakno1234@tcp(localhost:3306)/sports_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = db.AutoMigrate(&Province{}, &Sport{}, &Event{}, &Athlete{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	seedData(db)

	r := gin.Default()

	r.GET("/provinces", func(c *gin.Context) {
		filter := &ProvinceFilter{}
		response := pagination.PaginatedAPIResponseWithCustomFilter[Province](
			db, c, filter, "Provinces retrieved successfully",
		)
		c.JSON(response.Code, response)
	})

	r.GET("/provinces/with-athletes", func(c *gin.Context) {
		filter := &ProvinceFilter{}
		filter.BindPagination(c)
		c.ShouldBindQuery(filter)

		provinces, total, err := pagination.PaginatedQueryWithIncludable[Province](db, filter)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		paginationResponse := pagination.CalculatePagination(filter.GetPagination(), total)
		response := pagination.NewPaginatedResponse(200, "Provinces with athletes retrieved successfully", provinces, paginationResponse)

		c.JSON(200, response)
	})

	r.GET("/sports", func(c *gin.Context) {
		filter := &SportFilter{}
		response := pagination.PaginatedAPIResponseWithCustomFilter[Sport](
			db, c, filter, "Sports retrieved successfully",
		)
		c.JSON(response.Code, response)
	})

	r.GET("/sports/with-relations", func(c *gin.Context) {
		filter := &SportFilter{}
		filter.BindPagination(c)
		c.ShouldBindQuery(filter)

		sports, total, err := pagination.PaginatedQueryWithIncludable[Sport](db, filter)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		paginationResponse := pagination.CalculatePagination(filter.GetPagination(), total)
		response := pagination.NewPaginatedResponse(200, "Sports with relations retrieved successfully", sports, paginationResponse)

		c.JSON(200, response)
	})

	r.GET("/events", func(c *gin.Context) {
		filter := &EventFilter{}
		response := pagination.PaginatedAPIResponseWithCustomFilter[Event](
			db, c, filter, "Events retrieved successfully",
		)
		c.JSON(response.Code, response)
	})

	r.GET("/events/with-sport", func(c *gin.Context) {
		filter := &EventFilter{}
		filter.BindPagination(c)
		c.ShouldBindQuery(filter)

		events, total, err := pagination.PaginatedQueryWithIncludable[Event](db, filter)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		paginationResponse := pagination.CalculatePagination(filter.GetPagination(), total)
		response := pagination.NewPaginatedResponse(200, "Events with sport retrieved successfully", events, paginationResponse)

		c.JSON(200, response)
	})

	r.GET("/athletes", func(c *gin.Context) {
		filter := &AthleteFilter{}
		response := pagination.PaginatedAPIResponseWithCustomFilter[Athlete](
			db, c, filter, "Athletes retrieved successfully",
		)
		c.JSON(response.Code, response)
	})

	r.GET("/athletes/with-includes", func(c *gin.Context) {
		filter := &AthleteFilter{}
		filter.BindPagination(c)
		c.ShouldBindQuery(filter)

		athletes, total, err := pagination.PaginatedQueryWithIncludable[Athlete](db, filter)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		paginationResponse := pagination.CalculatePagination(filter.GetPagination(), total)
		response := pagination.NewPaginatedResponse(200, "Athletes with includes retrieved successfully", athletes, paginationResponse)

		c.JSON(200, response)
	})

	r.GET("/athletes/detailed", func(c *gin.Context) {
		filter := &AthleteFilter{}
		filter.BindPagination(c)
		c.ShouldBindQuery(filter)

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
	log.Println("GET /provinces - Filter: ?id=1&name=jakarta&code=JKT&search=jakarta&page=1&per_page=10")
	log.Println("GET /provinces/with-athletes - Same as above but with ?includes=Athletes")
	log.Println("GET /sports - Filter: ?id=1&name=sepak&category=team&search=sepak&page=1&per_page=10")
	log.Println("GET /sports/with-relations - Same as above but with ?includes=Athletes,Events")
	log.Println("GET /events - Filter: ?id=1&name=pon&location=jakarta&start_year=2024&search=pon&page=1&per_page=10")
	log.Println("GET /events/with-sport - Same as above but with ?includes=Sport")
	log.Println("GET /athletes - Filter: ?id=1&province_id=1&sport_id=1&event_id=1&min_age=18&max_age=30&search=name&page=1&per_page=10")
	log.Println("GET /athletes/with-includes - Same as above but with ?includes=Province,Sport,PlayersEvents")
	log.Println("GET /athletes/detailed - Same as athletes but with relationships loaded")
	log.Println("GET /provinces/:id/athletes - Athletes from specific province")
	log.Println("GET /sports/:id/athletes - Athletes from specific sport")
	log.Println("GET /events/:id/athletes - Athletes from specific event")

	r.Run(":8080")
}

func seedData(db *gorm.DB) {
	var count int64
	db.Model(&Province{}).Count(&count)
	if count > 0 {
		return
	}

	log.Println("Seeding database...")

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

	events := []Event{
		{
			Name:        "PON XXI Papua 2024",
			Description: "Pekan Olahraga Nasional XXI",
			StartDate:   time.Date(2024, 10, 15, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2024, 10, 30, 0, 0, 0, 0, time.UTC),
			Location:    "Papua",
			SportID:     1,
		},
		{
			Name:        "SEA Games 2023",
			Description: "Southeast Asian Games 2023",
			StartDate:   time.Date(2023, 5, 12, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2023, 5, 23, 0, 0, 0, 0, time.UTC),
			Location:    "Cambodia",
			SportID:     2,
		},
		{
			Name:        "Asian Games 2022",
			Description: "Asian Games Hangzhou 2022",
			StartDate:   time.Date(2022, 9, 10, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2022, 9, 25, 0, 0, 0, 0, time.UTC),
			Location:    "Hangzhou",
			SportID:     3,
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

	athletes := []Athlete{
		{Name: "Budi Santoso", ProvinceID: 1, SportID: 1, Age: 25, Gender: "Male", BirthDate: "1998-01-15", Height: 175, Image: "budi.jpg"},
		{Name: "Siti Nurhaliza", ProvinceID: 1, SportID: 2, Age: 23, Gender: "Female", BirthDate: "2000-05-20", Height: 165, Image: "siti.jpg"},
		{Name: "Ahmad Subandrio", ProvinceID: 2, SportID: 1, Age: 27, Gender: "Male", BirthDate: "1996-08-10", Height: 180, Image: "ahmad.jpg"},
		{Name: "Dewi Sartika", ProvinceID: 2, SportID: 3, Age: 24, Gender: "Female", BirthDate: "1999-12-05", Height: 168, Image: "dewi.jpg"},
		{Name: "Rudi Tabuti", ProvinceID: 3, SportID: 4, Age: 26, Gender: "Male", BirthDate: "1997-03-22", Height: 172, Image: "rudi.jpg"},
		{Name: "Maya Sari", ProvinceID: 3, SportID: 5, Age: 22, Gender: "Female", BirthDate: "2001-07-18", Height: 160, Image: "maya.jpg"},
		{Name: "Andi Lala", ProvinceID: 4, SportID: 1, Age: 28, Gender: "Male", BirthDate: "1995-11-30", Height: 178, Image: "andi.jpg"},
		{Name: "Rina Marlina", ProvinceID: 4, SportID: 2, Age: 21, Gender: "Female", BirthDate: "2002-04-14", Height: 163, Image: "rina.jpg"},
		{Name: "Agus Salim", ProvinceID: 5, SportID: 3, Age: 29, Gender: "Male", BirthDate: "1994-09-25", Height: 185, Image: "agus.jpg"},
		{Name: "Putri Indah", ProvinceID: 5, SportID: 4, Age: 20, Gender: "Female", BirthDate: "2003-02-08", Height: 158, Image: "putri.jpg"},
		{Name: "Joko Widodo", ProvinceID: 6, SportID: 5, Age: 30, Gender: "Male", BirthDate: "1993-06-12", Height: 170, Image: "joko.jpg"},
		{Name: "Sari Dewi", ProvinceID: 6, SportID: 6, Age: 19, Gender: "Female", BirthDate: "2004-10-03", Height: 155, Image: "sari.jpg"},
		{Name: "Bambang Pamungkas", ProvinceID: 7, SportID: 1, Age: 32, Gender: "Male", BirthDate: "1991-12-01", Height: 182, Image: "bambang.jpg"},
		{Name: "Taufik Hidayat", ProvinceID: 7, SportID: 4, Age: 33, Gender: "Male", BirthDate: "1990-08-16", Height: 176, Image: "taufik.jpg"},
		{Name: "Liliyana Natsir", ProvinceID: 1, SportID: 4, Age: 31, Gender: "Female", BirthDate: "1992-05-09", Height: 162, Image: "liliyana.jpg"},
	}

	for _, athlete := range athletes {
		db.Create(&athlete)
	}

	playersEvents := []PlayersEvents{
		{PlayerID: 1, PlayerType: "athlete", EventID: 1},
		{PlayerID: 2, PlayerType: "athlete", EventID: 1},
		{PlayerID: 3, PlayerType: "athlete", EventID: 2},
		{PlayerID: 4, PlayerType: "athlete", EventID: 2},
		{PlayerID: 5, PlayerType: "athlete", EventID: 3},
		{PlayerID: 6, PlayerType: "athlete", EventID: 3},
		{PlayerID: 7, PlayerType: "athlete", EventID: 4},
		{PlayerID: 8, PlayerType: "athlete", EventID: 4},
		{PlayerID: 9, PlayerType: "athlete", EventID: 1},
		{PlayerID: 10, PlayerType: "athlete", EventID: 2},
	}

	for _, playerEvent := range playersEvents {
		db.Create(&playerEvent)
	}

	log.Println("Database seeded successfully!")
}
