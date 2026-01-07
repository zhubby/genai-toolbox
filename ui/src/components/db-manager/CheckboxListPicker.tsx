"use client"

import * as React from "react"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"

export type CheckboxListItem = {
  id: string
  label: string
  meta?: string
}

export function CheckboxListPicker(props: {
  title: string
  items: CheckboxListItem[]
  selectedIds: string[]
  onChange: (next: string[]) => void
  className?: string
  placeholder?: string
}) {
  const { title, items, selectedIds, onChange, className, placeholder } = props
  const [search, setSearch] = React.useState("")

  const selected = React.useMemo(() => new Set(selectedIds), [selectedIds])

  const filtered = React.useMemo(() => {
    const q = search.trim().toLowerCase()
    if (!q) return items
    return items.filter((it) => it.label.toLowerCase().includes(q) || (it.meta || "").toLowerCase().includes(q))
  }, [items, search])

  const toggle = (id: string) => {
    const next = new Set(selectedIds)
    if (next.has(id)) next.delete(id)
    else next.add(id)
    onChange(Array.from(next))
  }

  const selectAllFiltered = () => {
    const next = new Set(selectedIds)
    for (const it of filtered) next.add(it.id)
    onChange(Array.from(next))
  }

  const clearAll = () => onChange([])

  return (
    <div className={cn("space-y-2", className)}>
      <div className="flex items-center justify-between gap-2">
        <div className="text-sm font-medium">
          {title} <span className="text-xs text-muted-foreground">({selectedIds.length})</span>
        </div>
        <div className="flex items-center gap-2 text-xs">
          <button className="text-muted-foreground hover:text-foreground" type="button" onClick={selectAllFiltered}>
            全选当前
          </button>
          <button className="text-muted-foreground hover:text-foreground" type="button" onClick={clearAll}>
            清空
          </button>
        </div>
      </div>
      <Input
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder={placeholder || "搜索..."}
        className="h-9 text-sm"
      />
      <div className="rounded-md border bg-background max-h-[260px] overflow-auto">
        {filtered.length === 0 ? (
          <div className="px-3 py-3 text-sm text-muted-foreground">无匹配项</div>
        ) : (
          <div className="divide-y">
            {filtered.map((it) => {
              const checked = selected.has(it.id)
              return (
                <label key={it.id} className="flex items-start gap-2 px-3 py-2 hover:bg-muted/30 cursor-pointer">
                  <input
                    type="checkbox"
                    className="mt-0.5"
                    checked={checked}
                    onChange={() => toggle(it.id)}
                  />
                  <div className="min-w-0">
                    <div className="text-sm font-medium truncate">{it.label}</div>
                    {it.meta ? <div className="text-xs text-muted-foreground truncate">{it.meta}</div> : null}
                  </div>
                </label>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}


