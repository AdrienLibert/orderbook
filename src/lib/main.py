from confluent_kafka.admin import AdminClient
from TopicManager import TopicManager

if __name__ == "__main__":

    admin_client = AdminClient({"bootstrap.servers": "bitnami-kafka:9092"})
    topic_manager = TopicManager('./src/lib/test.yaml', admin_client)
    topic_manager.sync_topics()
