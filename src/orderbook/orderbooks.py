# Simple OrderBook in Python using a Stack-like structure to manage buy and sell orders in bid and ask

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

    def size(self):
        return len(self)

    def __str__(self):
        return super().__str__()


class OrderBookError(Exception):
    pass


class SimpleOrderBook:

    _QUOTES_TOPIC = "orders.topic"
    _ORDER_STATUS_TOPIC = "order.status.topic"
    _PRICE_TOPIC = "order.last_price.topic"

    def __init__(self, bid: Stack, ask: Stack, kafka_client: KafkaClient):
        self.bid = bid
        self.ask = ask
        self.lock = threading.Lock()
        self.kafka_client = kafka_client

    def match(self, order: dict):
        print(order)
        in_, out_, action, comparator, order["quantity"] = (
            (self.bid, self.ask, "Buy", lambda x, y: x <= y, order["quantity"])
            if order["quantity"] > 0
            else (self.ask, self.bid, "Sell", lambda x, y: x >= y, -order["quantity"])
        )
        while (
            order["quantity"] > 0
            and out_
            and comparator(order["price"], out_[0]["price"])
        ):
            out_order = out_[0]  # can pop() for thread safe operations
            trade_quantity = min(order["quantity"], out_order["quantity"])
            print(
                f"Executed trade: {action} {trade_quantity} @ {out_order['price']} | "
                f"Left Order ID: {order['order_id']}, Right Order ID: {out_order['order_id']} | "
                f"Left Order Quantity: {order['quantity']}, Right Order Quantity: {out_order['quantity']}"
            )
            order["quantity"] -= trade_quantity
            out_order["quantity"] -= trade_quantity
            self.publish_trade(
                order["trader_id"],
                order["order_id"],
                out_order["order_id"],
                trade_quantity,
                out_order["price"],
                action,
            )
            if out_order["quantity"] == 0:
                out_.pop()
            if order["quantity"] == 0 or out_order["quantity"] == 0:
                self.publish_price(out_order["price"])
        if order["quantity"] > 0:
            in_.push(order)

    def publish_trade(
        self,
        trader_id: int,
        left_order_id: str,
        right_order_id: str,
        quantity: int,
        price: int,
        action: str,
    ):
        status = "closed" if quantity == 0 else "partial"
        self.kafka_client.produce(
            self._ORDER_STATUS_TOPIC,
            bytes(
                json.dumps(
                    {
                        "trader_id": trader_id,
                        "left_order_id": left_order_id,
                        "right_order_id": right_order_id,
                        "quantity": quantity,
                        "price": price,
                        "action": action,
                        "status": status,
                    }
                ),
                "utf-8",
            ),
        )

    def publish_price(self, price):
        message = bytes(json.dumps({"last quote price": price}), "utf-8")
        self.kafka_client.produce(self._PRICE_TOPIC, message)
        print(f"{self._PRICE_TOPIC}: last quote price '{price}'")

    def start(self):
        for msgs in self.kafka_client.consume(self._QUOTES_TOPIC):
            for order in msgs:
                self.match(order)

    def __str__(self):
        return f"OrderBook(bid: {self.bid}, ask: {self.ask})"
