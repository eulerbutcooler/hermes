"use client";

import { useParams, useRouter } from "next/navigation";
import { useEffect, useState, useCallback } from "react";
import { toast } from "sonner";
import Link from "next/link";

import { WorkflowCanvas } from "@/components/workflow/workflow-canvas";
import { WorkflowProvider, useWorkflow } from "@/components/workflow/workflow-provider";
import { getRelayDAG, createRelayDAG, updateRelayDAG } from "@/lib/api-dag";
import { graphToDAGPayload, dagPayloadToGraph } from "@/lib/workflow-serializer";
import { getRelay, triggerRelay } from "@/lib/api";
import type { RelayMeta } from "@/types/workflow";

// ─── Inner page (has access to WorkflowProvider context) ──────────────────────

function BuilderPageInner({ id }: { id: string }) {
  const router = useRouter();
  const {
    nodes,
    edges,
    relayMeta,
    loadGraph,
    setMeta,
    markSaved,
  } = useWorkflow();

  const [isSaving, setIsSaving] = useState(false);
  const [isLoading, setIsLoading] = useState(id !== "new");
  const [relayId, setRelayId] = useState<string | null>(id === "new" ? null : id);

  // ── Load existing relay into canvas ───────────────────────────────────────

  useEffect(() => {
    if (id === "new") return;

    async function load() {
      setIsLoading(true);
      try {
        // Try DAG endpoint first (when backend merges), fall back to regular relay
        let relay;
        try {
          relay = await getRelayDAG(id);
        } catch {
          relay = await getRelay(id);
        }

        const { nodes: loadedNodes, edges: loadedEdges } = dagPayloadToGraph(relay as Parameters<typeof dagPayloadToGraph>[0]);
        const meta: Partial<RelayMeta> = {
          name: relay.name,
          description: relay.description ?? "",
          triggerType: relay.trigger_type as RelayMeta["triggerType"],
          triggerConfig: relay.trigger_config ?? {},
        };
        loadGraph(loadedNodes, loadedEdges, meta);
      } catch (err) {
        toast.error(err instanceof Error ? err.message : "Failed to load relay");
      } finally {
        setIsLoading(false);
      }
    }

    load();
  }, [id, loadGraph]);

  // ── Save ──────────────────────────────────────────────────────────────────

  const handleSave = useCallback(async () => {
    if (!relayMeta.name.trim()) {
      toast.error("Relay name is required");
      return;
    }

    const hasTrigger = nodes.some((n) => n.type === "trigger");
    if (!hasTrigger) {
      toast.error("Add a trigger node before saving");
      return;
    }

    const hasActions = nodes.some((n) => n.type === "action" || n.type === "condition");
    if (!hasActions) {
      toast.error("Add at least one action node");
      return;
    }

    setIsSaving(true);
    try {
      const payload = graphToDAGPayload(nodes, edges, relayMeta);

      let saved;
      if (relayId) {
        saved = await updateRelayDAG(relayId, payload);
      } else {
        saved = await createRelayDAG(payload);
        setRelayId(saved.id);
        // Update URL without full navigation
        window.history.replaceState(
          {},
          "",
          `/dashboard/relays/builder/${saved.id}`
        );
      }

      markSaved();
      toast.success("Relay saved!");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to save relay");
    } finally {
      setIsSaving(false);
    }
  }, [nodes, edges, relayMeta, relayId, markSaved]);

  // ── Manual trigger ────────────────────────────────────────────────────────

  const handleTrigger = useCallback(async () => {
    if (!relayId) {
      toast.error("Save the relay first");
      return;
    }
    try {
      await triggerRelay(relayId, {});
      toast.success("Relay triggered!");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to trigger");
    }
  }, [relayId]);

  // ── Loading state ─────────────────────────────────────────────────────────

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center bg-[#0d0d0d]">
        <div className="flex flex-col items-center gap-3">
          <span className="h-8 w-8 animate-spin rounded-full border-2 border-white/10 border-t-orange-500" />
          <p className="text-sm text-zinc-500">Loading relay…</p>
        </div>
      </div>
    );
  }

  // ── Check if trigger is manual (for run button) ───────────────────────────

  const isManuatTrigger = relayMeta.triggerType === "manual";

  return (
    <div className="flex h-full flex-col overflow-hidden">
      {/* Top bar */}
      <header className="flex h-11 flex-shrink-0 items-center justify-between border-b border-white/10 bg-[#141414] px-4">
        <div className="flex items-center gap-3">
          {/* Back */}
          <Link
            href="/dashboard/relays"
            className="flex items-center gap-1.5 text-xs text-zinc-500 hover:text-white transition-colors"
          >
            <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 19.5 8.25 12l7.5-7.5" />
            </svg>
            Relays
          </Link>
          <span className="text-zinc-700">/</span>
          <span className="text-xs font-medium text-zinc-300">
            {relayMeta.name || "Untitled"}
          </span>
        </div>

        <div className="flex items-center gap-2">
          <span className="text-[10px] text-zinc-600 font-mono">
            {nodes.length} node{nodes.length !== 1 ? "s" : ""} · {edges.length} edge{edges.length !== 1 ? "s" : ""}
          </span>
          {relayId && (
            <Link
              href={`/dashboard/relays/${relayId}`}
              className="rounded-lg border border-white/10 px-2.5 py-1 text-xs text-zinc-400 hover:border-white/20 hover:text-white transition-colors"
            >
              Detail view
            </Link>
          )}
        </div>
      </header>

      {/* Canvas */}
      <div className="flex-1 overflow-hidden">
        <WorkflowCanvas
          onSave={handleSave}
          onTrigger={isManuatTrigger && relayId ? handleTrigger : undefined}
          isSaving={isSaving}
        />
      </div>
    </div>
  );
}

// ─── Page (wraps with WorkflowProvider) ───────────────────────────────────────

export default function BuilderPage() {
  const { id } = useParams<{ id: string }>();

  return (
    <WorkflowProvider>
      <BuilderPageInner id={id} />
    </WorkflowProvider>
  );
}
