package main

var KV = map[string]string{}

func Set(key string, value string) {
	KV[key] = value
}

func Get(key string) (string, bool) {
	value, found := KV[key]
	return value, found
}

func List() [][2]string {
	res := [][2]string{}
	for k, v := range KV {
		res = append(res, [2]string{k, v})
	}

	return res
}

func Delete(key string) {
	delete(KV, key)
}
