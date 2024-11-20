# This script is intended to be modified, the goal is to have manual producer to kafka
# Feel free to edit values / whatever for your testing purposes
from confluent_kafka import Producer,Consumer
from datetime import datetime, timezone
import threading
import random
import time
import sys
import os

from drgn.kafka import kafka_config,kafka_consumer_config

import uuid
import json

sys.path.append(os.path.join(os.path.dirname(__file__), "../src"))

from orderbook.OrderBook import OrderBook
from orderbook.Stack import Stack
from orderbook.Order import Order
# defined in topics_config.yaml. if changed, need to rebuild the image.

TOPIC = "orders.topic"

# def produce_orderbook(producer: Producer, orderbook: OrderBook):
#     msg = {
#         "bid": orderbook.get_bid().to_dict(),
#         "ask": orderbook.get_ask().to_dict(),
#     }

#     producer.produce(TOPIC, bytes(json.dumps(msg), "utf-8"))
#     producer.flush()
#     print(f"Message produced to topic '{TOPIC}': {msg}")


def produce_order(producer: Producer, order: Order):
    msg = {
        "order_id": order.get_order_id(),
        "order_type": order.get_order_type(),
        "price":  order.get_price(),
        "quantity": order.get_quantity(),
        "time": int(datetime.now(timezone.utc).timestamp() * 1000000000),  # nanosecond
    }

    producer.produce(TOPIC, bytes(json.dumps(msg), "utf-8"))
    producer.flush()
    print(f"Message produced to topic '{TOPIC}': {msg}")

def consume_messages(orderbook: OrderBook):

    consumer = Consumer(kafka_consumer_config)
    consumer.subscribe([TOPIC])
    print("TOPIC")

    try:
        while True:
            msg = consumer.poll(timeout=1.0)
            if msg is None:
                continue
            if msg.error():
                print(msg.error())
                continue

            value = json.loads(msg.value().decode('utf-8'))

            order_id = value["order_id"]
            order_type = value["order_type"]
            price = value["price"]
            quantity = value["quantity"]

            order = Order(order_id,price,quantity,order_type,)
            orderbook.add_order(order)

    except KeyboardInterrupt:
        print("Consumer stopped.")
    finally:
        consumer.close()


if __name__ == "__main__":

    producer = Producer(kafka_config)

    bid = Stack()
    ask = Stack(False)
    bid.push(10,3)
    bid.push(20,5)
    ask.push(15,2)
    ask.push(25,6)
    bid.size()
    ask.size()
    orderbook = OrderBook(bid,ask)

    for _ in range(10):
        order_id = str(uuid.uuid4())
        price = random.randint(40, 60)
        quantity = random.randint(1, 10)
        order_type = random.choice(["Buy", "Sell"])
        order = Order(order_id, price, quantity, order_type)
        produce_order(producer, order)

    consumer = Consumer(kafka_consumer_config)
    consume_messages(orderbook)