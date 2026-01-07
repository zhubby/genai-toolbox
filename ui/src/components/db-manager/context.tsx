"use client"

import * as React from "react"
import { listTools, type ExecuteSQLResponse, type SourceItem, type ToolItem } from "@/lib/api"

export type SelectedSource = Pick<SourceItem, "name" | "kind" | "config">

export type PreviewState =
  | { status: "idle"; data: null; error: null }
  | { status: "running"; data: ExecuteSQLResponse | null; error: null }
  | { status: "success"; data: ExecuteSQLResponse; error: null }
  | { status: "error"; data: ExecuteSQLResponse | null; error: string }

type DbManagerState = {
  selectedSource: SelectedSource | null
  selectSource: (source: SelectedSource | null) => void
  tools: ToolItem[]
  toolsLoading: boolean
  toolsError: string | null
  reloadTools: () => Promise<void>
  selectedToolName: string | null
  selectedTool: ToolItem | null
  selectTool: (toolName: string | null) => void
  preview: PreviewState
  setPreview: (next: PreviewState) => void
}

const DbManagerContext = React.createContext<DbManagerState | null>(null)

const LS_KEY = "tb.dbManager.selectedSourceName"

export function DbManagerProvider({ children }: { children: React.ReactNode }) {
  const [selectedSource, setSelectedSource] = React.useState<SelectedSource | null>(null)
  const [tools, setTools] = React.useState<ToolItem[]>([])
  const [toolsLoading, setToolsLoading] = React.useState(false)
  const [toolsError, setToolsError] = React.useState<string | null>(null)
  const [selectedToolName, setSelectedToolName] = React.useState<string | null>(null)
  const [preview, setPreview] = React.useState<PreviewState>({ status: "idle", data: null, error: null })

  // 仅持久化 name（source 详情由 Sidebar 点击时补齐）
  React.useEffect(() => {
    try {
      const saved = localStorage.getItem(LS_KEY)
      if (saved) setSelectedSource({ name: saved, kind: "", config: {} })
    } catch {
      // ignore
    }
  }, [])

  React.useEffect(() => {
    try {
      if (selectedSource?.name) localStorage.setItem(LS_KEY, selectedSource.name)
      else localStorage.removeItem(LS_KEY)
    } catch {
      // ignore
    }
  }, [selectedSource?.name])

  const reloadTools = React.useCallback(async () => {
    const sourceName = selectedSource?.name
    if (!sourceName) {
      setTools([])
      setToolsError(null)
      setSelectedToolName(null)
      return
    }
    setToolsLoading(true)
    setToolsError(null)
    try {
      const all = await listTools()
      const filtered = all.filter((t) => t.sourceName === sourceName)
      setTools(filtered)
    } catch (e: any) {
      setToolsError(e?.message || String(e))
    } finally {
      setToolsLoading(false)
    }
  }, [selectedSource?.name])

  React.useEffect(() => {
    void reloadTools()
  }, [reloadTools])

  const selectedTool = React.useMemo(() => {
    if (!selectedToolName) return null
    return tools.find((t) => t.name === selectedToolName) || null
  }, [tools, selectedToolName])

  // 如果 tools 列表变化导致当前选中项不存在，则清空选中
  React.useEffect(() => {
    if (selectedToolName && !selectedTool) setSelectedToolName(null)
  }, [selectedToolName, selectedTool])

  const selectSource = React.useCallback((next: SelectedSource | null) => {
    setSelectedSource(next)
    // 切换 source 时清空选中 tool，避免跨 source 误加载
    setSelectedToolName(null)
  }, [])

  const value = React.useMemo<DbManagerState>(
    () => ({
      selectedSource,
      selectSource,
      tools,
      toolsLoading,
      toolsError,
      reloadTools,
      selectedToolName,
      selectedTool,
      selectTool: setSelectedToolName,
      preview,
      setPreview,
    }),
    [selectedSource, selectSource, tools, toolsLoading, toolsError, reloadTools, selectedToolName, selectedTool, preview],
  )

  return <DbManagerContext.Provider value={value}>{children}</DbManagerContext.Provider>
}

export function useDbManager() {
  const ctx = React.useContext(DbManagerContext)
  if (!ctx) throw new Error("useDbManager must be used within DbManagerProvider")
  return ctx
}


