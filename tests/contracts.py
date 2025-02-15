import json
import pytest
from pathlib import Path
from jsonschema.protocols import Validator
from jsonschema import Draft202012Validator

SCHEMA_PATH = Path("./schemas")

def test_contracts():
    for file in SCHEMA_PATH.iterdir():
        f = open(SCHEMA_PATH.name + "/" + file.name, "rb")
        schema = json.load(f)

        try:
            Draft202012Validator.check_schema(schema)
        except Exception as e:
            pytest.fail(f"Validator.check_schema raised an error: {type(e)} - (e)")