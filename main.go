// Copyright 2013 The Prometheus Authors
// Copyright 2018 eBay Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"net/http"
	"net/url"

	"github.com/howeyc/fsnotify"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/unicell/trafficserver_exporter/collector"

	"github.com/prometheus/statsd_exporter/pkg/mapper"
)

func serveHTTP(listenAddress, metricsEndpoint string) {
	http.Handle(metricsEndpoint, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Apache Traffic Server Exporter</title></head>
			<body>
			<h1>Apache Traffic Server Exporter</h1>
			<p><a href="` + metricsEndpoint + `">Metrics</a></p>
			</body>
			</html>`))
	})
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

func watchConfig(fileName string, mapper *mapper.MetricMapper) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	err = watcher.WatchFlags(fileName, fsnotify.FSN_MODIFY)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case ev := <-watcher.Event:
			log.Infof("Config file changed (%s), attempting reload", ev)
			err = mapper.InitFromFile(fileName)
			if err != nil {
				log.Errorln("Error reloading config:", err)
				configLoads.WithLabelValues("failure").Inc()
			} else {
				log.Infoln("Config reloaded successfully")
				configLoads.WithLabelValues("success").Inc()
			}
			// Re-add the file watcher since it can get lost on some changes. E.g.
			// saving a file with vim results in a RENAME-MODIFY-DELETE event
			// sequence, after which the newly written file is no longer watched.
			_ = watcher.WatchFlags(fileName, fsnotify.FSN_MODIFY)
		case err := <-watcher.Error:
			log.Errorln("Error watching config:", err)
		}
	}
}

func main() {
	var (
		listenAddress   = kingpin.Flag("web.listen-address", "The address on which to expose the web interface and generated Prometheus metrics.").Default(":9122").Short('l').String()
		metricsEndpoint = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		mappingConfig   = kingpin.Flag("trafficserver.mapping-config", "Metric mapping configuration file name.").Short('c').String()
		timeout         = kingpin.Flag("timeout", "Timeout waiting for http request").Default("5s").Short('t').Duration()
		endpoint        = kingpin.Flag("endpoint", "Endpoint to fetch trafficserver statistics from").Required().String()
	)

	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("trafficserver_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting Traffic Server -> Prometheus Exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())
	log.Infoln("Accepting Prometheus Requests on", *listenAddress)

	go serveHTTP(*listenAddress, *metricsEndpoint)

	events := make(chan Events, 1024)
	defer close(events)

	url, err := url.Parse(*endpoint)
	if err != nil {
		log.Fatal(err)
	}

	httpClient := &http.Client{
		Timeout: *timeout,
	}

	prometheus.MustRegister(version.NewCollector("trafficserver_exporter"))
	prometheus.MustRegister(collector.NewGlobalCollector(httpClient, url))

	mapper := &mapper.MetricMapper{MappingsCount: mappingsCount}
	if *mappingConfig != "" {
		err := mapper.InitFromFile(*mappingConfig)
		if err != nil {
			log.Fatal("Error loading config:", err)
		}
		go watchConfig(*mappingConfig, mapper)
	}
	exporter := NewExporter(mapper)
	exporter.Listen(events)
}
