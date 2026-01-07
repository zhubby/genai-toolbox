"use client"

import * as React from "react"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area"
import { Button } from "@/components/ui/button"
import { Clock, CheckCircle2, ChevronLeft, ChevronRight, Download, Filter, AlertCircle, Loader2 } from "lucide-react"
import { useDbManager } from "./context"

export function ResultTable() {
  const { preview } = useDbManager()
  const columns = preview.data?.columns ?? []
  const rows = preview.data?.rows ?? []
  const rowCount = preview.data?.rowCount ?? 0
  const durationMs = preview.data?.durationMs

  const statusNode = React.useMemo(() => {
    switch (preview.status) {
      case "running":
        return (
          <span className="flex items-center gap-1.5 text-muted-foreground font-medium">
            <Loader2 className="w-3.5 h-3.5 animate-spin" />
            Running
          </span>
        )
      case "error":
        return (
          <span className="flex items-center gap-1.5 text-red-600 font-medium">
            <AlertCircle className="w-3.5 h-3.5" />
            Error
          </span>
        )
      case "success":
        return (
          <span className="flex items-center gap-1.5 text-green-600 font-medium">
            <CheckCircle2 className="w-3.5 h-3.5" />
            Success
          </span>
        )
      default:
        return <span className="text-muted-foreground">Preview</span>
    }
  }, [preview.status])

  return (
    <div className="flex flex-col h-full w-full max-w-full min-w-0 overflow-hidden bg-background">
      <div className="flex items-center justify-between px-4 py-2 border-b bg-muted/10 h-10 shrink-0">
         <div className="flex items-center gap-4 text-xs text-muted-foreground">
            {statusNode}
            {typeof durationMs === "number" ? (
              <span className="flex items-center gap-1.5">
                <Clock className="w-3.5 h-3.5" />
                {durationMs}ms
              </span>
            ) : null}
            <span className="text-foreground font-medium">{rowCount} rows</span>
         </div>
         <div className="flex items-center gap-2">
            <Button variant="ghost" size="icon" className="h-6 w-6 text-muted-foreground">
                <Filter className="h-3.5 w-3.5" />
            </Button>
            <Button variant="ghost" size="icon" className="h-6 w-6 text-muted-foreground">
                <Download className="h-3.5 w-3.5" />
            </Button>
         </div>
      </div>
      <ScrollArea className="flex-1 whitespace-nowrap">
        <Table>
          <TableHeader className="bg-muted/30 sticky top-0 z-10">
            {columns.length ? (
              <TableRow className="hover:bg-transparent border-b-muted/50">
                {columns.map((c) => (
                  <TableHead key={c} className="h-9">
                    {c}
                  </TableHead>
                ))}
              </TableRow>
            ) : null}
          </TableHeader>
          <TableBody>
            {rows.map((row, idx) => (
              <TableRow key={idx} className="border-b-muted/40 hover:bg-muted/20">
                {columns.map((_, cidx) => (
                  <TableCell key={cidx} className="py-2 text-xs font-mono text-muted-foreground">
                    {row?.[cidx] === null || row?.[cidx] === undefined ? "" : String(row[cidx])}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
        <ScrollBar orientation="horizontal" />
      </ScrollArea>
      <div className="border-t p-2 flex items-center justify-between bg-muted/5 text-xs text-muted-foreground shrink-0">
          <div>Page 1 of 5</div>
          <div className="flex items-center gap-1">
             <Button variant="outline" size="icon" className="h-7 w-7" disabled>
                 <ChevronLeft className="h-3.5 w-3.5" />
             </Button>
             <Button variant="outline" size="icon" className="h-7 w-7">
                 <ChevronRight className="h-3.5 w-3.5" />
             </Button>
          </div>
      </div>
    </div>
  )
}
