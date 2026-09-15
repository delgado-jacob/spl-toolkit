#include "libspl_toolkit.h"
#include <assert.h>
#include <string.h>

int main(void) {
    char models[] = "| tstats count from datamodel=Web by Web.status | tstats count from datamodel=Authentication by Authentication.user | tstats count from datamodel=Network_Traffic by Network_Traffic.src_ip";
    char macros[] = "`get_data` | eval result=`calculate_score(field1, field2)` | where result>10";
    for (int i = 0; i < 200; i++) {
        int handle = spl_mapper_new();
        assert(handle > 0);
        SPLQueryInfo *a = spl_mapper_discover_query(handle, models);
        assert(a && !a->error && a->data_models_count == 3);
        assert(strcmp(a->data_models[0], "Web") == 0);
        assert(strcmp(a->data_models[1], "Authentication") == 0);
        assert(strcmp(a->data_models[2], "Network_Traffic") == 0);
        spl_query_info_free(a);
        SPLQueryInfo *b = spl_mapper_discover_query(handle, macros);
        assert(b && !b->error && b->macros_count == 2);
        assert(strcmp(b->macros[0], "get_data") == 0);
        assert(strcmp(b->macros[1], "calculate_score") == 0);
        spl_query_info_free(b);
        char *error = spl_mapper_load_mappings(handle, "[");
        assert(error && strlen(error) > 0);
        spl_string_free(error);
        char rewrite_request[] = "{\"schema_version\":1,\"mode\":\"apply\",\"document\":{\"text\":\"search src=x\"},\"rules\":[{\"id\":\"map\",\"kind\":\"field\",\"source\":{\"name\":\"src\"},\"target\":{\"name\":\"user\"}}]}";
        SPLResult *rewrite = spl_mapper_rewrite(handle, rewrite_request);
        assert(rewrite && !rewrite->error && rewrite->result);
        assert(strstr(rewrite->result, "\"text\":\"search user=x\"") != NULL);
        spl_result_free(rewrite);
        SPLResult *batch = spl_mapper_rewrite_batch(handle, "{\"schema_version\":1,\"documents\":[{\"text\":\"FROM main\",\"language\":\"spl2\"}],\"rules\":[]}");
        assert(batch && !batch->error && batch->result);
        assert(strstr(batch->result, "\"reports\":[") != NULL);
        spl_result_free(batch);
        SPLResult *malformed = spl_mapper_rewrite(handle, NULL);
        assert(malformed && malformed->error && !malformed->result);
        spl_result_free(malformed);
        malformed = spl_mapper_rewrite_batch(handle, "{\"schema_version\":1,\"documents\":[],\"rules\":[]}");
        assert(malformed && malformed->error && !malformed->result);
        spl_result_free(malformed);
        SPLResult *requirements = spl_mapper_requirements_query(handle, "{\"text\":\"search host=web\"}");
        assert(requirements && !requirements->error && requirements->result);
        assert(strstr(requirements->result, "\"query_status\":\"valid\"") != NULL);
        spl_result_free(requirements);
        malformed = spl_mapper_requirements_query(handle, NULL);
        assert(malformed && malformed->error && !malformed->result);
        spl_result_free(malformed);
        spl_mapper_free(handle);
        SPLResult *closed = spl_mapper_map_query(handle, "search src_ip=1");
        assert(closed && closed->error);
        spl_result_free(closed);
        closed = spl_mapper_rewrite(handle, rewrite_request);
        assert(closed && closed->error && !closed->result);
        spl_result_free(closed);
        closed = spl_mapper_rewrite_batch(handle, NULL);
        assert(closed && closed->error && !closed->result);
        spl_result_free(closed);
        closed = spl_mapper_requirements_query(handle, "{\"text\":\"search host=web\"}");
        assert(closed && closed->error && !closed->result);
        spl_result_free(closed);
        spl_mapper_free(handle);
    }
    spl_result_free(NULL);
    spl_query_info_free(NULL);
    spl_string_free(NULL);
    return 0;
}
