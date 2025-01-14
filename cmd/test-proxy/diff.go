package main

import (
	"fmt"
	"sort"
	"strings"
)

type Diff struct {
	Missing []*Metric
	Extra   []*Metric
}

func DiffMetrics(expected, actual []*Metric) *Diff {
	keyers := metricKeyers(expected)
	expectedKeyMap := metricKeyMap(expected, keyers)
	actualKeyMap := metricKeyMap(actual, keyers)
	missing, extra := disjunct(expectedKeyMap, actualKeyMap)
	return &Diff{
		Missing: missing,
		Extra:   extra,
	}
}

// keyer returns whether or not it could generate a key and the key of the given metric
type keyer func(*Metric) (bool, string)

func metricKeyers(expected []*Metric) map[string][]keyer {
	keyersByMetric := map[string][]keyer{}
	for _, m := range expected {
		keyersByMetric[m.Name] = append(keyersByMetric[m.Name], metricKeyer(m))
	}
	return keyersByMetric
}

func metricKeyer(m *Metric) keyer {
	var keyers []keyer
	keyers = append(keyers, nameKey(m.Name))
	if m.Value != "" {
		keyers = append(keyers, valueKey(m.Value))
	}
	if m.Timestamp != "" {
		keyers = append(keyers, timestampKey(m.Timestamp))
	}
	keyers = append(keyers, tagsKey(m.Tags))
	return compositeKey(keyers...)
}

func compositeKey(keyers ...keyer) keyer {
	return func(metric *Metric) (bool, string) {
		var keys []string
		for _, keyer := range keyers {
			matched, key := keyer(metric)
			if !matched {
				return false, ""
			}
			keys = append(keys, key)
		}
		return true, strings.Join(keys, " ")
	}
}

func nameKey(expected string) keyer {
	return func(metric *Metric) (bool, string) {
		return metric.Name == expected, metric.Name
	}
}

func valueKey(expected string) keyer {
	return func(metric *Metric) (bool, string) {
		return metric.Value == expected, metric.Value
	}
}

func timestampKey(expected string) keyer {
	return func(metric *Metric) (bool, string) {
		return metric.Timestamp == expected, metric.Timestamp
	}
}

func tagNameKey(name string) keyer {
	return func(metric *Metric) (bool, string) {
		_, exists := metric.Tags[name]
		return exists, fmt.Sprintf("%s=*", name)
	}
}

func fullTagKey(name, value string) keyer {
	return func(metric *Metric) (bool, string) {
		return metric.Tags[name] == value, fmt.Sprintf("%s=%#v", name, metric.Tags[name])
	}
}

func tagsKey(tags map[string]string) keyer {
	tagNames := make([]string, 0, len(tags))
	for name := range tags {
		tagNames = append(tagNames, name)
	}
	sort.Strings(tagNames)
	keyers := make([]keyer, len(tags))
	for i, name := range tagNames {
		if tags[name] == "" {
			keyers[i] = tagNameKey(name)
		} else {
			keyers[i] = fullTagKey(name, tags[name])
		}
	}
	return compositeKey(keyers...)
}

func metricKeyMap(metrics []*Metric, keyers map[string][]keyer) map[string]*Metric {
	keyMap := map[string]*Metric{}
	for _, metric := range metrics {
		foundKeyers := keyers[metric.Name]
		found := false
		for _, foundKeyer := range foundKeyers {
			matched, key := foundKeyer(metric)
			if matched {
				keyMap[key] = metric
				found = true
				break
			}
		}
		if !found {
			_, key := metricKeyer(metric)(metric)
			keyMap[key] = metric
		}
	}
	return keyMap
}

func disjunct(a, b map[string]*Metric) (onlyInA []*Metric, onlyInB []*Metric) {
	for x := range a {
		if _, exists := b[x]; !exists {
			onlyInA = append(onlyInA, a[x])
		}
	}
	for y := range b {
		if _, exists := a[y]; !exists {
			onlyInB = append(onlyInB, b[y])
		}
	}
	return onlyInA, onlyInB
}
