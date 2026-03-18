CREATE TABLE IF NOT EXISTS api_request_metrics (
    id BIGSERIAL PRIMARY KEY,
    request_type VARCHAR(20) NOT NULL,     -- graphql | rest | grpc
    load INTEGER NOT NULL,                 -- in-flight requests at request time
    request_time TIMESTAMPTZ NOT NULL,
    response_time TIMESTAMPTZ NOT NULL,
    total_time_ms BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_request_metrics_request_type
ON api_request_metrics(request_type);

CREATE INDEX IF NOT EXISTS idx_api_request_metrics_request_time
ON api_request_metrics(request_time);