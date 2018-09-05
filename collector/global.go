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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
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
		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "json_parse_failures"),
			Help: "Number of errors while parsing JSON.",
		}),
	}
}

// implements prometheus.Collector interface
func (c *GlobalCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up.Desc()
	ch <- c.totalScrapes.Desc()
	ch <- c.jsonParseFailures.Desc()
}

// implements prometheus.Collector interface
func (c *GlobalCollector) Collect(ch chan<- prometheus.Metric) {
	c.totalScrapes.Inc()
	defer func() {
		ch <- c.up
		ch <- c.totalScrapes
		ch <- c.jsonParseFailures
	}()

	_, err := c.fetchAndDecode()
	if err != nil {
		c.up.Set(0)
		log.Errorln("Failed to fetch and decode JSON stats:", err)
		return
	}
	c.up.Set(1)
}

func (c *GlobalCollector) fetchAndDecode() (map[string]interface{}, error) {
	var stats map[string]interface{}

	u := *c.url
	u.Path = path.Join(u.Path, "/_stats")
	res, err := c.client.Get(u.String())
	if err != nil {
		return stats, fmt.Errorf("Failed to get cluster health from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			log.Errorln("Failed to close http.Client:", err)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return stats, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		c.jsonParseFailures.Inc()
		return stats, err
	}

	return stats, nil
}
