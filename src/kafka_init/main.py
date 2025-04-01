import json
import time
import uuid
import yaml

from confluent_kafka import Producer
from confluent_kafka.admin import AdminClient, NewTopic, TopicMetadata
from datetime import datetime, timezone
from enum import StrEnum
from drgn.config import env_config
from drgn.kafka import kafka_config


def load_yaml(path: str) -> dict:
    with open(path, "r") as file:
        return yaml.safe_load(file)


def delivery(err, msg):
    print(f"INFO: {msg.value()}")


class TopicConfig:
    def __init__(self, topic_dict: dict):
        self.topic_name = list(topic_dict.keys())[0]
        self.partitions = topic_dict[self.topic_name]["partition"]


class TopicAction(StrEnum):
    CREATE = "CREATE"
    READ = "READ"
    UPDATE = "UPDATE"
    DELETE = "DELETE"


class KafkaTopicSynchronizer:
    def __init__(self, admin_client: AdminClient, topics_config: list[TopicConfig]):
        self.admin_client = admin_client
        self.topics_config = topics_config

        self.__timeout = 2.0

    def get_config_state(self) -> dict[str, TopicConfig]:
        return {topic.topic_name: topic for topic in self.topics_config}

    def get_current_state(self) -> dict[str, TopicMetadata]:
        return self.admin_client.list_topics(timeout=self.__timeout).topics

    def get_target_state(
        self,
        config_state: dict[str, TopicConfig],
        current_state: dict[str, TopicMetadata],
    ):
        target_state = []

        for topic, topic_metadata in current_state.items():
            if topic not in config_state:
                target_state.append((topic, TopicAction.DELETE, None))
            else:
                if (
                    len(topic_metadata.partitions.keys())
                    != config_state[topic].partitions
                ):
                    target_state.append((topic, TopicAction.DELETE, None))
                    target_state.append(
                        (topic, TopicAction.CREATE, config_state[topic])
                    )
                else:
                    target_state.append((topic, TopicAction.READ, None))

        for topic, config in config_state.items():
            if topic not in current_state:
                target_state.append((topic, TopicAction.CREATE, config))

        return target_state

    def run(self):
        config_state = self.get_config_state()
        current_state = self.get_current_state()
        target_state = self.get_target_state(config_state, current_state)
        print(f"Target state and actions: {target_state}")
        for topic, action, config in target_state:
            if action == TopicAction.CREATE:
                self.create_topic(topic, config)
            elif action == TopicAction.UPDATE:
                self.update_topic(topic, config)
            elif action == TopicAction.DELETE:
                self.delete_topic(topic)
            elif action == TopicAction.READ:
                print(f"topic {topic} is already in target configuration")

    def create_topic(self, topic: str, config: TopicConfig):
        new_topic = NewTopic(
            topic=topic,
            num_partitions=config.partitions,
            replication_factor=1,
            config={},
        )  # add potential other configs here
        res = self.admin_client.create_topics(
            new_topics=[new_topic],
            operation_timeout=self.__timeout,
            request_timeout=self.__timeout,
        )
        for tp, futures in res.items():
            print(f"topic '{tp}' created succesfully: {futures.result()}")

    def update_topic(self, topic: str, config: TopicConfig):
        raise NotImplementedError()

    def delete_topic(self, topic: str):
        self.admin_client.delete_topics([topic])
        print(f"topic '{topic}' deleted successfully.")


class ColdStartOrders:

    def __init__(
        self, producer: Producer, mid_price: float, spread: float, quantity: int
    ):
        self.quantity = quantity
        self.producer = producer

        self._starting_bid = mid_price - spread / 2
        self._starting_ask = mid_price + spread / 2

    def produce_orders(self):
        now = int(datetime.now(timezone.utc).timestamp() * 1000000000)
        buy_order = {
            "order_id": str(uuid.uuid4()),
            "order_type": "limit",
            "price": self._starting_bid,
            "quantity": self.quantity,
            "timestamp": now,
        }
        sell_order = {
            "order_id": str(uuid.uuid4()),
            "order_type": "limit",
            "price": self._starting_ask,
            "quantity": -self.quantity,
            "timestamp": now,
        }

        self.producer.produce(
            "orders.topic", bytes(json.dumps(buy_order), "utf-8"), on_delivery=delivery
        )
        self.producer.poll()

        self.producer.produce(
            "orders.topic", bytes(json.dumps(sell_order), "utf-8"), on_delivery=delivery
        )
        self.producer.poll()

    def run(self):
        print(
            f"Creating initial orders with parameters: quantity {self.quantity} - best bid {self._starting_bid} - best ask {self._starting_ask}"
        )
        self.produce_orders()


def main(script: str = "init_topics"):
    admin_client = AdminClient(kafka_config)

    try:
        topics = admin_client.list_topics(timeout=1.0).topics
    except Exception as e:
        print(f"Kafka not ready: {e}")
        exit(1)

    if script == "init_topics":
        config = load_yaml(env_config["kafka"]["topics_config"])
        topic_configs = (
            [TopicConfig(topic) for topic in config["topics"]] if config else []
        )
        topic_manager = KafkaTopicSynchronizer(admin_client, topic_configs)
        topic_manager.run()

    if script == "init_orderbook":
        sleep = 1.0
        retries = 10
        while retries:
            topics = admin_client.list_topics(timeout=1.0).topics
            print(f"Topics are: {topics}")
            if "orders.topic" in topics.keys():
                mid_price = float(env_config["orderbook"]["mid_price"])
                spread = float(env_config["orderbook"]["spread"])
                quantity = int(env_config["orderbook"]["quantity"])
                producer = Producer(kafka_config)
                cold_start_producer = ColdStartOrders(
                    producer, mid_price, spread, quantity
                )
                cold_start_producer.run()
                break

            time.sleep(sleep)
            sleep *= 1.5  # exp backoff
            retries -= 1

if __name__ == "__main__":
    main()