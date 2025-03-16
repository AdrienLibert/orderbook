package co.orderbook.demos.flinkhelloworld;

import org.apache.flink.api.common.eventtime.WatermarkStrategy;
import org.apache.flink.api.common.typeinfo.Types;
import org.apache.flink.api.connector.source.util.ratelimit.RateLimiterStrategy;
import org.apache.flink.connector.datagen.source.DataGeneratorSource;
import org.apache.flink.connector.datagen.source.GeneratorFunction;
import org.apache.flink.streaming.api.datastream.DataStreamSource;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;
import org.apache.flink.streaming.api.functions.sink.PrintSink;

public class DataStreamJob {

	public static void main(String[] args) throws Exception {
		final StreamExecutionEnvironment env = StreamExecutionEnvironment.getExecutionEnvironment();
		GeneratorFunction<Long, String> generatorFunction = index -> "Number: " + index;
		long numberOfRecords = 100_000;

		DataGeneratorSource<String> source = new DataGeneratorSource<>(
				generatorFunction,
				numberOfRecords,
				RateLimiterStrategy.perSecond(1),
				Types.STRING);

		DataStreamSource<String> stream =
		        env.fromSource(source,
		        WatermarkStrategy.noWatermarks(),
		        "Generator Source");

		PrintSink<String> sink = new PrintSink<>(true);

		stream.sinkTo(sink);

		env.execute("Flink Java API Skeleton");
	}
}
