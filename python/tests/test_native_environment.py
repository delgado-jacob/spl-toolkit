"""Environment validation through the installed Python and owned C surfaces."""

import json
import os
from pathlib import Path

import pytest

from spl_toolkit import SPLMapper, SPLMapperError
from spl_toolkit.exceptions import MapperNotFoundError


FIXTURES = Path(os.environ["SPL_ENVIRONMENT_FIXTURES"])
assert FIXTURES.is_absolute(), "SPL_ENVIRONMENT_FIXTURES must be absolute"
CASES = json.loads((FIXTURES / "cases.json").read_text(encoding="utf-8"))


def open_mapper():
    options = {"library_path": os.environ["SPL_NATIVE_LIBRARY"]} if "SPL_NATIVE_LIBRARY" in os.environ else {}
    return SPLMapper(**options)


def raw_report(mapper, raw):
    pointer = mapper._lib.spl_mapper_validate_environment(mapper._mapper_id, raw)
    assert pointer
    try:
        assert pointer.contents.error is None
        assert pointer.contents.result is not None
        return json.loads(pointer.contents.result)
    finally:
        mapper._lib.spl_result_free(pointer)


@pytest.mark.parametrize("case", CASES, ids=lambda case: case["name"])
def test_environment_native_python_parity_and_owned_copy(case):
    request = {"schema_version": 1, "snapshot": case["snapshot"], "schema_bundle": case["schema_bundle"]}
    with open_mapper() as mapper:
        report = mapper.validate_environment(request)
        assert report == raw_report(mapper, json.dumps(request).encode("utf-8"))
        assert report["status"] == case["status"]
        report["diagnostics"].append({"code": "changed"})
        assert mapper.validate_environment(request) == raw_report(mapper, json.dumps(request).encode("utf-8"))


def test_environment_inline_malformed_is_request_diagnostic():
    raw = b'{"schema_version":1,"snapshot":{"schema_version":1,},"schema_bundle":{}}'
    with open_mapper() as mapper:
        report = raw_report(mapper, raw)
        assert report["status"] == "invalid"
        assert report["diagnostics"][0]["artifact"] == "request"
        assert report["diagnostics"][0]["code"] == "request_invalid"


def test_environment_python_rejects_unserializable_request_and_closed_mapper():
    mapper = open_mapper()
    try:
        with pytest.raises(SPLMapperError, match="Invalid environment validation request JSON"):
            mapper.validate_environment({"schema_version": 1, "snapshot": object()})
    finally:
        mapper.close()
    with pytest.raises(MapperNotFoundError, match="Mapper is closed"):
        mapper.validate_environment({"schema_version": 1})
