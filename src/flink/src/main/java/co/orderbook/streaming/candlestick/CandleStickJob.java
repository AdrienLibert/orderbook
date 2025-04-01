package co.orderbook.streaming.candlestick;

import org.apache.flink.table.api.*;

import static org.apache.flink.table.api.Expressions.$;
import static org.apache.flink.table.api.Expressions.lit;

public class CandleStickJob {

    public static Table compute_ticks(Table tradesTable) {
        return tradesTable
            .where($("action").isEqual("BUY")) // remove duplicates
            .window(Tumble.over(lit(5).seconds()).on($("event_time")).as("w"))
            .groupBy($("w"))
            .select(
                $("w").start().as("window_start"),
                $("w").end().as("window_end"),
                $("price").firstValue().as("open"),
                $("price").max().as("high"),
                $("price").min().as("low"),
                $("price").lastValue().as("close"),
                $("quantity").sum().as("volume"));
    }

    public static void main(String[] args) throws Exception {
        final EnvironmentSettings settings = EnvironmentSettings.inStreamingMode();
        final TableEnvironment tableEnv = TableEnvironment.create(settings);
        final TableConfig config = tableEnv.getConfig();
        config.set("table.exec.source.idle-timeout", "1000 ms");
        config.set("table.exec.resource.default-parallelism", "1");
        config.set("table.local-time-zone", "UTC");

        tableEnv.executeSql(
            "CREATE TABLE Trades (\n"+
            "    trade_id STRING,\n"+
            "    order_id STRING,\n"+
            "    quantity INT,\n"+
            "    price FLOAT,\n"+
            "    action STRING,\n"+
            "    status STRING,\n"+
            "    `timestamp` BIGINT,\n"+
            "    event_time AS TO_TIMESTAMP_LTZ(`timestamp`, 3),\n"+
            "    WATERMARK FOR event_time AS event_time - INTERVAL '5' SECOND\n"+
            ") WITH (\n"+
            "    'connector' = 'kafka',\n"+
            "    'topic' = 'trades.topic',\n"+
            "    'properties.bootstrap.servers' = 'bitnami-kafka.orderbook:9092',\n"+
            "    'properties.group.id' = 'candle-stick-job',\n"+
            "    'scan.startup.mode' = 'earliest-offset',\n"+
            "    'format' = 'json'\n"+
            ");"
        );
        
        tableEnv.executeSql(
            "CREATE TABLE TickDataSink (\n" +
            "    `window_start` TIMESTAMP(3),\n" +
            "    `window_end` TIMESTAMP(3),\n" +
            "    `open` FLOAT,\n" +
            "    `high` FLOAT,\n" +
            "    `low` FLOAT,\n" +
            "    `close` FLOAT,\n" +
            "    `volume` FLOAT\n" +
            ") WITH (\n" +
            "    'connector'    = 'jdbc',\n" +
            "    'url'          = 'jdbc:postgresql://postgres-postgresql:5432/analytics',\n" +
            "    'table-name'   = 'tick_data',\n" +
            "    'driver'       = 'org.postgresql.Driver',\n" +
            "    'username'     = 'postgres',\n" +
            "    'password'     = 'postgres'\n" +
            ")"
        );

        Table tradesTable = tableEnv.from("Trades");
        compute_ticks(tradesTable).executeInsert("TickDataSink");
    }
}