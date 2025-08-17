#!/usr/bin/env python3
"""
Basic usage example for SPL Toolkit Python bindings
"""

import sys
import os

# Add the package to path for testing
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from spl_toolkit import SPLMapper, SPLMapperError


def basic_mapping_example():
    """Demonstrate basic field mapping"""
    print("=== Basic Field Mapping Example ===")
    
    try:
        # Create a mapper
        mapper = SPLMapper()
        
        # Define some basic mappings
        mappings = [
            {"source": "src_ip", "target": "source_ip"},
            {"source": "dst_ip", "target": "destination_ip"},
            {"source": "src_port", "target": "source_port"},
            {"source": "dst_port", "target": "destination_port"},
        ]
        
        # Load the mappings
        mapper.load_mappings(mappings)
        
        # Test query
        query = "search src_ip=192.168.1.1 dst_port=80"
        print(f"Original query: {query}")
        
        # Map the query
        mapped_query = mapper.map_query(query)
        print(f"Mapped query: {mapped_query}")
        
    except SPLMapperError as e:
        print(f"Error: {e}")


def conditional_mapping_example():
    """Demonstrate conditional field mapping"""
    print("\n=== Conditional Field Mapping Example ===")
    
    try:
        # Configuration with conditional rules
        config = {
            "version": "1.0",
            "name": "Web Server Log Mapping",
            "mappings": [
                {"source": "ip", "target": "client_ip"}
            ],
            "rules": [
                {
                    "id": "apache_logs",
                    "name": "Apache Access Log Fields",
                    "conditions": [
                        {
                            "type": "sourcetype",
                            "operator": "equals", 
                            "value": "access_combined"
                        }
                    ],
                    "mappings": [
                        {"source": "clientip", "target": "source_address"},
                        {"source": "status", "target": "http_status_code"}
                    ],
                    "priority": 1,
                    "enabled": True
                }
            ]
        }
        
        # Create mapper with configuration
        mapper = SPLMapper(config=config)
        
        # Test query with context
        query = "search sourcetype=access_combined clientip=192.168.1.1 status=200"
        context = {"sourcetype": "access_combined"}
        
        print(f"Original query: {query}")
        print(f"Context: {context}")
        
        mapped_query = mapper.map_query_with_context(query, context)
        print(f"Mapped query: {mapped_query}")
        
    except SPLMapperError as e:
        print(f"Error: {e}")


def query_discovery_example():
    """Demonstrate query discovery functionality"""
    print("\n=== Query Discovery Example ===")
    
    try:
        mapper = SPLMapper()
        
        # Test different types of queries
        queries = [
            "search sourcetype=access_combined | stats count by src_ip",
            "| inputlookup ip_geo_lookup.csv | search country=US",
            "| datamodel Network_Traffic All_Traffic search | stats avg(bytes_in) by src_ip",
            "| tstats count from datamodel=Web where Web.status=200 by Web.src_ip",
        ]
        
        for i, query in enumerate(queries, 1):
            print(f"\nQuery {i}: {query}")
            
            # Discover information about the query
            info = mapper.discover_query(query)
            
            print(f"  Data Models: {info.data_models}")
            print(f"  Lookups: {info.lookups}")
            print(f"  Source Types: {info.source_types}")
            print(f"  Sources: {info.sources}")
            print(f"  Input Fields: {info.input_fields}")
            
    except SPLMapperError as e:
        print(f"Error: {e}")


def input_fields_example():
    """Demonstrate input field discovery"""
    print("\n=== Input Fields Discovery Example ===")
    
    try:
        mapper = SPLMapper()
        
        query = "search src_ip=192.168.1.1 | eval computed_field=src_ip+dst_port | stats count by src_ip, computed_field"
        print(f"Query: {query}")
        
        # Get input fields
        input_fields = mapper.get_input_fields(query)
        print(f"Input fields required: {input_fields}")
        
    except SPLMapperError as e:
        print(f"Error: {e}")


if __name__ == "__main__":
    print("SPL Toolkit Python Bindings - Examples")
    print("=" * 50)
    
    basic_mapping_example()
    conditional_mapping_example()
    query_discovery_example()
    input_fields_example()
    
    print("\nExamples completed!")