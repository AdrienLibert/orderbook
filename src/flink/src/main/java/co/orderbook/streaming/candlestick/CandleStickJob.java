package co.orderbook.streaming.candlestick;

import co.orderbook.streaming.models.Trade;
import co.orderbook.streaming.models.TradeDeserializationSchema;

import org.apache.flink.api.common.eventtime.WatermarkStrategy;
import org.apache.flink.connector.kafka.source.KafkaSource;
import org.apache.flink.connector.kafka.source.enumerator.initializer.OffsetsInitializer;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;
import org.apache.flink.streaming.api.windowing.assigners.TumblingProcessingTimeWindows;
import java.time.Duration;

public class CandleStickJob {
    public static void main(String[] args) throws Exception {
        // Set up the execution environment
        final StreamExecutionEnvironment env = StreamExecutionEnvironment.getExecutionEnvironment();

        KafkaSource<Trade> kafkaSource = KafkaSource.<Trade>builder()
            .setBootstrapServers("bitnami-kafka.orderbook:9092")
            .setTopics("trades.topic")
            .setGroupId("trade-consumer-flink-group")
            .setStartingOffsets(OffsetsInitializer.earliest())
            .setValueOnlyDeserializer(new TradeDeserializationSchema())
            .build();

        DataStream<Trade> stream = env.fromSource(kafkaSource, WatermarkStrategy.noWatermarks(), "Trades Topic Source");

        // Perform aggregation in a 5-second tumbling window
        stream
            .keyBy(Trade::getTrade_id)
            .window(TumblingProcessingTimeWindows.of(Duration.ofSeconds(5)))
            .reduce((trade1, trade2) -> {
                // TODO: make it a real agregation, this is for kafka testing purposos
                Trade aggregatedTrade = new Trade();
                aggregatedTrade.setAction("AGGREGATED");
                aggregatedTrade.setQuantity(trade1.getQuantity() + trade2.getQuantity());
                aggregatedTrade.setPrice((trade1.getPrice() + trade2.getPrice()) / 2); // Average price
                return aggregatedTrade;
            });

        // Execute the Flink job
        stream.print();
        env.execute("Trade Aggregation Job");
    }
}
