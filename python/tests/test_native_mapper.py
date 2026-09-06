"""Real-library tests for the Python mapper lifecycle."""

from __future__ import annotations

from concurrent.futures import ThreadPoolExecutor
import os
import threading

import pytest

from spl_toolkit import SPLMapper
from spl_toolkit.exceptions import ConfigurationError, MapperNotFoundError


def mapper_kwargs() -> dict[str, str]:
    return {"library_path": os.environ["SPL_NATIVE_LIBRARY"]} if "SPL_NATIVE_LIBRARY" in os.environ else {}


def test_close_and_context_manager():
    mapper = SPLMapper(**mapper_kwargs())
    with mapper as active:
        assert active is mapper
        active.load_mappings([{"source": "src_ip", "target": "source_ip"}])
        assert active.map_query("search src_ip=1") == "search source_ip=1"
    mapper.close()
    with pytest.raises(MapperNotFoundError):
        mapper.map_query("search src_ip=1")


def test_json_context_array_matches():
    config = {"version":"1.0", "mappings":[], "rules":[{
        "id":"web", "enabled":True, "priority":1,
        "conditions":[{"type":"sourcetype","operator":"regex","value":"^web_"}],
        "mappings":[{"source":"src_ip","target":"client_ip"}]}]}
    with SPLMapper(config=config, **mapper_kwargs()) as mapper:
        assert mapper.map_query_with_context("search src_ip=1", {"sourcetype":["mail","web_access"]}) == "search client_ip=1"


def test_invalid_mapping_errors_are_recoverable():
    with SPLMapper(**mapper_kwargs()) as mapper:
        for _ in range(100):
            with pytest.raises(ConfigurationError):
                mapper.load_mappings([{"source": 1, "target": "bad"}])
        mapper.load_mappings([{"source": "src_ip", "target": "source_ip"}])
        assert mapper.map_query("search src_ip=1") == "search source_ip=1"


def test_shared_mapper_supports_concurrent_mapping_and_discovery():
    with SPLMapper(**mapper_kwargs()) as mapper:
        mapper.load_mappings([{"source": "src_ip", "target": "source_ip"}])

        def work(index: int) -> tuple[str, list[str]]:
            mapped = mapper.map_query(f"search src_ip={index}")
            fields = mapper.discover_query(f"search src_ip={index} dst_ip=2").input_fields
            return mapped, fields

        with ThreadPoolExecutor(max_workers=16) as pool:
            results = list(pool.map(work, range(256)))

    assert [mapped for mapped, _ in results] == [f"search source_ip={index}" for index in range(256)]
    assert all("src_ip" in fields and "dst_ip" in fields for _, fields in results)


def test_separate_mappers_construct_and_close_concurrently():
    def work(index: int) -> str:
        with SPLMapper(**mapper_kwargs()) as mapper:
            mapper.load_mappings([{"source": "src_ip", "target": f"source_ip_{index}"}])
            return mapper.map_query("search src_ip=1")

    with ThreadPoolExecutor(max_workers=16) as pool:
        results = list(pool.map(work, range(64)))

    assert results == [f"search source_ip_{index}=1" for index in range(64)]


def test_native_workload_can_race_with_multiple_close_callers():
    mapper = SPLMapper(**mapper_kwargs())
    mapper.load_mappings([{"source": "src_ip", "target": "source_ip"}])
    start = threading.Event()

    def operate() -> int:
        start.wait(timeout=5)
        completed = 0
        while True:
            try:
                mapper.map_query("search src_ip=1")
                mapper.discover_query("search src_ip=1")
                completed += 1
            except MapperNotFoundError:
                return completed

    with ThreadPoolExecutor(max_workers=16) as pool:
        workers = [pool.submit(operate) for _ in range(12)]
        closers = [pool.submit(mapper.close) for _ in range(4)]
        start.set()
        completed = [worker.result(timeout=10) for worker in workers]
        for closer in closers:
            assert closer.result(timeout=10) is None

    assert all(count >= 0 for count in completed)
    with pytest.raises(MapperNotFoundError):
        mapper.discover_query("search src_ip=1")
