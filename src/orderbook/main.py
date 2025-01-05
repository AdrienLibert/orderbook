from orderbooks import (
    SimpleOrderBook,
    Stack,
)

from drgn.kafka import KafkaClient


def start():
    orderbook = SimpleOrderBook(
        Stack(), Stack(False), KafkaClient("orderbook.group.id")
    )
    orderbook.start()