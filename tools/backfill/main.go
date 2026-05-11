package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// Simple backfill tool: queries job_cables rows missing snapshots and calls
// WarehouseCore API to retrieve authoritative snapshots for each cable.

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL required")
	}
	rcBase := os.Getenv("WAREHOUSECORE_BASE_URL")
	if rcBase == "" {
		log.Fatal("WAREHOUSECORE_BASE_URL required")
	}
	apiKey := os.Getenv("WAREHOUSECORE_API_KEY")

	var limit int
	flag.IntVar(&limit, "limit", 100, "job_cables rows per run")
	flag.Parse()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT jobid, "cableID" FROM job_cables WHERE cable_snapshot IS NULL LIMIT $1`, limit)
	if err != nil {
		log.Fatalf("select job_cables: %v", err)
	}
	defer rows.Close()

	client := &http.Client{Timeout: 15 * time.Second}
	for rows.Next() {
		var (
			jobid   int
			cableID int
		)
		if err := rows.Scan(&jobid, &cableID); err != nil {
			log.Printf("scan: %v", err)
			continue
		}
		url := fmt.Sprintf("%s/admin/cables/%d", rcBase, cableID)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("request create: %v", err)
			continue
		}
		if apiKey != "" {
			req.Header.Set("X-API-Key", apiKey)
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("api call: %v", err)
			continue
		}
		func() {
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				log.Printf("api status %d for cable %d", resp.StatusCode, cableID)
				return
			}
			var snapshot map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
				log.Printf("decode: %v", err)
				return
			}
			bs, err := json.Marshal(snapshot)
			if err != nil {
				log.Printf("marshal: %v", err)
				return
			}
			if _, err := db.Exec(`UPDATE job_cables SET cable_snapshot = $1::jsonb WHERE jobid = $2 AND "cableID" = $3`, string(bs), jobid, cableID); err != nil {
				log.Printf("update job_cables %d/%d: %v", jobid, cableID, err)
				return
			}
			log.Printf("updated job_cables %d/%d", jobid, cableID)
		}()
	}
}
