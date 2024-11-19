# This script is intended to be modified, the goal is to have manual producer to kafka
# Feel free to edit values / whatever for your testing purposes
from confluent_kafka import Producer,Consumer
from datetime import datetime, timezone
import threading
import time
import sys
import os

from drgn.kafka import kafka_config,kafka_consumer_config

import uuid
import json

sys.path.append(os.path.join(os.path.dirname(__file__), "../src"))

from orderbook.OrderBook import OrderBook
from orderbook.Stack import Stack
from orderbook.Order import OrderBuy, OrderSell
# defined in topics_config.yaml. if changed, need to rebuild the image.

TOPIC = "orderbook"

def produce_orderbook(producer: Producer, orderbook: OrderBook):
    msg = {
        "bid": orderbook.get_bid().to_dict(),
        "ask": orderbook.get_ask().to_dict(),
    }

    producer.produce(TOPIC, bytes(json.dumps(msg), "utf-8"))
    producer.flush()
    print(f"Message produced to topic '{TOPIC}': {msg}")


def produce_buy_order(producer: Producer, order: OrderBuy):
    msg = {
        "order_id": order.get_order_id(),
        "order_type": "BUY",
        "price":  order.get_price(),
        "quantity": order.get_quantity(),
        "time": int(datetime.now(timezone.utc).timestamp() * 1000000000),  # nanosecond
    }

    producer.produce(TOPIC, bytes(json.dumps(msg), "utf-8"))
    producer.flush()
    print(f"Message produced to topic '{TOPIC}': {msg}")

def produce_sell_order(producer: Producer, order: OrderSell):
    msg = {
        "order_id": order.get_order_id(),
        "order_type": "BUY",
        "price": order.get_price(),
        "quantity": order.get_quantity(),
        "time": int(datetime.now(timezone.utc).timestamp() * 1000000000),  # nanosecond
    }

    producer.produce(TOPIC, bytes(json.dumps(msg), "utf-8"))
    producer.flush()
    print(f"Message produced to topic '{TOPIC}': {msg}")


def consume_messages():
    # Initialize Kafka Consumer
    consumer = Consumer(kafka_consumer_config)

    # Subscribe to the topic
    consumer.subscribe([TOPIC])

    print(f"Consumer subscribed to topic: {TOPIC}")


if __name__ == "__main__":
    # init thread level
    producer = Producer(kafka_config)

    order_buy = OrderBuy(str(uuid.uuid4()), 48, 6)
    produce_buy_order(producer, order_buy)
    order_sell = OrderSell(str(uuid.uuid4()), 50, 4)
    produce_sell_order(producer, order_sell)

    bid = Stack()
    ask = Stack(False)
    bid.push(10,3)
    bid.push(20,5)
    ask.push(15,2)
    ask.push(25,6)
    bid.size()
    ask.size()
    orderbook = OrderBook(bid,ask)
    produce_orderbook(producer, orderbook)
    consume_messages()
