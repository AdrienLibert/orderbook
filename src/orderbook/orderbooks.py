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
        book, opposite_book, action = (self.ask, self.bid, "Buy") if order["quantity"] > 0 else (self.bid, self.ask, "Sell")
        comparator = lambda p, q: p >= q if order["quantity"] > 0 else p <= q
        while abs(order["quantity"]) > 0 and book and comparator(order["price"], book.peek()["price"]):
            with self.lock:
                best_order = book.peek()
                trade_quantity = min(abs(order["quantity"]), best_order["quantity"])
                left_order_id = order["order_id"]
                right_order_id = best_order["order_id"]
                print(f"trade: {action} {trade_quantity} @ {best_order["price"]}")
                if order["quantity"] > 0:
                    order["quantity"] -= trade_quantity
                else:
                    order["quantity"] += trade_quantity 
                best_order["quantity"] -= trade_quantity
                self.publish_trade(
                    {
                "left_order_id": left_order_id,
                "right_order_id": right_order_id,
                "quantity": trade_quantity,
                "price": best_order["price"],
                "action": action,
                    }
                )
                if best_order["quantity"] == 0:
                    book.pop()
            if abs(order["quantity"]) > 0:
                opposite_book.push(order)

    def publish_trade(self, trade: dict):
        status = "closed" if trade["quantity"] == 0 else "partial"
        self.kafka_client.produce(
            self._ORDER_STATUS_TOPIC,
            bytes(json.dumps({
                "left_order_id": trade["left_order_id"],
                "right_order_id": trade["right_order_id"],
                "quantity": trade["quantity"],
                "price": trade["price"],
                "action": trade["action"],
                "status": status,
            }), "utf-8"),
        )
        print(f"{self._ORDER_STATUS_TOPIC}: "
          f"Left Order ID: {trade['left_order_id']} "
          f"Right Order ID: {trade['right_order_id']} "
          f"Quantity: {trade['quantity']} "
          f"@ {trade['price']} "
          f"Action: {trade['action']} "
          f"Status: {status}")

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
