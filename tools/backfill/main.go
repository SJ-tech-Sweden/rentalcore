package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

// Simple backfill tool: queries jobs needing denorm data and calls WarehouseCore API
// to retrieve authoritative snapshots and updates the jobs table.

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
	flag.IntVar(&limit, "limit", 100, "jobs per run")
	flag.Parse()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT jobid FROM jobs WHERE cable_snapshot IS NULL LIMIT $1", limit)
	if err != nil {
		log.Fatalf("select jobs: %v", err)
	}
	defer rows.Close()

	client := &http.Client{}
	for rows.Next() {
		var jobid int
		if err := rows.Scan(&jobid); err != nil {
			log.Printf("scan: %v", err)
			continue
		}
		url := fmt.Sprintf("%s/admin/jobs/%d/denorm", rcBase, jobid)
		req, _ := http.NewRequest("GET", url, nil)
		if apiKey != "" {
			req.Header.Set("X-API-Key", apiKey)
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("api call: %v", err)
			continue
		}
		if resp.StatusCode != 200 {
			log.Printf("api status %d for job %d", resp.StatusCode, jobid)
			resp.Body.Close()
			continue
		}
		var snapshot map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
			log.Printf("decode: %v", err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()
		bs, _ := json.Marshal(snapshot)
		if _, err := db.Exec("UPDATE jobs SET cable_snapshot = $1, updated_at = NOW() WHERE jobid = $2", string(bs), jobid); err != nil {
			log.Printf("update job %d: %v", jobid, err)
			continue
		}
		log.Printf("updated job %d", jobid)
	}
}
