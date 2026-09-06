#!/usr/bin/env python3
"""Runnable examples for the installed SPL Toolkit Python package."""

from spl_toolkit import SPLMapper, SPLMapperError


def main() -> int:
    mappings = [
        {"source": "src_ip", "target": "source_ip"},
        {"source": "dst_port", "target": "destination_port"},
    ]
    with SPLMapper() as mapper:
        mapper.load_mappings(mappings)
        query = "search src_ip=192.168.1.1 dst_port=80"
        print(f"Original query: {query}")
        print(f"Mapped query: {mapper.map_query(query)}")

    config = {
        "version": "1.0",
        "mappings": [{"source": "src_ip", "target": "base_ip"}],
        "rules": [
            {
                "id": "web",
                "enabled": True,
                "priority": 1,
                "conditions": [
                    {"type": "sourcetype", "operator": "equals", "value": "web"}
                ],
                "mappings": [{"source": "src_ip", "target": "source_ip"}],
            }
        ],
    }
    with SPLMapper(config=config) as mapper:
        query = "search sourcetype=web src_ip=1"
        context = {"sourcetype": ["mail", "web"]}
        print(f"Conditional mapping: {mapper.map_query_with_context(query, context)}")

    with SPLMapper() as mapper:
        info = mapper.discover_query("| inputlookup ip_geo_lookup.csv | search country=US")
        print(f"Lookups: {info.lookups}")
        print(f"Input fields: {info.input_fields}")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except SPLMapperError as error:
        print(f"SPL Toolkit error: {error}")
        raise SystemExit(1) from error
