package provider

import (
	"hash/fnv"

	"cuelang.org/go/pkg/strconv"
)

func expandStringList(list []interface{}) []string {
	vs := make([]string, len(list))
	for i, v := range list {
		vs[i] = v.(string)
	}
	return vs
}

func hashBytes(in []byte) string {
	hash := fnv.New64a()
	_, _ = hash.Write(in)
	return strconv.FormatUint(hash.Sum64(), 16)
}
