import type { RelayWithActions } from "@/types/relay";
import type {
  DAGEdgePayload,
  DAGNodePayload,
  RelayMeta,
  SaveRelayDAGRequest,
  WorkflowEdge,
  WorkflowNode,
} from "@/types/workflow";
import { autoLayout } from "@/lib/dag-utils";
import { v4 as uuidv4 } from "uuid";

// ─── Serialize: Canvas → API ───────────────────────────────────────────────────

export function graphToDAGPayload(
  nodes: WorkflowNode[],
  edges: WorkflowEdge[],
  meta: RelayMeta
): SaveRelayDAGRequest {
  const dagNodes: DAGNodePayload[] = [];
  const dagEdges: DAGEdgePayload[] = [];

  for (const node of nodes) {
    if (node.type === "trigger") {
      // Trigger is stored in relay-level fields, not as an action node
      continue;
    }
    if (node.type === "action") {
      const d = node.data as import("@/types/workflow").ActionNodeData;
      dagNodes.push({
        node_id: d.nodeId,
        action_type: d.actionType,
        config: d.config,
      });
    }
    if (node.type === "condition") {
      const d = node.data as import("@/types/workflow").ConditionNodeData;
      // Condition nodes are a special action type on the backend
      dagNodes.push({
        node_id: d.nodeId,
        action_type: "condition",
        config: d.condition as Record<string, unknown> ?? {},
      });
    }
  }

  for (const edge of edges) {
    const sourceNode = nodes.find((n) => n.id === edge.source);
    if (!sourceNode) continue;

    // Edges from trigger node don't go into relay_edges (trigger has no node_id)
    if (sourceNode.type === "trigger") continue;

    const sourceData = sourceNode.data as { nodeId: string };
    const targetNode = nodes.find((n) => n.id === edge.target);
    if (!targetNode) continue;
    const targetData = targetNode.data as { nodeId: string };

    let edgeCondition = edge.data?.condition ?? null;

    // Automatically generate condition for edges leaving a condition node
    if (sourceNode.type === "condition" && !edgeCondition && edge.data?.branch) {
      edgeCondition = {
        field: `{{ steps['${sourceData.nodeId}'].output.result }}`,
        operator: "==",
        value: edge.data.branch,
      };
    }

    dagEdges.push({
      parent_node_id: sourceData.nodeId,
      child_node_id: targetData.nodeId,
      condition: edgeCondition,
    });
  }

  return {
    name: meta.name,
    description: meta.description || undefined,
    trigger_type: meta.triggerType,
    trigger_config: meta.triggerConfig,
    actions: dagNodes,
    edges: dagEdges,
  };
}

// ─── Deserialize: API → Canvas ─────────────────────────────────────────────────

interface RelayWithDAG extends RelayWithActions {
  edges?: Array<{
    parent_node_id: string;
    child_node_id: string;
    condition?: import("@/types/workflow").ConditionData | null;
  }>;
}

export function dagPayloadToGraph(relay: RelayWithDAG): {
  nodes: WorkflowNode[];
  edges: WorkflowEdge[];
} {
  const nodes: WorkflowNode[] = [];
  const edges: WorkflowEdge[] = [];

  // Trigger node (always present, canvas id = "trigger")
  const triggerNode: import("@/types/workflow").TriggerNode = {
    id: "trigger",
    type: "trigger",
    position: { x: 0, y: 0 },
    data: {
      nodeId: "trigger",
      triggerType: relay.trigger_type as "webhook" | "manual" | "cron",
      triggerConfig: relay.trigger_config ?? {},
      label: relay.trigger_type === "webhook"
        ? "Webhook Trigger"
        : relay.trigger_type === "cron"
        ? "Cron Trigger"
        : "Manual Trigger",
    },
    deletable: false,
  };
  nodes.push(triggerNode);

  // Map node_id → canvas node id (canvas id = node_id for simplicity)
  for (const action of relay.actions ?? []) {
    const nodeId = action.node_id;
    const isCondition = action.action_type === "condition";

    if (isCondition) {
      const condNode: import("@/types/workflow").ConditionNode = {
        id: nodeId,
        type: "condition",
        position: { x: 0, y: 0 },
        data: {
          nodeId,
          label: "Condition",
          condition: action.config as unknown as import("@/types/workflow").ConditionData,
        },
      };
      nodes.push(condNode);
    } else {
      const actionNode: import("@/types/workflow").ActionNode = {
        id: nodeId,
        type: "action",
        position: { x: 0, y: 0 },
        data: {
          nodeId,
          actionType: action.action_type,
          config: { ...action.config },
          label: action.action_type,
        },
      };
      nodes.push(actionNode);
    }
  }

  // Edges from relay_edges
  const relayEdges = relay.edges ?? [];

  // Find root action nodes (those not referenced as child_node_id)
  const childNodeIds = new Set(relayEdges.map((e) => e.child_node_id));
  for (const action of relay.actions ?? []) {
    const nodeId = action.node_id;
    if (!childNodeIds.has(nodeId)) {
      // This is a root action → connect from trigger
      edges.push({
        id: `trigger-${nodeId}`,
        source: "trigger",
        target: nodeId,
        type: "workflow",
        data: {},
      });
    }
  }

  for (const e of relayEdges) {
    let sourceHandle: string | undefined = undefined;
    let branch: "true" | "false" | null = null;
    
    const parentNode = nodes.find((n) => n.id === e.parent_node_id);
    if (parentNode && parentNode.type === "condition") {
      if (e.condition?.value === "true") {
        sourceHandle = "true";
        branch = "true";
      } else if (e.condition?.value === "false") {
        sourceHandle = "false";
        branch = "false";
      }
    }

    edges.push({
      id: `${e.parent_node_id}-${e.child_node_id}`,
      source: e.parent_node_id,
      target: e.child_node_id,
      sourceHandle,
      type: "workflow",
      data: { condition: e.condition ?? null, branch },
    });
  }

  // Apply auto-layout
  const laidOut = autoLayout(nodes, edges);
  return { nodes: laidOut, edges };
}

// ─── Default empty workflow ────────────────────────────────────────────────────

export function createDefaultWorkflow(triggerType: "webhook" | "manual" | "cron" = "webhook"): {
  nodes: WorkflowNode[];
  edges: WorkflowEdge[];
} {
  const triggerNode: import("@/types/workflow").TriggerNode = {
    id: "trigger",
    type: "trigger",
    position: { x: 300, y: 80 },
    data: {
      nodeId: "trigger",
      triggerType,
      triggerConfig: {},
      label:
        triggerType === "webhook"
          ? "Webhook Trigger"
          : triggerType === "cron"
          ? "Cron Trigger"
          : "Manual Trigger",
    },
    deletable: false,
  };
  return { nodes: [triggerNode], edges: [] };
}

// ─── Unique node ID helper ─────────────────────────────────────────────────────

export function newNodeId(prefix: string): string {
  return `${prefix}_${uuidv4().replace(/-/g, "").slice(0, 8)}`;
}
