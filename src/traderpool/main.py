from trader import Trader
from drgn.kafka import KafkaClient
import threading

# def start():
#     trader = Trader(
#         eqlbm=100,                
#         limit_buy=105,            
#         limit_sell=95,           
#         aggressiveness_buy=0.2,   
#         aggressiveness_sell=0.3,
#         theta=-3.0,               
#         kafka_client=KafkaClient()
#     )   
#     trader.start()

def start_trader(index,action):
    print(f"Starting trader {index}")
    trader = Trader(
        id = index,
        first_action=action,
        eqlbm=102,                
        limit_buy=98 - index,      # Slightly different limits for each trader
        limit_sell=100 + index,      
        aggressiveness_buy=0.02,   
        aggressiveness_sell=0.03,
        theta=-3.0,                
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

    for t in threads:
        t.join()