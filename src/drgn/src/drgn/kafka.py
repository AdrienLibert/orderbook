import json
from confluent_kafka import Producer, Consumer, TopicPartition, KafkaError

from drgn.config import env_config
from typing import Generator

kafka_config = {
    "bootstrap.servers": env_config["kafka"]["bootstrap_servers"],
    "security.protocol": env_config["kafka"]["security_protocol"],
}


class KafkaClient:
    def __init__(self, group_id: str = __file__):
        self._consumer_config = kafka_config | {
            "group.id": group_id,
            "on_commit": lambda err, topics: print(err, topics),
            "enable.partition.eof": True,
        }
        self._producer_config = kafka_config
        print(self._producer_config)
        self._consumer = None
        self._producer = None

    @property
    def consumer(self) -> Consumer:
        if not self._consumer:  # lazy
            self._consumer = Consumer(self._consumer_config)
        return self._consumer

    @property
    def producer(self) -> Producer:
        if not self._producer:  # lazy
            self._producer = Producer(self._producer_config)
        return self._producer

    def _get_topic_partition(self, topic: str):
        return TopicPartition(topic, 0, 0)

    def consume(self, topic: str) -> Generator[list[dict], None, None]:
        self._consumer_config["auto.offset.reset"] = "earliest"
        self._consumer_config["enable.auto.commit"] = False
        self.consumer.assign([self._get_topic_partition(topic)])
        if isinstance(env_config, dict):
                consumer_config = env_config.get('consumer', {})
        else:
            try:
                consumer_config = dict(env_config['consumer']) if 'consumer' in env_config else {}
            except Exception as e:
                print(f"Error configuration: {e}")
                consumer_config = {}
        consume_size = int(consumer_config.get("consume_size", 10))
        consume_timeout = float(consumer_config.get("consume_timeout", 1.0))

        while True:
            msgs = self.consumer.consume(consume_size, timeout=consume_timeout)
            if not msgs:
                print("No messages received.")
                continue

            messages = []
            for msg in msgs:
                if msg.error():
                    if msg.error().code() == KafkaError._PARTITION_EOF:
                        continue 
                    else:
                        print(f"ERROR - KafkaException - {msg.error()}")
                else:
                    print(f"Message received: {msg.value()}")
                    messages.append(json.loads(msg.value().decode("utf-8")))

            if messages:
                yield messages
                self.consumer.commit()

    def produce(self, topic: str, message: bytes):
        self.producer.produce(topic, message)
        self.producer.flush()
