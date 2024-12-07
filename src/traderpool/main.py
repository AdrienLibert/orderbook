from trader import Trader
from drgn.kafka import KafkaClient

def start():
    trader = Trader(kafka_client=KafkaClient())
    trader.start()
