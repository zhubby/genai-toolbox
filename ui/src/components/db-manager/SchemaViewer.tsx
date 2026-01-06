"use client"

import * as React from "react"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Badge } from "@/components/ui/badge"
import { useDbManager } from "./context"

export function SchemaViewer() {
  const { selectedSource, tools, toolsLoading, toolsError } = useDbManager()

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
            return (
              <div
                key={t.name}
                className="rounded-md border bg-background px-3 py-2 hover:bg-muted/30 transition-colors"
              >
                <div className="text-sm font-medium truncate">
                  {desc ? (
                    <>
                      {desc} <span className="text-muted-foreground font-normal">({t.name})</span>
                    </>
                  ) : (
                    t.name
                  )}
                </div>
                <div className="text-xs text-muted-foreground mt-1 flex items-center gap-2">
                  <span className="font-mono">{t.kind}</span>
                </div>
              </div>
            )
          })}
        </div>
      </ScrollArea>
    </div>
  )
}
