"""
Tests for SPL Toolkit Python bindings
"""

import pytest
import json
import ctypes
from unittest.mock import patch, MagicMock

# For testing without the actual shared library
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from spl_toolkit import SPLMapper, QueryInfo, SPLMapperError
from spl_toolkit.exceptions import ParseError, ConfigurationError


class TestSPLMapper:
    """Test SPL Toolkit functionality"""
    
    def test_init_without_config(self):
        """Test initializing mapper without configuration"""
        # Since the shared library exists and works, this should succeed
        mapper = SPLMapper()
        assert mapper is not None
    
    def test_init_with_config(self):
        """Test initializing mapper with configuration"""
        config = {
            "version": "1.0",
            "mappings": [
                {"source": "src_ip", "target": "source_ip"}
            ]
        }
        
        # Since the shared library exists and works, this should succeed
        mapper = SPLMapper(config=config)
        assert mapper is not None
    
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
            
            mapper = SPLMapper()
            mapper.load_mappings(mappings)
            
            # Verify the function was called
            mock_lib.spl_mapper_load_mappings.assert_called_once()
    
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
            
            mapper = SPLMapper()
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
            
            mapper = SPLMapper()
            
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
            
            mapper = SPLMapper()
            info = mapper.discover_query("search sourcetype=access_combined src_ip=192.168.1.1")
            
            assert isinstance(info, QueryInfo)
            assert "Network_Traffic" in info.data_models
            assert "access_combined" in info.source_types
            assert "src_ip" in info.input_fields
            assert "dst_port" in info.input_fields


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