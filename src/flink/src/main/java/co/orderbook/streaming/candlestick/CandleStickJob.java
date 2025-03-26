package co.orderbook.streaming.candlestick;

import co.orderbook.streaming.models.Trade;
import co.orderbook.streaming.models.TradeDeserializationSchema;

import org.apache.flink.api.common.eventtime.WatermarkStrategy;
import org.apache.flink.connector.kafka.source.KafkaSource;
import org.apache.flink.connector.kafka.source.enumerator.initializer.OffsetsInitializer;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;
import org.apache.flink.streaming.api.windowing.assigners.TumblingProcessingTimeWindows;

import org.apache.flink.table.api.*;
import org.apache.flink.table.api.Expressions.*;
import org.apache.flink.types.Row;

import static org.apache.flink.table.api.Expressions.$;
import static org.apache.flink.table.api.Expressions.lit;

import java.time.Duration;
import java.time.LocalTime;

public class CandleStickJob {

    public static void main(String[] args) throws Exception {
        final EnvironmentSettings settings = EnvironmentSettings
            .newInstance()
            .inStreamingMode()
            .build();
        
        final TableEnvironment tableEnv = TableEnvironment.create(settings);

        tableEnv.executeSql(
            "CREATE TABLE Trades ("+
            "    trade_id STRING,"+
            "    order_id STRING,"+
            "    quantity INT,"+
            "    price FLOAT,"+
            "    action STRING,"+
            "    status STRING,"+
            "    `timestamp` BIGINT,"+
            "    event_time AS TO_TIMESTAMP(FROM_UNIXTIME(`timestamp`)),"+
            "    WATERMARK FOR event_time AS event_time - INTERVAL '6' SECOND"+
            ") WITH ("+
            "    'connector' = 'kafka',"+
            "    'topic' = 'trades.topic',"+
            "    'properties.bootstrap.servers' = 'bitnami-kafka.orderbook:9092',"+
            "    'properties.group.id' = 'candle-stick-job',"+
            "    'scan.startup.mode' = 'earliest-offset',"+
            "    'format' = 'json'"+
            ");"
        );

        Table trades = tableEnv.from("Trades");

        Table tickDataTable = trades
            .window(Tumble.over(lit(5).seconds()).on($("event_time")).as("w"))
            .groupBy($("w"))
            .select(
                $("w").start().as("windowStart"),
                $("w").end().as("windowEnd"),
                $("price").firstValue().as("open"),
                $("price").max().as("high"),
                $("price").min().as("low"),
                $("price").lastValue().as("close"),
                $("quantity").sum().as("volume")
            );

        TableResult result = tickDataTable.execute();
        result.print();
    }
}