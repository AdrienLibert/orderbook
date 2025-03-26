from drgn.kafka import kafka_config
from confluent_kafka.admin import AdminClient

if __name__ == "__main__":
    admin_client = AdminClient(kafka_config)

    print(admin_client.list_topics().topics)