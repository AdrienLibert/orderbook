from models import OrderBook, Stack, OrderbookKafka, MidPriceProducer, OrderStatusNotifier
from confluent_kafka import Consumer, TopicPartition
        
def main():
    orderbook = OrderBook(Stack(), Stack(False))
    mid_price_producer = MidPriceProducer(topic="order.last_price.topic")
    status_notifier = OrderStatusNotifier(topic="order.status.topic")
    kafka_orderbook = OrderbookKafka(topic="orders.topic", group_id="orderbook")
    kafka_orderbook.listen_to_kafka(orderbook, mid_price_producer, status_notifier)