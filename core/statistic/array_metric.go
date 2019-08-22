package statistic

// The basic metric statistic interface in Sentinel using a internal SlidingWindow.
type ArrayMetric struct {
	data         *SlidingWindow
	sampleCount  uint32
	intervalInMs uint32
}

func NewArrayMetric(sampleCount, intervalInMs uint32) *ArrayMetric {
	return &ArrayMetric{
		data:         NewSlidingWindow(sampleCount, intervalInMs),
		sampleCount:  sampleCount,
		intervalInMs: intervalInMs,
	}
}

func (m *ArrayMetric) GetIntervalInMs() uint32 {
	return m.intervalInMs
}

func (m *ArrayMetric) AddPass(count uint64) {
	m.data.AddCount(MetricEventPass, count)
}
func (m *ArrayMetric) AddBlock(count uint64) {
	m.data.AddCount(MetricEventBlock, count)
}
func (m *ArrayMetric) AddError(count uint64) {
	m.data.AddCount(MetricEventError, count)
}
func (m *ArrayMetric) AddSuccess(count uint64) {
	m.data.AddCount(MetricEventSuccess, count)
}

func (m *ArrayMetric) Request() uint64 {
	return m.Pass() + m.Block()
}
func (m *ArrayMetric) Pass() uint64 {
	return m.data.Count(MetricEventPass)
}
func (m *ArrayMetric) Block() uint64 {
	return m.data.Count(MetricEventBlock)
}
func (m *ArrayMetric) Success() uint64 {
	return m.data.Count(MetricEventSuccess)
}
func (m *ArrayMetric) Error() uint64 {
	return m.data.Count(MetricEventError)
}

// the unit of rt is millisecond
func (m *ArrayMetric) AddRt(rt uint64) {
	m.data.AddCount(MetricEventRt, rt)
}

func (m *ArrayMetric) Rt() uint64 {
	return m.data.Count(MetricEventRt)
}
