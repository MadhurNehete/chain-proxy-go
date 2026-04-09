package process

type mrvProcessObservability struct {
	counters map[string]uint64
}

func newMRVProcessObservability() *mrvProcessObservability {
	return &mrvProcessObservability{
		counters: make(map[string]uint64),
	}
}

func (m *mrvProcessObservability) increment(metric string) {
	m.counters[metric]++
}

func (m *mrvProcessObservability) snapshot() map[string]uint64 {
	result := make(map[string]uint64, len(m.counters))
	for key, value := range m.counters {
		result[key] = value
	}

	return result
}

func (m *mrvProcessObservability) reset() {
	m.counters = make(map[string]uint64)
}

var proxyMRVMetrics = newMRVProcessObservability()

func recordProxyMRVMetric(metric string) {
	proxyMRVMetrics.increment(metric)
}

func snapshotProxyMRVMetrics() map[string]uint64 {
	return proxyMRVMetrics.snapshot()
}

func resetProxyMRVMetrics() {
	proxyMRVMetrics.reset()
}
