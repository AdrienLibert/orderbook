from pyspark.sql import SparkSession
from pyspark.sql.functions import col, from_json
from pyspark.sql.types import StructType, StructField, StringType, LongType, DoubleType

# Initialize Spark session with Kafka packages
spark = SparkSession.builder \
    .appName("KafkaSparkReader") \
    .config("spark.jars", "/app/spark-sql-kafka-0-10_2.12-3.5.5.jar,/app/kafka-clients-3.5.2.jar,/app/spark-token-provider-kafka-0-10_2.12-3.5.5.jar,/app/commons-pool2-2.12.0.jar") \
    .getOrCreate()

# Log start
print("Starting Kafka Spark reader...")

# Kafka configuration
kafka_bootstrap_servers = "bitnami-kafka.orderbook.svc.cluster.local:9092"
kafka_topic = "trades.topic"

# Define schema for Kafka messages based on Trade struct
schema = StructType([
    StructField("trade_id", StringType(), True),
    StructField("order_id", StringType(), True),
    StructField("quantity", LongType(), True),
    StructField("price", DoubleType(), True),
    StructField("action", StringType(), True),
    StructField("status", StringType(), True),
    StructField("timestamp", LongType(), True)
])

try:
    # Read from Kafka
    df = spark \
        .readStream \
        .format("kafka") \
        .option("kafka.bootstrap.servers", kafka_bootstrap_servers) \
        .option("subscribe", kafka_topic) \
        .option("startingOffsets", "latest") \
        .load()

    # Cast key and parse JSON value
    df = df.select(
        col("key").cast("string"),
        from_json(col("value").cast("string"), schema).alias("parsed_value"),
        col("topic"),
        col("partition"),
        col("offset")
    )

    # Extract fields from parsed JSON
    df = df.select(
        col("key"),
        col("parsed_value.trade_id"),
        col("parsed_value.order_id"),
        col("parsed_value.quantity"),
        col("parsed_value.price"),
        col("parsed_value.action"),
        col("parsed_value.status"),
        col("parsed_value.timestamp"),
        col("topic"),
        col("partition"),
        col("offset")
    )

    # Print to console
    query = df \
        .writeStream \
        .outputMode("append") \
        .format("console") \
        .option("truncate", "false") \
        .start()

    # Log query start
    print(f"Streaming from Kafka topic {kafka_topic}...")

    # Wait for termination
    query.awaitTermination(timeout=60)  # Run for 60 seconds

except Exception as e:
    print(f"Error in Kafka streaming: {str(e)}")

finally:
    # Stop the query and Spark session
    if 'query' in locals():
        query.stop()
    spark.stop()
    print("Kafka Spark reader completed.")