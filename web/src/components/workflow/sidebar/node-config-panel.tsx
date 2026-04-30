"use client";

import { useWorkflow } from "@/components/workflow/workflow-provider";
import { ActionConfigFields } from "@/components/action-config-fields";
import type {
  ActionNodeData,
  ConditionData,
  ConditionNodeData,
  TriggerNodeData,
  WorkflowEdgeData,
} from "@/types/workflow";
import type { ActionType } from "@/types/relay";
import { useState } from "react";

const OPERATORS: ConditionData["operator"][] = [
  "==", "!=", ">", "<", ">=", "<=", "contains", "exists",
];

export function NodeConfigPanel() {
  const {
    nodes,
    edges,
    selectedNodeId,
    selectedEdgeId,
    updateNodeData,
    updateEdgeData,
    selectNode,
    selectEdge,
    relayMeta,
    setMeta,
  } = useWorkflow();

  const selectedNode = nodes.find((n) => n.id === selectedNodeId);
  const selectedEdge = edges.find((e) => e.id === selectedEdgeId);

  // ── Trigger config ──────────────────────────────────────────────────────

  if (!selectedNode && !selectedEdge) return null;

  const handleBack = () => {
    selectNode(null);
    selectEdge(null);
  };

  return (
    <div className="config-panel">
      {/* Header */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-white/8 flex-shrink-0">
        <button
          type="button"
          onClick={handleBack}
          className="flex h-6 w-6 items-center justify-center rounded text-zinc-500 hover:bg-white/5 hover:text-white transition-colors"
        >
          <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 19.5 8.25 12l7.5-7.5" />
          </svg>
        </button>
        <span className="text-xs font-semibold text-zinc-300">
          {selectedNode
            ? selectedNode.type === "trigger"
              ? "Trigger Config"
              : selectedNode.type === "condition"
              ? "Condition Config"
              : "Action Config"
            : "Edge Condition"}
        </span>
      </div>

      <div className="config-panel-body">
        {/* Trigger node config */}
        {selectedNode?.type === "trigger" && (
          <TriggerConfig
            data={selectedNode.data as TriggerNodeData}
            relayMeta={relayMeta}
            onChangeMeta={setMeta}
            onChangeData={(d) => updateNodeData(selectedNode.id, d)}
          />
        )}

        {/* Action node config */}
        {selectedNode?.type === "action" && (
          <ActionConfig
            data={selectedNode.data as ActionNodeData}
            onChange={(d) => updateNodeData(selectedNode.id, d)}
          />
        )}

        {/* Condition node config */}
        {selectedNode?.type === "condition" && (
          <ConditionNodeConfig
            data={selectedNode.data as ConditionNodeData}
            onChange={(d) => updateNodeData(selectedNode.id, d)}
          />
        )}

        {/* Edge condition config */}
        {selectedEdge && (
          <EdgeConditionConfig
            data={(selectedEdge.data ?? {}) as WorkflowEdgeData}
            onChange={(d) => updateEdgeData(selectedEdge.id, d)}
          />
        )}
      </div>
    </div>
  );
}

// ─── Trigger Config ────────────────────────────────────────────────────────────

const TRIGGER_OPTIONS: { value: TriggerNodeData["triggerType"]; label: string; desc: string }[] = [
  { value: "webhook", label: "Webhook", desc: "Receives an HTTP POST" },
  { value: "manual",  label: "Manual",  desc: "Triggered from dashboard" },
  { value: "cron",    label: "Scheduled", desc: "Runs on a cron schedule" },
];

function TriggerConfig({
  data,
  relayMeta,
  onChangeMeta,
  onChangeData,
}: {
  data: TriggerNodeData;
  relayMeta: import("@/types/workflow").RelayMeta;
  onChangeMeta: (m: Partial<import("@/types/workflow").RelayMeta>) => void;
  onChangeData: (d: Record<string, unknown>) => void;
}) {
  const handleTypeChange = (type: TriggerNodeData["triggerType"]) => {
    const label =
      type === "webhook" ? "Webhook Trigger"
      : type === "cron"    ? "Cron Trigger"
      : "Manual Trigger";
    onChangeData({ triggerType: type, triggerConfig: {}, label });
    onChangeMeta({ triggerType: type, triggerConfig: {} });
  };

  return (
    <div className="space-y-4">
      {/* Trigger type switcher */}
      <div>
        <p className="mb-2 text-[11px] font-medium text-zinc-400">Trigger type</p>
        <div className="grid grid-cols-3 gap-1 rounded-lg border border-white/8 bg-white/3 p-1">
          {TRIGGER_OPTIONS.map((opt) => (
            <button
              key={opt.value}
              type="button"
              onClick={() => handleTypeChange(opt.value)}
              className={`rounded px-2 py-1.5 text-[10px] font-medium transition-colors ${
                data.triggerType === opt.value
                  ? "bg-orange-500 text-white"
                  : "text-zinc-500 hover:text-zinc-300"
              }`}
            >
              {opt.label}
            </button>
          ))}
        </div>
        <p className="mt-1.5 text-[10px] text-zinc-600">
          {TRIGGER_OPTIONS.find((o) => o.value === data.triggerType)?.desc}
        </p>
      </div>

      <Field label="Relay Name">
        <input
          type="text"
          value={relayMeta.name}
          onChange={(e) => onChangeMeta({ name: e.target.value })}
          className="config-input"
          placeholder="My relay"
        />
      </Field>
      <Field label="Description">
        <input
          type="text"
          value={relayMeta.description}
          onChange={(e) => onChangeMeta({ description: e.target.value })}
          className="config-input"
          placeholder="Optional description"
        />
      </Field>

      {data.triggerType === "cron" && (
        <Field label="Cron Schedule">
          <input
            type="text"
            value={(data.triggerConfig?.schedule as string) ?? "0 9 * * *"}
            onChange={(e) => {
              const newConfig = { ...data.triggerConfig, schedule: e.target.value };
              onChangeData({ triggerConfig: newConfig });
              onChangeMeta({ triggerConfig: newConfig });
            }}
            className="config-input font-mono"
            placeholder="0 9 * * *"
          />
          <p className="mt-1 text-[10px] text-zinc-600">min hour day month weekday</p>
        </Field>
      )}

      {data.triggerType === "webhook" && (
        <p className="text-[11px] text-zinc-500 leading-relaxed">
          A unique webhook URL will be generated when you save this relay.
        </p>
      )}

      {data.triggerType === "manual" && (
        <p className="text-[11px] text-zinc-500 leading-relaxed">
          This relay will only run when you click <strong className="text-zinc-300">Run</strong> in the dashboard.
        </p>
      )}
    </div>
  );
}

// ─── Action Config ─────────────────────────────────────────────────────────────

function ActionConfig({
  data,
  onChange,
}: {
  data: ActionNodeData;
  onChange: (d: Record<string, unknown>) => void;
}) {
  const handleConfigChange = (key: string, value: unknown) => {
    const next = { ...data.config, [key]: value };
    if (value === undefined) delete next[key];
    onChange({ config: next });
  };

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <p className="text-[10px] text-zinc-500 uppercase tracking-wider font-semibold">
          {data.actionType.replace(/_/g, " ")}
        </p>
        <span className="text-[9px] font-mono text-zinc-600 bg-black/20 px-1.5 py-0.5 rounded" title="Use this ID in your template variables: {{ steps['node_id'].output... }}">
          ID: {data.nodeId}
        </span>
      </div>
      <ActionConfigFields
        type={data.actionType as ActionType}
        config={data.config}
        onChange={handleConfigChange}
      />
    </div>
  );
}

// ─── Condition Node Config ─────────────────────────────────────────────────────

function ConditionNodeConfig({
  data,
  onChange,
}: {
  data: ConditionNodeData;
  onChange: (d: Record<string, unknown>) => void;
}) {
  const cond: ConditionData = data.condition ?? { field: "", operator: "==", value: "" };

  const update = (partial: Partial<ConditionData>) => {
    onChange({ condition: { ...cond, ...partial } });
  };

  return (
    <div className="space-y-4">
      <Field label="Node Label">
        <input
          type="text"
          value={data.label}
          onChange={(e) => onChange({ label: e.target.value })}
          className="config-input"
          placeholder="Condition"
        />
      </Field>
      
      <div className="space-y-3 pt-2 border-t border-white/5">
        <Field label="Field path">
          <input
            type="text"
            value={cond.field}
            onChange={(e) => update({ field: e.target.value })}
            className="config-input font-mono"
            placeholder="payload.status"
          />
        </Field>
        <Field label="Operator">
          <select
            value={cond.operator}
            onChange={(e) => update({ operator: e.target.value as ConditionData["operator"] })}
            className="config-input"
          >
            {OPERATORS.map((op) => (
              <option key={op} value={op}>{op}</option>
            ))}
          </select>
        </Field>
        {cond.operator !== "exists" && (
          <Field label="Value">
            <input
              type="text"
              value={cond.value}
              onChange={(e) => update({ value: e.target.value })}
              className="config-input"
              placeholder="200"
            />
          </Field>
        )}
      </div>

      <p className="text-[11px] text-zinc-500 leading-relaxed">
        Connect edges to the True/False handles.
      </p>
    </div>
  );
}

// ─── Edge Condition Config ─────────────────────────────────────────────────────

function EdgeConditionConfig({
  data,
  onChange,
}: {
  data: WorkflowEdgeData;
  onChange: (d: Record<string, unknown>) => void;
}) {
  const [enabled, setEnabled] = useState(!!data.condition);

  const cond: ConditionData = data.condition ?? { field: "", operator: "==", value: "" };

  const update = (partial: Partial<ConditionData>) => {
    onChange({ condition: { ...cond, ...partial } });
  };

  const toggle = (on: boolean) => {
    setEnabled(on);
    onChange({ condition: on ? cond : null });
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-xs font-medium text-zinc-300">Edge Condition</p>
        <button
          type="button"
          onClick={() => toggle(!enabled)}
          className={`relative inline-flex h-5 w-9 flex-shrink-0 rounded-full border-2 border-transparent transition-colors ${
            enabled ? "bg-orange-500" : "bg-zinc-700"
          }`}
        >
          <span
            className={`inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform ${
              enabled ? "translate-x-4" : "translate-x-0"
            }`}
          />
        </button>
      </div>

      {enabled && (
        <div className="space-y-3">
          <Field label="Field path">
            <input
              type="text"
              value={cond.field}
              onChange={(e) => update({ field: e.target.value })}
              className="config-input font-mono"
              placeholder="payload.status or steps.http.status_code"
            />
          </Field>
          <Field label="Operator">
            <select
              value={cond.operator}
              onChange={(e) => update({ operator: e.target.value as ConditionData["operator"] })}
              className="config-input"
            >
              {OPERATORS.map((op) => (
                <option key={op} value={op}>{op}</option>
              ))}
            </select>
          </Field>
          {cond.operator !== "exists" && (
            <Field label="Value">
              <input
                type="text"
                value={cond.value}
                onChange={(e) => update({ value: e.target.value })}
                className="config-input"
                placeholder="200"
              />
            </Field>
          )}
        </div>
      )}

      {!enabled && (
        <p className="text-[11px] text-zinc-500">
          This edge always passes. Enable a condition to make it conditional.
        </p>
      )}
    </div>
  );
}

// ─── Field Helper ──────────────────────────────────────────────────────────────

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <label className="mb-1 block text-[11px] font-medium text-zinc-400">{label}</label>
      {children}
    </div>
  );
}
