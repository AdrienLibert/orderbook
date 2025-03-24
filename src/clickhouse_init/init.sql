CREATE DATABASE IF NOT EXISTS analytics;
CREATE TABLE orderbook_data
(
    timestamp DateTime,
    open Float64,
    high Float64,
    low Float64,
    close Float64,
    volume Float64
)
ENGINE = MergeTree
ORDER BY timestamp;