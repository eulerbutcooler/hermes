"use client";

import {
  useDeleteExecution,
  useDeleteRelay,
  useExecutionSteps,
  useExecutions,
  useRelay,
  useUpdateRelay,
  useUpdateRelayActions,
  useTriggerRelay,
} from "@/lib/queries";
import { ActionCard, SegmentedToggle } from "@/components/action-config-fields";
import {
  ACTION_LABELS,
  type Execution,
  type CreateRelayActionInput,
  type ExecutionStep,
  type HTTPRequestStepOutput,
  type TriggerType,
} from "@/types/relay";
import { useParams, useRouter } from "next/navigation";
import { useState } from "react";
import { toast } from "sonner";
import Link from "next/link";


function StatusBadge({ active }: { active: boolean }) {
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ${
        active ? "bg-green-500/10 text-green-400" : "bg-zinc-800 text-zinc-500"
      }`}
    >
      <span
        className={`h-1.5 w-1.5 rounded-full ${active ? "bg-green-400" : "bg-zinc-500"}`}
      />
      {active ? "Active" : "Inactive"}
    </span>
  );
}

function RunStatusBadge({ status }: { status: string }) {
  const map: Record<string, string> = {
    success: "bg-green-500/10 text-green-400",
    failed: "bg-red-500/10 text-red-400",
    running: "bg-yellow-500/10 text-yellow-400",
  };
  return (
    <span
      className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${map[status] ?? "bg-zinc-800 text-zinc-400"}`}
    >
      {status}
    </span>
  );
}

function TriggerBadge({ type }: { type: TriggerType }) {
  const map: Record<TriggerType, string> = {
    webhook: "bg-blue-500/10 text-blue-400",
    manual: "bg-purple-500/10 text-purple-400",
    cron: "bg-amber-500/10 text-amber-400",
  };
  const label: Record<TriggerType, string> = {
    webhook: "Webhook",
    manual: "Manual",
    cron: "Cron",
  };
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ${map[type]}`}
    >
      {label[type]}
    </span>
  );
}

function JsonBlock({ value }: { value: unknown }) {
  if (value == null) return <p className="text-xs text-zinc-500">No data</p>;

  return (
    <pre className="overflow-x-auto rounded-lg bg-black/30 p-3 text-xs text-zinc-300">
      {JSON.stringify(value, null, 2)}
    </pre>
  );
}

function HTTPRequestOutputView({ output }: { output: HTTPRequestStepOutput }) {
  return (
    <div className="space-y-3">
      <div className="grid gap-3 sm:grid-cols-3">
        <div className="rounded-lg bg-black/30 p-3">
          <p className="text-[11px] uppercase tracking-wide text-zinc-500">
            Status
          </p>
          <p className="mt-1 text-sm text-white">{output.status_code}</p>
        </div>
        <div className="rounded-lg bg-black/30 p-3">
          <p className="text-[11px] uppercase tracking-wide text-zinc-500">
            Content-Type
          </p>
          <p className="mt-1 break-all text-sm text-white">
            {output.content_type ?? "—"}
          </p>
        </div>
        <div className="rounded-lg bg-black/30 p-3">
          <p className="text-[11px] uppercase tracking-wide text-zinc-500">
            Duration
          </p>
          <p className="mt-1 text-sm text-white">
            {output.duration_ms != null ? `${output.duration_ms} ms` : "—"}
          </p>
        </div>
      </div>

      {output.headers && (
        <div>
          <p className="mb-2 text-xs font-medium text-zinc-400">Headers</p>
          <JsonBlock value={output.headers} />
        </div>
      )}

      {output.body_json !== undefined && (
        <div>
          <p className="mb-2 text-xs font-medium text-zinc-400">
            Response JSON
          </p>
          <JsonBlock value={output.body_json} />
        </div>
      )}

      {output.body_text && (
        <div>
          <p className="mb-2 text-xs font-medium text-zinc-400">
            Response Text
          </p>
          <pre className="overflow-x-auto rounded-lg bg-black/30 p-3 text-xs text-zinc-300">
            {output.body_text}
          </pre>
        </div>
      )}
    </div>
  );
}

function isHTTPRequestStepOutput(
  value: unknown,
): value is HTTPRequestStepOutput {
  if (!value || typeof value !== "object") return false;

  const candidate = value as Record<string, unknown>;
  return typeof candidate.status_code === "number";
}

function StepCard({ step, index }: { step: ExecutionStep; index: number }) {
  const isHTTPRequest = step.action_type === "http_request";

  return (
    <div className="rounded-xl border border-white/10 bg-[#161616] p-4">
      <div className="mb-4 flex items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <span className="flex h-7 w-7 items-center justify-center rounded-full bg-orange-500/10 text-xs font-bold text-orange-400">
            {index + 1}
          </span>
          <div>
            <p className="text-sm font-medium text-white">
              {ACTION_LABELS[step.action_type as keyof typeof ACTION_LABELS] ??
                step.action_type}
            </p>
            <p className="text-xs text-zinc-500">
              {new Date(step.started_at).toLocaleString()}
            </p>
          </div>
        </div>
        <RunStatusBadge status={step.status} />
      </div>

      <div className="space-y-4">
        <div>
          <p className="mb-2 text-xs font-medium text-zinc-400">Input</p>
          <JsonBlock value={step.input} />
        </div>

        <div>
          <p className="mb-2 text-xs font-medium text-zinc-400">Output</p>
          {isHTTPRequest && isHTTPRequestStepOutput(step.output) ? (
            <HTTPRequestOutputView output={step.output} />
          ) : (
            <JsonBlock value={step.output} />
          )}
        </div>

        {step.error_message && (
          <div>
            <p className="mb-2 text-xs font-medium text-red-400">Error</p>
            <pre className="overflow-x-auto rounded-lg bg-red-500/10 p-3 text-xs text-red-300">
              {step.error_message}
            </pre>
          </div>
        )}
      </div>
    </div>
  );
}
function ExecutionRow({
  execution,
  expanded,
  onToggle,
  onDelete,
}: {
  execution: Execution;
  expanded: boolean;
  onToggle: () => void;
  onDelete: (id: string) => void;
}) {
  const { data: steps, isLoading } = useExecutionSteps(execution.id, expanded);

  return (
    <div className="relative overflow-hidden rounded-xl border border-white/10 bg-[#1a1a1a]">
      {/* Delete X button */}
      <button
        type="button"
        onClick={(e) => {
          e.stopPropagation();
          onDelete(execution.id);
        }}
        className="absolute right-2 top-2 z-10 flex h-5 w-5 items-center justify-center rounded-full text-zinc-600 transition-colors hover:bg-white/10 hover:text-zinc-300"
        title="Delete execution"
      >
        <svg
          className="h-3 w-3"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2.5}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M6 18 18 6M6 6l12 12"
          />
        </svg>
      </button>

      <button
        type="button"
        onClick={onToggle}
        className="grid w-full grid-cols-[1fr_120px_180px] gap-4 px-4 py-3 text-left transition-colors hover:bg-white/5"
      >
        <span className="min-w-0">
          <span className="block truncate font-mono text-xs text-zinc-300">
            {execution.id}
          </span>
          {execution.event_id && (
            <span className="block truncate font-mono text-xs text-zinc-600">
              Event: {execution.event_id}
            </span>
          )}
        </span>
        <span className="flex items-center">
          <RunStatusBadge status={execution.status} />
        </span>
        <span className="flex items-center justify-end pr-6 text-xs text-zinc-500">
          {new Date(execution.started_at).toLocaleString()}
        </span>
      </button>

      {expanded && (
        <div className="border-t border-white/10 px-4 py-4">
          <div>
            <p className="mb-3 text-xs font-medium text-zinc-400">Steps</p>
            {isLoading ? (
              <div className="space-y-3">
                {[...Array(2)].map((_, i) => (
                  <div
                    key={i}
                    className="h-28 animate-pulse rounded-xl bg-white/5"
                  />
                ))}
              </div>
            ) : !steps || steps.length === 0 ? (
              <div className="rounded-lg bg-black/20 p-4 text-sm text-zinc-500">
                No steps recorded.
              </div>
            ) : (
              <div className="space-y-3">
                {steps.map((step, index) => (
                  <StepCard key={step.id} step={step} index={index} />
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

export default function RelayDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const { data: relay, isLoading, isError } = useRelay(id);
  const { data: executions } = useExecutions(id);
  const updateMutation = useUpdateRelay(id);
  const updateActionsMutation = useUpdateRelayActions(id);
  const deleteMutation = useDeleteRelay();
  const triggerMutation = useTriggerRelay(id);
  const deleteExecutionMutation = useDeleteExecution(id);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [copied, setCopied] = useState(false);
  const [triggerPayload, setTriggerPayload] = useState("{}");
  const [showTriggerPanel, setShowTriggerPanel] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [editTab, setEditTab] = useState<"general" | "actions">("general");
  const [editName, setEditName] = useState("");
  const [editDescription, setEditDescription] = useState("");
  const [editTriggerType, setEditTriggerType] =
    useState<TriggerType>("webhook");
  const [editCronSchedule, setEditCronSchedule] = useState("");
  const [editActions, setEditActions] = useState<CreateRelayActionInput[]>([]);
  const [expandedExecutionId, setExpandedExecutionId] = useState<string | null>(
    null,
  );

  const webhookUrl = relay?.webhook_url ?? "";

  const copyWebhook = async () => {
    await navigator.clipboard.writeText(webhookUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const toggleActive = async () => {
    if (!relay) return;
    try {
      await updateMutation.mutateAsync({ is_active: !relay.is_active });
      toast.success(relay.is_active ? "Relay deactivated" : "Relay activated");
    } catch {
      toast.error("Failed to update relay");
    }
  };

  const handleDelete = async () => {
    try {
      await deleteMutation.mutateAsync(id);
      toast.success("Relay deleted");
      router.push("/dashboard/relays");
    } catch {
      toast.error("Failed to delete relay");
    }
  };

  const handleDeleteExecution = (executionId: string) => {
    deleteExecutionMutation.mutate(executionId, {
      onError: (err) =>
        toast.error(
          err instanceof Error ? err.message : "Failed to delete execution",
        ),
    });
  };

  const handleSaveActions = async () => {
    try {
      await updateActionsMutation.mutateAsync(
        editActions.map((a, i) => ({ ...a, order_index: i })),
      );
      toast.success("Actions updated");
      setShowEditModal(false);
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : "Failed to update actions",
      );
    }
  };
  const handleEdit = async (e: React.SyntheticEvent) => {
    e.preventDefault();
    if (!editName.trim()) return toast.error("Name is required");
    if (editTriggerType === "cron" && !editCronSchedule.trim())
      return toast.error("Cron schedule is required");
    try {
      await updateMutation.mutateAsync({
        name: editName.trim(),
        description: editDescription.trim() || undefined,
        trigger_type: editTriggerType,
        trigger_config:
          editTriggerType === "cron"
            ? { schedule: editCronSchedule.trim() }
            : {},
      });
      toast.success("Relay updated");
      setShowEditModal(false);
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : "Failed to update relay",
      );
    }
  };

  const handleManualTrigger = async () => {
    let payload: Record<string, unknown> = {};
    try {
      payload = JSON.parse(triggerPayload);
    } catch {
      toast.error("Payload must be valid JSON");
      return;
    }
    try {
      await triggerMutation.mutateAsync(payload);
      toast.success("Relay triggered!");
      setShowTriggerPanel(false);
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : "Failed to trigger relay",
      );
    }
  };

  if (isLoading) {
    return (
      <div className="mx-auto w-full max-w-5xl space-y-4 p-8">
        {[...Array(4)].map((_, i) => (
          <div key={i} className="h-12 animate-pulse rounded-lg bg-white/5" />
        ))}
      </div>
    );
  }

  if (isError || !relay) {
    return (
      <div className="mx-auto w-full max-w-5xl p-8">
        <p className="text-sm text-red-400">Failed to load relay.</p>
      </div>
    );
  }

  return (
    <div className="mx-auto w-full max-w-5xl space-y-8 p-8">
      <div className="flex items-start justify-between gap-4">
        <div>
          <div className="mb-1 flex items-center gap-3">
            <h1 className="text-xl font-bold text-white">{relay.name}</h1>
            <StatusBadge active={relay.is_active} />
            <TriggerBadge type={relay.trigger_type} />
          </div>
          {relay.description && (
            <p className="text-sm text-zinc-500">{relay.description}</p>
          )}
        </div>
        <div className="flex shrink-0 items-center gap-2">
          {relay.trigger_type === "manual" && (
            <button
              type="button"
              onClick={() => setShowTriggerPanel(true)}
              className="rounded-lg bg-purple-500/10 border border-purple-500/30 px-3.5 py-1.5 text-xs font-medium text-purple-400 transition-colors hover:bg-purple-500/20 hover:text-purple-300"
            >
              Trigger
            </button>
          )}
          <Link
            href={`/dashboard/relays/builder/${relay.id}`}
            className="rounded-lg border border-orange-500/25 bg-orange-500/8 px-3.5 py-1.5 text-xs font-medium text-orange-400 transition-colors hover:bg-orange-500/15 hover:text-orange-300"
          >
            Open in Builder
          </Link>
          <button
            type="button"
            onClick={() => {
              setEditName(relay.name);
              setEditDescription(relay.description ?? "");
              setEditTriggerType(relay.trigger_type);
              setEditCronSchedule(
                (relay.trigger_config?.schedule as string) ?? "0 9 * * *",
              );
              setEditActions(
                relay.actions.map((a) => ({
                  action_type: a.action_type,
                  config: { ...a.config },
                  node_id: a.node_id,
                })),
              );
              setEditTab("general");
              setShowEditModal(true);
            }}
            className="rounded-lg border border-white/10 px-3.5 py-1.5 text-xs font-medium text-zinc-300 transition-colors hover:border-white/20 hover:text-white"
          >
            Edit
          </button>
          <button
            type="button"
            onClick={toggleActive}
            disabled={updateMutation.isPending}
            className="rounded-lg border border-white/10 px-3.5 py-1.5 text-xs font-medium text-zinc-300 transition-colors hover:border-white/20 hover:text-white disabled:opacity-50"
          >
            {relay.is_active ? "Deactivate" : "Activate"}
          </button>
          <button
            type="button"
            onClick={() => setShowDeleteConfirm(true)}
            className="rounded-lg border border-red-500/20 px-3.5 py-1.5 text-xs font-medium text-red-400 transition-colors hover:border-red-500/40 hover:text-red-300"
          >
            Delete
          </button>
        </div>
      </div>

      {/* Webhook URL — only for webhook trigger type */}
      {relay.trigger_type === "webhook" && (
        <div className="rounded-xl border border-white/10 bg-[#1a1a1a] p-4">
          <p className="mb-2 text-xs font-medium text-zinc-400">Webhook URL</p>
          <div className="flex items-center gap-2">
            <code className="flex-1 truncate rounded-lg bg-black/30 px-3 py-2 font-mono text-xs text-zinc-300">
              {webhookUrl}
            </code>
            <button
              type="button"
              onClick={copyWebhook}
              className="shrink-0 rounded-lg bg-white/5 px-3 py-2 text-xs text-zinc-400 transition-colors hover:bg-white/10 hover:text-white"
            >
              {copied ? "Copied!" : "Copy"}
            </button>
          </div>
        </div>
      )}

      {/* Cron info panel */}
      {relay.trigger_type === "cron" && (
        <div className="rounded-xl border border-amber-500/20 bg-amber-500/5 p-4">
          <p className="mb-1 text-xs font-medium text-amber-400">
            Cron Schedule
          </p>
          <code className="font-mono text-sm text-amber-300">
            {(relay.trigger_config?.schedule as string) ?? "—"}
          </code>
          <p className="mt-1.5 text-xs text-zinc-500">
            Runs automatically on this schedule. The worker checks every minute.
          </p>
        </div>
      )}

      {relay.actions && relay.actions.length > 0 && (
        <div>
          <h2 className="mb-3 text-sm font-medium text-zinc-300">
            Actions ({relay.actions.length})
          </h2>
          <div className="space-y-2">
            {relay.actions.map((action, i) => (
              <div
                key={action.id}
                className="flex items-center gap-3 rounded-xl border border-white/10 bg-[#1a1a1a] px-4 py-3"
              >
                <span className="flex h-6 w-6 items-center justify-center rounded-full bg-orange-500/10 text-xs font-bold text-orange-400">
                  {i + 1}
                </span>
                <span className="text-sm text-white">
                  {ACTION_LABELS[
                    action.action_type as keyof typeof ACTION_LABELS
                  ] ?? action.action_type}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      <div>
        <h2 className="mb-3 text-sm font-medium text-zinc-300">Executions</h2>
        {!executions || executions.length === 0 ? (
          <div className="rounded-xl border border-white/10 bg-[#1a1a1a] px-4 py-8 text-center">
            <p className="text-sm text-zinc-500">
              No executions yet — send a POST request to your webhook URL to see
              runs here.
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            {executions.map((execution) => (
              <ExecutionRow
                key={execution.id}
                execution={execution}
                expanded={expandedExecutionId === execution.id}
                onToggle={() =>
                  setExpandedExecutionId((current) =>
                    current === execution.id ? null : execution.id,
                  )
                }
                onDelete={handleDeleteExecution}
              />
            ))}
          </div>
        )}
      </div>
      {showEditModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4">
          <div className="w-full max-w-lg rounded-2xl border border-white/10 bg-[#1a1a1a]">
            {/* Modal header with tabs */}
            <div className="flex items-center justify-between border-b border-white/10 px-5 pt-4">
              <div className="flex gap-1">
                {(["general", "actions"] as const).map((tab) => (
                  <button
                    key={tab}
                    type="button"
                    onClick={() => setEditTab(tab)}
                    className={`rounded-t-md px-4 py-2 text-xs font-medium capitalize transition-colors ${
                      editTab === tab
                        ? "border-b-2 border-orange-500 text-white"
                        : "text-zinc-500 hover:text-zinc-300"
                    }`}
                  >
                    {tab}
                  </button>
                ))}
              </div>
              <button
                type="button"
                onClick={() => setShowEditModal(false)}
                className="mb-2 flex h-7 w-7 items-center justify-center rounded-lg text-zinc-500 transition-colors hover:bg-white/5 hover:text-white"
              >
                <svg
                  className="h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  strokeWidth={2}
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M6 18 18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>

            <div className="p-5">
              {editTab === "general" ? (
                <form onSubmit={handleEdit} className="space-y-4">
                  <div>
                    <label className="mb-1.5 block text-xs font-medium text-zinc-400">
                      Name <span className="text-orange-500">*</span>
                    </label>
                    <input
                      type="text"
                      value={editName}
                      onChange={(e) => setEditName(e.target.value)}
                      className="w-full rounded-lg border border-white/10 bg-black/30 px-3 py-2 text-sm text-white placeholder:text-zinc-600 focus:border-orange-500/50 focus:outline-none focus:ring-1 focus:ring-orange-500/50"
                    />
                  </div>
                  <div>
                    <label className="mb-1.5 block text-xs font-medium text-zinc-400">
                      Description
                    </label>
                    <input
                      type="text"
                      value={editDescription}
                      onChange={(e) => setEditDescription(e.target.value)}
                      placeholder="Optional"
                      className="w-full rounded-lg border border-white/10 bg-black/30 px-3 py-2 text-sm text-white placeholder:text-zinc-600 focus:border-orange-500/50 focus:outline-none focus:ring-1 focus:ring-orange-500/50"
                    />
                  </div>
                  <div>
                    <p className="mb-2 text-xs font-medium text-zinc-400">
                      Trigger type
                    </p>
                    <SegmentedToggle
                      label=""
                      value={editTriggerType}
                      onChange={(v) => setEditTriggerType(v as TriggerType)}
                      options={[
                        { label: "Webhook", value: "webhook" },
                        { label: "Manual", value: "manual" },
                        { label: "Cron", value: "cron" },
                      ]}
                    />
                  </div>
                  {editTriggerType === "cron" && (
                    <div>
                      <label className="mb-1.5 block text-xs font-medium text-zinc-400">
                        Schedule <span className="text-orange-500">*</span>
                      </label>
                      <input
                        type="text"
                        value={editCronSchedule}
                        onChange={(e) => setEditCronSchedule(e.target.value)}
                        placeholder="0 9 * * *"
                        className="w-full rounded-lg border border-white/10 bg-black/30 px-3 py-2 font-mono text-sm text-white placeholder:text-zinc-600 focus:border-orange-500/50 focus:outline-none focus:ring-1 focus:ring-orange-500/50"
                      />
                      <p className="mt-1 text-xs text-zinc-600">
                        minute hour day month weekday — e.g.{" "}
                        <code className="text-zinc-500">0 9 * * *</code> = daily
                        at 9am
                      </p>
                    </div>
                  )}

                  <div className="flex gap-3 pt-1">
                    <button
                      type="submit"
                      disabled={updateMutation.isPending}
                      className="flex-1 rounded-lg bg-orange-500 py-2 text-sm font-medium text-white transition-colors hover:bg-orange-600 disabled:opacity-50"
                    >
                      {updateMutation.isPending ? "Saving…" : "Save changes"}
                    </button>
                    <button
                      type="button"
                      onClick={() => setShowEditModal(false)}
                      className="flex-1 rounded-lg border border-white/10 py-2 text-sm text-zinc-300 transition-colors hover:border-white/20 hover:text-white"
                    >
                      Cancel
                    </button>
                  </div>
                </form>
              ) : (
                <div className="space-y-3">
                  <div className="max-h-[60vh] space-y-3 overflow-y-auto pr-1">
                    {editActions.map((action, i) => (
                      <ActionCard
                        key={i}
                        index={i}
                        action={action}
                        canRemove={editActions.length > 1}
                        onUpdate={(idx, updated) =>
                          setEditActions((prev) =>
                            prev.map((a, j) => (j === idx ? updated : a)),
                          )
                        }
                        onRemove={(idx) =>
                          setEditActions((prev) =>
                            prev.filter((_, j) => j !== idx)
                          )
                        }
                      />
                    ))}
                  </div>
                  <button
                    type="button"
                    onClick={() =>
                      setEditActions((prev) => [
                        ...prev,
                        {
                          action_type: "debug_log",
                          config: {},
                          node_id: `action_${Math.random().toString(36).substr(2, 6)}`,
                        },
                      ])
                    }
                    className="flex items-center gap-1.5 text-xs text-orange-400 transition-colors hover:text-orange-300"
                  >
                    <svg
                      className="h-3.5 w-3.5"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      strokeWidth={2}
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        d="M12 4.5v15m7.5-7.5h-15"
                      />
                    </svg>
                    Add action
                  </button>
                  <div className="flex gap-3 pt-1">
                    <button
                      type="button"
                      onClick={handleSaveActions}
                      disabled={updateActionsMutation.isPending}
                      className="flex-1 rounded-lg bg-orange-500 py-2 text-sm font-medium text-white transition-colors hover:bg-orange-600 disabled:opacity-50"
                    >
                      {updateActionsMutation.isPending
                        ? "Saving…"
                        : "Save actions"}
                    </button>
                    <button
                      type="button"
                      onClick={() => setShowEditModal(false)}
                      className="flex-1 rounded-lg border border-white/10 py-2 text-sm text-zinc-300 transition-colors hover:border-white/20 hover:text-white"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {showTriggerPanel && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70">
          <div className="w-full max-w-md rounded-2xl border border-white/10 bg-[#1a1a1a] p-6">
            <h3 className="mb-1 text-base font-semibold text-white">
              Trigger relay manually
            </h3>
            <p className="mb-4 text-sm text-zinc-500">
              Optionally provide a JSON payload to pass as the trigger data.
            </p>
            <textarea
              value={triggerPayload}
              onChange={(e) => setTriggerPayload(e.target.value)}
              rows={5}
              className="w-full rounded-lg border border-white/10 bg-black/30 px-3 py-2 font-mono text-xs text-zinc-300 focus:border-purple-500/50 focus:outline-none focus:ring-1 focus:ring-purple-500/50"
              spellCheck={false}
            />
            <div className="mt-4 flex gap-3">
              <button
                type="button"
                onClick={handleManualTrigger}
                disabled={triggerMutation.isPending}
                className="flex-1 rounded-lg bg-purple-500 py-2 text-sm font-medium text-white transition-colors hover:bg-purple-600 disabled:opacity-50"
              >
                {triggerMutation.isPending ? "Triggering…" : "Run now"}
              </button>
              <button
                type="button"
                onClick={() => setShowTriggerPanel(false)}
                className="flex-1 rounded-lg border border-white/10 py-2 text-sm text-zinc-300 transition-colors hover:border-white/20 hover:text-white"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
      {showDeleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70">
          <div className="w-full max-w-sm rounded-2xl border border-white/10 bg-[#1a1a1a] p-6">
            <h3 className="mb-1 text-base font-semibold text-white">
              Delete relay?
            </h3>
            <p className="mb-5 text-sm text-zinc-500">
              This will permanently delete{" "}
              <span className="font-medium text-white">{relay.name}</span> and
              all its executions.
            </p>
            <div className="flex gap-3">
              <button
                type="button"
                onClick={handleDelete}
                disabled={deleteMutation.isPending}
                className="flex-1 rounded-lg bg-red-500 py-2 text-sm font-medium text-white transition-colors hover:bg-red-600 disabled:opacity-50"
              >
                {deleteMutation.isPending ? "Deleting…" : "Delete"}
              </button>
              <button
                type="button"
                onClick={() => setShowDeleteConfirm(false)}
                className="flex-1 rounded-lg border border-white/10 py-2 text-sm text-zinc-300 transition-colors hover:border-white/20 hover:text-white"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
