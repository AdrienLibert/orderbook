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
        if any(o["order_id"] == order["order_id"] for o in self.bid + self.ask): # Change
            print("nn")
            return
        with self.lock:
            in_, out_, action, comparator, order["quantity"] = (
                (self.bid, self.ask, "buy", lambda x, y: x <= y, order["quantity"])
                if order["quantity"] > 0
                else (self.ask, self.bid, "sell", lambda x, y: x >= y, -order["quantity"])
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
                    f"Left Order Quantity: {order['quantity']}, Right Order Quantity: {out_order['quantity']} | "
                    f"Trader Left Id : {order['trader_id']}, Trader Right Id : {out_order['trader_id']}"
                )
                self.publish_trade(
                    trade_quantity,
                    order["trader_id"],
                    out_order["trader_id"],               
                    order["order_id"],
                    out_order["order_id"],
                    order["quantity"],
                    out_order["quantity"],
                    out_order["price"],
                )
                order["quantity"] -= trade_quantity
                out_order["quantity"] -= trade_quantity
                if out_order["quantity"] == 0:
                    out_.pop(0)
                    continue
                if order["quantity"] == 0 or out_order["quantity"] == 0:
                    self.publish_price(out_order["price"])
            if order["quantity"] > 0:
                in_.push(order)

    def publish_trade(
        self,
        trade_quantity: int,
        trader_left_id: int,
        trader_right_id: int,
        left_order_id: str,
        right_order_id: str,
        left_quantity: int,
        right_quantity: int,
        price: int,
    ):
        left_status = "closed" if left_quantity == trade_quantity else "partial"
        right_status = "closed" if right_quantity == trade_quantity else "partial"
        self.kafka_client.produce(
            self._ORDER_STATUS_TOPIC,
            bytes(
                json.dumps(
                    {
                        "trader_left_id": trader_left_id,
                        "trader_right_id": trader_right_id,
                        "left_order_id": left_order_id,
                        "right_order_id": right_order_id,
                        "left_quantity": left_quantity,
                        "right_quantity": right_quantity,
                        "price": price,
                        "left_status": left_status,
                        "right_status": right_status,
                    }
                ),
                "utf-8",
            ),
        )

    def publish_price(self, price):
        message = bytes(json.dumps({"last quote price": price}), "utf-8")
        self.kafka_client.produce(self._PRICE_TOPIC, message)

    def start(self):
        for msgs in self.kafka_client.consume(self._QUOTES_TOPIC):
            for order in msgs:
                self.match(order)

    def __str__(self):
        return f"OrderBook(bid: {self.bid}, ask: {self.ask})"
