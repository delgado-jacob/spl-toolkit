"""Native ABI ownership tests, isolated so allocator faults crash only a child."""

from __future__ import annotations

import ctypes
import os
from pathlib import Path
import subprocess
import sys

import pytest


THREE_MODELS = (
    "| tstats count from datamodel=Web by Web.status "
    "| tstats count from datamodel=Authentication by Authentication.user "
    "| tstats count from datamodel=Network_Traffic by Network_Traffic.src_ip"
)
MACROS = "`get_data` | eval result=`calculate_score(field1, field2)` | where result>10"


class SPLResult(ctypes.Structure):
    _fields_ = [("error", ctypes.c_char_p), ("result", ctypes.c_char_p)]


class SPLQueryInfoC(ctypes.Structure):
    _fields_ = [
        ("data_models", ctypes.POINTER(ctypes.c_char_p)),
        ("datasets", ctypes.POINTER(ctypes.c_char_p)),
        ("lookups", ctypes.POINTER(ctypes.c_char_p)),
        ("macros", ctypes.POINTER(ctypes.c_char_p)),
        ("sources", ctypes.POINTER(ctypes.c_char_p)),
        ("source_types", ctypes.POINTER(ctypes.c_char_p)),
        ("input_fields", ctypes.POINTER(ctypes.c_char_p)),
        ("data_models_count", ctypes.c_int),
        ("datasets_count", ctypes.c_int),
        ("lookups_count", ctypes.c_int),
        ("macros_count", ctypes.c_int),
        ("sources_count", ctypes.c_int),
        ("source_types_count", ctypes.c_int),
        ("input_fields_count", ctypes.c_int),
        ("error", ctypes.c_char_p),
    ]


def _library() -> ctypes.CDLL:
    library_path = Path(os.environ["SPL_NATIVE_LIBRARY"])
    assert library_path.is_absolute(), "SPL_NATIVE_LIBRARY must be absolute"
    lib = ctypes.CDLL(str(library_path))
    lib.spl_mapper_new.argtypes = []
    lib.spl_mapper_new.restype = ctypes.c_int
    lib.spl_mapper_free.argtypes = [ctypes.c_int]
    lib.spl_mapper_free.restype = None
    lib.spl_mapper_discover_query.argtypes = [ctypes.c_int, ctypes.c_char_p]
    lib.spl_mapper_discover_query.restype = ctypes.POINTER(SPLQueryInfoC)
    lib.spl_query_info_free.argtypes = [ctypes.POINTER(SPLQueryInfoC)]
    lib.spl_query_info_free.restype = None
    lib.spl_mapper_load_mappings.argtypes = [ctypes.c_int, ctypes.c_char_p]
    lib.spl_mapper_load_mappings.restype = ctypes.c_void_p
    lib.spl_mapper_map_query.argtypes = [ctypes.c_int, ctypes.c_char_p]
    lib.spl_mapper_map_query.restype = ctypes.POINTER(SPLResult)
    lib.spl_result_free.argtypes = [ctypes.POINTER(SPLResult)]
    lib.spl_result_free.restype = None
    return lib


def discover(query: str) -> dict[str, list[str]]:
    lib = _library()
    handle = lib.spl_mapper_new()
    assert handle > 0
    result = lib.spl_mapper_discover_query(handle, query.encode())
    try:
        assert result
        info = result.contents
        assert not info.error, info.error.decode() if info.error else "native discovery failed"

        def copy(values: ctypes.POINTER(ctypes.c_char_p), count: int) -> list[str]:
            return [values[index].decode() for index in range(count)]

        return {
            "data_models": copy(info.data_models, info.data_models_count),
            "datasets": copy(info.datasets, info.datasets_count),
            "lookups": copy(info.lookups, info.lookups_count),
            "macros": copy(info.macros, info.macros_count),
            "sources": copy(info.sources, info.sources_count),
            "source_types": copy(info.source_types, info.source_types_count),
            "input_fields": copy(info.input_fields, info.input_fields_count),
        }
    finally:
        lib.spl_query_info_free(result)
        lib.spl_mapper_free(handle)


def read_models(query: str) -> list[str]:
    return discover(query)["data_models"]


def read_macros(query: str) -> list[str]:
    return discover(query)["macros"]


def _complete_arrays() -> None:
    assert set(read_models(THREE_MODELS)) == {"Web", "Authentication", "Network_Traffic"}
    assert set(read_macros(MACROS)) == {"get_data", "calculate_score"}


def _repeated_and_empty_arrays() -> None:
    for _ in range(100):
        arrays = discover("| head 10")
        assert all(not values for values in arrays.values())


def _invalid_handle_and_cleanup() -> None:
    lib = _library()
    lib.spl_string_free.argtypes = [ctypes.c_void_p]
    lib.spl_string_free.restype = None
    result = lib.spl_mapper_discover_query(2_147_483_647, b"search index=main")
    try:
        assert result and result.contents.error == b"Mapper not found"
    finally:
        lib.spl_query_info_free(result)

    mapped = lib.spl_mapper_map_query(2_147_483_647, b"search index=main")
    try:
        assert mapped and mapped.contents.error == b"Mapper not found"
    finally:
        lib.spl_result_free(mapped)

    lib.spl_query_info_free(None)
    lib.spl_result_free(None)
    lib.spl_string_free(None)


def _standalone_error_ownership() -> None:
    lib = _library()
    lib.spl_string_free.argtypes = [ctypes.c_void_p]
    lib.spl_string_free.restype = None
    handle = lib.spl_mapper_new()
    assert handle > 0
    try:
        error = lib.spl_mapper_load_mappings(handle, b"[")
        assert error
        copied = ctypes.string_at(error).decode()
        assert copied
        lib.spl_string_free(error)
    finally:
        lib.spl_mapper_free(handle)


CASES = {
    "complete-arrays": _complete_arrays,
    "repeated-empty": _repeated_and_empty_arrays,
    "invalid-cleanup": _invalid_handle_and_cleanup,
    "standalone-error": _standalone_error_ownership,
}


@pytest.mark.parametrize("case", CASES)
def test_native_abi_in_child_process(case: str) -> None:
    completed = subprocess.run(
        [sys.executable, str(Path(__file__).resolve()), case],
        capture_output=True,
        text=True,
        env=os.environ.copy(),
        check=False,
    )
    assert completed.returncode == 0, (
        f"native ABI child {case!r} exited {completed.returncode}\n"
        f"stdout:\n{completed.stdout}\nstderr:\n{completed.stderr}"
    )


if __name__ == "__main__":
    CASES[sys.argv[1]]()
