package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/sirupsen/logrus"
)

var (
	showVersion             = flag.Bool("version", false, "Prints version information and exit")
	metricsPath             = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	listenAddress           = flag.String("web.listen-address", ":9190", "Address on which to expose metrics and web interface.")
	pgpoolHostname          = flag.String("pgpool.host", "127.0.0.1", "PgPool2 hostname")
	pgpoolPort              = flag.Int("pgpool.port", 9898, "PgPool2 port")
	pgpoolUsername          = flag.String("pgpool.username", "pcpadmin", "PgPool2 username")
	pgpoolPassword          = flag.String("pgpool.password", "", "PgPool2 password")
	pgpoolConnectionTimeout = flag.Int("pgpool.timeout", 10, "PgPool2 connection timeout in seconds")
)

func versionInfo() {
	fmt.Println(version.Print(exporterName))
	os.Exit(0)
}

func main() {
	flag.Parse()

	if *showVersion == true {
		versionInfo()
	}

	logrus.Infof("Starting %s %s...", exporterName, version.Version)

	pgpool := &PGPoolClient{
		Hostname:         *pgpoolHostname,
		Port:             *pgpoolPort,
		Username:         *pgpoolUsername,
		Password:         *pgpoolPassword,
		TimeoutInSeconds: *pgpoolConnectionTimeout,
	}
	err := pgpool.Validate()
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	exporter := NewExporter(pgpool)
	if err := prometheus.Register(exporter); err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("Listen address: %s", *listenAddress)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>` + exporterName + ` v` + version.Version + `</title></head>
			<body>
			<h1>` + exporterName + ` v` + version.Version + `</h1>
			<p><a href='` + *metricsPath + `'>Metrics</a></p>
			</body>
			</html>
		`))
	})
	logrus.Fatal(http.ListenAndServe(*listenAddress, nil))
}
