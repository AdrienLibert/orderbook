from trader import Trader
import random
import time
from drgn.kafka import KafkaClient
import threading

def start_trader(index,action):
    print(f"Starting trader {index}")
    trader = Trader(
        trader_id = index,
        limit_price=random.randint(95, 105),
        quantity= random.randint(5, 20) * action,
        kafka_client=KafkaClient()
    )   
    trader.start()

def start_multiple_traders(num_traders):
    threads = []
    j = -1
    for i in range(num_traders):
        j = -1 * j
        t = threading.Thread(target=start_trader, args=(i,j,))
        t.start()
        threads.append(t)
        time.sleep(3)

    for t in threads:
        t.join()