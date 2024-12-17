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

def start_trader(index):
    print(f"Starting trader {index}")
    trader = Trader(
        eqlbm=100,                
        limit_buy=105 + index,      # Slightly different limits for each trader
        limit_sell=95 - index,      
        aggressiveness_buy=0.2,   
        aggressiveness_sell=0.3,
        theta=-3.0,                
        kafka_client=KafkaClient()
    )   
    trader.start()

def start_multiple_traders(num_traders):
    threads = []
    for i in range(num_traders):
        t = threading.Thread(target=start_trader, args=(i,))
        t.start()
        threads.append(t)

    for t in threads:
        t.join()