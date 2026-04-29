"use client";

import { useReactFlow } from "@xyflow/react";
import { useWorkflow } from "@/components/workflow/workflow-provider";
import { autoLayout } from "@/lib/dag-utils";
import { toast } from "sonner";

interface CanvasToolbarProps {
  onSave: () => Promise<void>;
  onTrigger?: () => void;
  isSaving: boolean;
}

export function CanvasToolbar({ onSave, onTrigger, isSaving }: CanvasToolbarProps) {
  const { fitView } = useReactFlow();
  const {
    nodes,
    edges,
    relayMeta,
    setMeta,
    isDirty,
    undo,
    redo,
    canUndo,
    canRedo,
    loadGraph,
  } = useWorkflow();

  const handleAutoLayout = () => {
    const laidOut = autoLayout(nodes, edges);
    loadGraph(laidOut, edges, relayMeta);
    setTimeout(() => fitView({ padding: 0.2, duration: 400 }), 50);
    toast.success("Layout applied");
  };

  const handleFitView = () => {
    fitView({ padding: 0.2, duration: 400 });
  };

  return (
    <div className="workflow-toolbar">
      {/* Relay name */}
      <input
        type="text"
        value={relayMeta.name}
        onChange={(e) => setMeta({ name: e.target.value })}
        className="h-8 min-w-[160px] max-w-[200px] rounded-lg border border-white/10 bg-white/5 px-3 text-sm font-medium text-white placeholder:text-zinc-500 focus:border-orange-500/40 focus:outline-none"
        placeholder="Relay name"
      />

      <div className="h-4 w-px bg-white/10 mx-1" />

      {/* Undo / Redo */}
      <ToolbarButton
        onClick={undo}
        disabled={!canUndo}
        title="Undo (Ctrl+Z)"
        label={
          <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M9 15 3 9m0 0 6-6M3 9h12a6 6 0 0 1 0 12h-3" />
          </svg>
        }
      />
      <ToolbarButton
        onClick={redo}
        disabled={!canRedo}
        title="Redo (Ctrl+Shift+Z)"
        label={
          <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="m15 15 6-6m0 0-6-6m6 6H9a6 6 0 0 0 0 12h3" />
          </svg>
        }
      />

      <div className="h-4 w-px bg-white/10 mx-1" />

      {/* Layout / fit */}
      <ToolbarButton onClick={handleAutoLayout} title="Auto-layout" label={
        <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 6A2.25 2.25 0 0 1 6 3.75h2.25A2.25 2.25 0 0 1 10.5 6v2.25a2.25 2.25 0 0 1-2.25 2.25H6a2.25 2.25 0 0 1-2.25-2.25V6ZM3.75 15.75A2.25 2.25 0 0 1 6 13.5h2.25a2.25 2.25 0 0 1 2.25 2.25V18a2.25 2.25 0 0 1-2.25 2.25H6A2.25 2.25 0 0 1 3.75 18v-2.25ZM13.5 6a2.25 2.25 0 0 1 2.25-2.25H18A2.25 2.25 0 0 1 20.25 6v2.25A2.25 2.25 0 0 1 18 10.5h-2.25a2.25 2.25 0 0 1-2.25-2.25V6ZM13.5 15.75a2.25 2.25 0 0 1 2.25-2.25H18a2.25 2.25 0 0 1 2.25 2.25V18A2.25 2.25 0 0 1 18 20.25h-2.25A2.25 2.25 0 0 1 13.5 18v-2.25Z" />
        </svg>
      } />
      <ToolbarButton onClick={handleFitView} title="Fit view (F)" label={
        <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 3.75v4.5m0-4.5h4.5m-4.5 0L9 9M3.75 20.25v-4.5m0 4.5h4.5m-4.5 0L9 15M20.25 3.75h-4.5m4.5 0v4.5m0-4.5L15 9m5.25 11.25h-4.5m4.5 0v-4.5m0 4.5L15 15" />
        </svg>
      } />

      <div className="h-4 w-px bg-white/10 mx-1" />

      {/* Manual trigger */}
      {onTrigger && (
        <button
          type="button"
          onClick={onTrigger}
          className="flex items-center gap-1.5 rounded-lg border border-purple-500/30 bg-purple-500/10 px-3 py-1.5 text-xs font-medium text-purple-300 transition-colors hover:bg-purple-500/20 hover:text-purple-200"
        >
          <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 0 1 0 1.972l-11.54 6.347a1.125 1.125 0 0 1-1.667-.986V5.653Z" />
          </svg>
          Run
        </button>
      )}

      {/* Save */}
      <button
        type="button"
        onClick={onSave}
        disabled={isSaving}
        className={`flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-semibold transition-colors disabled:opacity-50 ${
          isDirty
            ? "bg-orange-500 text-white hover:bg-orange-600"
            : "bg-white/8 text-zinc-300 border border-white/10 hover:bg-white/12"
        }`}
      >
        {isSaving ? (
          <>
            <span className="h-3 w-3 animate-spin rounded-full border-2 border-white/30 border-t-white" />
            Saving…
          </>
        ) : (
          <>
            <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M17.593 3.322c1.1.128 1.907 1.077 1.907 2.185V21L12 17.25 4.5 21V5.507c0-1.108.806-2.057 1.907-2.185a48.507 48.507 0 0 1 11.186 0Z" />
            </svg>
            {isDirty ? "Save*" : "Saved"}
          </>
        )}
      </button>
    </div>
  );
}

function ToolbarButton({
  onClick,
  label,
  title,
  disabled = false,
}: {
  onClick: () => void;
  label: React.ReactNode;
  title: string;
  disabled?: boolean;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      title={title}
      className="flex h-7 w-7 items-center justify-center rounded-lg text-zinc-400 transition-colors hover:bg-white/8 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed"
    >
      {label}
    </button>
  );
}
