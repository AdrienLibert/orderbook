import os
from configparser import ConfigParser


def load_config():
    config_path = os.getenv("CONFIG_PATH")
    config = ConfigParser()
    if not config_path:  # dev mode
        config.read("task.conf")
    else:
        config.read(config_path)
    return config


env_config = load_config()
