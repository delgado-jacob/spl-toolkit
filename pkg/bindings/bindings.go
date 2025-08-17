package main

/*
#include <stdlib.h>
#include <string.h>

typedef struct {
    char* error;
    char* result;
} SPLResult;

typedef struct {
    char** data_models;
    char** datasets;
    char** lookups;
    char** macros;
    char** sources;
    char** source_types;
    char** input_fields;
    int data_models_count;
    int datasets_count;
    int lookups_count;
    int macros_count;
    int sources_count;
    int source_types_count;
    int input_fields_count;
    char* error;
} SPLQueryInfo;

static char* allocate_string(const char* str) {
    if (str == NULL) return NULL;
    size_t len = strlen(str) + 1;
    char* result = malloc(len);
    if (result) {
        strcpy(result, str);
    }
    return result;
}

static char** allocate_string_array(int count) {
    if (count <= 0) return NULL;
    return (char**)malloc(count * sizeof(char*));
}

static void free_string_array(char** arr, int count) {
    if (arr == NULL) return;
    for (int i = 0; i < count; i++) {
        if (arr[i]) free(arr[i]);
    }
    free(arr);
}
*/
import "C"

import (
	"encoding/json"
	"unsafe"

	"github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

// Global mapper instances (in practice, you'd want better memory management)
var mappers = make(map[int]*mapper.Mapper)
var nextMapperID = 1

//export spl_mapper_new
func spl_mapper_new() C.int {
	m := mapper.New()
	id := nextMapperID
	mappers[id] = m
	nextMapperID++
	return C.int(id)
}

//export spl_mapper_new_with_config
func spl_mapper_new_with_config(configJSON *C.char) C.int {
	jsonStr := C.GoString(configJSON)

	config, err := mapper.LoadMappingConfig([]byte(jsonStr))
	if err != nil {
		return -1 // Error creating mapper
	}

	m := mapper.NewWithConfig(config)
	id := nextMapperID
	mappers[id] = m
	nextMapperID++
	return C.int(id)
}

//export spl_mapper_free
func spl_mapper_free(mapperID C.int) {
	delete(mappers, int(mapperID))
}

//export spl_mapper_load_mappings
func spl_mapper_load_mappings(mapperID C.int, mappingsJSON *C.char) *C.char {
	m, exists := mappers[int(mapperID)]
	if !exists {
		return C.CString("Mapper not found")
	}

	jsonStr := C.GoString(mappingsJSON)
	err := m.LoadMappings([]byte(jsonStr))
	if err != nil {
		return C.CString(err.Error())
	}

	return nil // Success
}

//export spl_mapper_map_query
func spl_mapper_map_query(mapperID C.int, query *C.char) *C.SPLResult {
	result := (*C.SPLResult)(C.malloc(C.sizeof_SPLResult))
	result.error = nil
	result.result = nil

	m, exists := mappers[int(mapperID)]
	if !exists {
		result.error = C.CString("Mapper not found")
		return result
	}

	queryStr := C.GoString(query)
	mappedQuery, err := m.MapQuery(queryStr)
	if err != nil {
		result.error = C.CString(err.Error())
		return result
	}

	result.result = C.CString(mappedQuery)
	return result
}

//export spl_mapper_map_query_with_context
func spl_mapper_map_query_with_context(mapperID C.int, query *C.char, contextJSON *C.char) *C.SPLResult {
	result := (*C.SPLResult)(C.malloc(C.sizeof_SPLResult))
	result.error = nil
	result.result = nil

	m, exists := mappers[int(mapperID)]
	if !exists {
		result.error = C.CString("Mapper not found")
		return result
	}

	queryStr := C.GoString(query)
	contextStr := C.GoString(contextJSON)

	var context map[string]interface{}
	if err := json.Unmarshal([]byte(contextStr), &context); err != nil {
		result.error = C.CString("Invalid context JSON: " + err.Error())
		return result
	}

	mappedQuery, err := m.MapQueryWithContext(queryStr, context)
	if err != nil {
		result.error = C.CString(err.Error())
		return result
	}

	result.result = C.CString(mappedQuery)
	return result
}

//export spl_mapper_discover_query
func spl_mapper_discover_query(mapperID C.int, query *C.char) *C.SPLQueryInfo {
	result := (*C.SPLQueryInfo)(C.malloc(C.sizeof_SPLQueryInfo))

	// Initialize all fields
	result.data_models = nil
	result.datasets = nil
	result.lookups = nil
	result.macros = nil
	result.sources = nil
	result.source_types = nil
	result.input_fields = nil
	result.data_models_count = 0
	result.datasets_count = 0
	result.lookups_count = 0
	result.macros_count = 0
	result.sources_count = 0
	result.source_types_count = 0
	result.input_fields_count = 0
	result.error = nil

	m, exists := mappers[int(mapperID)]
	if !exists {
		result.error = C.CString("Mapper not found")
		return result
	}

	queryStr := C.GoString(query)
	info, err := m.DiscoverQuery(queryStr)
	if err != nil {
		result.error = C.CString(err.Error())
		return result
	}

	// Convert Go slices to C arrays
	result.data_models_count = C.int(len(info.DataModels))
	if len(info.DataModels) > 0 {
		result.data_models = C.allocate_string_array(result.data_models_count)
		for i, dm := range info.DataModels {
			result.data_models = (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(result.data_models)) + uintptr(i)*unsafe.Sizeof(*result.data_models)))
			*result.data_models = C.allocate_string(C.CString(dm))
		}
		// Reset pointer to beginning
		result.data_models = (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(result.data_models)) - uintptr(len(info.DataModels)-1)*unsafe.Sizeof(*result.data_models)))
	}

	result.datasets_count = C.int(len(info.Datasets))
	if len(info.Datasets) > 0 {
		result.datasets = C.allocate_string_array(result.datasets_count)
		for i, ds := range info.Datasets {
			dsPtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(result.datasets)) + uintptr(i)*unsafe.Sizeof(*result.datasets)))
			*dsPtr = C.allocate_string(C.CString(ds))
		}
	}

	result.lookups_count = C.int(len(info.Lookups))
	if len(info.Lookups) > 0 {
		result.lookups = C.allocate_string_array(result.lookups_count)
		for i, lookup := range info.Lookups {
			lookupPtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(result.lookups)) + uintptr(i)*unsafe.Sizeof(*result.lookups)))
			*lookupPtr = C.allocate_string(C.CString(lookup))
		}
	}

	result.source_types_count = C.int(len(info.SourceTypes))
	if len(info.SourceTypes) > 0 {
		result.source_types = C.allocate_string_array(result.source_types_count)
		for i, st := range info.SourceTypes {
			stPtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(result.source_types)) + uintptr(i)*unsafe.Sizeof(*result.source_types)))
			*stPtr = C.allocate_string(C.CString(st))
		}
	}

	result.sources_count = C.int(len(info.Sources))
	if len(info.Sources) > 0 {
		result.sources = C.allocate_string_array(result.sources_count)
		for i, src := range info.Sources {
			srcPtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(result.sources)) + uintptr(i)*unsafe.Sizeof(*result.sources)))
			*srcPtr = C.allocate_string(C.CString(src))
		}
	}

	result.input_fields_count = C.int(len(info.InputFields))
	if len(info.InputFields) > 0 {
		result.input_fields = C.allocate_string_array(result.input_fields_count)
		for i, field := range info.InputFields {
			fieldPtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(result.input_fields)) + uintptr(i)*unsafe.Sizeof(*result.input_fields)))
			*fieldPtr = C.allocate_string(C.CString(field))
		}
	}

	return result
}

//export spl_result_free
func spl_result_free(result *C.SPLResult) {
	if result == nil {
		return
	}
	if result.error != nil {
		C.free(unsafe.Pointer(result.error))
	}
	if result.result != nil {
		C.free(unsafe.Pointer(result.result))
	}
	C.free(unsafe.Pointer(result))
}

//export spl_query_info_free
func spl_query_info_free(info *C.SPLQueryInfo) {
	if info == nil {
		return
	}

	C.free_string_array(info.data_models, info.data_models_count)
	C.free_string_array(info.datasets, info.datasets_count)
	C.free_string_array(info.lookups, info.lookups_count)
	C.free_string_array(info.macros, info.macros_count)
	C.free_string_array(info.sources, info.sources_count)
	C.free_string_array(info.source_types, info.source_types_count)
	C.free_string_array(info.input_fields, info.input_fields_count)

	if info.error != nil {
		C.free(unsafe.Pointer(info.error))
	}

	C.free(unsafe.Pointer(info))
}

func main() {
	// Required for building as a shared library
}
