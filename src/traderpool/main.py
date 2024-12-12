from trader import Trader
from drgn.kafka import KafkaClient

def start():
    trader1 = Trader(
        eqlbm=100,                
        limit_buy=105,            
        limit_sell=95,           
        aggressiveness_buy=0.2,   
        aggressiveness_sell=0.3,
        theta=-3.0,               
        kafka_client=KafkaClient()
    )
    
    trader2 = Trader(
        eqlbm=100,                
        limit_buy=107,            
        limit_sell=98,           
        aggressiveness_buy=0.6,   
        aggressiveness_sell=0.8,
        theta=-3.0,               
        kafka_client=KafkaClient()
    )
    trader1.start()

    trader2.start()

    print("Target Buy Price:", trader1.target_buy)
    print("Target Sell Price:", trader1.target_sell)
    print("Target Buy Price:", trader2.target_buy)
    print("Target Sell Price:", trader2.target_sell)

