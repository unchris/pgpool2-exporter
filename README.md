# Pgpool2 Prometheus Exporter

## Arguments

* `web.telemetry-path` – Path under which to expose metrics
* `web.listen-address` – Address on which to expose metrics and web interface
* `pcp.passfile` – Path to the PCP password file containing hostname:port:username:password
* `pcp.host` – PCP hostname
* `pcp.port` – PCP port
* `pcp.username` – PCP username
* `pcp.password` – PCP password

## Metrics

* `pgpool2_last_scrape_error`
* `pgpool2_last_scrape_duration_seconds`
* `pgpool2_node_count`
* `pgpool2_node_info`
* `pgpool2_proc_count`
* `pgpool2_frontend_active_connections`
* `pgpool2_frontend_inactive_connections`
* `pgpool2_watchdog_nodes_total`
* `pgpool2_watchdog_nodes_remote`
* `pgpool2_watchdog_nodes_alive_remote`
* `pgpool2_watchdog_vip`
* `pgpool2_watchdog_quorum_state`
