"use client"

import * as React from "react"
import { X, Plus, RefreshCw, Trash2, Pencil } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Separator } from "@/components/ui/separator"
import { cn } from "@/lib/utils"
import {
  createToolset,
  deleteToolset,
  listToolsets,
  listTools,
  updateToolset,
  type ToolItem,
  type ToolsetItem,
} from "@/lib/api"
import { CheckboxListPicker, type CheckboxListItem } from "./CheckboxListPicker"

type Mode = "list" | "create" | "edit"

export function ToolsetsDialog(props: { open: boolean; onOpenChange: (open: boolean) => void }) {
  const { open, onOpenChange } = props

  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [toolsets, setToolsets] = React.useState<ToolsetItem[]>([])

  const [toolsLoading, setToolsLoading] = React.useState(false)
  const [toolsError, setToolsError] = React.useState<string | null>(null)
  const [tools, setTools] = React.useState<ToolItem[]>([])

  const [mode, setMode] = React.useState<Mode>("list")
  const [formError, setFormError] = React.useState<string | null>(null)
  const [saving, setSaving] = React.useState(false)
  const [name, setName] = React.useState("")
  const [selectedTools, setSelectedTools] = React.useState<string[]>([])
  const [editTarget, setEditTarget] = React.useState<string | null>(null)

  const [deleteTarget, setDeleteTarget] = React.useState<string | null>(null)

  const close = () => onOpenChange(false)

  const resetForm = React.useCallback(() => {
    setFormError(null)
    setSaving(false)
    setName("")
    setSelectedTools([])
    setEditTarget(null)
  }, [])

  const refreshToolsets = React.useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await listToolsets()
      setToolsets(data)
    } catch (e: any) {
      setError(e?.message || String(e))
    } finally {
      setLoading(false)
    }
  }, [])

  const refreshTools = React.useCallback(async () => {
    setToolsLoading(true)
    setToolsError(null)
    try {
      const data = await listTools()
      setTools(data)
    } catch (e: any) {
      setToolsError(e?.message || String(e))
    } finally {
      setToolsLoading(false)
    }
  }, [])

  React.useEffect(() => {
    if (!open) return
    void refreshToolsets()
    void refreshTools()
    setMode("list")
    resetForm()
    setDeleteTarget(null)
  }, [open, refreshToolsets, refreshTools, resetForm])

  React.useEffect(() => {
    if (!open) return
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") close()
    }
    document.addEventListener("keydown", onKeyDown)
    return () => document.removeEventListener("keydown", onKeyDown)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  const toolItems: CheckboxListItem[] = React.useMemo(() => {
    const sorted = [...tools].sort((a, b) => a.name.localeCompare(b.name))
    return sorted.map((t) => ({
      id: t.name,
      label: t.name,
      meta: t.kind ? `kind: ${t.kind}` : "",
    }))
  }, [tools])

  const openCreate = () => {
    resetForm()
    setMode("create")
  }

  const openEdit = (ts: ToolsetItem) => {
    resetForm()
    setMode("edit")
    setEditTarget(ts.name)
    setName(ts.name)
    setSelectedTools(Array.isArray(ts.toolNames) ? ts.toolNames : [])
  }

  const backToList = () => {
    setMode("list")
    resetForm()
  }

  const handleSave = async () => {
    setFormError(null)
    if (!name.trim()) {
      setFormError("Name 必填")
      return
    }
    if (!Array.isArray(selectedTools)) {
      setFormError("ToolNames 非法")
      return
    }
    setSaving(true)
    try {
      const payload = {
        name: name.trim(),
        toolNames: selectedTools.slice().sort(),
      }
      if (mode === "edit" && editTarget) {
        await updateToolset(editTarget, { toolNames: payload.toolNames })
      } else {
        await createToolset(payload)
      }
      await refreshToolsets()
      backToList()
    } catch (e: any) {
      setFormError(e?.message || String(e))
    } finally {
      setSaving(false)
    }
  }

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/40" onClick={() => !saving && close()} />
      <div className="relative z-10 w-[860px] max-w-[92vw] max-h-[88vh] rounded-md border bg-background shadow-lg overflow-hidden flex flex-col">
        <div className="flex items-center justify-between px-4 py-3 border-b">
          <div className="font-medium">工具集合管理</div>
          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => !saving && close()}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        {mode === "list" ? (
          <>
            <div className="px-4 py-3 flex items-center justify-between gap-2">
              <div className="text-sm text-muted-foreground">Toolsets（共 {toolsets.length} 个）</div>
              <div className="flex items-center gap-2">
                <Button variant="outline" size="sm" onClick={refreshToolsets} disabled={loading}>
                  <RefreshCw className={cn("h-4 w-4 mr-2", loading ? "animate-spin" : "")} />
                  刷新
                </Button>
                <Button size="sm" onClick={openCreate}>
                  <Plus className="h-4 w-4 mr-2" />
                  新建
                </Button>
              </div>
            </div>
            <Separator />
            <div className="flex-1 overflow-auto">
              <div className="p-4 space-y-3">
                {error && <div className="text-sm text-red-600">{error}</div>}
                {loading && <div className="text-sm text-muted-foreground">Loading...</div>}
                {!loading && !error && toolsets.length === 0 && (
                  <div className="text-sm text-muted-foreground italic">暂无 toolsets</div>
                )}
                <div className="space-y-2">
                  {toolsets
                    .slice()
                    .sort((a, b) => a.name.localeCompare(b.name))
                    .map((ts) => (
                      <div key={ts.name} className="rounded-md border bg-background px-3 py-2 flex items-start gap-3">
                        <div className="min-w-0 flex-1">
                          <div className="text-sm font-medium truncate">{ts.name}</div>
                          <div className="text-xs text-muted-foreground mt-1">
                            tools: {(ts.toolNames || []).length}
                            {ts.updatedAt ? ` · updated: ${ts.updatedAt}` : ""}
                          </div>
                          {(ts.toolNames || []).length ? (
                            <div className="text-[11px] text-muted-foreground mt-1 font-mono truncate">
                              {(ts.toolNames || []).slice(0, 6).join(", ")}
                              {(ts.toolNames || []).length > 6 ? ` (+${(ts.toolNames || []).length - 6})` : ""}
                            </div>
                          ) : null}
                        </div>
                        <div className="flex items-center gap-1">
                          <Button variant="ghost" size="icon" className="h-8 w-8" title="编辑" onClick={() => openEdit(ts)}>
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 text-red-600 hover:text-red-700"
                            title="删除"
                            onClick={() => setDeleteTarget(ts.name)}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </div>
                    ))}
                </div>

                {toolsError ? (
                  <div className="text-xs text-red-600">加载 tools 失败（成员选择将不可用）: {toolsError}</div>
                ) : toolsLoading ? (
                  <div className="text-xs text-muted-foreground">Loading tools...</div>
                ) : null}
              </div>
            </div>
          </>
        ) : (
          <>
            <div className="px-4 py-3 flex items-center justify-between gap-2">
              <div className="text-sm text-muted-foreground">{mode === "edit" ? "编辑 Toolset" : "新建 Toolset"}</div>
              <Button variant="outline" size="sm" onClick={backToList} disabled={saving}>
                返回列表
              </Button>
            </div>
            <Separator />
            <div className="flex-1 overflow-auto">
              <div className="p-4 space-y-4">
                {formError && <div className="text-sm text-red-600">{formError}</div>}
                <div className="grid grid-cols-3 gap-3 items-center">
                  <label className="text-sm text-muted-foreground">Name</label>
                  <div className="col-span-2">
                    <Input value={name} onChange={(e) => setName(e.target.value)} disabled={mode === "edit"} />
                  </div>
                </div>
                <CheckboxListPicker
                  title="选择 tools"
                  items={toolItems}
                  selectedIds={selectedTools}
                  onChange={(next) => setSelectedTools(next)}
                  placeholder="搜索 tool..."
                />
              </div>
            </div>
            <div className="px-4 py-3 border-t flex items-center justify-end gap-2">
              <Button variant="outline" onClick={backToList} disabled={saving}>
                取消
              </Button>
              <Button onClick={handleSave} disabled={saving}>
                {saving ? "保存中..." : "保存"}
              </Button>
            </div>
          </>
        )}
      </div>

      {deleteTarget && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center">
          <div className="absolute inset-0 bg-black/40" onClick={() => setDeleteTarget(null)} />
          <div className="relative z-10 w-[420px] max-w-[92vw] rounded-md border bg-background shadow-lg">
            <div className="flex items-center justify-between px-4 py-3 border-b">
              <div className="font-medium">删除 Toolset</div>
              <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setDeleteTarget(null)}>
                <X className="h-4 w-4" />
              </Button>
            </div>
            <div className="p-4 text-sm">
              确认删除 <span className="font-mono font-medium">{deleteTarget}</span> ？该操作不可恢复。
            </div>
            <div className="px-4 py-3 border-t flex items-center justify-end gap-2">
              <Button variant="outline" onClick={() => setDeleteTarget(null)} disabled={saving}>
                取消
              </Button>
              <Button
                variant="destructive"
                onClick={async () => {
                  const name = deleteTarget
                  try {
                    setSaving(true)
                    await deleteToolset(name)
                    setDeleteTarget(null)
                    await refreshToolsets()
                  } catch (e: any) {
                    alert(`Delete failed: ${e?.message || e}`)
                  } finally {
                    setSaving(false)
                  }
                }}
                disabled={saving}
              >
                删除
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}


