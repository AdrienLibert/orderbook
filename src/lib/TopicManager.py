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
        try:
            existing_topics = self.admin_client.list_topics().topics.keys()
        except Exception as e:
            raise ConnectionError("Failed to connect to Kafka broker. Ensure the broker is running and accessible.") from e
        target_state = []
        for topic in self.topics.get('topics', []):
            topic_name = list(topic.keys())[0] # Assuming each topic is a dictionary with one key-value pair
            if topic_name not in existing_topics:
                target_state.append((topic_name, TopicType.CREATE))
            else:
                target_state.append((topic_name, TopicType.UPDATE))
        target_topics = set([target[0] for target in target_state])
        topics_to_delete = set(existing_topics) - target_topics # Find topics to delete
        
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

    def config_topic(self, topic_name): # Get the configuration for a specific topic in the configuration file
        topic_config = None
        for tc in self.topics["topics"]:
            if topic_name in tc:
                topic_config = tc
                break
        return topic_config
    
    # Create the topic in Kafka
    def _create_topic(self, topic_name):   
            topic_config = self.config_topic(topic_name) # Get the configuration for the topic      
            topic_details = topic_config[topic_name] # Get the details for the topic
            new_topic = NewTopic( # Create a NewTopic object
                topic=topic_name,
                num_partitions=topic_details.get("partition", 1),
                replication_factor=1,
                config={k: str(v) for k, v in topic_details.items() if k not in ["partition"]} # Set the topic configuration (excluding partition)
            )
            try:
                self.admin_client.create_topics(new_topics=[new_topic]) # Create the topic in Kafka
                print(f"Topic '{topic_name}' created successfully.")
            except Exception as e:
                print(f"Failed to create topic '{topic_name}': {e}")
                
    # Update the topic in Kafka
    def _update_topic(self, topic_name):
        topic_config = self.config_topic(topic_name) # Get the configuration for the topic
        topic_details = topic_config[topic_name]
        configs = {k: str(v) for k, v in topic_details.items() if k not in ["partition"]} # Set the topic configuration (excluding partition)
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