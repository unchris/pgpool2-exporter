package main

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	"github.com/sirupsen/logrus"
)

const (
	namespace    = "pgpool2"
	exporterName = "pgpool2_exporter"
)

var (
	PoolLastScrapeError = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_scrape_error"),
		"Whether the last scrape of metrics from Pgpool2 resulted in an error (1 for error, 0 for success)",
		nil, nil,
	)
	PoolLastScrapeDuration = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_scrape_duration_seconds"),
		"Duration of the last scrape of metrics from Pgpool2",
		nil, nil,
	)
	PoolNodeCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "node_count"),
		"Displays the total number of database nodes",
		nil, nil,
	)
	PoolNodeInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "node_info"),
		"Displays the information of node",
		[]string{"id", "node", "port", "width"}, nil,
	)
	PoolProcCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "proc_count"),
		"Displays number of all Pgpool-II children processes",
		nil, nil,
	)
	PoolNumberActiveConnections = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "frontend_active_connections"),
		"Displays number of all active connections to all Pgpool-II children processes",
		[]string{"database"}, nil,
	)
	PoolNumberInactiveConnections = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "frontend_inactive_connections"),
		"Displays number of all inactive connections to all Pgpool-II children processes",
		[]string{"database"}, nil,
	)
)

type Exporter struct {
	pgpool *PGPoolClient
}

func init() {
	prometheus.MustRegister(version.NewCollector(exporterName))
}

func NewExporter(pgpool *PGPoolClient) *Exporter {
	return &Exporter{
		pgpool: pgpool,
	}
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	var scrapeError bool
	defer func(begun time.Time) {
		ch <- prometheus.MustNewConstMetric(
			PoolLastScrapeDuration,
			prometheus.GaugeValue,
			time.Since(begun).Seconds(),
		)
	}(time.Now())
	nodeCount, err := e.pgpool.ExecNodeCount()
	if err != nil {
		scrapeError = true
		logrus.Errorf("ExecNodeCount error: %v", err)
		ch <- prometheus.MustNewConstMetric(
			PoolNodeCount,
			prometheus.GaugeValue,
			0.0,
		)
	} else {
		ch <- prometheus.MustNewConstMetric(
			PoolNodeCount,
			prometheus.GaugeValue,
			float64(nodeCount),
		)
		for i := 0; i < nodeCount; i++ {
			nodeInfo, err := e.pgpool.ExecNodeInfo(i)
			if err != nil {
				scrapeError = true
				logrus.Errorf("ExecNodeInfo error: %v", err)
				continue
			}
			ch <- prometheus.MustNewConstMetric(
				PoolNodeInfo,
				prometheus.GaugeValue,
				float64(nodeInfo.Status),
				strconv.Itoa(i),
				nodeInfo.Hostname,
				strconv.Itoa(nodeInfo.Port),
				strconv.FormatFloat(nodeInfo.Weight, 'f', 6, 64),
			)
		}
	}
	procArr, err := e.pgpool.ExecProcCount()
	if err != nil {
		scrapeError = true
		logrus.Errorf("ExecProcCount error: %v", err)
		ch <- prometheus.MustNewConstMetric(
			PoolProcCount,
			prometheus.GaugeValue,
			0.0,
		)
	} else {
		ch <- prometheus.MustNewConstMetric(
			PoolProcCount,
			prometheus.GaugeValue,
			float64(len(procArr)),
		)
	}
	procInfoArr, err := e.pgpool.ExecProcInfo()
	if err != nil {
		scrapeError = true
		logrus.Errorf("ExecProcInfo error: %v", err)
	} else {
		procSummary := e.pgpool.ProcInfoSummary(procInfoArr)
		for database, counter := range procSummary.Active {
			ch <- prometheus.MustNewConstMetric(
				PoolNumberActiveConnections,
				prometheus.GaugeValue,
				float64(counter),
				database,
			)
		}
		for database, counter := range procSummary.Inactive {
			ch <- prometheus.MustNewConstMetric(
				PoolNumberInactiveConnections,
				prometheus.GaugeValue,
				float64(counter),
				database,
			)
		}
	}
	scrapeErrorFloat := 0.0
	if scrapeError {
		scrapeErrorFloat = 1.0
	}
	ch <- prometheus.MustNewConstMetric(
		PoolLastScrapeError,
		prometheus.GaugeValue,
		scrapeErrorFloat,
	)
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- PoolLastScrapeError
	ch <- PoolLastScrapeDuration
	ch <- PoolNodeCount
	ch <- PoolProcCount
	ch <- PoolNodeInfo
	ch <- PoolNumberActiveConnections
	ch <- PoolNumberInactiveConnections
}
