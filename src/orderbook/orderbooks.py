import json
from drgn.kafka import KafkaClient
import threading


class Stack(list):
    def __init__(self, is_bid=True):
        self.is_bid = is_bid

    def _sort(self):
        self.sort(key=lambda x: x["price"], reverse=self.is_bid)

    def push(self, order: dict):
        self.append(order)
        self._sort()

    def peek(self):
        if not self:
            return None
        return self[0]

    def size(self):
        return len(self)

    def __str__(self):
        return str(self)


class OrderBookError(Exception):
    pass


class SimpleOrderBook:
    def __init__(self, bid: Stack, ask: Stack, kafka_client: KafkaClient):
        self.bid = bid
        self.ask = ask
        self.lock = threading.Lock()
        self.kafka_client = kafka_client

        self._QUOTES_TOPIC = "orders.topic"
        self._ORDER_STATUS_TOPIC = "order.status.topic"
        self._PRICE_TOPIC = "order.last_price.topic"

    def match(self, order: dict):
        """
        Executes trades by matching buy orders with the lowest-priced ask
        and sell orders with the highest-priced bid.
        """
        with self.lock:
            is_buy = order["price"] > 0
            book, opposite_book = (self.ask, self.bid) if is_buy else (self.bid, self.ask)
            comparator = lambda p, q: p >= q if is_buy else abs(p) <= q
            action = "Buy" if is_buy else "Sell"
            while order["quantity"] > 0 and book and comparator(order["price"], book.peek()["price"]):
                best_order = book.peek()
                trade_quantity = min(order["quantity"], best_order["quantity"])
                print(f"trade: {action} {trade_quantity} @ {best_order['price']}")
                order["quantity"] -= trade_quantity
                best_order["quantity"] -= trade_quantity
                if best_order["quantity"] == 0:
                    book.pop()
            if order["quantity"] > 0:
                opposite_book.push(order)

    def publish_trade(self, order: dict):
        order_id = order["order_id"]
        if order["quantity"] == 0:
            status = "closed"
        else:
            status = "partial"
        self.kafka_client.produce(
            self._ORDER_STATUS_TOPIC,
            bytes(json.dumps({"order_id": order_id, "status": status}), "utf-8"),
        )
        print(f"{self._ORDER_STATUS_TOPIC}: {order_id} traded -> '{status}'")

    def publish_price(self, mid_price: float):
        message = bytes(json.dumps({"mid_price": mid_price}), "utf-8")
        self.kafka_client.produce(self._PRICE_TOPIC, message)
        print(f"{self._PRICE_TOPIC}: last price '{mid_price}'")

    def start(self):
        for msgs in self.kafka_client.consume(self._QUOTES_TOPIC):
            for order in msgs:
                self.match(order)
                self.publish_price(order["price"])
                self.publish_trade(order)

    def __str__(self):
        return f"OrderBook(bid: {self.bid}, ask: {self.ask})"
