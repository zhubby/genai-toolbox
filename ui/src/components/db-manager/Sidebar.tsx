"use client"

import * as React from "react"
import { Database, Search, Plus, RefreshCw, MoreVertical, X, Check } from "lucide-react"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { cn } from "@/lib/utils"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { listSources, createSource, updateSource, deleteSource, validateSource, type SourceItem } from "@/lib/api"
import { useDbManager } from "./context"

interface SidebarProps extends React.HTMLAttributes<HTMLDivElement> {}

export function Sidebar({ className, ...props }: SidebarProps) {
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const [sources, setSources] = React.useState<SourceItem[]>([]);
  const [createOpen, setCreateOpen] = React.useState(false);
  const [creating, setCreating] = React.useState(false);
  const [formError, setFormError] = React.useState<string | null>(null);
  const [validating, setValidating] = React.useState(false);
  const [validateResult, setValidateResult] = React.useState<{ ok: boolean; message: string } | null>(null);
  const [createMode, setCreateMode] = React.useState<"create" | "edit" | "copy">("create");
  const [editTarget, setEditTarget] = React.useState<{ name: string; kind: string } | null>(null);
  const [search, setSearch] = React.useState("")
  const [form, setForm] = React.useState({
    name: "",
    kind: "postgres" as "postgres" | "mysql",
    host: "127.0.0.1",
    port: 5432,
    database: "toolbox_db",
    user: "toolbox_user",
    password: "",
  });
  const [deleteTarget, setDeleteTarget] = React.useState<string | null>(null);
  const [sourcesByName, setSourcesByName] = React.useState<Record<string, SourceItem>>({});
  const [menuOpenFor, setMenuOpenFor] = React.useState<string | null>(null);

  const { selectedSource, selectSource } = useDbManager()

  const refresh = React.useCallback(async () => {
    setLoading(true); setError(null);
    try {
      const data = await listSources();
      const map: Record<string, SourceItem> = {};
      for (const s of data) map[s.name] = s;
      setSourcesByName(map);
      setSources(data);

      // 如果之前从 localStorage 恢复了 name，但没有 kind/config，则在这里补齐
      if (selectedSource?.name) {
        const existing = map[selectedSource.name]
        if (existing && (!selectedSource.kind || !selectedSource.config || Object.keys(selectedSource.config).length === 0)) {
          selectSource({ name: existing.name, kind: existing.kind, config: existing.config })
        }
      }
    } catch (e: any) {
      setError(e?.message || String(e));
    } finally {
      setLoading(false);
    }
  }, [selectSource, selectedSource?.name, selectedSource?.kind, selectedSource?.config]);

  React.useEffect(() => { refresh(); }, [refresh]);

  React.useEffect(() => {
    if (!menuOpenFor) return;
    const onDocClick = () => setMenuOpenFor(null);
    document.addEventListener("click", onDocClick);
    return () => document.removeEventListener("click", onDocClick);
  }, [menuOpenFor]);

  const onKindChange = (k: string) => {
    const kind = (k === "mysql" ? "mysql" : "postgres") as "postgres" | "mysql";
    setForm((f) => ({ ...f, kind, port: kind === "mysql" ? 3306 : 5432 }));
  };

  const openCreate = (mode: "create" | "edit" | "copy", sourceName?: string) => {
    setFormError(null);
    setValidateResult(null);
    setCreateMode(mode);

    if (!sourceName) {
      setEditTarget(null);
      setForm({ name: "", kind: "postgres", host: "127.0.0.1", port: 5432, database: "toolbox_db", user: "toolbox_user", password: "" });
      setCreateOpen(true);
      return;
    }

    const s = sourcesByName[sourceName];
    const kind = (s?.kind === "mysql" ? "mysql" : "postgres") as "postgres" | "mysql";
    const cfg = s?.config || {};
    const host = typeof cfg.host === "string" ? cfg.host : "127.0.0.1";
    const port = typeof cfg.port === "number" ? cfg.port : (kind === "mysql" ? 3306 : 5432);
    const database = typeof cfg.database === "string" ? cfg.database : "toolbox_db";
    const user = typeof cfg.user === "string" ? cfg.user : "toolbox_user";
    const password = typeof cfg.password === "string" ? cfg.password : "";

    if (mode === "edit") {
      setEditTarget({ name: sourceName, kind: s?.kind || kind });
      setForm({ name: sourceName, kind, host, port, database, user, password });
    } else {
      setEditTarget(null);
      setForm({ name: `${sourceName}-copy`, kind, host, port, database, user, password });
    }
    setCreateOpen(true);
  };

  const handleCreate = async () => {
    setFormError(null);
    setValidateResult(null);
    if (createMode !== "edit" && !form.name.trim()) { setFormError("Name 必填"); return; }
    if (!form.host.trim()) { setFormError("Host 必填"); return; }
    if (!form.database.trim()) { setFormError("Database 必填"); return; }
    if (!form.user.trim()) { setFormError("User 必填"); return; }
    if (!Number.isFinite(form.port) || form.port <= 0) { setFormError("Port 非法"); return; }

    try {
      setCreating(true);
      const cfg = {
        host: form.host.trim(),
        port: Number(form.port),
        database: form.database.trim(),
        user: form.user.trim(),
        password: form.password,
      };
      if (createMode === "edit" && editTarget) {
        await updateSource(editTarget.name, { kind: editTarget.kind, config: cfg });
      } else {
        await createSource({
          name: form.name.trim(),
          kind: form.kind,
          config: cfg,
        });
      }
      await refresh();
      setCreateOpen(false);
      setFormError(null);
      setForm({ name: "", kind: "postgres", host: "127.0.0.1", port: 5432, database: "toolbox_db", user: "toolbox_user", password: "" });
    } catch (e: any) {
      setFormError(e?.message || String(e));
    } finally {
      setCreating(false);
    }
  };

  const handleValidate = async () => {
    setFormError(null);
    setValidateResult(null);
    if (!form.kind) { setFormError("Kind 必填"); return; }
    if (!form.host.trim()) { setFormError("Host 必填"); return; }
    if (!form.database.trim()) { setFormError("Database 必填"); return; }
    if (!form.user.trim()) { setFormError("User 必填"); return; }
    if (!Number.isFinite(form.port) || form.port <= 0) { setFormError("Port 非法"); return; }

    try {
      setValidating(true);
      const res = await validateSource({
        kind: form.kind,
        config: {
          host: form.host.trim(),
          port: Number(form.port),
          database: form.database.trim(),
          user: form.user.trim(),
          password: form.password,
        },
      });
      if (res.available) {
        setValidateResult({ ok: true, message: "连接可用" });
      } else {
        setValidateResult({ ok: false, message: res.error || "连接不可用" });
      }
    } catch (e: any) {
      setValidateResult({ ok: false, message: e?.message || String(e) });
    } finally {
      setValidating(false);
    }
  };

  return (
    <div className={cn("flex flex-col h-full border-r bg-background min-w-0", className)} {...props}>
      <div className="p-4 space-y-4">
        <div className="flex items-center justify-between">
           <h2 className="text-sm font-semibold tracking-tight uppercase text-muted-foreground">Explorer</h2>
           <div className="flex gap-1">
             <Button variant="ghost" size="icon" className="h-6 w-6" title="New Source" onClick={() => openCreate("create")}>
               <Plus className="h-4 w-4" />
             </Button>
             <Button variant="ghost" size="icon" className="h-6 w-6" title="Refresh" onClick={refresh}>
               <RefreshCw className="h-4 w-4" />
             </Button>
           </div>
        </div>
        <div className="relative">
          <Search className="absolute left-2 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
          <Input
            placeholder="Search source..."
            className="pl-8 h-9 text-sm"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        {error && <div className="text-xs text-red-500">{error}</div>}
        {loading && <div className="text-xs text-muted-foreground">Loading...</div>}
      </div>
      <Separator />
      <ScrollArea className="flex-1">
        <div className="p-2">
          {sources
            .filter((s) => (search.trim() ? s.name.toLowerCase().includes(search.trim().toLowerCase()) : true))
            .sort((a, b) => a.name.localeCompare(b.name))
            .map((s) => {
              const isSelected = selectedSource?.name === s.name
              return (
                <div
                  key={s.name}
                  className={cn(
                    "flex items-center gap-2 px-2 py-1.5 rounded-sm cursor-pointer group",
                    isSelected ? "bg-muted" : "hover:bg-muted/50",
                  )}
                  onClick={() => selectSource({ name: s.name, kind: s.kind, config: s.config })}
                >
                  <Database className="w-4 h-4 text-blue-500" />
                  <span className="text-sm font-medium truncate flex-1">{s.name}</span>
                  {isSelected && <Check className="w-4 h-4 text-green-600 shrink-0" />}
                  <div className="relative">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6"
                      onClick={(e) => {
                        e.stopPropagation()
                        setMenuOpenFor((cur) => (cur === s.name ? null : s.name))
                      }}
                      title="操作"
                    >
                      <MoreVertical className="w-3 h-3" />
                    </Button>
                    {menuOpenFor === s.name && (
                      <div
                        className="absolute right-0 top-7 z-50 w-28 rounded-md border bg-background shadow-md"
                        onClick={(e) => e.stopPropagation()}
                      >
                        <button
                          className="w-full text-left px-3 py-2 text-sm hover:bg-muted"
                          onClick={() => {
                            setMenuOpenFor(null)
                            openCreate("edit", s.name)
                          }}
                        >
                          修改
                        </button>
                        <button
                          className="w-full text-left px-3 py-2 text-sm hover:bg-muted"
                          onClick={() => {
                            setMenuOpenFor(null)
                            openCreate("copy", s.name)
                          }}
                        >
                          复制
                        </button>
                        <button
                          className="w-full text-left px-3 py-2 text-sm text-red-600 hover:bg-muted"
                          onClick={() => {
                            setMenuOpenFor(null)
                            setDeleteTarget(s.name)
                          }}
                        >
                          删除
                        </button>
                      </div>
                    )}
                  </div>
                </div>
              )
            })}
          {sources.length === 0 && !loading && !error && (
            <div className="text-xs text-muted-foreground px-2 py-2 italic">No sources</div>
          )}
        </div>
      </ScrollArea>
      <Separator />
      <div className="p-2 bg-muted/20">
          <Button
            variant="outline"
            size="sm"
            className="w-full justify-start text-muted-foreground font-normal"
            onClick={() => openCreate("create")}
          >
              <Plus className="mr-2 h-4 w-4" />
              New Connection
          </Button>
      </div>

      {createOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="absolute inset-0 bg-black/40" onClick={() => !creating && setCreateOpen(false)} />
          <div className="relative z-10 w-[520px] max-w-[92vw] rounded-md border bg-background shadow-lg">
            <div className="flex items-center justify-between px-4 py-3 border-b">
              <div className="font-medium">
                {createMode === "edit" ? "修改 Source" : createMode === "copy" ? "复制 Source" : "创建 Source"}
              </div>
              <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => !creating && setCreateOpen(false)}>
                <X className="h-4 w-4" />
              </Button>
            </div>
            <div className="p-4 space-y-3">
              {formError && <div className="text-sm text-red-600">{formError}</div>}
              {validateResult && (
                <div className={cn("text-sm", validateResult.ok ? "text-green-600" : "text-red-600")}>
                  {validateResult.message}
                </div>
              )}
              <div className="grid grid-cols-3 gap-3 items-center">
                <label className="text-sm text-muted-foreground">Name</label>
                <div className="col-span-2">
                  <Input
                    value={form.name}
                    onChange={(e) => setForm({ ...form, name: e.target.value })}
                    placeholder="e.g. prod-db-01"
                    disabled={createMode === "edit"}
                  />
                </div>

                <label className="text-sm text-muted-foreground">Kind</label>
                <div className="col-span-2">
                  <Select value={form.kind} onValueChange={onKindChange}>
                    <SelectTrigger className="h-9" disabled={createMode === "edit"}>
                      <SelectValue placeholder="Select kind" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="postgres">PostgreSQL</SelectItem>
                      <SelectItem value="mysql">MySQL</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <label className="text-sm text-muted-foreground">Host</label>
                <div className="col-span-2"><Input value={form.host} onChange={(e) => setForm({ ...form, host: e.target.value })} placeholder="127.0.0.1" /></div>

                <label className="text-sm text-muted-foreground">Port</label>
                <div className="col-span-2"><Input type="number" value={form.port} onChange={(e) => setForm({ ...form, port: Number(e.target.value) })} /></div>

                <label className="text-sm text-muted-foreground">Database</label>
                <div className="col-span-2"><Input value={form.database} onChange={(e) => setForm({ ...form, database: e.target.value })} placeholder="toolbox_db" /></div>

                <label className="text-sm text-muted-foreground">User</label>
                <div className="col-span-2"><Input value={form.user} onChange={(e) => setForm({ ...form, user: e.target.value })} placeholder="toolbox_user" /></div>

                <label className="text-sm text-muted-foreground">Password</label>
                <div className="col-span-2"><Input type="password" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} placeholder="my-password" /></div>
              </div>
            </div>
            <div className="px-4 py-3 border-t flex items-center justify-end gap-2">
              <Button variant="outline" onClick={() => setCreateOpen(false)} disabled={creating || validating}>取消</Button>
              <Button variant="secondary" onClick={handleValidate} disabled={creating || validating}>
                {validating ? "验证中..." : "验证连接"}
              </Button>
              <Button onClick={handleCreate} disabled={creating || validating}>
                {creating ? (createMode === "edit" ? "保存中..." : "创建中...") : (createMode === "edit" ? "保存" : "创建")}
              </Button>
            </div>
          </div>
        </div>
      )}

      {deleteTarget && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="absolute inset-0 bg-black/40" onClick={() => setDeleteTarget(null)} />
          <div className="relative z-10 w-[420px] max-w-[92vw] rounded-md border bg-background shadow-lg">
            <div className="flex items-center justify-between px-4 py-3 border-b">
              <div className="font-medium">删除 Source</div>
              <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setDeleteTarget(null)}>
                <X className="h-4 w-4" />
              </Button>
            </div>
            <div className="p-4 text-sm">
              确认删除 <span className="font-mono font-medium">{deleteTarget}</span> ？该操作不可恢复。
            </div>
            <div className="px-4 py-3 border-t flex items-center justify-end gap-2">
              <Button variant="outline" onClick={() => setDeleteTarget(null)}>取消</Button>
              <Button variant="destructive" onClick={async () => {
                try {
                  await deleteSource(deleteTarget);
                  if (selectedSource?.name === deleteTarget) {
                    selectSource(null)
                  }
                  setDeleteTarget(null);
                  await refresh();
                } catch (e: any) {
                  alert(`Delete failed: ${e?.message || e}`);
                }
              }}>删除</Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
