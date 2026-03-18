package main

import (
	"context"
	"time"
)

type BenchmarkRequest struct {
	Message string `json:"message"`
	WorkMS  int32  `json:"work_ms"`
}

type BenchmarkResponse struct {
	Transport      string `json:"transport"`
	Message        string `json:"message"`
	RequestTime    string `json:"request_time"`
	ResponseTime   string `json:"response_time"`
	TotalTimeMS    int64  `json:"total_time_ms"`
	CurrentLoad    int    `json:"current_load"`
	FastestHint    string `json:"fastest_hint"`
	ProcessedValue string `json:"processed_value"`
}

type BenchmarkService struct {
	metrics *MetricsStore
}

func NewBenchmarkService(metrics *MetricsStore) *BenchmarkService {
	return &BenchmarkService{
		metrics: metrics,
	}
}

func (s *BenchmarkService) Handle(
	ctx context.Context,
	transport string,
	req BenchmarkRequest,
) (*BenchmarkResponse, error) {
	requestTime := time.Now().UTC()

	load := s.metrics.Increment(transport)
	defer s.metrics.Decrement(transport)

	if req.WorkMS > 0 {
		select {
		case <-time.After(time.Duration(req.WorkMS) * time.Millisecond):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	responseTime := time.Now().UTC()
	totalTimeMS := responseTime.Sub(requestTime).Milliseconds()

	if err := s.metrics.InsertMetric(
		ctx,
		transport,
		load,
		requestTime,
		responseTime,
		totalTimeMS,
	); err != nil {
		return nil, err
	}

	return &BenchmarkResponse{
		Transport:      transport,
		Message:        req.Message,
		RequestTime:    requestTime.Format(time.RFC3339Nano),
		ResponseTime:   responseTime.Format(time.RFC3339Nano),
		TotalTimeMS:    totalTimeMS,
		CurrentLoad:    load,
		FastestHint:    "Use PostgreSQL aggregation to identify whether REST or GraphQL is faster overall",
		ProcessedValue: "Processed: " + req.Message,
	}, nil
}
