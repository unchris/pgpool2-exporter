package main

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	"github.com/sirupsen/logrus"
)

const (
	namespace    = "pgpool2"
	exporterName = "pgpool2_exporter"
)

var (
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
	nodeCount, err := e.pgpool.ExecNodeCount()
	if err != nil {
		logrus.Error(err)
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
				logrus.Error(err)
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
		logrus.Error(err)
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
		logrus.Error(err)
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
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- PoolNodeCount
	ch <- PoolProcCount
	ch <- PoolNodeInfo
	ch <- PoolNumberActiveConnections
	ch <- PoolNumberInactiveConnections
}
