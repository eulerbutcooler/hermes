"use client";

import { useCallback, useEffect, useRef } from "react";
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  type Connection,
  type OnConnect,
  BackgroundVariant,
  useReactFlow,
  SelectionMode,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import "./nodes/node-styles.css";

import { useWorkflow } from "./workflow-provider";
import { nodeTypes } from "./nodes/node-types";
import { edgeTypes } from "./edges/edge-types";
import { EdgeMarkerDefs } from "./edges/conditional-edge";
import { Palette } from "./sidebar/palette";
import { NodeConfigPanel } from "./sidebar/node-config-panel";
import { CanvasToolbar } from "./toolbar/canvas-toolbar";

import type { PaletteItem, WorkflowEdge, WorkflowNode } from "@/types/workflow";
import { validateConnection } from "@/lib/dag-utils";
import { newNodeId } from "@/lib/workflow-serializer";
import { toast } from "sonner";
import { v4 as uuidv4 } from "uuid";

interface WorkflowCanvasProps {
  onSave: () => Promise<void>;
  onTrigger?: () => void;
  isSaving: boolean;
}

function CanvasInner({ onSave, onTrigger, isSaving }: WorkflowCanvasProps) {
  const reactFlowWrapper = useRef<HTMLDivElement>(null);
  const { screenToFlowPosition } = useReactFlow();

  const {
    nodes,
    edges,
    selectedNodeId,
    selectedEdgeId,
    onNodesChange,
    onEdgesChange,
    addNode,
    addEdge,
    selectNode,
    selectEdge,
    undo,
    redo,
  } = useWorkflow();

  const hasTrigger = nodes.some((n) => n.type === "trigger");
  const showConfigPanel = selectedNodeId !== null || selectedEdgeId !== null;

  // ── Keyboard shortcuts ─────────────────────────────────────────────────────

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement).tagName;
      if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;

      if ((e.ctrlKey || e.metaKey) && e.key === "z" && !e.shiftKey) {
        e.preventDefault();
        undo();
      }
      if ((e.ctrlKey || e.metaKey) && (e.key === "y" || (e.key === "z" && e.shiftKey))) {
        e.preventDefault();
        redo();
      }
      if ((e.ctrlKey || e.metaKey) && e.key === "s") {
        e.preventDefault();
        onSave();
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [undo, redo, onSave]);

  // ── Drag-and-drop from palette ─────────────────────────────────────────────

  const onDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
  }, []);

  const onDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      const raw = e.dataTransfer.getData("application/hermes-node");
      if (!raw) return;

      let item: PaletteItem;
      try {
        item = JSON.parse(raw);
      } catch {
        return;
      }

      // Prevent duplicate trigger
      if (item.type === "trigger" && hasTrigger) {
        toast.error("Only one trigger node is allowed per relay");
        return;
      }

      const position = screenToFlowPosition({
        x: e.clientX,
        y: e.clientY,
      });

      let newNode: WorkflowNode;

      if (item.type === "trigger") {
        newNode = {
          id: "trigger",
          type: "trigger",
          position,
          data: {
            nodeId: "trigger",
            triggerType: item.triggerType!,
            triggerConfig: {},
            label:
              item.triggerType === "webhook"
                ? "Webhook Trigger"
                : item.triggerType === "cron"
                  ? "Cron Trigger"
                  : "Manual Trigger",
          },
          deletable: false,
        } as WorkflowNode;
      } else if (item.type === "condition") {
        const nodeId = newNodeId("cond");
        newNode = {
          id: nodeId,
          type: "condition",
          position,
          data: {
            nodeId,
            label: "Condition",
            condition: { field: "", operator: "==", value: "" },
          },
        } as WorkflowNode;
      } else {
        const nodeId = newNodeId(item.actionType ?? "action");
        newNode = {
          id: nodeId,
          type: "action",
          position,
          data: {
            nodeId,
            actionType: item.actionType!,
            config: {},
            label: item.label,
          },
        } as WorkflowNode;
      }

      addNode(newNode);
    },
    [hasTrigger, screenToFlowPosition, addNode]
  );

  // ── Connection handler ─────────────────────────────────────────────────────

  const onConnect: OnConnect = useCallback(
    (connection: Connection) => {
      const error = validateConnection(
        connection.source!,
        connection.target!,
        nodes,
        edges
      );
      if (error) {
        toast.error(error);
        return;
      }

      // Determine branch type from handle
      const branch =
        connection.sourceHandle === "true"
          ? "true"
          : connection.sourceHandle === "false"
            ? "false"
            : null;

      const newEdge: WorkflowEdge = {
        id: `e-${uuidv4().slice(0, 8)}`,
        source: connection.source!,
        target: connection.target!,
        sourceHandle: connection.sourceHandle ?? undefined,
        targetHandle: connection.targetHandle ?? undefined,
        type: "workflow",
        data: { condition: null, branch },
      };

      addEdge(newEdge);
    },
    [nodes, edges, addEdge]
  );

  // ── Selection ─────────────────────────────────────────────────────────────

  const onNodeClick = useCallback(
    (_: React.MouseEvent, node: import("@xyflow/react").Node) => {
      selectNode(node.id);
    },
    [selectNode]
  );

  const onEdgeClick = useCallback(
    (_: React.MouseEvent, edge: import("@xyflow/react").Edge) => {
      selectEdge(edge.id);
    },
    [selectEdge]
  );

  const onPaneClick = useCallback(() => {
    selectNode(null);
    selectEdge(null);
  }, [selectNode, selectEdge]);

  return (
    <div className="flex h-full w-full overflow-hidden" style={{ background: "#0d0d0d" }}>
      {/* SVG markers for edge arrows */}
      <EdgeMarkerDefs />

      {/* Left sidebar: palette or config panel */}
      <div className="flex-shrink-0 h-full" style={{ zIndex: 5 }}>
        {showConfigPanel ? <NodeConfigPanel /> : <Palette hasTrigger={hasTrigger} />}
      </div>

      {/* Canvas */}
      <div className="flex-1 h-full relative" ref={reactFlowWrapper}>
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onDrop={onDrop}
          onDragOver={onDragOver}
          onNodeClick={onNodeClick}
          onEdgeClick={onEdgeClick}
          onPaneClick={onPaneClick}
          nodeTypes={nodeTypes}
          edgeTypes={edgeTypes}
          selectionMode={SelectionMode.Partial}
          fitView
          fitViewOptions={{ padding: 0.25, maxZoom: 1.2 }}
          minZoom={0.2}
          maxZoom={2}
          deleteKeyCode={["Delete", "Backspace"]}
          defaultEdgeOptions={{
            type: "workflow",
            data: { condition: null },
          }}
          proOptions={{ hideAttribution: true }}
        >
          <Background
            variant={BackgroundVariant.Dots}
            gap={28}
            size={1.5}
            color="rgba(255,255,255,0.2)"
          />
          <Controls
            position="bottom-right"
            showInteractive={false}
            style={{ bottom: 80 }}
          />
          <MiniMap
            position="bottom-right"
            style={{ bottom: 80, right: 60 }}
            pannable
            zoomable
          />

          {/* Empty state */}
          {nodes.length === 0 && (
            <div
              style={{
                position: "absolute",
                top: "50%",
                left: "50%",
                transform: "translate(-50%, -50%)",
                pointerEvents: "none",
                textAlign: "center",
              }}
            >
              <p className="text-zinc-600 text-sm font-medium">
                Drag a trigger from the sidebar to start
              </p>
              <p className="text-zinc-700 text-xs mt-1">
                Then add actions and connect them
              </p>
            </div>
          )}

          {/* Floating toolbar */}
          <CanvasToolbar
            onSave={onSave}
            onTrigger={onTrigger}
            isSaving={isSaving}
          />
        </ReactFlow>
      </div>
    </div>
  );
}

// ─── Public export wraps with ReactFlowProvider ────────────────────────────────

import { ReactFlowProvider } from "@xyflow/react";

export function WorkflowCanvas(props: WorkflowCanvasProps) {
  return (
    <ReactFlowProvider>
      <CanvasInner {...props} />
    </ReactFlowProvider>
  );
}
