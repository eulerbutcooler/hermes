"use client";

import {
  createContext,
  useCallback,
  useContext,
  useRef,
  useState,
} from "react";
import {
  applyNodeChanges,
  applyEdgeChanges,
  type NodeChange,
  type EdgeChange,
} from "@xyflow/react";
import type {
  RelayMeta,
  WorkflowEdge,
  WorkflowNode,
} from "@/types/workflow";


// ─── Context Types ─────────────────────────────────────────────────────────────

interface WorkflowContextValue {
  nodes: WorkflowNode[];
  edges: WorkflowEdge[];
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  relayMeta: RelayMeta;
  isDirty: boolean;

  // Node / edge state setters for React Flow
  onNodesChange: (changes: NodeChange[]) => void;
  onEdgesChange: (changes: EdgeChange[]) => void;

  // Actions
  addNode: (node: WorkflowNode) => void;
  removeNode: (id: string) => void;
  updateNodeData: (id: string, data: Record<string, unknown>) => void;
  addEdge: (edge: WorkflowEdge) => void;
  removeEdge: (id: string) => void;
  updateEdgeData: (id: string, data: Record<string, unknown>) => void;
  selectNode: (id: string | null) => void;
  selectEdge: (id: string | null) => void;
  setMeta: (meta: Partial<RelayMeta>) => void;
  loadGraph: (nodes: WorkflowNode[], edges: WorkflowEdge[], meta?: Partial<RelayMeta>) => void;
  undo: () => void;
  redo: () => void;
  canUndo: boolean;
  canRedo: boolean;
  markSaved: () => void;
}

const WorkflowContext = createContext<WorkflowContextValue | null>(null);

// ─── History ──────────────────────────────────────────────────────────────────

interface HistoryEntry {
  nodes: WorkflowNode[];
  edges: WorkflowEdge[];
}

// ─── Provider ─────────────────────────────────────────────────────────────────

export function WorkflowProvider({
  children,
  initialMeta,
}: {
  children: React.ReactNode;
  initialMeta?: Partial<RelayMeta>;
}) {
  const [nodes, setNodes] = useState<WorkflowNode[]>([]);
  const [edges, setEdges] = useState<WorkflowEdge[]>([]);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [selectedEdgeId, setSelectedEdgeId] = useState<string | null>(null);
  const [isDirty, setIsDirty] = useState(false);
  const [relayMeta, setRelayMeta] = useState<RelayMeta>({
    name: initialMeta?.name ?? "Untitled Relay",
    description: initialMeta?.description ?? "",
    triggerType: initialMeta?.triggerType ?? "webhook",
    triggerConfig: initialMeta?.triggerConfig ?? {},
  });

  // Undo/redo stacks
  const past = useRef<HistoryEntry[]>([]);
  const future = useRef<HistoryEntry[]>([]);

  const pushHistory = useCallback(
    (prevNodes: WorkflowNode[], prevEdges: WorkflowEdge[]) => {
      past.current.push({ nodes: prevNodes, edges: prevEdges });
      if (past.current.length > 50) past.current.shift(); // cap history
      future.current = [];
      setIsDirty(true);
    },
    []
  );

  // ── React Flow native change handlers ──

  const onNodesChange = useCallback(
    (changes: NodeChange[]) => {
      setNodes((nds) => applyNodeChanges(changes, nds) as WorkflowNode[]);
    },
    []
  );

  const onEdgesChange = useCallback(
    (changes: EdgeChange[]) => {
      setEdges((eds) => applyEdgeChanges(changes, eds) as WorkflowEdge[]);
    },
    []
  );

  // ── CRUD ──

  const addNode = useCallback(
    (node: WorkflowNode) => {
      setNodes((nds) => {
        pushHistory(nds, edges);
        return [...nds, node];
      });
    },
    [edges, pushHistory]
  );

  const removeNode = useCallback(
    (id: string) => {
      setNodes((nds) => {
        pushHistory(nds, edges);
        return nds.filter((n) => n.id !== id);
      });
      setEdges((eds) => eds.filter((e) => e.source !== id && e.target !== id));
      setSelectedNodeId((cur) => (cur === id ? null : cur));
    },
    [edges, pushHistory]
  );

  const updateNodeData = useCallback(
    (id: string, data: Record<string, unknown>) => {
      setNodes((nds) => {
        return nds.map((n) =>
          n.id === id ? ({ ...n, data: { ...n.data, ...data } } as WorkflowNode) : n
        );
      });
      setIsDirty(true);
    },
    []
  );

  const addEdge = useCallback(
    (edge: WorkflowEdge) => {
      setEdges((eds) => {
        pushHistory(nodes, eds);
        return [...eds, edge];
      });
    },
    [nodes, pushHistory]
  );

  const removeEdge = useCallback(
    (id: string) => {
      setEdges((eds) => {
        pushHistory(nodes, eds);
        return eds.filter((e) => e.id !== id);
      });
    },
    [nodes, pushHistory]
  );

  const updateEdgeData = useCallback(
    (id: string, data: Record<string, unknown>) => {
      setEdges((eds) =>
        eds.map((e) =>
          e.id === id ? { ...e, data: { ...e.data, ...data } } : e
        )
      );
      setIsDirty(true);
    },
    []
  );

  const selectNode = useCallback((id: string | null) => {
    setSelectedNodeId(id);
    setSelectedEdgeId(null);
  }, []);

  const selectEdge = useCallback((id: string | null) => {
    setSelectedEdgeId(id);
    setSelectedNodeId(null);
  }, []);

  const setMeta = useCallback((meta: Partial<RelayMeta>) => {
    setRelayMeta((prev) => ({ ...prev, ...meta }));
    setIsDirty(true);
  }, []);

  const loadGraph = useCallback(
    (
      newNodes: WorkflowNode[],
      newEdges: WorkflowEdge[],
      meta?: Partial<RelayMeta>
    ) => {
      past.current = [];
      future.current = [];
      setNodes(newNodes);
      setEdges(newEdges);
      if (meta) setRelayMeta((prev) => ({ ...prev, ...meta }));
      setSelectedNodeId(null);
      setSelectedEdgeId(null);
      setIsDirty(false);
    },
    []
  );

  const undo = useCallback(() => {
    const prev = past.current.pop();
    if (!prev) return;
    future.current.push({ nodes, edges });
    setNodes(prev.nodes);
    setEdges(prev.edges);
    setIsDirty(true);
  }, [nodes, edges]);

  const redo = useCallback(() => {
    const next = future.current.pop();
    if (!next) return;
    past.current.push({ nodes, edges });
    setNodes(next.nodes);
    setEdges(next.edges);
    setIsDirty(true);
  }, [nodes, edges]);

  const markSaved = useCallback(() => setIsDirty(false), []);

  return (
    <WorkflowContext.Provider
      value={{
        nodes,
        edges,
        selectedNodeId,
        selectedEdgeId,
        relayMeta,
        isDirty,
        onNodesChange,
        onEdgesChange,
        addNode,
        removeNode,
        updateNodeData,
        addEdge,
        removeEdge,
        updateEdgeData,
        selectNode,
        selectEdge,
        setMeta,
        loadGraph,
        undo,
        redo,
        canUndo: past.current.length > 0,
        canRedo: future.current.length > 0,
        markSaved,
      }}
    >
      {children}
    </WorkflowContext.Provider>
  );
}

export function useWorkflow(): WorkflowContextValue {
  const ctx = useContext(WorkflowContext);
  if (!ctx) throw new Error("useWorkflow must be used inside <WorkflowProvider>");
  return ctx;
}
