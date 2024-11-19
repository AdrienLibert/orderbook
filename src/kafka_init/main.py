import yaml
from confluent_kafka.admin import AdminClient, NewTopic, TopicMetadata
from enum import StrEnum
from drgn.config import env_config
from drgn.kafka import kafka_config


def load_yaml(path: str) -> dict:
    with open(path, "r") as file:
        return yaml.safe_load(file)


class TopicConfig:
    def __init__(self, topic_dict: dict):
        self.topic_name = list(topic_dict.keys())[0]
        self.partitions = topic_dict[self.topic_name]["partition"]


class TopicAction(StrEnum):
    CREATE = "CREATE"
    READ = "READ"
    UPDATE = "UPDATE"
    DELETE = "DELETE"


class KafkaTopicSynchronizer:
    def __init__(self, admin_client: AdminClient, topics_config: list[TopicConfig]):
        self.admin_client = admin_client
        self.topics_config = topics_config

        self.__timeout = 2.0

    def get_config_state(self) -> dict[str, TopicConfig]:
        return {topic.topic_name: topic for topic in self.topics_config}

    def get_current_state(self) -> dict[str, TopicMetadata]:
        return self.admin_client.list_topics(timeout=self.__timeout).topics

    def get_target_state(
        self,
        config_state: dict[str, TopicConfig],
        current_state: dict[str, TopicMetadata],
    ):
        target_state = []

        for topic, topic_metadata in current_state.items():
            if topic not in config_state:
                target_state.append((topic, TopicAction.DELETE, None))
            else:
                if (
                    len(topic_metadata.partitions.keys())
                    != config_state[topic].partitions
                ):
                    target_state.append((topic, TopicAction.DELETE, None))
                    target_state.append(
                        (topic, TopicAction.CREATE, config_state[topic])
                    )
                else:
                    target_state.append((topic, TopicAction.READ, None))

        for topic, config in config_state.items():
            if topic not in current_state:
                target_state.append((topic, TopicAction.CREATE, config))

        return target_state

    def run(self):
        config_state = self.get_config_state()
        current_state = self.get_current_state()
        target_state = self.get_target_state(config_state, current_state)
        print(f"Target state and actions: {target_state}")
        for topic, action, config in target_state:
            if action == TopicAction.CREATE:
                self.create_topic(topic, config)
            elif action == TopicAction.UPDATE:
                self.update_topic(topic, config)
            elif action == TopicAction.DELETE:
                self.delete_topic(topic)
            elif action == TopicAction.READ:
                print(f"topic {topic} is already in target configuration")

    def create_topic(self, topic: str, config: TopicConfig):
        new_topic = NewTopic(
            topic=topic,
            num_partitions=config.partitions,
            replication_factor=1,
            config={},
        )  # add potential other configs here
        self.admin_client.create_topics(
            new_topics=[new_topic],
            operation_timeout=self.__timeout,
            request_timeout=self.__timeout,
        )
        print(f"topic '{topic}' created succesfully")

    def update_topic(self, topic: str, config: TopicConfig):
        raise NotImplementedError()

    def delete_topic(self, topic: str):
        self.admin_client.delete_topics([topic])
        print(f"topic '{topic}' deleted successfully.")


def main():
    config = load_yaml(env_config["kafka"]["topics_config"])
    topics_config = [TopicConfig(topic) for topic in config["topics"]] if config else []

    admin_client = AdminClient(kafka_config)
    topic_manager = KafkaTopicSynchronizer(admin_client, topics_config)
    topic_manager.run()
