from confluent_kafka import Consumer, TopicPartition
import json
from drgn.kafka import kafka_config
import time
import random
import sys
import os

sys.path.append(os.path.join(os.path.dirname(__file__), "../src"))

from orderbook.OrderBook import OrderBook
from orderbook.Stack import Stack
from orderbook.Order import Order
# defined in topics_config.yaml. if changed, need to rebuild the image.
TOPIC = "orders.topic"


def consume_orders(consumer: Consumer, orderbook: OrderBook):
    while True:
        msg = consumer.poll(1)

        if msg:
            if msg.error():
                print(msg.error)
            else:
                order_data = json.loads(msg.value().decode('utf-8'))   
                order_id = order_data.get('order_id')
                order_type = order_data.get('order_type')
                price = order_data.get('price')
                quantity = order_data.get('quantity')
                timestamp = order_data.get('time') 
                order = Order(order_id,price,quantity,order_type)
                if order_type == 'Buy':
                    orderbook.match_buy_order(order)
                else:
                    orderbook.match_sell_order(order)
                print(orderbook)
        else:
            print("no message")
        time.sleep(1)
        consumer.commit()


if __name__ == "__main__":

    bid = Stack()
    ask = Stack(False)
    bid.push(10,3)
    bid.push(20,5)
    ask.push(15,2)
    ask.push(25,6)
    bid.size()
    ask.size()
    orderbook = OrderBook(bid,ask)
    # init thread level
    consumer = Consumer(
        kafka_config
        | {
            "group.id": "test",
            "on_commit": lambda x: print(x),
            "auto.offset.reset": "earliest",
        }
    )

    consumer.assign([TopicPartition(TOPIC, 0, 0)])

    consume_orders(consumer,orderbook)