package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	eventGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "strfry_event_total",
			Help: "The total number of handled events",
		},
		[]string{"kind"},
	)

	dbSizeGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "strfry_dbsize_total",
			Help: "strfry DB size",
		},
	)
)

func scrape() {
	go func() {
		for {
			fetchDbSize()
			fetchEvent()
			time.Sleep(30 * time.Second)
		}
	}()
}

func fetchDbSize() {
	go func() {
		dbFilePath := "/app/strfry-db/data.mdb"

		fileInfo, statErr := os.Stat(dbFilePath)

		if statErr != nil {
			fmt.Println(statErr)
			return
		}

		dbSizeGauge.Set(float64(fileInfo.Size()))
	}()
}

func fetchEvent() {
	go func() {
		kinds := []string{"0", "1", "2", "3", "4", "5", "6", "7", "40", "41", "42", "43", "44", "1984", "9735", "10000", "10001", "30000", "30001"}
		for _, kind := range kinds {
			searchOpts := fmt.Sprintf(`{"kinds": [%s]}`, kind)
			out, err := exec.Command("/app/strfry", "scan", "--count", searchOpts).Output()
			if err != nil {
				fmt.Println(string(out))
				fmt.Println(err)
				continue
			}
			kindCount, _ := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
			eventGauge.With(prometheus.Labels{"kind": kind}).Set(float64(kindCount))
		}

	}()
}

func main() {
	scrape()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}
