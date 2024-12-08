from trader import Trader
from drgn.kafka import KafkaClient

def start():
    trader = Trader(
        eqlbm=100,                
        limit_buy=105,            
        limit_sell=95,           
        aggressiveness_buy=0.2,   
        aggressiveness_sell=0.3,
        theta=-3.0,               
        kafka_client=KafkaClient()
    )

    trader.update_target_prices()
    trader.start()

    print("Target Buy Price:", trader.target_buy)
    print("Target Sell Price:", trader.target_sell)

