package main

/*
#include <stdlib.h>

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
	"runtime"
	"unsafe"

	"github.com/delgado-jacob/spl-toolkit/internal/buildinfo"
	"github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

var registry = newMapperRegistry()

func cStrings(values []string) **C.char {
	if len(values) == 0 {
		return nil
	}
	base := (**C.char)(C.malloc(C.size_t(len(values)) * C.size_t(unsafe.Sizeof(uintptr(0)))))
	items := unsafe.Slice(base, len(values))
	for i, value := range values {
		items[i] = C.CString(value)
	}
	return base
}

//export spl_mapper_new
func spl_mapper_new() C.int {
	m := mapper.New()
	id, ok := registry.add(m)
	if !ok {
		return -1
	}
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
	id, ok := registry.add(m)
	if !ok {
		return -1
	}
	return C.int(id)
}

//export spl_mapper_free
func spl_mapper_free(mapperID C.int) {
	registry.remove(int(mapperID))
}

//export spl_mapper_load_mappings
func spl_mapper_load_mappings(mapperID C.int, mappingsJSON *C.char) *C.char {
	m, exists := registry.get(int(mapperID))
	if !exists {
		return C.CString("Mapper not found")
	}
	defer runtime.KeepAlive(m)

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

	m, exists := registry.get(int(mapperID))
	if !exists {
		result.error = C.CString("Mapper not found")
		return result
	}
	defer runtime.KeepAlive(m)

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

	m, exists := registry.get(int(mapperID))
	if !exists {
		result.error = C.CString("Mapper not found")
		return result
	}
	defer runtime.KeepAlive(m)

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

	m, exists := registry.get(int(mapperID))
	if !exists {
		result.error = C.CString("Mapper not found")
		return result
	}
	defer runtime.KeepAlive(m)

	queryStr := C.GoString(query)
	info, err := m.DiscoverQuery(queryStr)
	if err != nil {
		result.error = C.CString(err.Error())
		return result
	}

	result.data_models_count = C.int(len(info.DataModels))
	result.data_models = cStrings(info.DataModels)
	result.datasets_count = C.int(len(info.Datasets))
	result.datasets = cStrings(info.Datasets)
	result.lookups_count = C.int(len(info.Lookups))
	result.lookups = cStrings(info.Lookups)
	result.macros_count = C.int(len(info.Macros))
	result.macros = cStrings(info.Macros)
	result.sources_count = C.int(len(info.Sources))
	result.sources = cStrings(info.Sources)
	result.source_types_count = C.int(len(info.SourceTypes))
	result.source_types = cStrings(info.SourceTypes)
	result.input_fields_count = C.int(len(info.InputFields))
	result.input_fields = cStrings(info.InputFields)

	return result
}

//export spl_string_free
func spl_string_free(value *C.char) {
	C.free(unsafe.Pointer(value))
}

//export spl_toolkit_version
func spl_toolkit_version() *C.char {
	return C.CString(buildinfo.Version)
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
