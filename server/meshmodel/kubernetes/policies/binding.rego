package meshmodel_policy

import data.meshmodel_policy.extract_components
import data.meshmodel_policy.contains
import data.meshmodel_policy.get_array_pos
import future.keywords.in
# import data.mo

# binding_types := ["mou/nt", "permission"]

evaluate[results] {
    binding_types := { type |
        selector := data.selectors.allow.from[_].match
    
        selector[key]

        key != "self"
        type = key
    }

    print(binding_types, "ppp")
    some type in binding_types
    print(type)
    # has_key(data, binding_type)
    # results := object.keys(data[binding_type].selectors)
    results = binding_relationship with data.binding_comp as type
}

binding_relationship[results] {

    print(data.selectors, "\n\n")
    # object.get(data, ["selectors"], "") != ""
    # print(data)
    # pattern := json.marshal(input.services)    
    from_selectors := { kind: selectors |
        s := data.selectors.allow.from[_]
        kind := s.kind
        selectors := s
    }

    to_selectors :=  { kind: selectors |
        t := data.selectors.allow.to[_]
        kind := t.kind
        selectors := t
    }

    # print(input.services)
    # contains "selectors.from" components only, eg: Role/ClusterRole(s) comps only
    from := extract_components(input.services, from_selectors)
    to := extract_components(input.services, to_selectors)

    # binding_selectors := { data.bindingResource : [] }
    binding_resources := extract_components(input.services, [{"kind": data.binding_comp}])

    # print("from:", from, "\n\n", "to:", to, "\n\n", "binding_resources\n\n",  binding_resources)
    # print("from-selectors: ", from_selectors, "to-selectors: ", to_selectors, "binding-selectors: ", data.binding_comp)
    services_map := { service.traits.meshmap.id: service |
        service := input.services[_]
    }

    # print(services_map)
    # results = [ result |
        some i, j, k
            resource := from[i]
            # some j
                binding_resource := binding_resources[j]

            print(resource.type, binding_resource.type, from_selectors[resource.type])
            r := is_match(resource, binding_resource, from_selectors[resource.type])
            r == true  
            print("[[[]]]")
            # some k
                to_resource := to[k]
                q := is_match(to_resource, binding_resource, to_selectors[to_resource.type])
                q == true
                print("line 48: ")

        results = {
            "from": {
                "id": resource.traits.meshmap.id
            },
            "to": {
                "id": to_resource.traits.meshmap.id
            },
            "binded_by": {
                "id": binding_resource.traits.meshmap.id
            }
        }
    # ]
    print(results)
}

is_match(resource1, resource2, from_selectors) {
    print("from: ", from_selectors)
    # match = true
    some i
    match_from := from_selectors.match.self[i]
    match_to := from_selectors.match[data.binding_comp][i]
    ans := is_feasible(match_from, match_to, resource1, resource2) 

    print("ans", ans)
    # value1 := object.get(resource1, match_from, "")
    # value2 := object.get(resource2, match_to, "/")
    # print("r1:", resource1.type, "v1:", value1, "r2:", resource2.type, "v2:", value2, "mf: ", match_from, "mt: ", match_to)
    
    #     value1 != value2
    #     match = false
    
    # print(match)
} 

is_feasible(from, to, resource1, resource2) {
    not contains(to, "_")
    print("\n\n\n 82", from, to, resource1,resource2)
    object.get(resource1, from, "") == object.get(resource2, to, "")
}

is_feasible(from, to, resource1, resource2) {
    match(from, to, resource1, resource2)    
}

is_feasible(from, to, resource1, resource2) {
    match(to, from, resource2, resource1)    
}

match(from, to, resource1, resource2) {
    print(to, ";;;")
    contains(to, "_")

    index := get_array_pos(to)
    prefix_path := array.slice(to, 0, index)
    suffix_path := array.slice(to, index + 1, count(to))

    resource_val := object.get(resource2, prefix_path, "")
    print("line96: ", index, prefix_path, suffix_path, resource_val)
    value := [val |
        
        v := resource_val[_]
        val := object.get(v, suffix_path, "")
    ]

    some i in value 
     i == object.get(resource1, from, "")
}