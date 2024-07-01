package cacher

import (
	"database/sql"
	"net/url"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/logger"
)

type CacheTime int

var cacheTime CacheTime

func init() {
	var err error
	common.DB, err = sql.Open("sqlite3", "./cache.db")
	if err != nil {
		panic(err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS cache (
		endpoint TEXT,
		module TEXT,
		last_scanned INTEGER,
		PRIMARY KEY (endpoint, module)
	);`
	if _, err := common.DB.Exec(createTable); err != nil {
		panic(err)
	}

	cacheTime = 15
}

func SetCacheTime(time int) {
	cacheTime = CacheTime(time)
}

func cleanURL(endpoint string) string {
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}
	parsedURL.Fragment = ""
	cleanedURL := strings.TrimRight(parsedURL.String(), "/")
	return cleanedURL
}

func CheckCache(endpoint, module string) bool {
	cleaned := cleanURL(endpoint)
	if CanScan(cleaned, module) {
		logger.Debugf("Scanning %s for %s\n", endpoint, module)
		UpdateScanTime(cleaned, module)
		return true
	} else {
		logger.Debugf("Skipping %s for %s\n", endpoint, module)
		return false
	}

}

func UpdateScanTime(endpoint, module string) {
	currentTime := time.Now().Unix()

	tx, err := common.DB.Begin()
	if err != nil {
		logger.Errorf("Error starting transaction: %v\n", err)
		return
	}

	query := `REPLACE INTO cache (endpoint, module, last_scanned) VALUES (?, ?, ?);`
	_, err = tx.Exec(query, endpoint, module, currentTime)
	if err != nil {
		tx.Rollback()
		logger.Errorf("Error updating cache: %v\n", err)
		return
	}

	if err = tx.Commit(); err != nil {
		logger.Errorf("Error committing transaction: %v\n", err)
	}
}

func CanScan(endpoint, module string) bool {
	var LastScanned int64
	currentTime := time.Now().Unix()

	query := `SELECT last_scanned FROM cache WHERE endpoint = ? AND module = ?;`
	err := common.DB.QueryRow(query, endpoint, module).Scan(&LastScanned)
	if err != nil && err != sql.ErrNoRows {
		logger.Errorf("Error querying cache: %v\n", err)
		return true
	}

	if err == sql.ErrNoRows {
		return true
	}

	return currentTime-LastScanned > int64(cacheTime)*60

}

func GetFirstPath(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		logger.Debugf("Error parsing URL: %v", err)
		return "", err
	}

	hostname := parsedURL.Hostname()
	pathSegments := strings.Split(parsedURL.Path, "/")
	var fistPathSegment string
	if len(pathSegments) > 1 {
		fistPathSegment = pathSegments[1]
	}
	return hostname + "/" + fistPathSegment, nil
}
