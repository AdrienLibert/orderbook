# This script is intended to be modified, the goal is to have manual producer to kafka
# Feel free to edit values / whatever for your testing purposes
from confluent_kafka import Producer
from datetime import datetime, timezone

from drgn.kafka import kafka_config

import uuid
import json
import time
import random


# defined in topics_config.yaml. if changed, need to rebuild the image.
TOPIC = "orders.topic"


def produce_buy_order(producer: Producer):
    msg = {
        "order_id": str(uuid.uuid4()),
        "price": random.randint(45, 55),
        "quantity": random.randint(1, 50),
        "time": int(datetime.now(timezone.utc).timestamp() * 1000000000),  # nanosecond
    }

    producer.produce(TOPIC, bytes(json.dumps(msg), "utf-8"))
    producer.flush()


def produce_sell_order(producer: Producer):
    msg = {
        "order_id": str(uuid.uuid4()),
        "price": -random.randint(45, 55),
        "quantity": random.randint(1, 50),
        "time": int(datetime.now(timezone.utc).timestamp() * 1000000000),  # nanosecond
    }

    producer.produce(TOPIC, bytes(json.dumps(msg), "utf-8"))
    producer.flush()


if __name__ == "__main__":
    # init thread level
    producer = Producer(kafka_config)

    produce_buy_order(producer)
    time.sleep(1)
    produce_sell_order(producer)
