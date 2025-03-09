package co.orderbook.streaming.candlechart;

import co.orderbook.streaming.models.Trade;

import org.apache.flink.api.common.eventtime.WatermarkStrategy;
import org.apache.flink.api.common.serialization.SimpleStringSchema;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;
import org.apache.flink.streaming.api.windowing.time.Time;
import org.apache.flink.streaming.connectors.kafka.FlinkKafkaConsumer;
import java.util.Properties;

public class TradeAggregationJob {
    public static void main(String[] args) throws Exception {
        // Set up the execution environment
        final StreamExecutionEnvironment env = StreamExecutionEnvironment.getExecutionEnvironment();

        // Configure Kafka consumer properties
        Properties properties = new Properties();
        properties.setProperty("bootstrap.servers", "localhost:9092");
        properties.setProperty("group.id", "trade-consumer-group");

        // Create a Kafka consumer
        FlinkKafkaConsumer<Trade> kafkaConsumer = new FlinkKafkaConsumer<>(
            "trades.topic",
            new TradeDeserializationSchema(),
            properties
        );

        // Assign timestamps and watermarks
        kafkaConsumer.assignTimestampsAndWatermarks(
            WatermarkStrategy.forMonotonousTimestamps()
                .withTimestampAssigner((event, timestamp) -> System.currentTimeMillis())
        );

        // Add the Kafka source to the environment
        DataStream<Trade> tradeStream = env.addSource(kafkaConsumer);

        // Perform aggregation in a 5-second tumbling window
        tradeStream
            .keyBy(Trade::getAction)
            .window(TumblingProcessingTimeWindows.of(Time.seconds(5)))
            .reduce((trade1, trade2) -> {
                Trade aggregatedTrade = new Trade();
                aggregatedTrade.setAction(trade1.getAction());
                aggregatedTrade.setQuantity(trade1.getQuantity() + trade2.getQuantity());
                aggregatedTrade.setPrice((trade1.getPrice() + trade2.getPrice()) / 2); // Average price
                return aggregatedTrade;
            })
            .print();

        // Execute the Flink job
        env.execute("Trade Aggregation Job");
    }
}
