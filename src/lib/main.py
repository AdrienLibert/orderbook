from confluent_kafka.admin import AdminClient
from TopicManager import TopicManager

if __name__ == "__main__":

    admin_client = AdminClient({"bootstrap.servers": "localhost:9092"})
    topic_manager = TopicManager(admin_client, 'config.yaml')
    topic_manager.sync_topics()
