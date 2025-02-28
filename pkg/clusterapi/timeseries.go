package clusterapi

type Query struct {
	Name             string   `json:"name"`
	Downsampler      *string  `json:"downsampler,omitempty"`
	SourceAggregator *string  `json:"sourceAggregator,omitempty"`
	Derivative       *string  `json:"derivative,omitempty"`
	Sources          []string `json:"sources"`
}

type DataPoint struct {
	TimestampNanos string  `json:"timestampNanos"`
	Value          float64 `json:"value"`
}

type QueryResult struct {
	Query      Query       `json:"query"`
	Datapoints []DataPoint `json:"datapoints"`
}

type TimeseriesQueryResponse struct {
	Results []QueryResult `json:"results"`
}

type TimeseriesQueryRequest struct {
	StartNanos int64   `json:"start_nanos"`
	EndNanos   int64   `json:"end_nanos"`
	Queries    []Query `json:"queries"`
}
