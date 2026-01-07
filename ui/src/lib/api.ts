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

export type PromptItem = {
  name: string
  kind: string
  config: Record<string, any>
  createdAt?: string
  updatedAt?: string
}

export type ToolsetItem = {
  name: string
  toolNames: string[]
  createdAt?: string
  updatedAt?: string
}

export type PromptsetItem = {
  name: string
  promptNames: string[]
  createdAt?: string
  updatedAt?: string
}

const BASE = "/tb-api/api/config";
const SQL_BASE = "/tb-api/api/sql";

export type ExecuteSQLResponse = {
  columns: string[]
  rows: any[][]
  rowCount: number
  durationMs: number
}

export async function executeSQL(body: {
  source: string
  statement: string
  parameters?: any[] | Record<string, any>
  readOnly?: boolean
}): Promise<ExecuteSQLResponse> {
  const res = await fetch(`${SQL_BASE}/execute`, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`executeSQL failed: ${res.status}`)
  return res.json()
}

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

export async function listToolKinds(prefix?: string): Promise<string[]> {
  const qs = prefix ? `?prefix=${encodeURIComponent(prefix)}` : ""
  const res = await fetch(`${BASE}/tools/kinds${qs}`, { cache: "no-store" })
  if (!res.ok) throw new Error(`listToolKinds failed: ${res.status}`)
  return res.json()
}

export async function listPrompts(): Promise<PromptItem[]> {
  const res = await fetch(`${BASE}/prompts/`, { cache: "no-store" })
  if (!res.ok) throw new Error(`listPrompts failed: ${res.status}`)
  return res.json()
}

export async function listToolsets(): Promise<ToolsetItem[]> {
  const res = await fetch(`${BASE}/toolsets/`, { cache: "no-store" })
  if (!res.ok) throw new Error(`listToolsets failed: ${res.status}`)
  return res.json()
}

export async function createToolset(body: { name: string; toolNames: string[] }): Promise<ToolsetItem> {
  const res = await fetch(`${BASE}/toolsets/`, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`createToolset failed: ${res.status}`)
  return res.json()
}

export async function getToolset(name: string): Promise<ToolsetItem> {
  const res = await fetch(`${BASE}/toolsets/${encodeURIComponent(name)}/`)
  if (!res.ok) throw new Error(`getToolset failed: ${res.status}`)
  return res.json()
}

export async function updateToolset(name: string, body: { toolNames: string[] }): Promise<ToolsetItem> {
  const res = await fetch(`${BASE}/toolsets/${encodeURIComponent(name)}/`, {
    method: "PUT",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`updateToolset failed: ${res.status}`)
  return res.json()
}

export async function deleteToolset(name: string): Promise<void> {
  const res = await fetch(`${BASE}/toolsets/${encodeURIComponent(name)}/`, {
    method: "DELETE",
  })
  if (!res.ok && res.status !== 204) throw new Error(`deleteToolset failed: ${res.status}`)
}

export async function listPromptsets(): Promise<PromptsetItem[]> {
  const res = await fetch(`${BASE}/promptsets/`, { cache: "no-store" })
  if (!res.ok) throw new Error(`listPromptsets failed: ${res.status}`)
  return res.json()
}

export async function createPromptset(body: { name: string; promptNames: string[] }): Promise<PromptsetItem> {
  const res = await fetch(`${BASE}/promptsets/`, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`createPromptset failed: ${res.status}`)
  return res.json()
}

export async function getPromptset(name: string): Promise<PromptsetItem> {
  const res = await fetch(`${BASE}/promptsets/${encodeURIComponent(name)}/`)
  if (!res.ok) throw new Error(`getPromptset failed: ${res.status}`)
  return res.json()
}

export async function updatePromptset(name: string, body: { promptNames: string[] }): Promise<PromptsetItem> {
  const res = await fetch(`${BASE}/promptsets/${encodeURIComponent(name)}/`, {
    method: "PUT",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`updatePromptset failed: ${res.status}`)
  return res.json()
}

export async function deletePromptset(name: string): Promise<void> {
  const res = await fetch(`${BASE}/promptsets/${encodeURIComponent(name)}/`, {
    method: "DELETE",
  })
  if (!res.ok && res.status !== 204) throw new Error(`deletePromptset failed: ${res.status}`)
}
