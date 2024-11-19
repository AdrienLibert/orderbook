from drgn.config import env_config

kafka_config = {
    "bootstrap.servers": env_config["kafka"]["bootstrap_servers"],
    "security.protocol": env_config["kafka"]["security_protocol"],
}

kafka_consumer_config = kafka_config | {
    "group.id": "orderbook_consumer_group",
    "auto.offset.reset": "earliest",   
}