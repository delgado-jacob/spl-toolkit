package mapper

import (
	"testing"
)

func TestNewParser(t *testing.T) {
	p := NewParser()
	if p == nil {
		t.Fatal("Expected non-nil parser")
	}
}

func TestParse(t *testing.T) {
	p := NewParser()

	// Test empty query
	_, err := p.Parse("")
	if err == nil {
		t.Error("Expected error for empty query")
	}

	// Test simple query
	query := "search index=main"
	ast, err := p.Parse(query)
	if err != nil {
		t.Fatalf("Failed to parse query: %v", err)
	}

	if ast == nil {
		t.Fatal("Expected non-nil AST")
	}

	if ast.Type != "query" {
		t.Errorf("Expected root node type 'query', got %s", ast.Type)
	}

	// Check that we have children (init_command should be present)
	if len(ast.Children) == 0 {
		t.Error("Expected query to have children")
	}

	// Check that first child is init_command
	if len(ast.Children) > 0 && ast.Children[0].Type != "init_command" {
		t.Errorf("Expected first child to be 'init_command', got %s", ast.Children[0].Type)
	}
}

func TestValidateQuery(t *testing.T) {
	p := NewParser()

	// Test valid query
	err := p.ValidateQuery("search index=main")
	if err != nil {
		t.Errorf("Expected valid query to pass validation: %v", err)
	}

	// Test invalid query
	err = p.ValidateQuery("")
	if err == nil {
		t.Error("Expected empty query to fail validation")
	}
}
