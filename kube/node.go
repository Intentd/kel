package kube

import "text/template"

// nodeFuncs — Node
//
// Domain: node
//
// Read a Node: readiness, allocatable CPU and memory, conditions, taints, and
// where it is (zone, region, instance type, from the well-known topology
// labels); and which node a pod is scheduled on.
func nodeFuncs() template.FuncMap {
	return template.FuncMap{
		"nodeReady":             NodeReady,
		"nodeAllocatableCPU":    NodeAllocatableCPU,
		"nodeAllocatableMemory": NodeAllocatableMemory,
		"nodeCondition":         NodeCondition,
		"nodeTaints":            NodeTaints,
		"podNodeName":           PodNodeName,
		"nodeZone":              NodeZone,
		"nodeRegion":            NodeRegion,
		"nodeInstanceType":      NodeInstanceType,
	}
}

// NodeReady returns `true` when the Node's Ready condition status is `"True"`.
//
// Keywords: node, kubernetes, ready, condition, boolean, health, cluster
//
// Example:
//
//	{{ nodeReady .node }}  → true
func NodeReady(obj interface{}) bool {
	return NodeCondition(obj, "Ready") == "True"
}

// NodeAllocatableCPU returns `status.allocatable.cpu`.
//
// Keywords: node, kubernetes, cpu, allocatable, resource, capacity, string
//
// Example:
//
//	{{ nodeAllocatableCPU .node }}  → "3920m"
func NodeAllocatableCPU(obj interface{}) string {
	return nodeAllocatable(obj, "cpu")
}

// NodeAllocatableMemory returns `status.allocatable.memory`.
//
// Keywords: node, kubernetes, memory, allocatable, resource, capacity, string
//
// Example:
//
//	{{ nodeAllocatableMemory .node }}  → "15032020Ki"
func NodeAllocatableMemory(obj interface{}) string {
	return nodeAllocatable(obj, "memory")
}

// NodeCondition returns the `status` string of the named node condition, or
// `""` when absent — common types are `Ready`, `MemoryPressure`, and
// `DiskPressure`.
//
// Keywords: node, kubernetes, condition, status, string, pressure, health
//
// Example:
//
//	{{ nodeCondition .node "MemoryPressure" }}  → "False"
func NodeCondition(obj interface{}, condType string) string {
	status := Status(obj)
	conditions, _ := status["conditions"].([]interface{})
	for _, c := range conditions {
		cm, _ := c.(map[string]interface{})
		if cm == nil {
			continue
		}
		if t, _ := cm["type"].(string); t == condType {
			s, _ := cm["status"].(string)
			return s
		}
	}
	return ""
}

// NodeTaints returns the keys of the node's taints.
//
// Keywords: node, kubernetes, taints, scheduling, string, toleration, placement
//
// Example:
//
//	{{ nodeTaints .node }}  → [dedicated]
func NodeTaints(node interface{}) []string {
	taints, _ := nested(object(node), "spec", "taints").([]interface{})
	var keys []string
	for _, t := range taints {
		if k := nestedString(object(t), "key"); k != "" {
			keys = append(keys, k)
		}
	}
	return keys
}

func nodeAllocatable(obj interface{}, resource string) string {
	status := Status(obj)
	allocatable, _ := status["allocatable"].(map[string]interface{})
	if allocatable == nil {
		return ""
	}
	v, _ := allocatable[resource].(string)
	return v
}

// ── Where a pod runs ──────────────────────────────────────────────────────

// PodNodeName returns the node a pod is scheduled on; "" while pending.
//
// Keywords: node, pod, topology, name, string, scheduled, placement
//
// Example:
//
//	{{ podNodeName .pod }}  → "ip-10-0-1-5.us-east-2.compute.internal"
func PodNodeName(pod interface{}) string {
	return nestedString(object(pod), "spec", "nodeName")
}

// NodeZone returns the node's zone (label topology.kubernetes.io/zone).
//
// Keywords: node, zone
//
// Example:
//
//	{{ nodeZone .node }}  → "us-east-2a"
func NodeZone(node interface{}) string {
	return nodeLabel(node, "topology.kubernetes.io/zone")
}

// NodeRegion returns the node's region (label topology.kubernetes.io/region).
//
// Keywords: node, region
//
// Example:
//
//	{{ nodeRegion .node }}  → "us-east-2"
func NodeRegion(node interface{}) string {
	return nodeLabel(node, "topology.kubernetes.io/region")
}

// NodeInstanceType returns the node's instance type (label
// node.kubernetes.io/instance-type).
//
// Keywords: instance, node, type
//
// Example:
//
//	{{ nodeInstanceType .node }}  → "m5.large"
func NodeInstanceType(node interface{}) string {
	return nodeLabel(node, "node.kubernetes.io/instance-type")
}

func nodeLabel(node interface{}, key string) string {
	labels, _ := nested(object(node), "metadata", "labels").(map[string]interface{})
	v, _ := labels[key].(string)
	return v
}
