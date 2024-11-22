import json
from confluent_kafka import Consumer, TopicPartition, Producer
from drgn.kafka import kafka_config

class Stack:
    def __init__(self, is_bid=True):
        self._list = []
        self.is_bid = is_bid

    def is_empty(self):
        return len(self._list) == 0

    def sort(self):
        self._list.sort(key=lambda x: x["price"], reverse=self.is_bid)

    def push(self, price, quantity):
        self._list.append({"price": price, "quantity": quantity})
        self.sort()

    def pop(self):
        if self.is_empty():
            return None
        return self._list.pop(0)

    def peek(self):
        if self.is_empty():
            return None
        return self._list[0]

    def size(self):
        return len(self._list)

    def __str__(self):
        return str(self._list)

    def to_dict(self):
        return {"order": self._list}

    def __repr__(self):
        return json.dumps(self.to_dict())

class OrderBookError(Exception):
    pass

class MidPriceProducer:
    def __init__(self, topic: str):
        self.producer = Producer(kafka_config)
        self.topic = topic

    def produce_mid_price(self, bid_price, ask_price):
        if not bid_price:
            mid_price = ask_price
        elif not ask_price:
            mid_price = bid_price 
        else:
            mid_price = (bid_price + ask_price) / 2
        message = json.dumps({"mid_price": mid_price})
        self.producer.produce(self.topic, message)
        self.producer.flush()
        print(f"Produced mid_price: {mid_price} to topic {self.topic}")

class OrderStatusNotifier:
    def __init__(self, topic: str):
        self.producer = Producer(kafka_config)
        self.topic = topic

    def trade_quantity(self, order):
        if order["quantity"] == 0:
            self.notify_order_status(order["order_id"],"Sucessful")
        else :
            self.notify_order_status(order["order_id"],"Waiting")


    def notify_order_status(self, order_id: str, status:str):
        message = json.dumps({"order_id": order_id, "status": status})
        self.producer.produce(self.topic, message)
        self.producer.flush()
        print(f"Produced order status: {status} for order {order_id} to topic {self.topic}")

class OrderBook:
    def __init__(self, bid, ask):
        self.bid = bid
        self.ask = ask

    def get_bid(self):
        return self.bid

    def get_ask(self):
        return self.ask

    def _get_best_bid(self):
        return self.bid.peek() if not self.bid.is_empty() else None

    def _get_best_ask(self):
        return self.ask.peek() if not self.ask.is_empty() else None

    def add_order(self, order):
        print(order)
        if order["order_type"] == "buy":
            self.match_buy_order(order)
        elif order["order_type"] == "sell":
            self.match_sell_order(order)
        else:
            raise ValueError("Invalid order type")

        print("OrderBook state:")
        print(f"Bid: {self.bid}")
        print(f"Ask: {self.ask}")

    def match_buy_order(self, order):
        while (
            order["quantity"] > 0
            and not self.ask.is_empty()
            and order["price"] >= self.ask.peek()["price"]
        ):
            best_ask = self.ask.peek()
            min_trade_quantity = min(order["quantity"], best_ask["quantity"])
            print(f"Trade: Buy {min_trade_quantity} @ {best_ask['price']}")
           
            order["quantity"] -= min_trade_quantity
            best_ask["quantity"] -= min_trade_quantity
            if best_ask["quantity"] == 0:
                self.ask.pop()
        if order["quantity"] > 0:
            self.bid.push(order["price"], order["quantity"])

    def match_sell_order(self, order):
        while (
            order["quantity"] > 0
            and not self.bid.is_empty()
            and order["price"] <= self.bid.peek()["price"]
        ):
            best_bid = self.bid.peek()
            min_trade_quantity = min(order["quantity"], best_bid["quantity"])
            print(f"Trade: Sell {min_trade_quantity} @ {best_bid['price']}")
            order["quantity"] -= min_trade_quantity
            best_bid["quantity"] -= min_trade_quantity
            if best_bid["quantity"] == 0:
                self.bid.pop()
        if order["quantity"] > 0:
            self.ask.push(order["price"], order["quantity"])

    def __str__(self):
        return f"Bid: {self.bid}, Ask: {self.ask}"
    
class OrderbookKafka:
    def __init__(self, topic: str, group_id: str):
        self.consumer = Consumer(
            kafka_config
            | {
                "group.id": group_id,
                "auto.offset.reset": "latest",
                "on_commit": lambda err, topics: print(err, topics),
            }
        )
        self.topic = topic
        self.consumer.assign([TopicPartition(topic, 0, 0)])

    def listen_to_kafka(self, orderbook: OrderBook, mid_price_producer: MidPriceProducer, status_notifier: OrderStatusNotifier):
        print(f"Listening to Kafka topic '{self.topic}'...")
        try:
            while True:
                print("Polling Kafka...")
                msg = self.consumer.poll(1.0)
                
                if msg is None:
                    continue
                if msg.error():
                    print(f"Consumer error: {msg.error()}")
                    continue
                else:
                    order = self.parse_order(msg.value().decode("utf-8"))
                    if order:
                        orderbook.add_order(order)
                        mid_price_producer.produce_mid_price(orderbook._get_best_bid(), orderbook._get_best_ask())
                        status_notifier.trade_quantity(order)
                self.consumer.commit()
        except KeyboardInterrupt:
            print("Stopping consumer...")
        finally:
            self.consumer.close()
            
    @staticmethod
    def parse_order(message: str):
        try:
            return json.loads(message)
        except Exception as e:
            print(f"Failed to parse message: {message}. Error: {e}")
            return None
