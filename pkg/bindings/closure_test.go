package main

import "testing"

func TestClosureExportReturnsOwnedErrors(t *testing.T) {
	handle := spl_mapper_new()
	result := spl_mapper_closure_query(handle, nil)
	if result == nil || result.error == nil || result.result != nil {
		t.Fatal("malformed request did not return owned error")
	}
	spl_result_free(result)
	spl_mapper_free(handle)
	result = spl_mapper_closure_query(handle, nil)
	if result == nil || result.error == nil || result.result != nil {
		t.Fatal("closed handle did not return owned error")
	}
	spl_result_free(result)
}
