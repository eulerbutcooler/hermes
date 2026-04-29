"use client";

import { useRouter } from "next/navigation";
import { useId, useState, useEffect } from "react";
import { toast } from "sonner";
import { useConnections, useCreateRelay, useSecrets } from "@/lib/queries";
import { ActionCard, SegmentedToggle } from "@/components/action-config-fields";
import {
  ACTION_LABELS,
  ACTION_TYPES,
  type ActionType,
  type CreateRelayActionInput,
  type TriggerType,
} from "@/types/relay";

const TEMPLATE_HINT =
  "Supports {{ payload.x }}, {{ prev.output.x }}, {{ steps[0].output.x }}";

export default function NewRelayPage() {
  const router = useRouter();
  const createMutation = useCreateRelay();
  const nameId = useId();
  const descId = useId();
  const scheduleId = useId();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [triggerType, setTriggerType] = useState<TriggerType>("webhook");
  const [cronSchedule, setCronSchedule] = useState("0 9 * * *");
  const [actions, setActions] = useState<CreateRelayActionInput[]>([
    { action_type: "debug_log", config: {}, node_id: `action_${Math.random().toString(36).substr(2, 6)}` },
  ]);

  const addAction = () =>
    setActions((prev) => [
      ...prev,
      { action_type: "debug_log", config: {}, node_id: `action_${Math.random().toString(36).substr(2, 6)}` },
    ]);

  const updateAction = (index: number, updated: CreateRelayActionInput) =>
    setActions((prev) => prev.map((a, i) => (i === index ? updated : a)));

  const removeAction = (index: number) =>
    setActions((prev) =>
      prev.filter((_, i) => i !== index)
    );

  const handleSubmit = async (e: React.SyntheticEvent) => {
    e.preventDefault();
    if (!name.trim()) return toast.error("Name is required");
    if (triggerType === "cron" && !cronSchedule.trim())
      return toast.error("Cron schedule is required");
    try {
      const triggerConfig =
        triggerType === "cron" ? { schedule: cronSchedule.trim() } : {};
      const relay = await createMutation.mutateAsync({
        name,
        description,
        trigger_type: triggerType,
        trigger_config: triggerConfig,
        actions,
      });
      toast.success("Relay created!");
      router.push(`/dashboard/relays/${relay.id}`);
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : "Failed to create relay",
      );
    }
  };

  return (
    <div className="mx-auto w-full max-w-5xl p-8">
      <div className="mb-8">
        <h1 className="text-xl font-bold text-white">New Relay</h1>
        <p className="mt-0.5 text-sm text-zinc-500">
          Configure a trigger and its actions
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Basic info */}
        <div className="rounded-xl border border-white/10 bg-[#1a1a1a] p-5 space-y-4">
          <h2 className="text-sm font-medium text-zinc-300">Basic info</h2>
          <div>
            <label
              htmlFor={nameId}
              className="mb-1.5 block text-sm font-medium text-zinc-300"
            >
              Name <span className="text-orange-500">*</span>
            </label>
            <input
              id={nameId}
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="My relay"
              className="w-full rounded-lg border border-white/10 bg-white/5 px-4 py-2.5 text-sm text-white placeholder:text-zinc-500 focus:border-orange-500/50 focus:outline-none focus:ring-1 focus:ring-orange-500/50"
            />
          </div>
          <div>
            <label
              htmlFor={descId}
              className="mb-1.5 block text-sm font-medium text-zinc-300"
            >
              Description
            </label>
            <input
              id={descId}
              type="text"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Optional description"
              className="w-full rounded-lg border border-white/10 bg-white/5 px-4 py-2.5 text-sm text-white placeholder:text-zinc-500 focus:border-orange-500/50 focus:outline-none focus:ring-1 focus:ring-orange-500/50"
            />
          </div>
        </div>

        {/* Trigger */}
        <div className="rounded-xl border border-white/10 bg-[#1a1a1a] p-5 space-y-4">
          <h2 className="text-sm font-medium text-zinc-300">Trigger</h2>
          <SegmentedToggle
            label="Trigger type"
            value={triggerType}
            onChange={(v) => setTriggerType(v as TriggerType)}
            options={[
              { label: "Webhook", value: "webhook" },
              { label: "Manual", value: "manual" },
              { label: "Cron", value: "cron" },
            ]}
          />
          {triggerType === "webhook" && (
            <p className="text-xs text-zinc-500">
              A unique webhook URL will be generated. Send a POST request to it
              to fire this relay.
            </p>
          )}
          {triggerType === "manual" && (
            <p className="text-xs text-zinc-500">
              This relay can only be triggered manually from the dashboard or
              via the API.
            </p>
          )}
          {triggerType === "cron" && (
            <div>
              <label
                htmlFor={scheduleId}
                className="mb-1.5 block text-sm font-medium text-zinc-300"
              >
                Schedule <span className="text-orange-500">*</span>
              </label>
              <input
                id={scheduleId}
                type="text"
                value={cronSchedule}
                onChange={(e) => setCronSchedule(e.target.value)}
                placeholder="0 9 * * *"
                className="w-full rounded-lg border border-white/10 bg-white/5 px-4 py-2.5 font-mono text-sm text-white placeholder:text-zinc-500 focus:border-orange-500/50 focus:outline-none focus:ring-1 focus:ring-orange-500/50"
              />
              <p className="mt-1.5 text-xs text-zinc-500">
                Standard 5-field cron:{" "}
                <code className="text-zinc-400">
                  minute hour day month weekday
                </code>
                . Example: <code className="text-zinc-400">0 9 * * *</code> =
                every day at 9am.
              </p>
            </div>
          )}
        </div>

        {/* Actions */}
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-medium text-zinc-300">Actions</h2>
            <button
              type="button"
              onClick={addAction}
              className="flex items-center gap-1.5 text-xs text-orange-400 hover:text-orange-300 transition-colors"
            >
              <svg
                aria-hidden="true"
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
          </div>
          {actions.map((action, i) => (
            <ActionCard
              // biome-ignore lint/suspicious/noArrayIndexKey: order is stable here
              key={i}
              index={i}
              action={action}
              onUpdate={updateAction}
              onRemove={removeAction}
              canRemove={actions.length > 1}
            />
          ))}
        </div>

        {/* Submit */}
        <div className="flex items-center gap-3 pt-2">
          <button
            type="submit"
            disabled={createMutation.isPending}
            className="rounded-lg bg-orange-500 px-5 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-orange-600 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {createMutation.isPending ? "Creating…" : "Create relay"}
          </button>
          <button
            type="button"
            onClick={() => router.back()}
            className="rounded-lg border border-white/10 px-5 py-2.5 text-sm text-zinc-400 transition-colors hover:border-white/20 hover:text-white"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}
