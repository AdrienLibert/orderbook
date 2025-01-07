from trader import Trader
import random
import time
import os
from drgn.kafka import KafkaClient
import threading

def start():
    threads = []
    j = -1
    for i in range(int(os.getenv('NUM_TRADERS', 10))):
        j = -1 * j
        def trader_task(index, action):
            trader = Trader(
                trader_id=index,
                limit_price=random.randint(95, 105),
                quantity=random.randint(5, 20) * action,
                kafka_client=KafkaClient()
            )
            trader.start()
        t = threading.Thread(target=trader_task, args=(i, j))
        t.start()
        threads.append(t)
        time.sleep(3)
    for t in threads:
        t.join()
