export type SourceItem = {
  name: string
  kind: string
  config: Record<string, any>
  createdAt?: string
  updatedAt?: string
}

const BASE = "/tb-api/api/config";

export async function listSources(params?: { dbPath?: string }): Promise<SourceItem[]> {
  const q = params?.dbPath ? `?dbPath=${encodeURIComponent(params.dbPath)}` : "";
  const res = await fetch(`${BASE}/sources/${q}`, { cache: "no-store" });
  if (!res.ok) throw new Error(`listSources failed: ${res.status}`);
  return res.json();
}

export async function createSource(body: { name: string; kind: string; config?: Record<string, any> }, params?: { dbPath?: string }): Promise<SourceItem> {
  const q = params?.dbPath ? `?dbPath=${encodeURIComponent(params.dbPath)}` : "";
  const res = await fetch(`${BASE}/sources/${q}`, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`createSource failed: ${res.status}`);
  return res.json();
}

export async function deleteSource(name: string, params?: { dbPath?: string }): Promise<void> {
  const q = params?.dbPath ? `?dbPath=${encodeURIComponent(params.dbPath)}` : "";
  const res = await fetch(`${BASE}/sources/${encodeURIComponent(name)}/${q}`, {
    method: "DELETE",
  });
  if (!res.ok && res.status !== 204) throw new Error(`deleteSource failed: ${res.status}`);
}

export async function getSource(name: string, params?: { dbPath?: string }): Promise<SourceItem> {
  const q = params?.dbPath ? `?dbPath=${encodeURIComponent(params.dbPath)}` : "";
  const res = await fetch(`${BASE}/sources/${encodeURIComponent(name)}/${q}`);
  if (!res.ok) throw new Error(`getSource failed: ${res.status}`);
  return res.json();
}

export async function updateSource(name: string, body: { kind: string; config?: Record<string, any> }, params?: { dbPath?: string }): Promise<SourceItem> {
  const q = params?.dbPath ? `?dbPath=${encodeURIComponent(params.dbPath)}` : "";
  const res = await fetch(`${BASE}/sources/${encodeURIComponent(name)}/${q}`, {
    method: "PUT",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`updateSource failed: ${res.status}`);
  return res.json();
}
