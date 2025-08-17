"""
Exceptions for SPL Toolkit
"""


class SPLMapperError(Exception):
    """Base exception for SPL Toolkit errors"""
    pass


class MapperNotFoundError(SPLMapperError):
    """Raised when a mapper instance is not found"""
    pass


class ParseError(SPLMapperError):
    """Raised when SPL query parsing fails"""
    pass


class ConfigurationError(SPLMapperError):
    """Raised when configuration is invalid"""
    pass