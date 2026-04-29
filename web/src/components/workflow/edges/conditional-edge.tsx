"use client";

import {
  BaseEdge,
  EdgeLabelRenderer,
  getBezierPath,
  type EdgeProps,
  type Edge,
} from "@xyflow/react";
import type { WorkflowEdgeData } from "@/types/workflow";
import { useWorkflow } from "@/components/workflow/workflow-provider";

type WorkflowEdge = Edge<WorkflowEdgeData>;

export function ConditionalEdge({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  data,
  selected,
  sourceHandleId,
}: EdgeProps<WorkflowEdge>) {
  const { removeEdge } = useWorkflow();

  const [edgePath, labelX, labelY] = getBezierPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  });

  // Color based on branch type
  const isTrueBranch = sourceHandleId === "true";
  const isFalseBranch = sourceHandleId === "false";
  const hasCondition = !!data?.condition;

  const strokeColor = isTrueBranch
    ? "#22c55e"
    : isFalseBranch
    ? "#ef4444"
    : "#f97316";

  const conditionLabel = hasCondition
    ? `${data!.condition!.field} ${data!.condition!.operator} ${data!.condition!.value}`
    : isTrueBranch
    ? "True"
    : isFalseBranch
    ? "False"
    : null;

  return (
    <>
      {/* Animated edge path */}
      <BaseEdge
        id={id}
        path={edgePath}
        style={{
          stroke: strokeColor,
          strokeWidth: selected ? 2.5 : 1.8,
          strokeDasharray: "0",
          filter: selected ? `drop-shadow(0 0 4px ${strokeColor}80)` : undefined,
          transition: "stroke-width 0.15s, filter 0.15s",
        }}
        markerEnd={`url(#arrow-${isTrueBranch ? "green" : isFalseBranch ? "red" : "orange"})`}
      />

      {/* Label + delete button */}
      <EdgeLabelRenderer>
        <div
          style={{
            position: "absolute",
            transform: `translate(-50%, -50%) translate(${labelX}px,${labelY}px)`,
            pointerEvents: "all",
          }}
          className="nodrag nopan"
        >
          <div className="flex items-center gap-1 group">
            {/* Condition label */}
            {conditionLabel && (
              <span
                className={`px-2 py-0.5 rounded-full text-[10px] font-medium border backdrop-blur-sm ${
                  isTrueBranch
                    ? "bg-green-500/15 text-green-300 border-green-500/30"
                    : isFalseBranch
                    ? "bg-red-500/15 text-red-300 border-red-500/30"
                    : "bg-orange-500/15 text-orange-300 border-orange-500/30"
                }`}
              >
                {conditionLabel}
              </span>
            )}

            {/* Delete edge button (shows on hover) */}
            <button
              type="button"
              onClick={() => removeEdge(id)}
              className="flex h-4 w-4 items-center justify-center rounded-full bg-zinc-800 border border-white/10 text-zinc-500 opacity-0 group-hover:opacity-100 hover:bg-red-500/20 hover:text-red-400 hover:border-red-500/30 transition-all"
              title="Remove connection"
            >
              <svg className="h-2.5 w-2.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 18 18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
      </EdgeLabelRenderer>
    </>
  );
}

// SVG marker defs — embed once in the canvas
export function EdgeMarkerDefs() {
  return (
    <svg style={{ position: "absolute", width: 0, height: 0 }}>
      <defs>
        {[
          { id: "arrow-orange", color: "#f97316" },
          { id: "arrow-green", color: "#22c55e" },
          { id: "arrow-red", color: "#ef4444" },
        ].map(({ id, color }) => (
          <marker
            key={id}
            id={id}
            viewBox="0 0 10 10"
            refX="8"
            refY="5"
            markerWidth="6"
            markerHeight="6"
            orient="auto-start-reverse"
          >
            <path d="M 0 0 L 10 5 L 0 10 z" fill={color} />
          </marker>
        ))}
      </defs>
    </svg>
  );
}
