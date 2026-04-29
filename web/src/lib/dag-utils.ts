import type { WorkflowEdge, WorkflowNode } from "@/types/workflow";
import Dagre from "dagre";

// ─── Cycle Detection ───────────────────────────────────────────────────────────

export function detectCycle(nodes: WorkflowNode[], edges: WorkflowEdge[]): boolean {
  const adj: Map<string, string[]> = new Map(nodes.map((n) => [n.id, []]));
  for (const e of edges) {
    adj.get(e.source)?.push(e.target);
  }

  const color = new Map<string, "white" | "gray" | "black">(
    nodes.map((n) => [n.id, "white"])
  );

  function dfs(id: string): boolean {
    color.set(id, "gray");
    for (const neighbor of adj.get(id) ?? []) {
      if (color.get(neighbor) === "gray") return true; // back-edge → cycle
      if (color.get(neighbor) === "white" && dfs(neighbor)) return true;
    }
    color.set(id, "black");
    return false;
  }

  for (const node of nodes) {
    if (color.get(node.id) === "white" && dfs(node.id)) return true;
  }
  return false;
}

// ─── Topological Sort (Kahn's algorithm) ──────────────────────────────────────

export function topologicalSort(
  nodes: WorkflowNode[],
  edges: WorkflowEdge[]
): WorkflowNode[] {
  const inDegree = new Map<string, number>(nodes.map((n) => [n.id, 0]));
  const adj = new Map<string, string[]>(nodes.map((n) => [n.id, []]));

  for (const e of edges) {
    inDegree.set(e.target, (inDegree.get(e.target) ?? 0) + 1);
    adj.get(e.source)?.push(e.target);
  }

  const queue: string[] = [];
  for (const [id, deg] of inDegree) {
    if (deg === 0) queue.push(id);
  }

  const sorted: WorkflowNode[] = [];
  const nodeMap = new Map(nodes.map((n) => [n.id, n]));

  while (queue.length > 0) {
    const id = queue.shift()!;
    const node = nodeMap.get(id);
    if (node) sorted.push(node);
    for (const neighbor of adj.get(id) ?? []) {
      const deg = (inDegree.get(neighbor) ?? 1) - 1;
      inDegree.set(neighbor, deg);
      if (deg === 0) queue.push(neighbor);
    }
  }

  return sorted;
}

// ─── Auto-Layout with Dagre ───────────────────────────────────────────────────

const NODE_WIDTH = 280;
const NODE_HEIGHT = 120;

export function autoLayout(
  nodes: WorkflowNode[],
  edges: WorkflowEdge[]
): WorkflowNode[] {
  if (nodes.length === 0) return nodes;

  const g = new Dagre.graphlib.Graph();
  g.setGraph({ rankdir: "TB", ranksep: 80, nodesep: 60 });
  g.setDefaultEdgeLabel(() => ({}));

  for (const node of nodes) {
    const h = node.type === "condition" ? 100 : NODE_HEIGHT;
    g.setNode(node.id, { width: NODE_WIDTH, height: h });
  }
  for (const edge of edges) {
    g.setEdge(edge.source, edge.target);
  }

  Dagre.layout(g);

  return nodes.map((node) => {
    const { x, y } = g.node(node.id);
    return {
      ...node,
      position: {
        x: x - NODE_WIDTH / 2,
        y: y - NODE_HEIGHT / 2,
      },
    };
  });
}

// ─── Validate New Connection ───────────────────────────────────────────────────

/** Returns an error string if the connection is invalid, null if OK. */
export function validateConnection(
  sourceId: string,
  targetId: string,
  nodes: WorkflowNode[],
  edges: WorkflowEdge[]
): string | null {
  if (sourceId === targetId) return "Cannot connect a node to itself";

  const sourceNode = nodes.find((n) => n.id === sourceId);
  const targetNode = nodes.find((n) => n.id === targetId);
  if (!sourceNode || !targetNode) return "Node not found";

  // Triggers cannot be targets
  if (targetNode.type === "trigger") return "Cannot connect to the trigger node";

  // No duplicate edges
  const duplicate = edges.some(
    (e) => e.source === sourceId && e.target === targetId
  );
  if (duplicate) return "Connection already exists";

  // Cycle check (add the hypothetical edge and test)
  const hypothetical: WorkflowEdge = {
    id: `__test__`,
    source: sourceId,
    target: targetId,
    data: {},
  };
  if (detectCycle(nodes, [...edges, hypothetical])) return "Connection would create a cycle";

  return null;
}
