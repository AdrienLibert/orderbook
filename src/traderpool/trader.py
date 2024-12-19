import math
import random
import uuid
import json
from datetime import datetime, timezone
from drgn.kafka import KafkaClient
import time

class Trader:
    def __init__(self, id: int ,first_action: int,eqlbm: float, limit_buy: float, limit_sell: float, aggressiveness_buy: float, aggressiveness_sell: float, theta: float, kafka_client: KafkaClient):
        self.id = id
        self.first_action = first_action
        self.eqlbm = eqlbm
        self.limit_buy = limit_buy 
        self.limit_sell = limit_sell
        self.aggressiveness_buy = aggressiveness_buy
        self.aggressiveness_sell = aggressiveness_sell
        self.theta = theta
        self.kafka_client = kafka_client

        self.buy_quantity = 10
        self.sell_quantity = 10

        self._QUOTES_TOPIC = "orders.topic"
        self._ORDER_STATUS_TOPIC = "order.status.topic"
        self._PRICE_TOPIC = "order.last_price.topic"

        self.target_buy = None 
        self.target_sell = None 

    def update_target_prices(self):
        if self.theta - 1 == 0:
            raise ValueError("Theta cannot be 1 to avoid division by zero.")

        if self.limit_buy > self.eqlbm:
            self.target_buy = self.eqlbm + (self.limit_buy - self.eqlbm) * (
                (math.exp(self.aggressiveness_buy * self.theta) - 1) / (math.exp(self.theta) - 1)
            )
            print(self.target_buy)
        else:
            self.target_buy = self.limit_buy * (
                1 - (math.exp(-self.aggressiveness_buy * self.theta) - 1) / (math.exp(self.theta) - 1)
            )
            print(self.target_buy)

        if self.limit_sell < self.eqlbm:
            self.target_sell = self.eqlbm + (self.limit_sell - self.eqlbm) * (
                (math.exp(-self.aggressiveness_sell * self.theta) - 1) / (math.exp(self.theta) - 1)
            )
            print(self.target_sell)
        else: 
            self.target_sell = self.limit_sell * (
                (math.exp(self.aggressiveness_sell * self.theta) - 1) / (math.exp(self.theta) - 1)
            )
            print(self.target_sell)

    def produce_buy_order(self):
        msg = {
            "trader_id": self.id,
            "order_id": str(uuid.uuid4()),
            "order_type": "limit",
            "side": "buy",
            "price": self.target_buy,
            "quantity": self.buy_quantity,
            "time": int(datetime.now(timezone.utc).timestamp() * 1e9),
        }
        self.kafka_client.produce(self._QUOTES_TOPIC, bytes(json.dumps(msg),"utf-8"))
        self.consume_trade()


    def produce_sell_order(self):
        msg = {
            "trader_id": self.id,
            "order_id": str(uuid.uuid4()),
            "order_type": "limit",
            "side": "sell",
            "price": self.target_sell,
            "quantity": -self.sell_quantity, 
            "time": int(datetime.now(timezone.utc).timestamp() * 1e9),
        }
        self.kafka_client.produce(self._QUOTES_TOPIC, bytes(json.dumps(msg),"utf-8"))
        self.consume_trade()


    def consume_last_price(self):
        for msgs in self.kafka_client.consume(self._PRICE_TOPIC):
            for last_price in msgs:
                print(f"Received price update: {last_price}")

    def consume_trade(self):
        for msgs in self.kafka_client.consume(self._ORDER_STATUS_TOPIC):
            for msg in msgs:
                print(msg)
                if msg["trader_id"] == self.id:
                    status = msg["status"]
                    print(f"Received Trade Update:")
                    print(f"Status: {status}")
                    while status != "closed":
                        time.sleep(1)
                        self.consume_trade()
                    self.consume_last_price()  
                    if msg["action"] == "buy":
                        self.produce_sell_order()
                    else:
                        self.produce_sell_order()                    

    def start(self):
        while True:
            self.update_target_prices()
            print("bug")
            if self.first_action == 1:
                self.produce_buy_order()
            else:
                self.produce_sell_order()
            time.sleep(10)