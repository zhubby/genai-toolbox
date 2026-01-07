"use client"

import * as React from "react"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Badge } from "@/components/ui/badge"
import { useDbManager } from "./context"

export function SchemaViewer() {
  const { selectedSource, tools, toolsLoading, toolsError, selectedToolName, selectTool } = useDbManager()

  return (
    <div className="flex flex-col h-full bg-muted/5 min-w-0">
      <div className="p-3 border-b font-medium text-sm flex items-center justify-between">
        <span className="truncate">
          Tools{selectedSource?.name ? `: ${selectedSource.name}` : ""}
        </span>
        {selectedSource?.kind ? <Badge variant="secondary" className="text-[10px] uppercase">{selectedSource.kind}</Badge> : null}
      </div>
      <ScrollArea className="flex-1">
        <div className="p-3 space-y-2">
          {!selectedSource?.name && (
            <div className="text-sm text-muted-foreground">请选择左侧一个 source</div>
          )}
          {selectedSource?.name && toolsLoading && (
            <div className="text-sm text-muted-foreground">Loading tools...</div>
          )}
          {toolsError && (
            <div className="text-sm text-red-600">{toolsError}</div>
          )}
          {selectedSource?.name && !toolsLoading && !toolsError && tools.length === 0 && (
            <div className="text-sm text-muted-foreground">该 source 暂无 tools</div>
          )}
          {tools.map((t) => {
            const desc = typeof t.config?.description === "string" ? t.config.description.trim() : ""
            const paramsRaw = Array.isArray(t.config?.parameters) ? t.config.parameters : []
            const paramNames = paramsRaw
              .map((p: any) => (typeof p?.name === "string" ? p.name.trim() : ""))
              .filter(Boolean)
            const isSelected = selectedToolName === t.name
            return (
              <div
                key={t.name}
                className={`rounded-md border bg-background px-3 py-2 hover:bg-muted/30 transition-colors cursor-pointer ${
                  isSelected ? "ring-1 ring-primary/40" : ""
                }`}
                onClick={() => selectTool(t.name)}
              >
                <div className="text-sm font-medium truncate">
                  {t.name} <span className="text-muted-foreground font-normal font-mono">({t.kind})</span>
                </div>
                {desc ? <div className="text-xs text-muted-foreground mt-1">{desc}</div> : null}
                {paramNames.length ? (
                  <div className="text-[11px] text-muted-foreground mt-1 font-mono truncate">
                    params: {paramNames.slice(0, 3).join(", ")}
                    {paramNames.length > 3 ? ` (+${paramNames.length - 3})` : ""}
                  </div>
                ) : null}
              </div>
            )
          })}
        </div>
      </ScrollArea>
    </div>
  )
}
