from order import Order
from drgn.kafka import kafka_config
from drgn.env import config_from_env
from confluent_kafka import Consumer, TopicPartition

class OrderBookKafka:
    def __init__(self, topic: str, group_id: str):
        self.consumer = Consumer(
            kafka_config
            | {
                "group.id": group_id,
                "auto.offset.reset": "earliest",
                "on_commit": lambda err, topics: print(err, topics),
            }
        )
        self.topic = topic
        self.consumer.assign([TopicPartition(topic, 0, 0)])

    def listen_to_kafka(self, orderbook):
        print(f"Listening to Kafka topic '{self.topic}'...")
        try:
            while True:
                print("Polling Kafka...")
                msg = self.consumer.poll(1.0)
                if msg is None:
                    continue
                if msg.error():
                    print(f"Consumer error: {msg.error()}")
                    continue
                order = self.parse_order(msg.value().decode("utf-8"))
                if order:
                    orderbook.add_order(order)
        except KeyboardInterrupt:
            print("Stopping consumer...")
        finally:
            self.consumer.close()

    @staticmethod
    def parse_order(message: str):
    
        try:
            import json
            order = json.loads(message)
            return Order(order["order_id"], order["price"], order["quantity"], order["order_type"])
        except Exception as e:
            print(f"Failed to parse message: {message}. Error: {e}")
            return None