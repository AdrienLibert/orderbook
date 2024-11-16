import yaml
from confluent_kafka.admin import AdminClient, NewTopic

def load_topics():
    with open('config.yaml', 'r') as file:
        topics = yaml.safe_load(file)
    return topics

def sync_topics(admin_client, topics_config):
    existing_topics = admin_client.list_topics(timeout=10).topics.keys()
    topics_create = []
    topics_to_delete = list(set(existing_topics) - set(topic_name for topic in topics_config.get("topics", []) for topic_name in topic.keys()))

    for topic_config in topics_config.get("topics", []):
        topic_name = list(topic_config.keys())[0]
        topic_details = topic_config[topic_name]
        if topic_name not in existing_topics:
            new_topic = NewTopic(
                topic=topic_name,
                num_partitions=topic_details.get("partition", 1),
                replication_factor=1,
                config={k: str(v) for k, v in topic_details.items() if k not in ["partition"]}
            )
            topics_create.append(new_topic)
    if topics_create:
        futures = admin_client.create_topics(new_topics=topics_create)
        for topic, future in futures.items():
            try:
                future.result()
                print(f"Created topic: {topic}")
            except Exception as e:
                print(f"Failed to create topic {topic}: {e}")
    if topics_to_delete:
        futures = admin_client.delete_topics(topics=topics_to_delete, operation_timeout=30)
        for topic, future in futures.items():
            try:
                future.result()
                print(f"Deleted topic: {topic}")
            except Exception as e:
                print(f"Failed to delete topic {topic}: {e}")

def main():
    topics_config = load_topics()
    admin_client = AdminClient({"bootstrap.servers": "localhost:9092"})
    sync_topics(admin_client, topics_config)

if __name__ == "__main__":
    main()