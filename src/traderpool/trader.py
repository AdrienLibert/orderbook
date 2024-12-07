import math
import random
import uuid
import json
from datetime import datetime, timezone
from drgn.kafka import KafkaClient
import math

class Trader:
    def __init__(self, eqlbm, limit_buy, limit_sell, aggressiveness_buy, aggressiveness_sell, theta, marketMax, kafka_client: KafkaClient):
        self.eqlbm = eqlbm
        self.limit_buy = limit_buy 
        self.limit_sell = limit_sell
        self.aggressiveness_buy = aggressiveness_buy
        self.aggressiveness_sell = aggressiveness_sell
        self.theta = theta
        self.marketMax = marketMax
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
        else:
            self.target_buy = self.limit_buy * (
                1 - (math.exp(-self.aggressiveness_buy * self.theta) - 1) / (math.exp(self.theta) - 1)
            )

        if self.limit_sell < self.eqlbm:
            self.target_sell = self.eqlbm + (self.limit_sell - self.eqlbm) * (
                (math.exp(-self.aggressiveness_sell * self.theta) - 1) / (math.exp(self.theta) - 1)
            )
        else: 
            self.target_sell = self.limit_sell + (self.marketMax - self.limit_sell) * (
                (math.exp(self.aggressiveness_sell * self.theta) - 1) / (math.exp(self.theta) - 1)
            )

    def produce_buy_order(self):
        msg = {
            "order_id": str(uuid.uuid4()),
            "order_type": "limit",
            "side": "buy",
            "price": self.target_buy,
            "quantity": self.buy_quantity,
            "time": int(datetime.now(timezone.utc).timestamp() * 1e9),
        }
        self.kafka_client.produce(self._QUOTES_TOPIC, bytes(json.dumps(msg),"utf-8"))
        self.kafka_client.flush()


    def produce_sell_order(self):
        msg = {
            "order_id": str(uuid.uuid4()),
            "order_type": "limit",
            "side": "sell",
            "price": self.target_sell,
            "quantity": -self.sell_quantity, 
            "time": int(datetime.now(timezone.utc).timestamp() * 1e9),
        }
        self.kafka_client.produce(self._QUOTES_TOPIC, bytes(json.dumps(msg),"utf-8"))
        self.kafka_client.flush()

    def consume_last_price(self):
        self.kafka_client.subscribe([self._PRICE_TOPIC])
        while True:
            msg = self.kafka_client.poll(1.0)
            if msg is None:
                continue
            data = json.loads(msg.value().decode('utf-8'))
            print(f"Received price update: {data}")

    def consume_trade(self):
        self.kafka_client.subscribe([self._ORDER_STATUS_TOPIC])
        while True:
            msg = self.kafka_client.poll(1.0)
            if msg is None:
                continue
            data = json.loads(msg.value().decode('utf-8'))
            print(f"Received trade update: {data}")
    
    def start(self):
        self.produce_buy_order()