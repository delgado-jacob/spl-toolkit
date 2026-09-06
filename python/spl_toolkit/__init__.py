"""
SPL Toolkit Python Bindings

This package provides Python bindings for the SPL Toolkit library,
enabling programmatic analysis and manipulation of Splunk SPL queries.
"""

from importlib.metadata import PackageNotFoundError, version

try:
    __version__ = version("spl-toolkit")
except PackageNotFoundError:
    __version__ = "dev"

from .mapper import SPLMapper, QueryInfo
from .exceptions import SPLMapperError

__all__ = ["SPLMapper", "QueryInfo", "SPLMapperError", "__version__"]
