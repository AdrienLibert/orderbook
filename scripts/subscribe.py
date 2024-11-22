# This script is intended to be modified, the goal is to have manual producer to kafka
# Feel free to edit values / whatever for your testing purposes
from confluent_kafka import Consumer, TopicPartition

from drgn.kafka import kafka_config
import time


# defined in topics_config.yaml. if changed, need to rebuild the image.
TOPIC = "orders.topic"


def consume_orders(consumer: Consumer):
    while True:
        msg = consumer.poll(1)

        if msg:
            if msg.error():
                print(msg.error)
            else:
                print(msg.timestamp(), msg.offset(), msg.value())
        else:
            print("no message")
        time.sleep(1)
        consumer.commit()


if __name__ == "__main__":
    # init thread level
    consumer = Consumer(
        kafka_config
        | {
            "group.id": "test",
            "on_commit": lambda err, topics: print(err, topics),
            "auto.offset.reset": "earliest",
        }
    )

    consumer.assign([TopicPartition(TOPIC, 0, 0)])

    consume_orders(consumer)
