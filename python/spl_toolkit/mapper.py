"""
Python bindings for SPL Toolkit
"""

import ctypes
import json
import os
from typing import Dict, List, Optional, Any
from dataclasses import dataclass

from .exceptions import SPLMapperError, MapperNotFoundError, ParseError, ConfigurationError


@dataclass
class QueryInfo:
    """Information discovered from a SPL query"""
    data_models: List[str]
    datasets: List[str] 
    lookups: List[str]
    macros: List[str]
    sources: List[str]
    source_types: List[str]
    input_fields: List[str]


class SPLResult(ctypes.Structure):
    """C structure for SPL operation results"""
    _fields_ = [
        ("error", ctypes.c_char_p),
        ("result", ctypes.c_char_p),
    ]


class SPLQueryInfoC(ctypes.Structure):
    """C structure for SPL query information"""
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


class SPLMapper:
    """Python wrapper for SPL Toolkit"""
    
    def __init__(self, config: Optional[Dict[str, Any]] = None, library_path: Optional[str] = None):
        """
        Initialize SPL Toolkit
        
        Args:
            config: Optional mapping configuration dictionary
            library_path: Optional path to the shared library
        """
        # Load the shared library
        if library_path is None:
            # Try to find the library in common locations
            possible_paths = [
                "./libspl_toolkit.so",
                "./libspl_toolkit.dylib", 
                os.path.join(os.path.dirname(__file__), "libspl_toolkit.so"),
                os.path.join(os.path.dirname(__file__), "libspl_toolkit.dylib"),
            ]
            
            library_path = None
            for path in possible_paths:
                if os.path.exists(path):
                    library_path = path
                    break
            
            if library_path is None:
                raise SPLMapperError("Could not find SPL Toolkit shared library")
        
        try:
            self._lib = ctypes.CDLL(library_path)
        except OSError as e:
            raise SPLMapperError(f"Failed to load library {library_path}: {e}")
        
        # Define function signatures
        self._setup_function_signatures()
        
        # Create mapper instance
        if config is None:
            self._mapper_id = self._lib.spl_mapper_new()
        else:
            config_json = json.dumps(config).encode('utf-8')
            self._mapper_id = self._lib.spl_mapper_new_with_config(config_json)
            
        if self._mapper_id < 0:
            raise ConfigurationError("Failed to create mapper with provided configuration")
    
    def _setup_function_signatures(self):
        """Setup ctypes function signatures"""
        # spl_mapper_new
        self._lib.spl_mapper_new.argtypes = []
        self._lib.spl_mapper_new.restype = ctypes.c_int
        
        # spl_mapper_new_with_config
        self._lib.spl_mapper_new_with_config.argtypes = [ctypes.c_char_p]
        self._lib.spl_mapper_new_with_config.restype = ctypes.c_int
        
        # spl_mapper_free
        self._lib.spl_mapper_free.argtypes = [ctypes.c_int]
        self._lib.spl_mapper_free.restype = None
        
        # spl_mapper_load_mappings
        self._lib.spl_mapper_load_mappings.argtypes = [ctypes.c_int, ctypes.c_char_p]
        self._lib.spl_mapper_load_mappings.restype = ctypes.c_char_p
        
        # spl_mapper_map_query
        self._lib.spl_mapper_map_query.argtypes = [ctypes.c_int, ctypes.c_char_p]
        self._lib.spl_mapper_map_query.restype = ctypes.POINTER(SPLResult)
        
        # spl_mapper_map_query_with_context
        self._lib.spl_mapper_map_query_with_context.argtypes = [ctypes.c_int, ctypes.c_char_p, ctypes.c_char_p]
        self._lib.spl_mapper_map_query_with_context.restype = ctypes.POINTER(SPLResult)
        
        # spl_mapper_discover_query
        self._lib.spl_mapper_discover_query.argtypes = [ctypes.c_int, ctypes.c_char_p]
        self._lib.spl_mapper_discover_query.restype = ctypes.POINTER(SPLQueryInfoC)
        
        # spl_result_free
        self._lib.spl_result_free.argtypes = [ctypes.POINTER(SPLResult)]
        self._lib.spl_result_free.restype = None
        
        # spl_query_info_free
        self._lib.spl_query_info_free.argtypes = [ctypes.POINTER(SPLQueryInfoC)]
        self._lib.spl_query_info_free.restype = None
    
    def __del__(self):
        """Cleanup mapper instance"""
        if hasattr(self, '_mapper_id') and hasattr(self, '_lib'):
            self._lib.spl_mapper_free(self._mapper_id)
    
    def load_mappings(self, mappings: List[Dict[str, str]]) -> None:
        """
        Load field mappings from a list of source->target dictionaries
        
        Args:
            mappings: List of mappings, each with 'source' and 'target' keys
        """
        mappings_json = json.dumps(mappings).encode('utf-8')
        error = self._lib.spl_mapper_load_mappings(self._mapper_id, mappings_json)
        
        if error:
            error_msg = error.decode('utf-8')
            raise SPLMapperError(f"Failed to load mappings: {error_msg}")
    
    def map_query(self, query: str) -> str:
        """
        Apply field mappings to a SPL query
        
        Args:
            query: SPL query string
            
        Returns:
            Mapped query string
        """
        query_bytes = query.encode('utf-8')
        result_ptr = self._lib.spl_mapper_map_query(self._mapper_id, query_bytes)
        
        if not result_ptr:
            raise SPLMapperError("Failed to map query")
        
        try:
            result = result_ptr.contents
            
            if result.error:
                error_msg = result.error.decode('utf-8')
                raise ParseError(f"Query mapping failed: {error_msg}")
            
            if result.result:
                return result.result.decode('utf-8')
            else:
                return ""
                
        finally:
            self._lib.spl_result_free(result_ptr)
    
    def map_query_with_context(self, query: str, context: Dict[str, Any]) -> str:
        """
        Apply field mappings to a SPL query with explicit context
        
        Args:
            query: SPL query string
            context: Context information for conditional mappings
            
        Returns:
            Mapped query string
        """
        query_bytes = query.encode('utf-8')
        context_json = json.dumps(context).encode('utf-8')
        result_ptr = self._lib.spl_mapper_map_query_with_context(self._mapper_id, query_bytes, context_json)
        
        if not result_ptr:
            raise SPLMapperError("Failed to map query with context")
        
        try:
            result = result_ptr.contents
            
            if result.error:
                error_msg = result.error.decode('utf-8')
                raise ParseError(f"Query mapping failed: {error_msg}")
            
            if result.result:
                return result.result.decode('utf-8')
            else:
                return ""
                
        finally:
            self._lib.spl_result_free(result_ptr)
    
    def discover_query(self, query: str) -> QueryInfo:
        """
        Analyze a SPL query and discover information about it
        
        Args:
            query: SPL query string
            
        Returns:
            QueryInfo object with discovered information
        """
        query_bytes = query.encode('utf-8')
        result_ptr = self._lib.spl_mapper_discover_query(self._mapper_id, query_bytes)
        
        if not result_ptr:
            raise SPLMapperError("Failed to discover query information")
        
        try:
            result = result_ptr.contents
            
            if result.error:
                error_msg = result.error.decode('utf-8')
                raise ParseError(f"Query discovery failed: {error_msg}")
            
            # Convert C arrays to Python lists
            def extract_string_array(ptr, count):
                if not ptr or count <= 0:
                    return []
                return [ptr[i].decode('utf-8') for i in range(count)]
            
            return QueryInfo(
                data_models=extract_string_array(result.data_models, result.data_models_count),
                datasets=extract_string_array(result.datasets, result.datasets_count),
                lookups=extract_string_array(result.lookups, result.lookups_count),
                macros=extract_string_array(result.macros, result.macros_count),
                sources=extract_string_array(result.sources, result.sources_count),
                source_types=extract_string_array(result.source_types, result.source_types_count),
                input_fields=extract_string_array(result.input_fields, result.input_fields_count),
            )
            
        finally:
            self._lib.spl_query_info_free(result_ptr)
    
    def get_input_fields(self, query: str) -> List[str]:
        """
        Get all input fields required for a query
        
        Args:
            query: SPL query string
            
        Returns:
            List of field names
        """
        info = self.discover_query(query)
        return info.input_fields