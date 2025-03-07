package tsdb

type QueryAgg string

const (
	QueryAggSum      QueryAgg = "sum"
	QueryAggAvg      QueryAgg = "avg"
	QueryAggMin      QueryAgg = "min"
	QueryAggMax      QueryAgg = "max"
	QueryAggFirst    QueryAgg = "first"
	QueryAggLast     QueryAgg = "last"
	QueryAggVariance QueryAgg = "variance"
)

type QueryDerivative string

const (
	None                  QueryDerivative = "none"
	Derivative            QueryDerivative = "derivative"
	NonNegativeDerivative QueryDerivative = "non-negative-derivative"
)

type Datapoint struct {
	TimestampNanos int64   `json:"timestamp_nanos"`
	Value          float64 `json:"value"`
}

type Query struct {
	Name              string          `json:"name"`
	DownSampler       QueryAgg        `json:"downsampler"`
	SourceAggeregator QueryAgg        `json:"source_aggregator"`
	Derivative        QueryDerivative `json:"derivative"`
	Sources           []string        `json:"sources"`
}

type Result struct {
	Query      Query       `json:"query"`
	Datapoints []Datapoint `json:"datapoints"`
}

type TSQueryRequest struct {
	StartNanos  int64   `json:"start_nanos"`
	EndNanos    int64   `json:"end_nanos"`
	Queries     []Query `json:"queries"`
	SampleNanos int64   `json:"sample_nanos"`
}

type TSQueryResponse struct {
	Results []Result `json:"results"`
}
