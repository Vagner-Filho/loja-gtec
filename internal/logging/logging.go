package logging

import (
	"database/sql"
	"encoding/json"
	"log"
)

var db *sql.DB

func SetDatabase(database *sql.DB) {
	db = database
}

func LogError(source, errorType, message string, details interface{}) {
	if db == nil {
		log.Printf("Error logging unavailable - database not set: source=%s, type=%s, msg=%s", source, errorType, message)
		return
	}

	var detailsJSON []byte
	var err error
	if details != nil {
		detailsJSON, err = json.Marshal(details)
		if err != nil {
			log.Printf("Failed to marshal error details: %v", err)
			detailsJSON = []byte("{}")
		}
	} else {
		detailsJSON = []byte("{}")
	}

	_, err = db.Exec(`
		INSERT INTO error_logs (source, error_type, message, details)
		VALUES ($1, $2, $3, $4)
	`, source, errorType, message, detailsJSON)

	if err != nil {
		log.Printf("Failed to insert error log: %v", err)
	}
}
