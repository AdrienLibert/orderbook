import time
from datetime import datetime, timezone
from drgn.env import config_from_env
import orderbook_kafka
import orderbook
import stack

if __name__ == "__main__":
    # Inject credentails to application
    print("Running entrypoint of orderbook")
    config_from_env()
    bid_stack = stack.Stack()
    ask_stack = stack.Stack()
    orderbook = orderbook_kafka.OrderBook(bid_stack, ask_stack)
    kafka_orderbook = orderbook_kafka.OrderBookKafka(topic="orders.topic", group_id="orderbook")
    kafka_orderbook.listen_to_kafka(orderbook)