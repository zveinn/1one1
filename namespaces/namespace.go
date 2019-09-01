package namespaces

var ns map[int]string

func Init() {
	ns = make(map[int]string)
	ns[1] = "cpu.percentage"
	ns[2] = "disk.percentage"
	ns[3] = "memory.percentage"
	ns[4] = "network.in"
	ns[5] = "network.out"
}
func GetIndexFromNamespace(namespace string) int {
	for i, v := range ns {
		if v == namespace {
			return i
		}
	}
	return 0
}

func GetNamespaceFromIndex(index int) (namespace string) {
	namespace, ok := ns[index]
	if !ok {
		return ""
	}
	return
}
func GetIndexesFromNamespaces(namespaces []string) []int {
	var indexes []int
	for i, listNamespace := range ns {
		for _, namespace := range namespaces {
			if listNamespace == namespace {
				indexes = append(indexes, i)
			}
		}
	}
	return indexes
}
func MakeMapFromNamespaces(namespaces []string) map[int]string {
	outgoingMap := make(map[int]string)
	for i, v := range ns {
		for _, iv := range namespaces {
			if v == iv {
				outgoingMap[i] = iv
			}
		}
	}
	return outgoingMap
}
