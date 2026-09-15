"""
Python bindings for SPL Toolkit
"""

import ctypes
import json
import os
import sys
import threading
from contextlib import contextmanager
from typing import Dict, List, Optional, Any
from dataclasses import dataclass

from .exceptions import SPLMapperError, MapperNotFoundError, ParseError, ConfigurationError
from . import __version__


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
        self._condition = threading.Condition()
        self._active_calls = 0
        self._closing = False
        self._closed = False
        self._mapper_id = None
        self._lib = None

        if library_path is None:
            if os.name == "nt":
                suffix = ".dll"
            elif sys.platform == "darwin":
                suffix = ".dylib"
            else:
                suffix = ".so"
            candidate = os.path.join(os.path.dirname(os.path.abspath(__file__)), f"libspl_toolkit{suffix}")
            if not os.path.exists(candidate):
                raise ConfigurationError("Could not find SPL Toolkit shared library")
            library_path = candidate
        
        try:
            self._lib = ctypes.CDLL(library_path)
        except OSError as e:
            raise ConfigurationError(f"Failed to load library {library_path}: {e}") from e
        
        # Define function signatures
        self._setup_function_signatures()
        self._native_version = self._read_native_version()
        if __version__ != "dev" and self._native_version != __version__:
            raise ConfigurationError(
                f"Native library version {self._native_version!r} does not match package version {__version__!r}"
            )
        
        # Create mapper instance
        if config is None:
            mapper_id = self._lib.spl_mapper_new()
        else:
            config_json = json.dumps(config).encode('utf-8')
            mapper_id = self._lib.spl_mapper_new_with_config(config_json)
            
        if mapper_id < 0:
            raise ConfigurationError("Failed to create mapper with provided configuration")
        self._mapper_id = mapper_id
    
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
        self._lib.spl_mapper_load_mappings.restype = ctypes.c_void_p

        # spl_string_free
        self._lib.spl_string_free.argtypes = [ctypes.c_void_p]
        self._lib.spl_string_free.restype = None

        self._lib.spl_toolkit_version.argtypes = []
        self._lib.spl_toolkit_version.restype = ctypes.c_void_p
        
        # spl_mapper_map_query
        self._lib.spl_mapper_map_query.argtypes = [ctypes.c_int, ctypes.c_char_p]
        self._lib.spl_mapper_map_query.restype = ctypes.POINTER(SPLResult)
        
        # spl_mapper_map_query_with_context
        self._lib.spl_mapper_map_query_with_context.argtypes = [ctypes.c_int, ctypes.c_char_p, ctypes.c_char_p]
        self._lib.spl_mapper_map_query_with_context.restype = ctypes.POINTER(SPLResult)
        
        # Owned canonical JSON results
        self._lib.spl_mapper_analyze_query.argtypes = [ctypes.c_int, ctypes.c_char_p]
        self._lib.spl_mapper_analyze_query.restype = ctypes.POINTER(SPLResult)
        self._lib.spl_mapper_requirements_query.argtypes = [ctypes.c_int, ctypes.c_char_p]
        self._lib.spl_mapper_requirements_query.restype = ctypes.POINTER(SPLResult)
        self._lib.spl_mapper_validate_fields.argtypes = [ctypes.c_int, ctypes.c_char_p]
        self._lib.spl_mapper_validate_fields.restype = ctypes.POINTER(SPLResult)
        self._lib.spl_mapper_validate_fields_batch.argtypes = [ctypes.c_int, ctypes.c_char_p]
        self._lib.spl_mapper_validate_fields_batch.restype = ctypes.POINTER(SPLResult)
        self._lib.spl_mapper_validate_schema.argtypes = [ctypes.c_int, ctypes.c_char_p]
        self._lib.spl_mapper_validate_schema.restype = ctypes.POINTER(SPLResult)
        self._lib.spl_mapper_validate_schema_batch.argtypes = [ctypes.c_int, ctypes.c_char_p]
        self._lib.spl_mapper_validate_schema_batch.restype = ctypes.POINTER(SPLResult)
        self._lib.spl_mapper_rewrite.argtypes = [ctypes.c_int, ctypes.c_char_p]
        self._lib.spl_mapper_rewrite.restype = ctypes.POINTER(SPLResult)
        self._lib.spl_mapper_rewrite_batch.argtypes = [ctypes.c_int, ctypes.c_char_p]
        self._lib.spl_mapper_rewrite_batch.restype = ctypes.POINTER(SPLResult)
        for operation in ("scan_corpus", "export_graph", "export_sarif", "impact_schema",
                          "impact_mapping", "document_view"):
            native = getattr(self._lib, "spl_mapper_" + operation)
            native.argtypes = [ctypes.c_int, ctypes.c_char_p]
            native.restype = ctypes.POINTER(SPLResult)
        self._lib.spl_mapper_capabilities.argtypes = [ctypes.c_int]
        self._lib.spl_mapper_capabilities.restype = ctypes.POINTER(SPLResult)
        self._lib.spl_mapper_capabilities_for.argtypes = [ctypes.c_int, ctypes.c_char_p]
        self._lib.spl_mapper_capabilities_for.restype = ctypes.POINTER(SPLResult)

        # spl_mapper_discover_query
        self._lib.spl_mapper_discover_query.argtypes = [ctypes.c_int, ctypes.c_char_p]
        self._lib.spl_mapper_discover_query.restype = ctypes.POINTER(SPLQueryInfoC)
        
        # spl_result_free
        self._lib.spl_result_free.argtypes = [ctypes.POINTER(SPLResult)]
        self._lib.spl_result_free.restype = None
        
        # spl_query_info_free
        self._lib.spl_query_info_free.argtypes = [ctypes.POINTER(SPLQueryInfoC)]
        self._lib.spl_query_info_free.restype = None

    def _read_native_version(self) -> str:
        pointer = self._lib.spl_toolkit_version()
        if not pointer:
            raise ConfigurationError("Native library returned no version")
        try:
            return ctypes.string_at(pointer).decode("utf-8")
        finally:
            self._lib.spl_string_free(pointer)

    @property
    def native_version(self) -> str:
        """Version embedded in the loaded native library."""
        return self._native_version
    
    @contextmanager
    def _operation(self):
        with self._condition:
            if self._closing or self._closed or self._mapper_id is None:
                raise MapperNotFoundError("Mapper is closed")
            self._active_calls += 1
            handle = self._mapper_id
        try:
            yield handle
        finally:
            with self._condition:
                self._active_calls -= 1
                self._condition.notify_all()

    def close(self) -> None:
        """Release the native mapper after any admitted operations finish."""
        with self._condition:
            if self._closed:
                return
            if self._closing:
                while not self._closed:
                    self._condition.wait()
                return

            self._closing = True
            while self._active_calls:
                self._condition.wait()
            mapper_id = self._mapper_id
            self._mapper_id = None

        try:
            if mapper_id is not None and self._lib is not None:
                self._lib.spl_mapper_free(mapper_id)
        finally:
            with self._condition:
                self._closed = True
                self._condition.notify_all()

    def __enter__(self) -> "SPLMapper":
        with self._condition:
            if self._closing or self._closed or self._mapper_id is None:
                raise MapperNotFoundError("Mapper is closed")
        return self

    def __exit__(self, exc_type, exc, traceback) -> None:
        self.close()
        return None

    def __del__(self):
        """Defensively release the mapper without surfacing finalizer errors."""
        try:
            self.close()
        except Exception:
            pass
    
    def load_mappings(self, mappings: List[Dict[str, str]]) -> None:
        """
        Load field mappings from a list of source->target dictionaries
        
        Args:
            mappings: List of mappings, each with 'source' and 'target' keys
        """
        mappings_json = json.dumps(mappings).encode('utf-8')
        with self._operation() as mapper_id:
            error_pointer = self._lib.spl_mapper_load_mappings(mapper_id, mappings_json)
            if error_pointer:
                try:
                    message = ctypes.string_at(error_pointer).decode("utf-8")
                finally:
                    self._lib.spl_string_free(error_pointer)
                raise ConfigurationError(message)
    
    def map_query(self, query: str, *, language="spl", profile="splunkd", version="current") -> str:
        """
        Apply field mappings to a SPL query
        
        Args:
            query: SPL query string
            
        Returns:
            Mapped query string
        """
        self._legacy_selectors(language, profile, version)
        query_bytes = query.encode('utf-8')
        with self._operation() as mapper_id:
            result_ptr = self._lib.spl_mapper_map_query(mapper_id, query_bytes)
            if not result_ptr:
                raise SPLMapperError("Failed to map query")

            try:
                result = result_ptr.contents

                if result.error:
                    error_msg = result.error.decode('utf-8')
                    raise ParseError(f"Query mapping failed: {error_msg}")

                return result.result.decode('utf-8') if result.result else ""
            finally:
                self._lib.spl_result_free(result_ptr)
    
    def map_query_with_context(self, query: str, context: Dict[str, Any], *, language="spl", profile="splunkd", version="current") -> str:
        """
        Apply field mappings to a SPL query with explicit context
        
        Args:
            query: SPL query string
            context: Context information for conditional mappings
            
        Returns:
            Mapped query string
        """
        self._legacy_selectors(language, profile, version)
        query_bytes = query.encode('utf-8')
        context_json = json.dumps(context).encode('utf-8')
        with self._operation() as mapper_id:
            result_ptr = self._lib.spl_mapper_map_query_with_context(mapper_id, query_bytes, context_json)

            if not result_ptr:
                raise SPLMapperError("Failed to map query with context")

            try:
                result = result_ptr.contents

                if result.error:
                    error_msg = result.error.decode('utf-8')
                    raise ParseError(f"Query mapping failed: {error_msg}")

                return result.result.decode('utf-8') if result.result else ""
            finally:
                self._lib.spl_result_free(result_ptr)
    
    def analyze_query(self, query: str, *, language: str = "spl", profile: str = "splunkd",
                      version: str = "current", source_id: str = "") -> dict:
        """Return the canonical source-aware report, including invalid/incomplete findings.

        Unsupported document options raise SPLMapperError. Source text is preserved.
        """
        document = {"text": query, "language": language, "profile": profile,
                    "version": version, "source_id": source_id}
        with self._operation() as handle:
            try:
                payload = json.dumps(document, allow_nan=False).encode("utf-8")
            except (TypeError, ValueError) as error:
                raise SPLMapperError(f"Invalid document JSON: {error}") from error
            pointer = self._lib.spl_mapper_analyze_query(handle, payload)
            if not pointer:
                raise SPLMapperError("Native analysis returned no result")
            try:
                if pointer.contents.error:
                    raise SPLMapperError(pointer.contents.error.decode("utf-8"))
                return json.loads(pointer.contents.result.decode("utf-8"))
            finally:
                self._lib.spl_result_free(pointer)

    def requirements_query(
        self,
        query: str,
        *,
        language: str = "spl",
        profile: str = "splunkd",
        version: str = "current",
        source_id: str = "",
    ) -> dict[str, Any]:
        """Return the canonical direct requirements for a query document."""
        document = {"text": query, "language": language, "profile": profile,
                    "version": version, "source_id": source_id}
        return self._validate_fields_request(
            self._lib.spl_mapper_requirements_query, document, operation="requirements")

    def validate_fields(self, query, catalog, *, language='spl', profile='splunkd',
                        version='current', source_id='') -> dict:
        """Validate a query against a field catalog using the canonical Go report.

        Query findings return valid/invalid/incomplete reports. Invalid catalogs
        and unsupported document options raise SPLMapperError.
        """
        document = {"text": query, "language": language, "profile": profile,
                    "version": version, "source_id": source_id}
        return self._validate_fields_request(
            self._lib.spl_mapper_validate_fields, {"document": document, "catalog": catalog})

    def validate_fields_batch(self, documents, catalog) -> dict:
        """Validate a nonempty list of query document dictionaries in order."""
        return self._validate_fields_request(
            self._lib.spl_mapper_validate_fields_batch, {"documents": documents, "catalog": catalog})

    def validate_schema(self, query, target, *, language='spl', profile='splunkd',
                        version='current', source_id='') -> dict:
        """Validate a query against an explicit JSON Schema or OCSF target.

        Query findings return valid/invalid/incomplete reports. Invalid targets
        and unsupported document options raise SPLMapperError.
        """
        document = {"text": query, "language": language, "profile": profile,
                    "version": version, "source_id": source_id}
        return self._validate_fields_request(
            self._lib.spl_mapper_validate_schema, {"document": document, "target": target})

    def validate_schema_batch(self, documents, target) -> dict:
        """Validate a nonempty list of documents against an explicit schema target."""
        return self._validate_fields_request(
            self._lib.spl_mapper_validate_schema_batch, {"documents": documents, "target": target})

    def rewrite(self, query, rules, *, mode="preview", validation_target=None,
                language="spl", profile="splunkd", version="current", source_id="") -> dict:
        """Preview or apply explicit safe rules using the canonical Go report."""
        request = {"schema_version": 1, "mode": mode, "rules": rules,
                   "document": {"text": query, "language": language, "profile": profile,
                                "version": version, "source_id": source_id}}
        if validation_target is not None:
            request["validation_target"] = validation_target
        return self._validate_fields_request(self._lib.spl_mapper_rewrite, request, operation="rewrite")

    def rewrite_batch(self, documents, rules, *, mode="preview", validation_target=None) -> dict:
        """Rewrite document dictionaries in order, retaining per-document status."""
        request = {"schema_version": 1, "mode": mode, "rules": rules, "documents": documents}
        if validation_target is not None:
            request["validation_target"] = validation_target
        return self._validate_fields_request(self._lib.spl_mapper_rewrite_batch, request, operation="rewrite")

    def scan_corpus(self, request: dict) -> dict:
        """Scan an inline corpus request with explicit document IDs and optional target."""
        return self._validate_fields_request(self._lib.spl_mapper_scan_corpus, request, operation="corpus scan")

    def export_graph(self, request: dict) -> dict:
        """Export canonical graph evidence from an inline corpus request."""
        return self._validate_fields_request(self._lib.spl_mapper_export_graph, request, operation="graph export")

    def export_sarif(self, request: dict) -> dict:
        """Export a SARIF 2.1.0 object from an inline corpus request."""
        return self._validate_fields_request(self._lib.spl_mapper_export_sarif, request, operation="SARIF export")

    def impact_schema(self, request: dict) -> dict:
        """Compare explicit before/after targets against identical inline documents."""
        return self._validate_fields_request(self._lib.spl_mapper_impact_schema, request, operation="schema impact")

    def impact_mapping(self, request: dict) -> dict:
        """Compare explicit before/after rule sets using canonical rewrite previews."""
        return self._validate_fields_request(self._lib.spl_mapper_impact_mapping, request, operation="mapping impact")

    def document_view(self, document: dict) -> dict:
        """Return a detached advanced view of a canonical QueryDocument dictionary."""
        return self._validate_fields_request(self._lib.spl_mapper_document_view, document, operation="document view")

    def _validate_fields_request(self, native, request, *, operation="validation") -> dict:
        with self._operation() as handle:
            try:
                payload = json.dumps(request, allow_nan=False).encode("utf-8")
            except (TypeError, ValueError) as error:
                raise SPLMapperError(f"Invalid {operation} request JSON: {error}") from error
            pointer = native(handle, payload)
            if not pointer:
                raise SPLMapperError(f"Native {operation} returned no result")
            try:
                if pointer.contents.error:
                    raise SPLMapperError(pointer.contents.error.decode("utf-8"))
                return json.loads(pointer.contents.result.decode("utf-8"))
            finally:
                self._lib.spl_result_free(pointer)

    def capabilities(self, *, language="spl", profile="splunkd", version="current") -> dict:
        """Return the native analysis capability manifest."""
        with self._operation() as handle:
            if (language, profile, version) == ("spl", "splunkd", "current"):
                pointer = self._lib.spl_mapper_capabilities(handle)
            else:
                try:
                    payload = json.dumps({"language": language, "profile": profile, "version": version},
                                         allow_nan=False).encode("utf-8")
                except (TypeError, ValueError) as error:
                    raise SPLMapperError(f"Invalid capability options JSON: {error}") from error
                pointer = self._lib.spl_mapper_capabilities_for(handle, payload)
            if not pointer:
                raise SPLMapperError("Native capabilities returned no result")
            try:
                if pointer.contents.error:
                    raise SPLMapperError(pointer.contents.error.decode("utf-8"))
                return json.loads(pointer.contents.result.decode("utf-8"))
            finally:
                self._lib.spl_result_free(pointer)

    def _legacy_selectors(self, language, profile, version):
        if (language, profile, version) != ("spl", "splunkd", "current"):
            manifest = self.capabilities(language=language, profile=profile, version=version)
            if manifest["language"] == "spl2":
                raise SPLMapperError("unsupported_dialect_for_operation: SPL2 requires analyze_query, validate_fields, validate_schema, or rewrite")

    def discover_query(self, query: str, *, language="spl", profile="splunkd", version="current") -> QueryInfo:
        """
        Analyze a SPL query and discover information about it
        
        Args:
            query: SPL query string
            
        Returns:
            QueryInfo object with discovered information
        """
        self._legacy_selectors(language, profile, version)
        query_bytes = query.encode('utf-8')
        with self._operation() as mapper_id:
            result_ptr = self._lib.spl_mapper_discover_query(mapper_id, query_bytes)

            if not result_ptr:
                raise SPLMapperError("Failed to discover query information")

            try:
                result = result_ptr.contents
            
                if result.error:
                    error_msg = result.error.decode('utf-8')
                    raise ParseError(f"Query discovery failed: {error_msg}")
            
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
    
    def get_input_fields(self, query: str, *, language="spl", profile="splunkd", version="current") -> List[str]:
        """
        Get all input fields required for a query
        
        Args:
            query: SPL query string
            
        Returns:
            List of field names
        """
        info = self.discover_query(query, language=language, profile=profile, version=version)
        return info.input_fields
