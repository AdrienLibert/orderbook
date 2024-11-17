import yaml


class TopicLoader:

    def __init__(self, config_file):
        self.config_file = config_file

    def load_topics(self):
        try:
            with open(self.config_file, 'r') as file:
                topics = yaml.safe_load(file)
            return topics
        except FileNotFoundError:
            raise print(f"File {self.config_file} not found")
