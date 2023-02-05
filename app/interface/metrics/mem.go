package metrics

type Metrics struct {
	TotalWords    int
	FilteredWords int
}

func New() *Metrics {
	return &Metrics{
		TotalWords:    0,
		FilteredWords: 0,
	}
}

func (m *Metrics) IncWords() {
	m.TotalWords++
}

func (m *Metrics) IncFilteredWords() {
	m.FilteredWords++
}

func (m *Metrics) GetMetrics() map[string]any {
	return map[string]any{
		"total_words":    m.TotalWords,
		"filtered_words": m.FilteredWords,
	}
}
