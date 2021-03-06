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
	"strconv"
	"strings"

	. "github.com/unicell/trafficserver_exporter/event"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const (
	namespace = "trafficserver"
)

type StatsCollector struct {
	client *http.Client
	url    *url.URL
	ch     chan<- Events

	up                              prometheus.Gauge
	totalScrapes, jsonParseFailures prometheus.Counter
}

func NewStatsCollector(client *http.Client, url *url.URL, ch chan<- Events) *StatsCollector {
	subsystem := ""
	return &StatsCollector{
		client: client,
		url:    url,
		ch:     ch,
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
func (c *StatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up.Desc()
	ch <- c.totalScrapes.Desc()
	ch <- c.jsonParseFailures.Desc()
}

// implements prometheus.Collector interface
func (c *StatsCollector) Collect(ch chan<- prometheus.Metric) {
	c.totalScrapes.Inc()
	defer func() {
		ch <- c.up
		ch <- c.totalScrapes
		ch <- c.jsonParseFailures
	}()

	data, err := c.fetchAndDecode()
	if err != nil {
		c.up.Set(0)
		log.Errorln("Failed to fetch and decode JSON data:", err)
		return
	}

	stats, ok := data["global"]
	if !ok {
		c.up.Set(0)
		log.Errorln("Failed to read global key from JSON data:", err)
		return
	}

	events := Events{}
	for k, v := range stats.(map[string]interface{}) {
		ev, err := c.buildEvent(k, v)
		if err != nil || ev == nil {
			continue
		}
		events = append(events, ev)
	}
	c.ch <- events
	c.up.Set(1)
}

func (c *StatsCollector) fetchAndDecode() (map[string]interface{}, error) {
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

func (c *StatsCollector) buildEvent(k string, v interface{}) (Event, error) {
	if strings.Contains(k, "total") || strings.Contains(k, "count") {
		value, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse into float type: %v", v)
		}
		return NewGaugeEvent(k, value, map[string]string{}, false), nil
	} else if strings.Contains(k, "version") ||
		strings.Contains(k, "hostname") ||
		!strings.HasPrefix(k, "proxy") {
		return nil, fmt.Errorf("Not interested metric")
	} else {
		value, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			log.Infoln("Failed to parse: ", k, " -> ", v)
			return nil, fmt.Errorf("Failed to parse into float type: %v", v)
		}
		return NewGaugeEvent(k, value, map[string]string{}, false), nil
	}

	return nil, fmt.Errorf("No Event built")
}
