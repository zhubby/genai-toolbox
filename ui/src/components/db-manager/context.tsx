"use client"

import * as React from "react"
import { listTools, type SourceItem, type ToolItem } from "@/lib/api"

export type SelectedSource = Pick<SourceItem, "name" | "kind" | "config">

type DbManagerState = {
  selectedSource: SelectedSource | null
  selectSource: (source: SelectedSource | null) => void
  tools: ToolItem[]
  toolsLoading: boolean
  toolsError: string | null
  reloadTools: () => Promise<void>
}

const DbManagerContext = React.createContext<DbManagerState | null>(null)

const LS_KEY = "tb.dbManager.selectedSourceName"

export function DbManagerProvider({ children }: { children: React.ReactNode }) {
  const [selectedSource, setSelectedSource] = React.useState<SelectedSource | null>(null)
  const [tools, setTools] = React.useState<ToolItem[]>([])
  const [toolsLoading, setToolsLoading] = React.useState(false)
  const [toolsError, setToolsError] = React.useState<string | null>(null)

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

  const value = React.useMemo<DbManagerState>(
    () => ({
      selectedSource,
      selectSource: setSelectedSource,
      tools,
      toolsLoading,
      toolsError,
      reloadTools,
    }),
    [selectedSource, tools, toolsLoading, toolsError, reloadTools],
  )

  return <DbManagerContext.Provider value={value}>{children}</DbManagerContext.Provider>
}

export function useDbManager() {
  const ctx = React.useContext(DbManagerContext)
  if (!ctx) throw new Error("useDbManager must be used within DbManagerProvider")
  return ctx
}


