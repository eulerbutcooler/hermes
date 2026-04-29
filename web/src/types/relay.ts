export interface Relay {
  id: string;
  user_id: string;
  name: string;
  description: string;
  webhook_path: string;
  webhook_url: string;
  trigger_type: TriggerType;
  trigger_config: Record<string, unknown>;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface RelayAction {
  id: string;
  relay_id: string;
  node_id: string;
  action_type: string;
  config: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface RelayWithActions extends Relay {
  actions: RelayAction[];
}

export interface HTTPRequestStepOutput {
  status_code: number;
  content_type?: string;
  headers?: Record<string, string>;
  body_json?: unknown;
  body_text?: string;
  duration_ms?: number;
}

export interface Execution {
  id: string;
  relay_id: string;
  event_id?: string;
  status: "running" | "success" | "failed";
  trigger_payload?: Record<string, unknown>;
  error_message?: string;
  started_at: string;
  finished_at?: string;
  steps?: ExecutionStep[];
}

export interface ExecutionStep {
  id: string;
  execution_id: string;
  node_id: string;
  action_type: string;
  status: "running" | "success" | "failed";
  input?: Record<string, unknown>;
  output?: Record<string, unknown>;
  error_message?: string;
  started_at: string;
  finished_at?: string;
}

export interface RelayEdge {
  parent_node_id: string;
  child_node_id: string;
  condition?: Record<string, unknown>;
}

export interface Secret {
  id: string;
  user_id: string;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface CreateRelayActionInput {
  node_id: string;
  action_type: string;
  config: Record<string, unknown>;
}

export interface CreateRelayRequest {
  name: string;
  description?: string;
  actions: CreateRelayActionInput[];
  edges?: RelayEdge[];
  trigger_type?: TriggerType;
  trigger_config?: Record<string, unknown>;
}

export interface UpdateRelayRequest {
  name?: string;
  description?: string;
  is_active?: boolean;
  trigger_type?: TriggerType;
  trigger_config?: Record<string, unknown>;
}

export interface CreateSecretRequest {
  name: string;
  value: string;
}

export interface Connection {
  id: string;
  user_id: string;
  provider: "google" | "microsoft";
  account_email: string;
  scopes: string;
  created_at: string;
  updated_at: string;
}

export interface UpdateRelayActionsRequest {
  actions: CreateRelayActionInput[];
  edges?: RelayEdge[];
}

export const ACTION_TYPES = [
  "discord_send",
  "slack_send",
  "http_request",
  "email_send",
  "debug_log",
] as const;

export type ActionType = (typeof ACTION_TYPES)[number];

export const ACTION_LABELS: Record<ActionType, string> = {
  discord_send: "Discord",
  slack_send: "Slack",
  http_request: "HTTP Request",
  email_send: "Email",
  debug_log: "Debug Log",
};

export type TriggerType = "webhook" | "manual" | "cron";

export const TRIGGER_LABELS: Record<TriggerType, string> = {
  webhook: "Webhook",
  manual: "Manual",
  cron: "Cron",
};
