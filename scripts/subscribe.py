# This script is intended to be modified, the goal is to have manual producer to kafka
# Feel free to edit values / whatever for your testing purposes
from confluent_kafka import Consumer, TopicPartition, KafkaError

from drgn.kafka import kafka_config
import time
import uuid


# defined in topics_config.yaml. if changed, need to rebuild the image.
TOPIC = "orders.topic"

def commit(err, topics):
    print(f"error on commit: {err}: {topics}")

def consume_orders(consumer: Consumer):
    while True:
        msgs = consumer.consume()

        if len(msgs) > 0:
            for msg in msgs:
                if msg.error():
                    code = msg.error().code() 
                    if code == KafkaError._PARTITION_EOF:
                        print(f"No more msgs")
                    else:
                        print(code)
                else:
                    print(msg.timestamp(), msg.offset(), msg.value())
                    consumer.commit(message=msg)
        else:
            print("no message")
        time.sleep(0.5)


if __name__ == "__main__":
    # init thread level
    consumer = Consumer(
        kafka_config
        | {
            "group.id": str(uuid.uuid4()),
            "on_commit": commit,
            "enable.auto.commit": False,
            "enable.partition.eof": True,
        }
    )
    consumer.assign([TopicPartition(TOPIC, 0, -1)])

    consume_orders(consumer)