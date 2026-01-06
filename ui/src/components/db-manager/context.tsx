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

  React.useEffect(() => {
    let cancelled = false
    async function run() {
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
        if (!cancelled) setTools(filtered)
      } catch (e: any) {
        if (!cancelled) setToolsError(e?.message || String(e))
      } finally {
        if (!cancelled) setToolsLoading(false)
      }
    }
    run()
    return () => {
      cancelled = true
    }
  }, [selectedSource?.name])

  const value = React.useMemo<DbManagerState>(
    () => ({
      selectedSource,
      selectSource: setSelectedSource,
      tools,
      toolsLoading,
      toolsError,
    }),
    [selectedSource, tools, toolsLoading, toolsError],
  )

  return <DbManagerContext.Provider value={value}>{children}</DbManagerContext.Provider>
}

export function useDbManager() {
  const ctx = React.useContext(DbManagerContext)
  if (!ctx) throw new Error("useDbManager must be used within DbManagerProvider")
  return ctx
}


