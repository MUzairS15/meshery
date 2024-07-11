package common

import rego.v1

resolve_path(arr, mutated) := path if {
	arr_contains(arr, "_")
	index := array_path_position(arr)

	prefix_path := array.slice(arr, 0, index)
	suffix_path := array.slice(arr, index + 1, count(arr))

	value_to_patch := object.get(mutated, prefix_path, "")

	index_to_patch := array_index_to_patch(count(value_to_patch))

	intermediate_path := array.concat(prefix_path, [index_to_patch])

	path = array.concat(intermediate_path, suffix_path)
}

resolve_path(arr, mutated) := path if {
	not arr_contains(arr, "_")
	path = arr
}

array_index_to_patch(no_of_elements) := index if {
	no_of_elements == 0

	# 0 based array indexing is followed
	index = "0"
}

array_index_to_patch(no_of_elements) := index if {
	not no_of_elements == 0

	# 0 based array indexing is followed
	index = format_int(no_of_elements - 1, 10)
}

arr_contains(arr, key) if {
	some element in arr
	key == element
} else := false

# Returns the index at which the mutated or mutator paths indicates the presence of an object of type: "array".
# eg: [spec, containers, _, env]:
#  It indicates that the type of "containers" field inside the configuration is of type "array".
# Hence try to append the configuration at the last available index.
array_path_position(arr_path) := index if {
	some index
	arr_path[index] == "_"
}

extract_components(services, selectors) := {component.traits.meshmap.id: component |
	selector := selectors[_]
	service := services[_]

	is_relationship_feasible(selector, service.type)
	component := service
}

is_relationship_feasible(selector, compType) if {
	selector.kind == "*"
}

is_relationship_feasible(selector, compType) if {
	selector.kind == compType
}

has_key(x, k) if {
	x[k]
}
