import math
import random
import time
import uuid
import json
from datetime import datetime, timezone
from drgn.kafka import KafkaClient

market_max = 105

# According AA strategy

class Trader:

    _QUOTES_TOPIC = "orders.topic"
    _ORDER_STATUS_TOPIC = "order.status.topic"
    _PRICE_TOPIC = "order.last_price.topic"
    
    def __init__(self, trader_id: int, limit_price: int, quantity: int, kafka_client: KafkaClient):
        self.trader_id = trader_id
        self.limit_price = limit_price
        self.quantity = quantity
        self.kafka_client = kafka_client

        self.n_last_trades = 5
        self.ema_param = 2 / float(self.n_last_trades + 1)
        self.max_newton_iter = 10
        self.max_newton_error = 0.0001

        self.equilibrium_price = None
        self.theta = -1.0 * (5.0 * random.random())
        self.last_trades = []
        self.aggressiveness_buy = -1.0 * (0.3 * random.random())
        self.aggressiveness_sell = -1.0 * (0.3 * random.random())

    def newton_method(self): #Newton-Raphson
        theta_est = self.theta
        right_side = (self.theta * (self.limit_price - self.equilibrium_price)) / (math.exp(self.theta) - 1)
        max_, right_side = (0, (self.theta * (self.limit_price - self.equilibrium_price)) / (math.exp(self.theta) - 1)) if self.quantity > 0 else (market_max, (self.theta * (self.equilibrium_price - self.limit_price)) / (math.exp(self.theta) - 1))
        for _ in range(self.max_newton_iter):
            exp_theta_est = math.exp(theta_est)
            exp_theta_est_minus_one = exp_theta_est - 1
            f_theta = (theta_est * (max_ - self.equilibrium_price))/ exp_theta_est_minus_one - right_side #f(θ)= θ⋅eqlbm / exp(θ)−1 −right_side = 0
            if abs(f_theta) <= self.max_newton_error:
                break
            df_theta = (max_ - self.equilibrium_price / exp_theta_est_minus_one) - (
                exp_theta_est * (max_ - self.equilibrium_price) * theta_est
            ) / (exp_theta_est_minus_one ** 2)
            theta_est -= f_theta / df_theta
        return theta_est if theta_est != 0 else 0.000001

    def update_target_prices(self):
        def calculate_target(limit_price: int, equilibrium_price: int, aggressiveness: int, theta: int, is_buy=True):
            if limit_price < equilibrium_price:
                if aggressiveness >= 0:
                    return limit_price
                else:
                    factor = (math.exp(-aggressiveness * theta) - 1) / (math.exp(theta) - 1)
                    return limit_price * (1 - factor) if is_buy else limit_price + (market_max - limit_price) * factor
            else:
                if aggressiveness >= 0:
                    factor = (math.exp(aggressiveness * theta) - 1) / (math.exp(theta) - 1)
                    return equilibrium_price + (limit_price - equilibrium_price) * factor if is_buy else \
                        limit_price + (equilibrium_price - limit_price) * (1 - factor)
                else:
                    theta_est = self.newton_method()
                    factor = (math.exp(-aggressiveness * theta_est) - 1) / (math.exp(theta_est) - 1)
                    return equilibrium_price * (1 - factor) if is_buy else \
                        equilibrium_price + (market_max - equilibrium_price) * factor

        self.equilibrium_price = (
            self.limit_price if self.equilibrium_price is None
            else self.ema_param * self.limit_price + (1 - self.ema_param) * self.equilibrium_price
        ) #EMA

        self.target_buy = calculate_target(
            self.limit_price, self.equilibrium_price, self.aggressiveness_buy, self.theta, is_buy=True
        )
        self.target_sell = calculate_target(
            self.limit_price, self.equilibrium_price, self.aggressiveness_sell, self.theta, is_buy=False
        )

    def produce_order(self):
        self.update_target_prices()
        msg = {
            "trader_id": self.trader_id,
            "order_id": str(uuid.uuid4()),
            "order_type": "limit",
            "price": self.target_buy if self.quantity > 0 else self.target_sell,
            "quantity": self.quantity,
            "time": int(datetime.now(timezone.utc).timestamp() * 1e9),
        }
        print(msg)
        self.kafka_client.produce(self._QUOTES_TOPIC, bytes(json.dumps(msg),"utf-8"))
        time.sleep(15)
        self.consume_trade()

    def consume_last_price(self):
        for msgs in self.kafka_client.consume(self._PRICE_TOPIC):
            for last_price in msgs:
                print(f"Received price update: {last_price}")

    def consume_trade(self):
        for msgs in self.kafka_client.consume(self._ORDER_STATUS_TOPIC):
            for msg in msgs:
                while (msg["trader_left_id"] == self.trader_id and msg["left_status"] != "closed") or \
                (msg["trader_right_id"] == self.trader_id and msg["right_status"] != "closed"):
                    time.sleep(15)
                self.quantity = -1 * random.randint(5,self.quantity) if self.quantity > 0 else random.randint(5,abs(self.quantity))
                self.produce_order()

    def start(self):
        self.produce_order()