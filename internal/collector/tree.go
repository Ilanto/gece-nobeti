package collector

// BuildProcessTree converts a flat process list into a hierarchical tree.
// ParentPID 0 means orphan (init process is special-cased).
func BuildProcessTree(processes []ProcessInfo) []*ProcessNode {
	if len(processes) == 0 {
		return nil
	}

	// Build lookup map
	pidMap := make(map[uint32]*ProcessInfo)
	for i := range processes {
		pidMap[processes[i].PID] = &processes[i]
	}

	// Find all PIDs
	allPIDs := make(map[uint32]bool)
	for _, p := range processes {
		allPIDs[p.PID] = true
	}

	// Build children map
	children := make(map[uint32][]*ProcessInfo)
	for i := range processes {
		ppid := processes[i].ParentPID
		children[ppid] = append(children[ppid], &processes[i])
	}

	// Recursive builder
	var build func(p *ProcessInfo, depth int, isOrphan bool) *ProcessNode
	build = func(p *ProcessInfo, depth int, isOrphan bool) *ProcessNode {
		node := &ProcessNode{
			Process:  *p,
			Depth:    depth,
			IsOrphan: isOrphan,
		}
		childList := children[p.PID]
		if len(childList) > 0 {
			node.Children = make([]*ProcessNode, 0, len(childList))
			for _, c := range childList {
				node.Children = append(node.Children, build(c, depth+1, false))
			}
		}
		return node
	}

	// Find roots: processes whose ParentPID is not in the pid map (orphans)
	// but PID 1 (init) is always a root even if we can't find its parent
	var roots []*ProcessNode
	for i := range processes {
		p := &processes[i]
		ppid := p.ParentPID
		if ppid == 0 || !allPIDs[ppid] {
			// Orphan or reparented to init
			isOrphan := ppid != 0 && !allPIDs[ppid] && p.PID != 1
			roots = append(roots, build(p, 0, isOrphan))
		}
	}

	return roots
}