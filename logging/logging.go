// Setup a Logging HTTP Handler to control logging level
package logging

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type LoggingResponse struct {
	Status   int    `json:"status"`
	Hostname string `json:"host"`
	Level    string `json:"level,omitempty"`
	Error    string `json:"error,omitempty"`
}

func LoggingHandler(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()

	if r.Method == http.MethodPost {
		newLevel := r.URL.Query().Get("level")
		level, err := log.ParseLevel(newLevel)
		if err != nil {
			writeResponse(w,
				http.StatusBadRequest,
				&LoggingResponse{
					Hostname: hostname,
					Error:    err.Error(),
				},
			)
			return
		}
		log.SetLevel(level)
	}

	writeResponse(w,
		http.StatusOK,
		&LoggingResponse{
			Hostname: hostname,
			Level:    log.StandardLogger().GetLevel().String(),
		},
	)

	return
}

func writeResponse(w http.ResponseWriter, statusCode int, resp *LoggingResponse) {
	resp.Status = statusCode
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	enc := json.NewEncoder(w)
	_ = enc.Encode(resp)
}
