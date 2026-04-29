import type { Edge, Node } from "@xyflow/react";

// ─── Node Data Types ───────────────────────────────────────────────────────────

export interface TriggerNodeData extends Record<string, unknown> {
  nodeId: string;
  triggerType: "webhook" | "manual" | "cron";
  triggerConfig: Record<string, unknown>;
  label: string;
}

export interface ActionNodeData extends Record<string, unknown> {
  nodeId: string;
  actionType: string;
  config: Record<string, unknown>;
  label: string;
}

export interface ConditionNodeData extends Record<string, unknown> {
  nodeId: string;
  label: string;
}

// ─── Canvas Node / Edge Types ──────────────────────────────────────────────────

export type TriggerNode = Node<TriggerNodeData, "trigger">;
export type ActionNode = Node<ActionNodeData, "action">;
export type ConditionNode = Node<ConditionNodeData, "condition">;

export type WorkflowNode = TriggerNode | ActionNode | ConditionNode;

export interface ConditionData {
  field: string;       // e.g. "steps.http.status_code" or "payload.type"
  operator: "==" | "!=" | ">" | "<" | ">=" | "<=" | "contains" | "exists";
  value: string;
}

export interface WorkflowEdgeData extends Record<string, unknown> {
  condition?: ConditionData | null;
  branch?: "true" | "false" | null; // for condition node outputs
}

export type WorkflowEdge = Edge<WorkflowEdgeData>;

// ─── Serialization / API Payloads ──────────────────────────────────────────────

export interface DAGNodePayload {
  node_id: string;
  action_type: string;
  config: Record<string, unknown>;
}

export interface DAGEdgePayload {
  parent_node_id: string;
  child_node_id: string;
  condition: ConditionData | null;
}

export interface SaveRelayDAGRequest {
  name: string;
  description?: string;
  trigger_type: "webhook" | "manual" | "cron";
  trigger_config: Record<string, unknown>;
  actions: DAGNodePayload[];
  edges: DAGEdgePayload[];
}

// ─── Relay Meta (for toolbar / modal) ─────────────────────────────────────────

export interface RelayMeta {
  name: string;
  description: string;
  triggerType: "webhook" | "manual" | "cron";
  triggerConfig: Record<string, unknown>;
}

// ─── Node palette config ───────────────────────────────────────────────────────

export interface PaletteItem {
  type: "trigger" | "action" | "condition";
  actionType?: string;
  triggerType?: "webhook" | "manual" | "cron";
  label: string;
  description: string;
  icon: string;
  color: string; // tailwind text color class
  bgColor: string; // tailwind bg color class
}
