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

package collector

import (
	"net/http"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "trafficserver"
)

type GlobalCollector struct {
	client *http.Client
	url    *url.URL

	up                              prometheus.Gauge
	totalScrapes, jsonParseFailures prometheus.Counter
}

func NewGlobalCollector(client *http.Client, url *url.URL) *GlobalCollector {
	subsystem := ""
	return &GlobalCollector{
		client: client,
		url:    url,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "up"),
			Help: "Was the last scrape of the ElasticSearch cluster health endpoint successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "total_scrapes"),
			Help: "Current total ElasticSearch cluster health scrapes.",
		}),
	}
}

// implements prometheus.Collector interface
func (c *GlobalCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up.Desc()
	ch <- c.totalScrapes.Desc()
}

// implements prometheus.Collector interface
func (c *GlobalCollector) Collect(ch chan<- prometheus.Metric) {
	c.totalScrapes.Inc()
	defer func() {
		ch <- c.up
		ch <- c.totalScrapes
	}()

	c.up.Set(1)
}
