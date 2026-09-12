package main

import "testing"

func TestRewriteExportsReturnOwnedErrors(t *testing.T) {
	handle := spl_mapper_new()
	defer spl_mapper_free(handle)
	for _, batch := range []bool{false, true} {
		result := spl_mapper_rewrite(handle, nil)
		if batch {
			spl_result_free(result)
			result = spl_mapper_rewrite_batch(handle, nil)
		}
		if result == nil || result.error == nil || result.result != nil {
			t.Fatalf("batch=%v: malformed request did not return an owned error", batch)
		}
		spl_result_free(result)
	}
	spl_mapper_free(handle)
	for _, batch := range []bool{false, true} {
		result := spl_mapper_rewrite(handle, nil)
		if batch {
			spl_result_free(result)
			result = spl_mapper_rewrite_batch(handle, nil)
		}
		if result == nil || result.error == nil || result.result != nil {
			t.Fatalf("batch=%v: closed handle did not return an owned error", batch)
		}
		spl_result_free(result)
	}
}

func TestOwnedRewriteAdmissionSurvivesConcurrentDestroy(t *testing.T) {
	handle := spl_mapper_new()
	admitted, destroyed, complete := make(chan struct{}), make(chan struct{}), make(chan struct{})
	go func() {
		defer close(complete)
		result := ownedMapperJSONResult(handle, func() (any, error) {
			close(admitted)
			<-destroyed
			return map[string]string{"status": "valid"}, nil
		})
		defer spl_result_free(result)
		if result == nil || result.error != nil || result.result == nil {
			t.Error("destroy canceled the admitted operation")
		}
	}()
	<-admitted
	spl_mapper_free(handle)
	close(destroyed)
	<-complete
	result := ownedMapperJSONResult(handle, func() (any, error) {
		t.Error("closed handle admitted another operation")
		return nil, nil
	})
	defer spl_result_free(result)
	if result == nil || result.error == nil || result.result != nil {
		t.Fatal("closed handle did not return an owned error")
	}
}
