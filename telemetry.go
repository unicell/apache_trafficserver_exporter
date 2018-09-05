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
	"github.com/prometheus/client_golang/prometheus"
)

var (
	eventStats = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "trafficserver_exporter_events_total",
			Help: "The total number of events seen.",
		},
		[]string{"type"},
	)
	eventsUnmapped = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "trafficserver_exporter_events_unmapped_total",
		Help: "The total number of events no mapping was found for.",
	})
	configLoads = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "trafficserver_exporter_config_reloads_total",
			Help: "The number of configuration reloads.",
		},
		[]string{"outcome"},
	)
	mappingsCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "trafficserver_exporter_loaded_mappings",
		Help: "The current number of configured metric mappings.",
	})
	conflictingEventStats = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "trafficserver_exporter_events_conflict_total",
			Help: "The total number of events with conflicting names.",
		},
		[]string{"type"},
	)
)

func init() {
	prometheus.MustRegister(eventStats)
	prometheus.MustRegister(eventsUnmapped)
	prometheus.MustRegister(configLoads)
	prometheus.MustRegister(mappingsCount)
	prometheus.MustRegister(conflictingEventStats)
}
