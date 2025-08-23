"""
SPL Toolkit Python Bindings

This package provides Python bindings for the SPL Toolkit library,
enabling programmatic analysis and manipulation of Splunk SPL queries.
"""

from .mapper import SPLMapper, QueryInfo
from .exceptions import SPLMapperError

__version__ = "0.1.1"
__all__ = ["SPLMapper", "QueryInfo", "SPLMapperError"]