package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
)

type response struct {
	Water int `json:"water"`
	Wind  int `json:"wind"`
}

var id int

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("[ERROR] Error loading .env file")
		panic(err)
	}

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dsn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := sql.Open(dbDriver, dsn)
	if err != nil {
		log.Fatal("[ERROR] failed connecting to database: ", err)
	}
	defer db.Close()

	dataChan := make(chan struct {
		water int
		wind  int
	})

	go func() {
		for {
			water := rand.Intn(100) + 1
			wind := rand.Intn(100) + 1
			dataChan <- struct {
				water int
				wind  int
			}{
				water: water,
				wind:  wind,
			}
			time.Sleep(15 * time.Second)
		}
	}()

	go func() {
		for {
			data := <-dataChan
			dataResponse := response{
				Water: data.water,
				Wind:  data.wind,
			}

			jsonData, err := json.MarshalIndent(dataResponse, "", "  ")
			if err != nil {
				log.Println("[ERROR] marshaling data to JSON : ", err)
				continue
			}

			waterStatus := getStatus(data.water, 5, 8)
			windStatus := getStatus(data.wind, 6, 15)
			fmt.Println(string(jsonData))
			fmt.Printf("status water : %s \n", waterStatus)
			fmt.Printf("status wind : %s \n", windStatus)

			resultQuery, errQuery := db.Query("SELECT id FROM weather LIMIT 1")
			if errQuery != nil {
				log.Printf("[ERROR] failed fetch data : %v\n", err)
			}

			if !resultQuery.Next() {
				_, errExec := db.Exec("INSERT INTO weather (water, wind) VALUES ($1, $2)", data.water, data.wind)
				if errExec != nil {
					log.Printf("[ERROR] failed insert data : %v\n", err)
				}
			} else {
				err := resultQuery.Scan(&id)
				if err != nil {
					log.Printf("[ERROR] failed fetch data : %v\n", err)
				}

				_, errExec := db.Exec("UPDATE weather SET water=$1, wind=$2 WHERE id=$3", data.water, data.wind, id)
				if errExec != nil {
					log.Printf("[ERROR] failed updated data : %v\n", err)
				}
			}
		}
	}()

	select {}
}

func getStatus(value, low, high int) string {
	if value < low {
		return "aman"
	} else if value >= low && value <= high {
		return "siaga"
	} else {
		return "bahaya"
	}
}
