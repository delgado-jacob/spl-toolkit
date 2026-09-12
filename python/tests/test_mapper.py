"""
Tests for SPL Toolkit Python bindings
"""

import pytest
import json
import ctypes
import threading
from unittest.mock import patch, MagicMock

# For testing without the actual shared library
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from spl_toolkit import SPLMapper, QueryInfo, SPLMapperError, __version__
from spl_toolkit.exceptions import ParseError, ConfigurationError, MapperNotFoundError
from spl_toolkit.mapper import SPLResult


MOCK_LIBRARY = "/mock/libspl_toolkit.dylib"


def configure_mock_version(mock_lib, version=None):
    if version is None:
        version = __version__.encode()
    native_version = ctypes.create_string_buffer(version)
    pointer = ctypes.cast(native_version, ctypes.c_void_p).value
    mock_lib.spl_toolkit_version.return_value = pointer
    mock_lib._native_version_owner = native_version


def mapper_kwargs():
    return {"library_path": os.environ["SPL_NATIVE_LIBRARY"]} if "SPL_NATIVE_LIBRARY" in os.environ else {}


@pytest.mark.parametrize("batch", [False, True])
@pytest.mark.parametrize("target", [None, {"kind": "field_list", "catalog": {"fields": ["user"]}}])
def test_rewrite_wrapper_preserves_request_and_frees_result(batch, target):
    rules = [{"id": "map", "kind": "field", "source": {"name": "src"}, "target": {"name": "user"}}]
    document = {"text": "\tFROM main | table src\r\n", "language": "spl2", "profile": "splunkd",
                "version": "current", "source_id": "é😀\x00"}
    with patch("ctypes.CDLL") as loader:
        lib = loader.return_value
        configure_mock_version(lib)
        lib.spl_mapper_new.return_value = 7
        result = SPLResult(error=None, result=b'{"schema_version":1,"status":"incomplete","reports":[]}')
        operation = "rewrite_batch" if batch else "rewrite"
        native = getattr(lib, "spl_mapper_" + operation)
        native.return_value = ctypes.pointer(result)
        with SPLMapper(library_path=MOCK_LIBRARY) as mapper:
            assert callable(getattr(mapper, operation, None)), "missing Python rewrite request boundary"
            if batch:
                actual = mapper.rewrite_batch([document], rules, mode="apply", validation_target=target)
            else:
                actual = mapper.rewrite(document["text"], rules, mode="apply", validation_target=target,
                                        **{k: v for k, v in document.items() if k != "text"})
        assert actual == {"schema_version": 1, "status": "incomplete", "reports": []}
        expected = {"schema_version": 1, "mode": "apply", "rules": rules}
        expected.update({"documents": [document]} if batch else {"document": document})
        if target is not None:
            expected["validation_target"] = target
        assert native.call_count == 1
        handle, payload = native.call_args.args
        assert handle == 7 and json.loads(payload) == expected
        lib.spl_result_free.assert_called_once_with(native.return_value)


class TestSPLMapper:
    """Test SPL Toolkit functionality"""
    
    def test_init_without_config(self):
        """Test initializing mapper without configuration"""
        # Since the shared library exists and works, this should succeed
        mapper = SPLMapper(**mapper_kwargs())
        assert mapper is not None
        mapper.close()
    
    def test_init_with_config(self):
        """Test initializing mapper with configuration"""
        config = {
            "version": "1.0",
            "mappings": [
                {"source": "src_ip", "target": "source_ip"}
            ]
        }
        
        # Since the shared library exists and works, this should succeed
        mapper = SPLMapper(config=config, **mapper_kwargs())
        assert mapper is not None
        mapper.close()
    
    def test_load_mappings(self):
        """Test loading field mappings"""
        mappings = [
            {"source": "src_ip", "target": "source_ip"},
            {"source": "dst_ip", "target": "destination_ip"},
        ]
        
        # Mock the library for testing
        with patch('ctypes.CDLL') as mock_cdll:
            mock_lib = MagicMock()
            mock_cdll.return_value = mock_lib
            configure_mock_version(mock_lib)
            mock_lib.spl_mapper_new.return_value = 1
            mock_lib.spl_mapper_load_mappings.return_value = None  # Success
            
            mapper = SPLMapper(library_path=MOCK_LIBRARY)
            mapper.load_mappings(mappings)
            
            # Verify the function was called
            mock_lib.spl_mapper_load_mappings.assert_called_once()

    def test_load_mappings_copies_and_frees_native_error(self):
        with patch("ctypes.CDLL") as mock_cdll:
            mock_lib = MagicMock()
            mock_cdll.return_value = mock_lib
            configure_mock_version(mock_lib)
            mock_lib.spl_mapper_new.return_value = 1
            native_error = ctypes.create_string_buffer(b"invalid source")
            error_pointer = ctypes.cast(native_error, ctypes.c_void_p).value
            mock_lib.spl_mapper_load_mappings.return_value = error_pointer

            mapper = SPLMapper(library_path=MOCK_LIBRARY)
            with pytest.raises(ConfigurationError, match="invalid source"):
                mapper.load_mappings([{"source": 1, "target": "bad"}])

            mock_lib.spl_string_free.assert_any_call(error_pointer)
            mapper.close()
    
    def test_map_query(self):
        """Test mapping a SPL query"""
        with patch('ctypes.CDLL') as mock_cdll:
            # Setup mock library
            mock_lib = MagicMock()
            mock_cdll.return_value = mock_lib
            configure_mock_version(mock_lib)
            mock_lib.spl_mapper_new.return_value = 1
            
            # Mock result structure
            mock_result = MagicMock()
            mock_result.error = None
            mock_result.result = b"search source_ip=192.168.1.1"
            
            mock_result_ptr = MagicMock()
            mock_result_ptr.contents = mock_result
            mock_lib.spl_mapper_map_query.return_value = mock_result_ptr
            
            mapper = SPLMapper(library_path=MOCK_LIBRARY)
            result = mapper.map_query("search src_ip=192.168.1.1")
            
            assert result == "search source_ip=192.168.1.1"
    
    def test_map_query_with_error(self):
        """Test mapping query that results in error"""
        with patch('ctypes.CDLL') as mock_cdll:
            mock_lib = MagicMock()
            mock_cdll.return_value = mock_lib
            configure_mock_version(mock_lib)
            mock_lib.spl_mapper_new.return_value = 1
            
            # Mock error result
            mock_result = MagicMock()
            mock_result.error = b"Parse error"
            mock_result.result = None
            
            mock_result_ptr = MagicMock()
            mock_result_ptr.contents = mock_result
            mock_lib.spl_mapper_map_query.return_value = mock_result_ptr
            
            mapper = SPLMapper(library_path=MOCK_LIBRARY)
            
            with pytest.raises(ParseError):
                mapper.map_query("invalid query")
    
    def test_discover_query(self):
        """Test query discovery functionality"""
        with patch('ctypes.CDLL') as mock_cdll:
            mock_lib = MagicMock()
            mock_cdll.return_value = mock_lib
            configure_mock_version(mock_lib)
            mock_lib.spl_mapper_new.return_value = 1
            
            # Mock discovery result
            mock_result = MagicMock()
            mock_result.error = None
            mock_result.data_models_count = 1
            mock_result.source_types_count = 1 
            mock_result.input_fields_count = 2
            
            # Mock string arrays
            mock_result.data_models = (ctypes.c_char_p * 1)(b"Network_Traffic")
            mock_result.source_types = (ctypes.c_char_p * 1)(b"access_combined")
            mock_result.input_fields = (ctypes.c_char_p * 2)(b"src_ip", b"dst_port")
            
            # Set other counts to 0
            mock_result.datasets_count = 0
            mock_result.lookups_count = 0
            mock_result.macros_count = 0
            mock_result.sources_count = 0
            
            mock_result_ptr = MagicMock()
            mock_result_ptr.contents = mock_result
            mock_lib.spl_mapper_discover_query.return_value = mock_result_ptr
            
            mapper = SPLMapper(library_path=MOCK_LIBRARY)
            info = mapper.discover_query("search sourcetype=access_combined src_ip=192.168.1.1")
            
            assert isinstance(info, QueryInfo)
            assert "Network_Traffic" in info.data_models
            assert "access_combined" in info.source_types
            assert "src_ip" in info.input_fields
            assert "dst_port" in info.input_fields

    def test_load_failure_is_configuration_error_and_finalizer_is_safe(self):
        with patch("ctypes.CDLL", side_effect=OSError("not a native library")):
            with pytest.raises(ConfigurationError, match="Failed to load library"):
                SPLMapper(library_path=MOCK_LIBRARY)

    def test_default_lookup_checks_only_package_directory(self):
        with patch("os.path.exists", return_value=False) as exists:
            with pytest.raises(ConfigurationError, match="Could not find"):
                SPLMapper()

        checked = [call.args[0] for call in exists.call_args_list]
        package_dir = os.path.dirname(os.path.abspath(sys.modules["spl_toolkit.mapper"].__file__))
        assert checked
        assert all(os.path.dirname(path) == package_dir for path in checked)

    def test_concurrent_close_waits_and_frees_once(self):
        entered = threading.Event()
        release = threading.Event()

        with patch("ctypes.CDLL") as mock_cdll:
            mock_lib = MagicMock()
            mock_cdll.return_value = mock_lib
            configure_mock_version(mock_lib)
            mock_lib.spl_mapper_new.return_value = 7

            def blocking_map(_handle, _query):
                entered.set()
                assert release.wait(timeout=5)
                result = SPLResult(error=None, result=b"search source_ip=1")
                pointer = ctypes.pointer(result)
                pointer._result_owner = result
                return pointer

            mock_lib.spl_mapper_map_query.side_effect = blocking_map
            mapper = SPLMapper(library_path=MOCK_LIBRARY)
            operation_errors = []

            def run_operation():
                try:
                    mapper.map_query("search src_ip=1")
                except Exception as error:
                    operation_errors.append(error)

            operation = threading.Thread(target=run_operation)
            operation.start()
            assert entered.wait(timeout=5)

            closers = [threading.Thread(target=mapper.close) for _ in range(3)]
            for closer in closers:
                closer.start()

            with mapper._condition:
                assert mapper._condition.wait_for(lambda: mapper._closing, timeout=5)

            with pytest.raises(MapperNotFoundError, match="closed"):
                mapper.discover_query("search src_ip=1")

            release.set()
            operation.join(timeout=5)
            for closer in closers:
                closer.join(timeout=5)

            assert not operation.is_alive()
            assert all(not closer.is_alive() for closer in closers)
            assert operation_errors == []
            mock_lib.spl_mapper_free.assert_called_once_with(7)

    def test_context_manager_rejects_closed_mapper(self):
        with patch("ctypes.CDLL") as mock_cdll:
            mock_lib = MagicMock()
            mock_cdll.return_value = mock_lib
            configure_mock_version(mock_lib)
            mock_lib.spl_mapper_new.return_value = 3
            mapper = SPLMapper(library_path=MOCK_LIBRARY)
            mapper.close()

            with pytest.raises(MapperNotFoundError, match="closed"):
                mapper.__enter__()

    def test_native_version_is_copied_freed_and_read_only(self):
        with patch("ctypes.CDLL") as mock_cdll:
            mock_lib = MagicMock()
            mock_cdll.return_value = mock_lib
            configure_mock_version(mock_lib)
            mock_lib.spl_mapper_new.return_value = 1

            mapper = SPLMapper(library_path=MOCK_LIBRARY)

            assert mapper.native_version == __version__
            mock_lib.spl_string_free.assert_any_call(mock_lib.spl_toolkit_version.return_value)
            with pytest.raises(AttributeError):
                mapper.native_version = "changed"
            mapper.close()

    def test_released_package_rejects_mismatched_explicit_library(self):
        if __version__ == "dev":
            pytest.skip("uninstalled source mode accepts and reports the native version")
        with patch("ctypes.CDLL") as mock_cdll:
            mock_lib = MagicMock()
            mock_cdll.return_value = mock_lib
            configure_mock_version(mock_lib, b"wrong-version")

            with pytest.raises(ConfigurationError, match="does not match package version"):
                SPLMapper(library_path=MOCK_LIBRARY)


class TestQueryInfo:
    """Test QueryInfo dataclass"""
    
    def test_query_info_creation(self):
        """Test creating QueryInfo instance"""
        info = QueryInfo(
            data_models=["Network_Traffic"],
            datasets=[],
            lookups=["ip_geo"],
            macros=[],
            sources=[],
            source_types=["access_combined"],
            input_fields=["src_ip", "dst_port"]
        )
        
        assert info.data_models == ["Network_Traffic"]
        assert info.lookups == ["ip_geo"]
        assert info.source_types == ["access_combined"]
        assert len(info.input_fields) == 2


class TestExceptions:
    """Test custom exceptions"""
    
    def test_spl_mapper_error(self):
        """Test base SPL mapper error"""
        with pytest.raises(SPLMapperError):
            raise SPLMapperError("Test error")
    
    def test_parse_error(self):
        """Test parse error"""
        with pytest.raises(ParseError):
            raise ParseError("Parse failed")
        
        # ParseError should also be caught as SPLMapperError
        with pytest.raises(SPLMapperError):
            raise ParseError("Parse failed")
    
    def test_configuration_error(self):
        """Test configuration error"""
        with pytest.raises(ConfigurationError):
            raise ConfigurationError("Config invalid")
        
        # ConfigurationError should also be caught as SPLMapperError
        with pytest.raises(SPLMapperError):
            raise ConfigurationError("Config invalid")


if __name__ == "__main__":
    pytest.main([__file__])
