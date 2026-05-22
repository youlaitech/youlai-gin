package utils

// BuildTreeSimple 构建树形结构（泛型方法，支持任意类型）
func BuildTreeSimple[T any](
	items []T,
	getID func(T) int64,
	getParentID func(T) int64,
	setChildren func(*T, []T),
) []T {
	if len(items) == 0 {
		return []T{}
	}

	itemMap := make(map[int64]*T)
	for i := range items {
		id := getID(items[i])
		itemMap[id] = &items[i]
	}

	childrenMap := make(map[int64][]T)
	for i := range items {
		parentID := getParentID(items[i])
		childrenMap[parentID] = append(childrenMap[parentID], items[i])
	}

	var setChildrenRecursive func(*T)
	setChildrenRecursive = func(node *T) {
		id := getID(*node)
		if children, ok := childrenMap[id]; ok {
			// 递归处理子节点
			for i := range children {
				setChildrenRecursive(&children[i])
			}
			setChildren(node, children)
		}
	}

	var roots []T
	for i := range items {
		if getParentID(items[i]) == 0 {
			item := items[i]
			setChildrenRecursive(&item)
			roots = append(roots, item)
		}
	}

	// 如果根节点被过滤掉，使用 parentIds - ids 作为递归起点
	if len(roots) == 0 {
		rootIDSet := make(map[int64]struct{})
		for i := range items {
			pid := getParentID(items[i])
			if pid == 0 {
				continue
			}
			if _, ok := itemMap[pid]; !ok {
				rootIDSet[pid] = struct{}{}
			}
		}

		for i := range items {
			pid := getParentID(items[i])
			if _, ok := rootIDSet[pid]; ok {
				item := items[i]
				setChildrenRecursive(&item)
				roots = append(roots, item)
			}
		}
	}

	return roots
}