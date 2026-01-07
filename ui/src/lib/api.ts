export type SourceItem = {
  name: string
  kind: string
  config: Record<string, any>
  createdAt?: string
  updatedAt?: string
}

export type ValidateSourceResult = {
  available: boolean
  error?: string
}

export type ToolItem = {
  name: string
  kind: string
  sourceName?: string
  config: Record<string, any>
  createdAt?: string
  updatedAt?: string
}

const BASE = "/tb-api/api/config";

export async function listSources(): Promise<SourceItem[]> {
  const res = await fetch(`${BASE}/sources/`, { cache: "no-store" });
  if (!res.ok) throw new Error(`listSources failed: ${res.status}`);
  return res.json();
}

export async function createSource(body: { name: string; kind: string; config?: Record<string, any> }): Promise<SourceItem> {
  const res = await fetch(`${BASE}/sources/`, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`createSource failed: ${res.status}`);
  return res.json();
}

export async function validateSource(body: { kind: string; config?: Record<string, any> }): Promise<ValidateSourceResult> {
  const res = await fetch(`${BASE}/sources/validate`, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`validateSource failed: ${res.status}`);
  return res.json();
}

export async function deleteSource(name: string): Promise<void> {
  const res = await fetch(`${BASE}/sources/${encodeURIComponent(name)}/`, {
    method: "DELETE",
  });
  if (!res.ok && res.status !== 204) throw new Error(`deleteSource failed: ${res.status}`);
}

export async function getSource(name: string): Promise<SourceItem> {
  const res = await fetch(`${BASE}/sources/${encodeURIComponent(name)}/`);
  if (!res.ok) throw new Error(`getSource failed: ${res.status}`);
  return res.json();
}

export async function updateSource(name: string, body: { kind: string; config?: Record<string, any> }): Promise<SourceItem> {
  const res = await fetch(`${BASE}/sources/${encodeURIComponent(name)}/`, {
    method: "PUT",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`updateSource failed: ${res.status}`);
  return res.json();
}

export async function listTools(): Promise<ToolItem[]> {
  const res = await fetch(`${BASE}/tools/`, { cache: "no-store" })
  if (!res.ok) throw new Error(`listTools failed: ${res.status}`)
  return res.json()
}

export async function createTool(body: {
  name: string
  kind: string
  sourceName?: string | null
  config?: Record<string, any>
}): Promise<ToolItem> {
  const res = await fetch(`${BASE}/tools/`, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`createTool failed: ${res.status}`)
  return res.json()
}

export async function updateTool(
  name: string,
  body: { kind: string; sourceName?: string | null; config?: Record<string, any> },
): Promise<ToolItem> {
  const res = await fetch(`${BASE}/tools/${encodeURIComponent(name)}/`, {
    method: "PUT",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`updateTool failed: ${res.status}`)
  return res.json()
}
