# This script is intended to be modified, the goal is to have manual producer to kafka
# Feel free to edit values / whatever for your testing purposes
from confluent_kafka import Producer
from datetime import datetime, timezone

from drgn.kafka import kafka_config

import uuid
import json
import time
import random
import argparse

# defined in topics_config.yaml. if changed, need to rebuild the image.
TOPIC = "orders.topic"


def delivery(err, msg):
    print(f"INFO: {msg.value()}")


def produce_buy_order(
    producer: Producer, min_quantity: int = -50, max_quantity: int = 50
):
    qty = random.randint(int(min_quantity), int(max_quantity))
    msg = {
        "order_id": str(uuid.uuid4()),
        "order_type": "limit",
        "price": random.randint(45, 60),
        "quantity": qty if qty != 0 else 1,
        "timestamp": int(
            datetime.now(timezone.utc).timestamp() * 1000000000
        ),  # nanosecond
    }

    producer.produce(TOPIC, bytes(json.dumps(msg), "utf-8"), on_delivery=delivery)
    producer.poll()


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--wait-for-input", dest="wait_for_input", default=False, action="store_true"
    )
    parser.add_argument("--min-qty", dest="min_quantity", default=-50)
    parser.add_argument("--max-qty", dest="max_quantity", default=50)
    args = parser.parse_args()

    producer = Producer(kafka_config)

    count = 0
    while True:
        produce_buy_order(producer, args.min_quantity, args.max_quantity)
        if args.wait_for_input:
            _ = input()
        else:
            time.sleep(1.0)
