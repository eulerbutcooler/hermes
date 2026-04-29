"use client";

import { Handle, Position, type NodeProps, type Node } from "@xyflow/react";
import type { ActionNodeData } from "@/types/workflow";
import { useWorkflow } from "@/components/workflow/workflow-provider";

type ActionNodeType = Node<ActionNodeData, "action">;

const ACTION_STYLES: Record<
  string,
  { icon: React.ReactNode; color: string; bg: string; border: string }
> = {
  discord_send: {
    icon: (
      <svg className="h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
        <path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057.099 18.082.114 18.105.132 18.12a19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028c.462-.63.874-1.295 1.226-1.994a.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03z" />
      </svg>
    ),
    color: "text-indigo-400",
    bg: "bg-indigo-500/10",
    border: "border-indigo-500/30",
  },
  slack_send: {
    icon: (
      <svg className="h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
        <path d="M5.042 15.165a2.528 2.528 0 0 1-2.52 2.523A2.528 2.528 0 0 1 0 15.165a2.527 2.527 0 0 1 2.522-2.52h2.52v2.52zm1.271 0a2.527 2.527 0 0 1 2.521-2.52 2.527 2.527 0 0 1 2.521 2.52v6.313A2.528 2.528 0 0 1 8.834 24a2.528 2.528 0 0 1-2.521-2.522v-6.313zM8.834 5.042a2.528 2.528 0 0 1-2.521-2.52A2.528 2.528 0 0 1 8.834 0a2.528 2.528 0 0 1 2.521 2.522v2.52H8.834zm0 1.271a2.528 2.528 0 0 1 2.521 2.521 2.528 2.528 0 0 1-2.521 2.521H2.522A2.528 2.528 0 0 1 0 8.834a2.528 2.528 0 0 1 2.522-2.521h6.312zm10.122 2.521a2.528 2.528 0 0 1 2.522-2.521A2.528 2.528 0 0 1 24 8.834a2.528 2.528 0 0 1-2.522 2.521h-2.522V8.834zm-1.268 0a2.528 2.528 0 0 1-2.523 2.521 2.527 2.527 0 0 1-2.52-2.521V2.522A2.527 2.527 0 0 1 15.165 0a2.528 2.528 0 0 1 2.523 2.522v6.312zm-2.523 10.122a2.528 2.528 0 0 1 2.523 2.522A2.528 2.528 0 0 1 15.165 24a2.527 2.527 0 0 1-2.52-2.522v-2.522h2.52zm0-1.268a2.527 2.527 0 0 1-2.52-2.523 2.526 2.526 0 0 1 2.52-2.52h6.313A2.527 2.527 0 0 1 24 15.165a2.528 2.528 0 0 1-2.522 2.523h-6.313z" />
      </svg>
    ),
    color: "text-green-400",
    bg: "bg-green-500/10",
    border: "border-green-500/30",
  },
  http_request: {
    icon: (
      <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M12 21a9.004 9.004 0 0 0 8.716-6.747M12 21a9.004 9.004 0 0 1-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 0 1 7.843 4.582M12 3a8.997 8.997 0 0 0-7.843 4.582m15.686 0A11.953 11.953 0 0 1 12 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0 1 21 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0 1 12 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 0 1 3 12c0-1.605.42-3.113 1.157-4.418" />
      </svg>
    ),
    color: "text-orange-400",
    bg: "bg-orange-500/10",
    border: "border-orange-500/30",
  },
  email_send: {
    icon: (
      <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 0 1-2.25 2.25h-15a2.25 2.25 0 0 1-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25m19.5 0v.243a2.25 2.25 0 0 1-1.07 1.916l-7.5 4.615a2.25 2.25 0 0 1-2.36 0L3.32 8.91a2.25 2.25 0 0 1-1.07-1.916V6.75" />
      </svg>
    ),
    color: "text-sky-400",
    bg: "bg-sky-500/10",
    border: "border-sky-500/30",
  },
  debug_log: {
    icon: (
      <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M6.75 7.5 3 12m0 0 3.75 4.5M3 12h18" />
      </svg>
    ),
    color: "text-zinc-400",
    bg: "bg-zinc-500/10",
    border: "border-zinc-500/30",
  },
};

const ACTION_LABELS: Record<string, string> = {
  discord_send: "Discord",
  slack_send: "Slack",
  http_request: "HTTP Request",
  email_send: "Email",
  debug_log: "Debug Log",
};

export function ActionNode({ id, data, selected }: NodeProps<ActionNodeType>) {
  const { removeNode, selectNode } = useWorkflow();
  const style = ACTION_STYLES[data.actionType] ?? ACTION_STYLES.debug_log;
  const label = ACTION_LABELS[data.actionType] ?? data.actionType;

  // Check if minimally configured
  const isConfigured = Object.keys(data.config).some(
    (k) => data.config[k] !== "" && data.config[k] !== undefined
  );

  return (
    <div
      className={`workflow-node ${selected ? "selected" : ""}`}
      onClick={() => selectNode(id)}
    >
      {/* Input handle */}
      <Handle
        type="target"
        position={Position.Top}
        id="in"
        style={{ top: -6 }}
      />

      {/* Header */}
      <div className="workflow-node-header">
        <div className={`workflow-node-icon ${style.bg} border ${style.border}`}>
          <span className={style.color}>{style.icon}</span>
        </div>
        <span className="workflow-node-label">{label}</span>
        <div className="flex items-center gap-1.5">
          {/* Config status dot */}
          <span
            title={isConfigured ? "Configured" : "Needs configuration"}
            className={`h-2 w-2 rounded-full flex-shrink-0 ${
              isConfigured ? "bg-green-400" : "bg-amber-400 animate-pulse"
            }`}
          />
          {/* Delete */}
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              removeNode(id);
            }}
            className="flex h-5 w-5 items-center justify-center rounded text-zinc-600 hover:bg-red-500/15 hover:text-red-400 transition-colors"
            title="Remove node"
          >
            <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18 18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>

      {/* Body preview */}
      <div className="workflow-node-body">
        {data.actionType === "debug_log" ? (
          <p className="text-[11px] text-zinc-500">Logs payload — no config needed</p>
        ) : isConfigured ? (
          <p className="text-[11px] text-zinc-500 truncate">
            {getConfigPreview(data.actionType, data.config)}
          </p>
        ) : (
          <p className="text-[11px] text-amber-500/80">⚠ Click to configure</p>
        )}
      </div>

      {/* Output handle */}
      <Handle
        type="source"
        position={Position.Bottom}
        id="out"
        style={{ bottom: -6 }}
      />
    </div>
  );
}

function getConfigPreview(actionType: string, config: Record<string, unknown>): string {
  switch (actionType) {
    case "discord_send":
    case "slack_send":
      return String(config.message_template || config.webhook_url || "Configured");
    case "http_request":
      return `${config.method ?? "POST"} ${config.url ?? ""}`;
    case "email_send":
      return `To: ${config.to ?? ""}`;
    default:
      return "Configured";
  }
}
