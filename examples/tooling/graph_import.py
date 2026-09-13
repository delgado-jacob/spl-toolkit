"""Import canonical graph JSON without merging revision-qualified node IDs."""
import json
import sys

with open(sys.argv[1], encoding="utf-8") as source:
    graph = json.load(source)
if graph["schema_version"] != 1:
    raise ValueError("unsupported graph version")
nodes = {node["id"]: node for node in graph["nodes"]}
if len(nodes) != len(graph["nodes"]):
    raise ValueError("duplicate node identity")
adjacency = {identity: [] for identity in nodes}
for edge in graph["edges"]:
    if edge["from"] not in nodes or edge["to"] not in nodes:
        raise ValueError("unresolved graph endpoint")
    adjacency[edge["from"]].append(edge)  # Keep relation, conditionality and evidence.
print(json.dumps({"status": graph["status"], "coverage": graph["coverage"],
                  "execution_complete": graph["execution_complete"], "adjacency": adjacency}, indent=2))
