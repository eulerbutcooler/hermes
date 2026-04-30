import type { SaveRelayDAGRequest } from "@/types/workflow";
import type { RelayWithActions } from "@/types/relay";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8081/api/v1";

function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("token");
}

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getToken();
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
  });
  const contentType = res.headers.get("content-type") ?? "";
  const json = contentType.includes("application/json") ? await res.json() : null;
  if (!res.ok) {
    throw new Error(
      (json as { error?: string } | null)?.error ??
        `Request failed: ${res.status} ${res.statusText}`
    );
  }
  return json as T;
}

// ─── DAG relay endpoints (when friend's branch merges) ────────────────────────

/** Create a new relay with full DAG payload */
export async function createRelayDAG(
  payload: SaveRelayDAGRequest
): Promise<RelayWithActions> {
  const res = await apiFetch<{ success: boolean; data: RelayWithActions }>(
    "/relays",
    { method: "POST", body: JSON.stringify(payload) }
  );
  return res.data;
}

/** Update an existing relay's DAG (nodes + edges) */
export async function updateRelayDAG(
  id: string,
  payload: SaveRelayDAGRequest
): Promise<RelayWithActions> {
  // Update metadata first
  await apiFetch(`/relays/${id}`, {
    method: "PUT",
    body: JSON.stringify({
      name: payload.name,
      description: payload.description,
      trigger_type: payload.trigger_type,
      trigger_config: payload.trigger_config,
    }),
  });

  // Then update DAG (actions + edges)
  const res = await apiFetch<{ success: boolean; data: RelayWithActions }>(
    `/relays/${id}/actions`,
    { 
      method: "PUT", 
      body: JSON.stringify({
        actions: payload.actions,
        edges: payload.edges,
      }) 
    }
  );
  return res.data;
}

/** Get a relay with full DAG (nodes + edges) */
export async function getRelayDAG(id: string): Promise<RelayWithActions & {
  edges?: Array<{
    parent_node_id: string;
    child_node_id: string;
    condition?: import("@/types/workflow").ConditionData | null;
  }>;
}> {
  const res = await apiFetch<{ success: boolean; data: RelayWithActions }>(
    `/relays/${id}`
  );
  return res.data as RelayWithActions & {
    edges?: Array<{
      parent_node_id: string;
      child_node_id: string;
      condition?: import("@/types/workflow").ConditionData | null;
    }>;
  };
}
