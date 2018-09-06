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

package event

import (
	"github.com/prometheus/statsd_exporter/pkg/mapper"
)

type Event interface {
	MetricName() string
	Value() float64
	Labels() map[string]string
	MetricType() mapper.MetricType
}

type CounterEvent struct {
	metricName string
	value      float64
	labels     map[string]string
}

func NewCounterEvent(metricName string, value float64, labels map[string]string) *CounterEvent {
	return &CounterEvent{
		metricName: metricName,
		value:      value,
		labels:     labels,
	}
}
func (c *CounterEvent) MetricName() string            { return c.metricName }
func (c *CounterEvent) Value() float64                { return c.value }
func (c *CounterEvent) Labels() map[string]string     { return c.labels }
func (c *CounterEvent) MetricType() mapper.MetricType { return mapper.MetricTypeCounter }

type GaugeEvent struct {
	metricName string
	value      float64
	labels     map[string]string
	relative   bool
}

func NewGaugeEvent(metricName string, value float64, labels map[string]string, relative bool) *GaugeEvent {
	return &GaugeEvent{
		metricName: metricName,
		value:      value,
		labels:     labels,
		relative:   relative,
	}
}
func (g *GaugeEvent) MetricName() string            { return g.metricName }
func (g *GaugeEvent) Value() float64                { return g.value }
func (c *GaugeEvent) Labels() map[string]string     { return c.labels }
func (c *GaugeEvent) Relative() bool                { return c.relative }
func (c *GaugeEvent) MetricType() mapper.MetricType { return mapper.MetricTypeGauge }

type TimerEvent struct {
	metricName string
	value      float64
	labels     map[string]string
}

func NewTimerEvent(metricName string, value float64, labels map[string]string) *TimerEvent {
	return &TimerEvent{
		metricName: metricName,
		value:      value,
		labels:     labels,
	}
}
func (t *TimerEvent) MetricName() string            { return t.metricName }
func (t *TimerEvent) Value() float64                { return t.value }
func (c *TimerEvent) Labels() map[string]string     { return c.labels }
func (c *TimerEvent) MetricType() mapper.MetricType { return mapper.MetricTypeTimer }

type Events []Event
