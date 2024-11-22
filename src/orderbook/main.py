from models import OrderBook, Stack, OrderbookKafka
from confluent_kafka import Consumer, TopicPartition
        
def main():
    orderbook = OrderBook(Stack(), Stack(False))
    kafka_orderbook = OrderbookKafka(topic="orders.topic", group_id="orderbook")
    kafka_orderbook.listen_to_kafka(orderbook)