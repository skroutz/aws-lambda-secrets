import pytest
import tftest
from pathlib import Path

import requests
import json

TERRAFORM_DIR = "tests/terraform"
CONTAINER_RESP = None
PLAN = None
# tf.apply(output=True, tf_non_interactive=True)
# tf.destroy()

@pytest.fixture
def plan():
    global CONTAINER_RESP
    global PLAN

    file_path = Path(__file__).resolve()
    base_dir = file_path.parent.parent.parent.absolute()
    tf = tftest.TerraformTest(tfdir=TERRAFORM_DIR, basedir=base_dir)
    tf.setup()
    # tf.apply(output=True, tf_non_interactive=True)

    if not PLAN:
        PLAN = tf.plan(output=True)

    lambda_url = PLAN.outputs['lambda-container-url']
    CONTAINER_RESP = requests.get(lambda_url)

    yield PLAN
    # tf.destroy(auto_approve=True, tf_non_interactive=True)

def test_connectivity(plan):
    assert CONTAINER_RESP.status_code == 200


def test_secret_plain(plan):
    resp_json = CONTAINER_RESP.json()

    assert resp_json["LAMBDASECRETS_PLAIN"] == plan.variables["secret-plain"]
    assert resp_json["LAMBDASECRETS_PLAIN_TYPE"] == "PLAIN"


def test_secret_json(plan):
    resp_json = CONTAINER_RESP.json()

    assert json.loads(resp_json["LAMBDASECRETS_JSON"]) == json.loads(plan.variables["secret-json"])
    assert resp_json["LAMBDASECRETS_JSON_TYPE"] == "JSON"


def test_secret_multiline(plan):
    resp_json = CONTAINER_RESP.json()

    assert resp_json["LAMBDASECRETS_MULTILINE"] == plan.variables["secret-multiline"]
    assert resp_json["LAMBDASECRETS_MULTILINE_TYPE"] == "PLAIN"


def test_secret_binary(plan):
    resp_json = CONTAINER_RESP.json()

    assert resp_json["LAMBDASECRETS_BINARY"] == plan.variables["secret-binary"]
    assert resp_json["LAMBDASECRETS_BINARY_TYPE"] == "BINARY"
