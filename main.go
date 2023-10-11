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
	eventCountGauge = promauto.NewGaugeVec(
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

	connectionEstablishedGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "strfry_open_connections",
			Help: "The total number of open connections",
		},
	)
)

func scrape() {
	go func() {
		for {
			fetchDbSize()
			fetchEventCount()
			fetchConnectionEstablishedCount()
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

func fetchEventCount() {
	go func() {
		kinds := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "40", "41", "42", "43", "44", "1984", "9735", "10000", "10001", "10002", "30000", "30001", "30008", "30009", "30023"}
		for _, kind := range kinds {
			searchOpts := fmt.Sprintf(`{"kinds": [%s]}`, kind)
			out, err := exec.Command("/app/strfry", "scan", "--count", searchOpts).Output()
			if err != nil {
				fmt.Println(string(out))
				fmt.Println(err)
				continue
			}
			kindCount, _ := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
			eventCountGauge.With(prometheus.Labels{"kind": kind}).Set(float64(kindCount))
		}

	}()
}

func countRune(s string, r rune) int {
	count := 0
	for _, c := range s {
		if c == r {
			count++
		}
	}
	return count
}

func fetchConnectionEstablishedCount() {
	go func() {
		out, err := exec.Command("/usr/bin/lsof", "-ni:7777", "-sTCP:ESTABLISHED").Output()
		if err != nil {
			fmt.Println(string(out))
			fmt.Println(err)
			return
		}
		connectionCount := countRune(string(out), '\n')
		connectionEstablishedGauge.Set(float64(connectionCount))
	}()
}

func main() {
	scrape()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}
