CREATE DATABASE IF NOT EXISTS analytics;
CREATE TABLE tick_data
(
    timestamp DateTime,
    open Float64,
    high Float64,
    low Float64,
    close Float64,
    volume integer
)
ENGINE = MergeTree
ORDER BY timestamp;