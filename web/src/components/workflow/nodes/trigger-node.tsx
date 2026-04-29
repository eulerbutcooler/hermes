"use client";

import { Handle, Position, type NodeProps, type Node } from "@xyflow/react";
import type { TriggerNodeData } from "@/types/workflow";

type TriggerNodeType = Node<TriggerNodeData, "trigger">;


const TRIGGER_STYLES: Record<
  string,
  { bg: string; border: string; badge: string; badgeBg: string; icon: React.ReactNode }
> = {
  webhook: {
    bg: "bg-blue-500/10",
    border: "border-blue-500/40",
    badge: "text-blue-300",
    badgeBg: "bg-blue-500/15",
    icon: (
      <svg className="h-4 w-4 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M13.19 8.688a4.5 4.5 0 0 1 1.242 7.244l-4.5 4.5a4.5 4.5 0 0 1-6.364-6.364l1.757-1.757m13.35-.622 1.757-1.757a4.5 4.5 0 0 0-6.364-6.364l-4.5 4.5a4.5 4.5 0 0 0 1.242 7.244" />
      </svg>
    ),
  },
  manual: {
    bg: "bg-purple-500/10",
    border: "border-purple-500/40",
    badge: "text-purple-300",
    badgeBg: "bg-purple-500/15",
    icon: (
      <svg className="h-4 w-4 text-purple-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 0 1 0 1.972l-11.54 6.347a1.125 1.125 0 0 1-1.667-.986V5.653Z" />
      </svg>
    ),
  },
  cron: {
    bg: "bg-amber-500/10",
    border: "border-amber-500/40",
    badge: "text-amber-300",
    badgeBg: "bg-amber-500/15",
    icon: (
      <svg className="h-4 w-4 text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
      </svg>
    ),
  },
};

export function TriggerNode({ data, selected }: NodeProps<TriggerNodeType>) {
  const style = TRIGGER_STYLES[data.triggerType] ?? TRIGGER_STYLES.webhook;

  return (
    <div className={`workflow-node ${selected ? "selected" : ""}`}>
      {/* Header */}
      <div className={`workflow-node-header ${style.bg} rounded-t-[13px]`}>
        <div className={`workflow-node-icon ${style.bg} border ${style.border}`}>
          {style.icon}
        </div>
        <span className="workflow-node-label">{data.label}</span>
        <span className={`workflow-node-badge ${style.badge} ${style.badgeBg}`}>
          Trigger
        </span>
      </div>

      {/* Body */}
      <div className="workflow-node-body">
        {data.triggerType === "webhook" && (
          <p className="text-[11px] text-zinc-500 leading-relaxed">
            Fires when a POST request is received at the webhook URL
          </p>
        )}
        {data.triggerType === "manual" && (
          <p className="text-[11px] text-zinc-500 leading-relaxed">
            Triggered manually from the dashboard or API
          </p>
        )}
        {data.triggerType === "cron" && (
          <div>
            <p className="text-[11px] text-zinc-500 mb-1">Schedule</p>
            <code className="text-[11px] text-amber-300 font-mono bg-amber-500/10 px-2 py-0.5 rounded">
              {(data.triggerConfig?.schedule as string) || "0 9 * * *"}
            </code>
          </div>
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
