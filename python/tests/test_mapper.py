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

from spl_toolkit import SPLMapper, QueryInfo, SPLMapperError
from spl_toolkit.exceptions import ParseError, ConfigurationError, MapperNotFoundError
from spl_toolkit.mapper import SPLResult


MOCK_LIBRARY = "/mock/libspl_toolkit.dylib"


class TestSPLMapper:
    """Test SPL Toolkit functionality"""
    
    def test_init_without_config(self):
        """Test initializing mapper without configuration"""
        # Since the shared library exists and works, this should succeed
        mapper = SPLMapper(library_path=os.environ["SPL_NATIVE_LIBRARY"])
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
        mapper = SPLMapper(config=config, library_path=os.environ["SPL_NATIVE_LIBRARY"])
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
            mock_lib.spl_mapper_new.return_value = 1
            native_error = ctypes.create_string_buffer(b"invalid source")
            error_pointer = ctypes.cast(native_error, ctypes.c_void_p).value
            mock_lib.spl_mapper_load_mappings.return_value = error_pointer

            mapper = SPLMapper(library_path=MOCK_LIBRARY)
            with pytest.raises(ConfigurationError, match="invalid source"):
                mapper.load_mappings([{"source": 1, "target": "bad"}])

            mock_lib.spl_string_free.assert_called_once_with(error_pointer)
            mapper.close()
    
    def test_map_query(self):
        """Test mapping a SPL query"""
        with patch('ctypes.CDLL') as mock_cdll:
            # Setup mock library
            mock_lib = MagicMock()
            mock_cdll.return_value = mock_lib
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
            mock_lib.spl_mapper_new.return_value = 3
            mapper = SPLMapper(library_path=MOCK_LIBRARY)
            mapper.close()

            with pytest.raises(MapperNotFoundError, match="closed"):
                mapper.__enter__()


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
