from orderbooks import (
    SimpleOrderBook,
    Stack,
)

from drgn.kafka import KafkaClient


def start():
    bid = Stack()
    ask = Stack(False)
    # bid.push({"price": 100, "quantity": 10, "order_id": "B1"})
    # bid.push({"price": 105, "quantity": 5, "order_id": "B2"})

    # ask.push({"price": 110, "quantity": 15, "order_id": "A1"})
    # ask.push({"price": 115, "quantity": 10, "order_id": "A2"})
    orderbook = SimpleOrderBook(
        bid, ask, KafkaClient("orderbook.group.id")
    )
    orderbook.start()
