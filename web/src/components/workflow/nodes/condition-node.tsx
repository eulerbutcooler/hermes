"use client";

import { Handle, Position, type NodeProps, type Node } from "@xyflow/react";
import type { ConditionNodeData } from "@/types/workflow";
import { useWorkflow } from "@/components/workflow/workflow-provider";

type ConditionNodeType = Node<ConditionNodeData, "condition">;


export function ConditionNode({ id, data, selected }: NodeProps<ConditionNodeType>) {
  const { removeNode, selectNode } = useWorkflow();

  return (
    <div
      className={`condition-node ${selected ? "selected" : ""}`}
      onClick={() => selectNode(id)}
    >
      {/* Input handle */}
      <Handle
        type="target"
        position={Position.Top}
        id="in"
        style={{ top: -6 }}
      />

      {/* Content */}
      <div className="p-3">
        {/* Header */}
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-purple-500/15 border border-purple-500/30">
              <svg className="h-3.5 w-3.5 text-purple-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M7.5 21 3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />
              </svg>
            </div>
            <span className="text-xs font-semibold text-purple-300">{data.label}</span>
          </div>
          <button
            type="button"
            onClick={(e) => { e.stopPropagation(); removeNode(id); }}
            className="flex h-5 w-5 items-center justify-center rounded text-zinc-600 hover:bg-red-500/15 hover:text-red-400 transition-colors"
          >
            <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18 18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        {data.condition?.field ? (
          <div className="mt-1 p-1.5 rounded bg-black/30 border border-white/5 font-mono text-[9px] text-zinc-400 break-all">
            {data.condition.field} {data.condition.operator} {data.condition.value}
          </div>
        ) : (
          <p className="text-[10px] text-zinc-500 leading-relaxed">
            Click to define condition. Connect branches below.
          </p>
        )}

        {/* Branch handle labels */}
        <div className="flex justify-between mt-3 px-1">
          <span className="text-[10px] font-medium text-green-400">True →</span>
          <span className="text-[10px] font-medium text-red-400">← False</span>
        </div>
      </div>

      {/* True branch source handle (left-bottom) */}
      <Handle
        type="source"
        position={Position.Bottom}
        id="true"
        className="handle-true"
        style={{ bottom: -6, left: "30%" }}
      />

      {/* False branch source handle (right-bottom) */}
      <Handle
        type="source"
        position={Position.Bottom}
        id="false"
        className="handle-false"
        style={{ bottom: -6, left: "70%" }}
      />
    </div>
  );
}
