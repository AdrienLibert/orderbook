from enum import Enum
from TopicLoader import TopicLoader
from confluent_kafka.admin import NewTopic

# Enum to represent the state of a topic: Create, Update, or Delete
class TopicType(Enum):
    CREATE = 'Create'
    UPDATE = 'Update'
    DELETE = 'Delete'

class TopicManager:
    # Initialize the TopicManager with an admin client and configuration file.
    def __init__(self, config_file, admin_client):
        self.config_file = config_file
        self.admin_client = admin_client
        self.topics = TopicLoader(config_file).load_topics()
        
    # Determine the target state (CREATE, UPDATE, DELETE) for each topic.
    def manage_topics(self):
        existing_topics = self.admin_client.list_topics().topics.keys()
        target_state = []
        for topic in self.topics.get('topics', []):
            topic_name = list(topic.keys())[0]
            if topic_name not in existing_topics:
                target_state.append((topic_name, TopicType.CREATE))
            else:
                target_state.append((topic_name, TopicType.UPDATE))
        topics_to_delete = set(existing_topics) - set([topic[0] for topic in target_state])
        
        for topic in topics_to_delete:
            target_state.append((topic, TopicType.DELETE))
        return target_state    # List of topics with their target states

    # Synchronize the topics with the target state
    def synchrone_topics(self):
        target_state = self.manage_topics()
        for topic, action in target_state:
            if action == TopicType.CREATE:
                self.create_topic(topic)
            elif action == TopicType.UPDATE:
                self.update_topic(topic)
            elif action == TopicType.DELETE:
                self.delete_topic(topic)
            pass

    def config_topic(self, topic_name):
        topic_config = None
        for tc in self.topics_config["topics"]:
            if topic_name in tc:
                topic_config = tc
                break
        return topic_config
    
    # Create the topic in Kafka
    def _create_topic(self, topic_name):   
            topic_config = self.config_topic(topic_name)       
            topic_details = topic_config[topic_name]
            new_topic = NewTopic(
                topic=topic_name,
                num_partitions=topic_details.get("partition", 1),
                replication_factor=1,
                config={k: str(v) for k, v in topic_details.items() if k not in ["partition"]}
            )
            try:
                self.admin_client.create_topics(new_topics=[new_topic])
                print(f"Topic '{topic_name}' created successfully.")
            except Exception as e:
                print(f"Failed to create topic '{topic_name}': {e}")
                
    # Update the topic in Kafka
    def _update_topic(self, topic_name):
        topic_config = self.config_topic(topic_name)       
        topic_details = topic_config[topic_name]
        configs = {k: str(v) for k, v in topic_details.items() if k not in ["partition"]}
        try:
            self.admin_client.alter_configs({"topic": topic_name, "config": configs})
            print(f"Topic '{topic_name}' updated successfully.")
        except Exception as e:
            print(f"Failed to update topic '{topic_name}': {e}")

    # Delete the topic from Kafka
    def _delete_topic(self, topic_name):
        try:
            self.admin_client.delete_topics([topic_name])
            print(f"Topic '{topic_name}' deleted successfully.")
        except Exception as e:
            print(f"Failed to delete topic '{topic_name}': {e}")