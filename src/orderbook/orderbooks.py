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
        book, opposite_book, action,comparator = (self.ask, self.bid, "Buy") if order["quantity"] > 0 else (self.bid, self.ask, "Sell")
        comparator = lambda p, q: p >= q if order["quantity"] > 0 else p <= q
        while abs(order["quantity"]) > 0 and book and comparator(order["price"], book.peek()["price"]):
            with self.lock: #TODO: allocate specific thread for publish operations
                right_order = book.peek()
                trade_quantity = min(abs(order["quantity"]), right_order["quantity"])
                print(f"Executed trade: {action} {trade_quantity} @ {right_order["price"]} | "
                    f"Left Order ID: {order["order_id"]}, Right Order ID: {right_order["order_id"]} | "
                    f"Left Order Quantity: {order['quantity']}, Right Order Quantity: {right_order['quantity']}")
                if order["quantity"] > 0:
                    order["quantity"] -= trade_quantity
                else:
                    order["quantity"] += trade_quantity 
                right_order["quantity"] -= trade_quantity
                self.publish_trade(order["order_id"],right_order["order_id"],trade_quantity,right_order["price"],action)
                if right_order["quantity"] == 0:
                    book.pop()
            if abs(order["quantity"]) > 0:
                opposite_book.push(order)
            self.publish_price()

    def publish_trade(self, left_order_id : str, right_order_id : str, quantity : int, price : int, action : str):
        status = "closed" if quantity == 0 else "partial"
        self.kafka_client.produce(
            self._ORDER_STATUS_TOPIC,
            bytes(json.dumps({
                "left_order_id":left_order_id,
                "right_order_id": right_order_id,
                "quantity": quantity,
                "price": price,
                "action": action,
                "status": status,
            }), "utf-8"),
        )
        print(f"{self._ORDER_STATUS_TOPIC}: "
          f"Left Order ID: {left_order_id} "
          f"Right Order ID: {right_order_id} "
          f"Quantity: {quantity} "
          f"@ {price} "
          f"Action: {action} "
          f"Status: {status}")
        
    def calculate_mid_price(self):
        """
        Calculate the mid-price based on the best bid and best ask prices.
        """
        best_bid = self.bid.peek()["price"]
        best_ask = self.ask.peek()["price"]
        if best_bid is not None and best_ask is not None:
            return (best_bid + best_ask) / 2
        elif best_bid is not None:
            return best_bid
        elif best_ask is not None:
            return best_ask
        else:
            return None

    def publish_price(self):
        mid_price = self.calculate_mid_price()
        if mid_price is not None:
            message = bytes(json.dumps({"mid_price": mid_price}), "utf-8")
            self.kafka_client.produce(self._PRICE_TOPIC, message)
            print(f"{self._PRICE_TOPIC}: last price '{mid_price}'")
        else:
            print(f"{self._PRICE_TOPIC}: No valid mid-price to publish (order books are empty).")

    def start(self):
        for msgs in self.kafka_client.consume(self._QUOTES_TOPIC):
            for order in msgs:
                self.match(order)

    def __str__(self):
        return f"OrderBook(bid: {self.bid}, ask: {self.ask})"
