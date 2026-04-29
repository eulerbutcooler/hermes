"use client";

import { useState, useCallback } from "react";
import type { PaletteItem } from "@/types/workflow";

const PALETTE_ITEMS: PaletteItem[] = [
  // Triggers
  {
    type: "trigger",
    triggerType: "webhook",
    label: "Webhook",
    description: "Receives an HTTP POST",
    icon: "webhook",
    color: "text-blue-400",
    bgColor: "bg-blue-500/10",
  },
  {
    type: "trigger",
    triggerType: "manual",
    label: "Manual",
    description: "Triggered from dashboard",
    icon: "play",
    color: "text-purple-400",
    bgColor: "bg-purple-500/10",
  },
  {
    type: "trigger",
    triggerType: "cron",
    label: "Scheduled",
    description: "Runs on a cron schedule",
    icon: "clock",
    color: "text-amber-400",
    bgColor: "bg-amber-500/10",
  },
  // Logic
  {
    type: "condition",
    label: "Condition",
    description: "If / else branching",
    icon: "branch",
    color: "text-purple-400",
    bgColor: "bg-purple-500/10",
  },
  // Actions
  {
    type: "action",
    actionType: "discord_send",
    label: "Discord",
    description: "Send a Discord message",
    icon: "discord",
    color: "text-indigo-400",
    bgColor: "bg-indigo-500/10",
  },
  {
    type: "action",
    actionType: "slack_send",
    label: "Slack",
    description: "Send a Slack message",
    icon: "slack",
    color: "text-green-400",
    bgColor: "bg-green-500/10",
  },
  {
    type: "action",
    actionType: "http_request",
    label: "HTTP Request",
    description: "Call any external API",
    icon: "globe",
    color: "text-orange-400",
    bgColor: "bg-orange-500/10",
  },
  {
    type: "action",
    actionType: "email_send",
    label: "Email",
    description: "Send an email via OAuth",
    icon: "email",
    color: "text-sky-400",
    bgColor: "bg-sky-500/10",
  },
  {
    type: "action",
    actionType: "debug_log",
    label: "Debug Log",
    description: "Log payload for testing",
    icon: "debug",
    color: "text-zinc-400",
    bgColor: "bg-zinc-500/10",
  },
];

function PaletteIcon({ name, className }: { name: string; className: string }) {
  switch (name) {
    case "webhook":
      return <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}><path strokeLinecap="round" strokeLinejoin="round" d="M13.19 8.688a4.5 4.5 0 0 1 1.242 7.244l-4.5 4.5a4.5 4.5 0 0 1-6.364-6.364l1.757-1.757m13.35-.622 1.757-1.757a4.5 4.5 0 0 0-6.364-6.364l-4.5 4.5a4.5 4.5 0 0 0 1.242 7.244" /></svg>;
    case "play":
      return <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}><path strokeLinecap="round" strokeLinejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 0 1 0 1.972l-11.54 6.347a1.125 1.125 0 0 1-1.667-.986V5.653Z" /></svg>;
    case "clock":
      return <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}><path strokeLinecap="round" strokeLinejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>;
    case "branch":
      return <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}><path strokeLinecap="round" strokeLinejoin="round" d="M7.5 21 3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" /></svg>;
    case "discord":
      return <svg className={className} viewBox="0 0 24 24" fill="currentColor"><path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057.099 18.082.114 18.105.132 18.12a19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028c.462-.63.874-1.295 1.226-1.994a.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03z" /></svg>;
    case "slack":
      return <svg className={className} viewBox="0 0 24 24" fill="currentColor"><path d="M5.042 15.165a2.528 2.528 0 0 1-2.52 2.523A2.528 2.528 0 0 1 0 15.165a2.527 2.527 0 0 1 2.522-2.52h2.52v2.52zm1.271 0a2.527 2.527 0 0 1 2.521-2.52 2.527 2.527 0 0 1 2.521 2.52v6.313A2.528 2.528 0 0 1 8.834 24a2.528 2.528 0 0 1-2.521-2.522v-6.313zM8.834 5.042a2.528 2.528 0 0 1-2.521-2.52A2.528 2.528 0 0 1 8.834 0a2.528 2.528 0 0 1 2.521 2.522v2.52H8.834zm0 1.271a2.528 2.528 0 0 1 2.521 2.521 2.528 2.528 0 0 1-2.521 2.521H2.522A2.528 2.528 0 0 1 0 8.834a2.528 2.528 0 0 1 2.522-2.521h6.312zm10.122 2.521a2.528 2.528 0 0 1 2.522-2.521A2.528 2.528 0 0 1 24 8.834a2.528 2.528 0 0 1-2.522 2.521h-2.522V8.834zm-1.268 0a2.528 2.528 0 0 1-2.523 2.521 2.527 2.527 0 0 1-2.52-2.521V2.522A2.527 2.527 0 0 1 15.165 0a2.528 2.528 0 0 1 2.523 2.522v6.312zm-2.523 10.122a2.528 2.528 0 0 1 2.523 2.522A2.528 2.528 0 0 1 15.165 24a2.527 2.527 0 0 1-2.52-2.522v-2.522h2.52zm0-1.268a2.527 2.527 0 0 1-2.52-2.523 2.526 2.526 0 0 1 2.52-2.52h6.313A2.527 2.527 0 0 1 24 15.165a2.528 2.528 0 0 1-2.522 2.523h-6.313z" /></svg>;
    case "globe":
      return <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}><path strokeLinecap="round" strokeLinejoin="round" d="M12 21a9.004 9.004 0 0 0 8.716-6.747M12 21a9.004 9.004 0 0 1-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 0 1 7.843 4.582M12 3a8.997 8.997 0 0 0-7.843 4.582m15.686 0A11.953 11.953 0 0 1 12 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0 1 21 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0 1 12 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 0 1 3 12c0-1.605.42-3.113 1.157-4.418" /></svg>;
    case "email":
      return <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}><path strokeLinecap="round" strokeLinejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 0 1-2.25 2.25h-15a2.25 2.25 0 0 1-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25m19.5 0v.243a2.25 2.25 0 0 1-1.07 1.916l-7.5 4.615a2.25 2.25 0 0 1-2.36 0L3.32 8.91a2.25 2.25 0 0 1-1.07-1.916V6.75" /></svg>;
    default:
      return <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}><path strokeLinecap="round" strokeLinejoin="round" d="M6.75 7.5 3 12m0 0 3.75 4.5M3 12h18" /></svg>;
  }
}

interface PaletteProps {
  hasTrigger: boolean;
}

export function Palette({ hasTrigger }: PaletteProps) {
  const [search, setSearch] = useState("");

  const onDragStart = useCallback(
    (e: React.DragEvent, item: PaletteItem) => {
      e.dataTransfer.setData("application/hermes-node", JSON.stringify(item));
      e.dataTransfer.effectAllowed = "move";
    },
    []
  );

  const filtered = PALETTE_ITEMS.filter((item) =>
    item.label.toLowerCase().includes(search.toLowerCase())
  );

  const triggers = filtered.filter((i) => i.type === "trigger");
  const logic = filtered.filter((i) => i.type === "condition");
  const actions = filtered.filter((i) => i.type === "action");

  return (
    <div className="workflow-palette flex flex-col">
      {/* Header */}
      <div className="p-4 border-b border-white/8">
        <h2 className="text-xs font-semibold text-zinc-300 mb-3 uppercase tracking-wider">
          Components
        </h2>
        <input
          type="text"
          placeholder="Search…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full rounded-lg border border-white/10 bg-white/5 px-3 py-1.5 text-xs text-white placeholder:text-zinc-600 focus:border-orange-500/40 focus:outline-none"
        />
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-4">
        {/* Triggers */}
        {triggers.length > 0 && (
          <Section label="Triggers">
            {triggers.map((item) => (
              <PaletteCard
                key={`${item.type}-${item.triggerType}`}
                item={item}
                disabled={hasTrigger && item.type === "trigger"}
                onDragStart={onDragStart}
              />
            ))}
          </Section>
        )}

        {/* Logic */}
        {logic.length > 0 && (
          <Section label="Logic">
            {logic.map((item) => (
              <PaletteCard
                key={item.type}
                item={item}
                disabled={false}
                onDragStart={onDragStart}
              />
            ))}
          </Section>
        )}

        {/* Actions */}
        {actions.length > 0 && (
          <Section label="Actions">
            {actions.map((item) => (
              <PaletteCard
                key={item.actionType}
                item={item}
                disabled={false}
                onDragStart={onDragStart}
              />
            ))}
          </Section>
        )}
      </div>

      {/* Hint */}
      <div className="p-3 border-t border-white/8">
        <p className="text-[10px] text-zinc-600 leading-relaxed text-center">
          Drag components onto the canvas
        </p>
      </div>
    </div>
  );
}

function Section({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <p className="mb-2 text-[10px] font-semibold text-zinc-500 uppercase tracking-wider px-1">
        {label}
      </p>
      <div className="space-y-1.5">{children}</div>
    </div>
  );
}

function PaletteCard({
  item,
  disabled,
  onDragStart,
}: {
  item: PaletteItem;
  disabled: boolean;
  onDragStart: (e: React.DragEvent, item: PaletteItem) => void;
}) {
  return (
    <div
      draggable={!disabled}
      onDragStart={disabled ? undefined : (e) => onDragStart(e, item)}
      className={`palette-item ${disabled ? "opacity-40 cursor-not-allowed" : ""}`}
      title={disabled ? "Trigger already added" : item.description}
    >
      <div className={`flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-lg ${item.bgColor}`}>
        <PaletteIcon name={item.icon} className={`h-3.5 w-3.5 ${item.color}`} />
      </div>
      <div className="min-w-0">
        <p className="text-xs font-medium text-white">{item.label}</p>
        <p className="text-[10px] text-zinc-500 truncate">{item.description}</p>
      </div>
    </div>
  );
}
