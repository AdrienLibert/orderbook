import datetime
import json

from pathlib import Path
from jsonschema import validate

SCHEMA_PATH = Path("./schemas")

ts = int(datetime.datetime.now(datetime.timezone.utc).timestamp() * 1000)

valid_payload = {
    "order": {
        "order_id": "abc-abc",
        "order_type": "Sell",
        "price": 10.0,
        "quantity": 10.0,
        "timestamp": ts,
    },
    "pricepoint": {"price": 10.0},
    "trade": {
        "trade_id": "abc-123",
        "order_id": "abc-abc",
        "quantity": 10.0,
        "price": 10.0,
        "action": "buy",
        "status": "closed",
        "timestamp": ts,
    },
}


def test_contracts():
    for file in SCHEMA_PATH.iterdir():
        f = open(SCHEMA_PATH.name + "/" + file.name, "rb")
        schema = json.load(f)
        schema_name = schema["metadata"]["name"]

        try:
            validate(valid_payload[schema_name], schema)
        except Exception as e:
            raise Exception(
                f"Unable to validate '{schema_name}' defined from {file}: {e}"
            )
