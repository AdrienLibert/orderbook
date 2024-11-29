import json
from drgn.kafka import KafkaClient


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
        self.kafka_client = kafka_client

        self._QUOTES_TOPIC = "orders.topic"
        self._ORDER_STATUS_TOPIC = "order.status.topic"
        self._PRICE_TOPIC = "order.last_price.topic"

    def match(self, order: dict): #The match function executes trades by matching buy orders with the lowest-priced ask and sell orders with the highest-priced bid
        if order["price"] > 0:
            while (
                order["quantity"] > 0
                and self.ask
                and order["price"] >= self.ask.peek()["price"]
            ):
                best_ask = self.ask.peek()
                if best_ask:
                    min_trade_quantity = min(order["quantity"], best_ask["quantity"])
                    print(f"trade: Buy {min_trade_quantity} @ {best_ask['price']}")
                    order["quantity"] -= min_trade_quantity
                    best_ask["quantity"] -= min_trade_quantity
                    if best_ask["quantity"] == 0:
                        self.ask.pop()
            if order["quantity"] > 0:
                self.bid.push(order)
        elif order["price"] < 0:
            while (
                order["quantity"] > 0
                and self.bid
                and abs(order["price"]) <= self.bid.peek()["price"]
            ):
                best_bid = self.bid.peek()
                if best_bid:
                    min_trade_quantity = min(order["quantity"], best_bid["quantity"])
                    print(f"trade: Sell {min_trade_quantity} @ {best_bid['price']}")
                    order["quantity"] -= min_trade_quantity
                    best_bid["quantity"] -= min_trade_quantity
                    if best_bid["quantity"] == 0:
                        self.bid.pop()
            if order["quantity"] > 0:
                self.ask.push(order)

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
