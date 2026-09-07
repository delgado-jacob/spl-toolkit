from __future__ import annotations

import json
import os
from pathlib import Path
import socket
import subprocess
import time
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

import pytest

from spl_toolkit import SPLMapper, __version__


CATEGORIES = (
    "datamodels",
    "datasets",
    "lookups",
    "macros",
    "sources",
    "sourcetypes",
    "input_fields",
)


def required_absolute_path(name: str) -> Path:
    value = os.environ.get(name)
    if not value:
        pytest.fail(f"{name} is required")
    path = Path(value)
    if not path.is_absolute():
        pytest.fail(f"{name} must be an absolute path: {value}")
    if not path.exists():
        pytest.fail(f"{name} does not exist: {path}")
    return path


@pytest.fixture(scope="session")
def cli_path() -> Path:
    return required_absolute_path("SPL_CLI")


@pytest.fixture(scope="session")
def fixture_path() -> Path:
    return required_absolute_path("SPL_FIXTURES")


@pytest.fixture(scope="session")
def cases(fixture_path: Path) -> list[dict]:
    fixture = json.loads(fixture_path.read_text(encoding="utf-8"))
    assert fixture["version"] == "1"
    return fixture["cases"]


def available_port() -> int:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as listener:
        listener.bind(("127.0.0.1", 0))
        return listener.getsockname()[1]


def get_json(url: str) -> dict:
    with urlopen(url, timeout=1) as response:
        assert response.status == 200
        return json.load(response)


def stop_child(child: subprocess.Popen) -> None:
    if child.poll() is not None:
        return
    child.terminate()
    try:
        child.wait(timeout=5)
    except subprocess.TimeoutExpired:
        child.kill()
        child.wait(timeout=5)


@pytest.fixture(scope="session")
def server_url(tmp_path_factory: pytest.TempPathFactory):
    server = required_absolute_path("SPL_SERVER")
    last_error = "server did not start"
    for attempt in range(3):
        port = available_port()
        log_path = tmp_path_factory.mktemp(f"server-{attempt}") / "server.log"
        log = log_path.open("w+", encoding="utf-8")
        env = os.environ.copy()
        env["PORT"] = str(port)
        child = subprocess.Popen(
            [str(server)],
            cwd=log_path.parent,
            env=env,
            stdout=log,
            stderr=subprocess.STDOUT,
            text=True,
        )
        base_url = f"http://127.0.0.1:{port}/api/v1"
        deadline = time.monotonic() + 5
        while time.monotonic() < deadline:
            if child.poll() is not None:
                break
            try:
                get_json(base_url + "/health")
                try:
                    yield base_url
                finally:
                    stop_child(child)
                    log.close()
                return
            except (OSError, URLError, AssertionError):
                time.sleep(0.05)
        stop_child(child)
        log.seek(0)
        last_error = log.read()
        log.close()
    pytest.fail(f"server failed to start after 3 attempts: {last_error}")


def post_json(base_url: str, endpoint: str, payload: dict) -> tuple[int, dict]:
    request = Request(
        base_url + endpoint,
        data=json.dumps(payload).encode("utf-8"),
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urlopen(request, timeout=5) as response:
            return response.status, json.load(response)
    except HTTPError as error:
        return error.code, json.load(error)


def normalize_discovery(value: dict) -> dict[str, list[str]]:
    return {category: sorted(value.get(category) or []) for category in CATEGORIES}


def python_discovery(info) -> dict[str, list[str]]:
    return normalize_discovery(
        {
            "datamodels": info.data_models,
            "datasets": info.datasets,
            "lookups": info.lookups,
            "macros": info.macros,
            "sources": info.sources,
            "sourcetypes": info.source_types,
            "input_fields": info.input_fields,
        }
    )


def test_health_and_openapi_report_installed_version(server_url: str) -> None:
    health = get_json(server_url + "/health")
    assert health == {"status": "healthy", "version": __version__, "service": "spl-toolkit-api"}
    openapi = get_json(server_url + "/openapi.json")
    assert openapi["openapi"] == "3.1.0"
    assert openapi["info"]["version"] == __version__


def test_mapping_parity(cases: list[dict], cli_path: Path, tmp_path: Path, server_url: str) -> None:
    for case in (case for case in cases if "mapped" in case):
        with SPLMapper(config=case["config"]) as python_mapper:
            if "context" in case:
                python_result = python_mapper.map_query_with_context(case["query"], case["context"])
            else:
                python_result = python_mapper.map_query(case["query"])
        assert python_result == case["mapped"], case["id"]

        request = {"query": case["query"], "config": case["config"]}
        if "context" in case:
            request["context"] = case["context"]
        status, response = post_json(server_url, "/query/map", request)
        assert status == 200, case["id"]
        assert response["mapped_query"] == case["mapped"], case["id"]

        if "context" not in case:
            config_path = tmp_path / f"{case['id']}.json"
            config_path.write_text(json.dumps(case["config"]), encoding="utf-8")
            completed = subprocess.run(
                [str(cli_path), "map", "--config", str(config_path), "--query", case["query"]],
                check=False,
                capture_output=True,
            )
            assert (completed.returncode, completed.stdout, completed.stderr) == (
                0,
                case["mapped"].encode("utf-8") + b"\n",
                b"",
            ), case["id"]


def test_discovery_parity(cases: list[dict], cli_path: Path, server_url: str) -> None:
    for case in (case for case in cases if "discovery" in case):
        expected = normalize_discovery(case["discovery"])
        with SPLMapper() as python_mapper:
            assert python_discovery(python_mapper.discover_query(case["query"])) == expected, case["id"]

        status, response = post_json(server_url, "/query/discover", {"query": case["query"]})
        assert status == 200, case["id"]
        assert normalize_discovery(response["query_info"]) == expected, case["id"]

        completed = subprocess.run(
            [str(cli_path), "discover", "--query", case["query"], "--format", "json"],
            check=False,
            capture_output=True,
            text=True,
        )
        assert (completed.returncode, completed.stderr) == (0, ""), case["id"]
        assert normalize_discovery(json.loads(completed.stdout)) == expected, case["id"]
