import os
from configparser import ConfigParser


def config_from_env():
    config = ConfigParser()
    for key, value in os.environ.items():
        if key.startswith("OB__"):
            parts = key.split("__")
            if not len(parts) == 3:  # wrong pattern
                continue
            section = parts[1].lower()
            name = parts[2].lower()

            if not config.has_section(section):
                config.add_section(section)

            config[section][name] = value

    with open("/app/task.conf", "w+") as f:
        config.write(f)
