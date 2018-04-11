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
	PGPoolNodeCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "pcp_node_count"),
		"Displays the total number of database nodes",
		nil, nil,
	)
	PGPoolNodeInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "pcp_node_info"),
		"Displays the information of node",
		[]string{"id", "hostname", "port", "status", "width"}, nil,
	)
	PGPoolProcCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "pcp_proc_count"),
		"Displays count of all Pgpool-II children processes",
		nil, nil,
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
			PGPoolNodeCount,
			prometheus.GaugeValue,
			0.0,
		)
	} else {
		ch <- prometheus.MustNewConstMetric(
			PGPoolNodeCount,
			prometheus.GaugeValue,
			float64(nodeCount),
		)
		for i := 0; i < nodeCount; i++ {
			nodeInfo, err := e.pgpool.ExecNodeInfo(i)
			if err != nil {
				logrus.Error(err)
			} else {
				ch <- prometheus.MustNewConstMetric(
					PGPoolNodeInfo,
					prometheus.GaugeValue,
					1.0,
					strconv.Itoa(i),
					nodeInfo.Hostname,
					strconv.Itoa(nodeInfo.Port),
					strconv.Itoa(nodeInfo.Status),
					strconv.FormatFloat(nodeInfo.Weight, 'f', 6, 64),
				)
			}
		}
	}
	procArr, err := e.pgpool.ExecProcCount()
	if err != nil {
		logrus.Error(err)
		ch <- prometheus.MustNewConstMetric(
			PGPoolProcCount,
			prometheus.GaugeValue,
			0.0,
		)
	} else {
		ch <- prometheus.MustNewConstMetric(
			PGPoolProcCount,
			prometheus.GaugeValue,
			float64(len(procArr)),
		)
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- PGPoolNodeCount
	ch <- PGPoolProcCount
	ch <- PGPoolNodeInfo
}
