"""Owned native closure result and Python wrapper behavior."""

from copy import deepcopy
import json
import os
from pathlib import Path

import pytest

from spl_toolkit import SPLMapper, SPLMapperError


ROOT = Path(__file__).resolve().parents[2]
CASES = json.loads((ROOT / "testdata/closure/cases.json").read_text(encoding="utf-8"))["cases"]


def mapper_kwargs():
    if "SPL_NATIVE_LIBRARY" in os.environ:
        return {"library_path": os.environ["SPL_NATIVE_LIBRARY"]}
    built = ROOT / "build/libspl_toolkit.dylib"
    return {"library_path": str(built)} if built.is_file() else {}


def test_closure_native_owned_result_and_errors():
    request = deepcopy(CASES[0]["request"])
    with SPLMapper(**mapper_kwargs()) as mapper:
        payload = json.dumps(request).encode("utf-8")
        native = mapper._lib.spl_mapper_closure_query
        pointer = native(mapper._mapper_id, payload)
        try:
            assert pointer and not pointer.contents.error
            expected = json.loads(pointer.contents.result)
            copied = deepcopy(expected)
        finally:
            mapper._lib.spl_result_free(pointer)
        assert copied == mapper.closure_query(request)
        assert copied["status"] == "valid"
        pointer = native(mapper._mapper_id, b"{}")
        try:
            assert pointer and pointer.contents.error and not pointer.contents.result
        finally:
            mapper._lib.spl_result_free(pointer)
        with pytest.raises(SPLMapperError):
            mapper.closure_query({"schema_version": 1})
        handle = mapper._mapper_id
        lib = mapper._lib
    pointer = lib.spl_mapper_closure_query(handle, payload)
    try:
        assert pointer and pointer.contents.error and not pointer.contents.result
    finally:
        lib.spl_result_free(pointer)


def test_closure_python_reports_content_statuses():
    with SPLMapper(**mapper_kwargs()) as mapper:
        assert mapper.closure_query(CASES[0]["request"])["status"] == "valid"
        assert mapper.closure_query(CASES[1]["request"])["status"] == "incomplete"
