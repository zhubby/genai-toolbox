"use client"

import * as React from "react"
import { Database, Search, Server, Table2, Plus, RefreshCw, ChevronRight, ChevronDown, MoreVertical, X } from "lucide-react"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { cn } from "@/lib/utils"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { listSources, createSource, deleteSource, type SourceItem } from "@/lib/api"

interface SidebarProps extends React.HTMLAttributes<HTMLDivElement> {}

type Instance = {
  id: string
  name: string
  type: string
  status: "online" | "offline"
  databases: { name: string; tables: string[] }[]
}

// 将后端 /sources 数据映射为简单的分组视图：按 kind 分组展示，后续可扩展到真实 DB 枚举。
function mapSourcesToInstances(sources: SourceItem[]): Instance[] {
  const groups = new Map<string, SourceItem[]>();
  for (const s of sources) {
    const k = s.kind || "unknown";
    if (!groups.has(k)) groups.set(k, []);
    groups.get(k)!.push(s);
  }
  const out: Instance[] = [];
  for (const [kind, list] of groups) {
    out.push({
      id: kind,
      name: kind,
      type: kind,
      status: "online",
      databases: list.map((s) => ({ name: s.name, tables: [] })),
    });
  }
  return out;
}

export function Sidebar({ className, ...props }: SidebarProps) {
  const [expanded, setExpanded] = React.useState<Record<string, boolean>>({});
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const [instances, setInstances] = React.useState<Instance[]>([]);
  const [createOpen, setCreateOpen] = React.useState(false);
  const [creating, setCreating] = React.useState(false);
  const [formError, setFormError] = React.useState<string | null>(null);
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

  const toggleExpand = (id: string) => {
    setExpanded(prev => ({ ...prev, [id]: !prev[id] }));
  };

  const refresh = React.useCallback(async () => {
    setLoading(true); setError(null);
    try {
      const data = await listSources();
      setInstances(mapSourcesToInstances(data));
    } catch (e: any) {
      setError(e?.message || String(e));
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => { refresh(); }, [refresh]);

  const onKindChange = (k: string) => {
    const kind = (k === "mysql" ? "mysql" : "postgres") as "postgres" | "mysql";
    setForm((f) => ({ ...f, kind, port: kind === "mysql" ? 3306 : 5432 }));
  };

  const handleCreate = async () => {
    setFormError(null);
    if (!form.name.trim()) { setFormError("Name 必填"); return; }
    if (!form.host.trim()) { setFormError("Host 必填"); return; }
    if (!form.database.trim()) { setFormError("Database 必填"); return; }
    if (!form.user.trim()) { setFormError("User 必填"); return; }
    if (!Number.isFinite(form.port) || form.port <= 0) { setFormError("Port 非法"); return; }

    try {
      setCreating(true);
      await createSource({
        name: form.name.trim(),
        kind: form.kind,
        config: {
          host: form.host.trim(),
          port: Number(form.port),
          database: form.database.trim(),
          user: form.user.trim(),
          password: form.password,
        },
      });
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

  return (
    <div className={cn("flex flex-col h-full border-r bg-background min-w-0", className)} {...props}>
      <div className="p-4 space-y-4">
        <div className="flex items-center justify-between">
           <h2 className="text-sm font-semibold tracking-tight uppercase text-muted-foreground">Explorer</h2>
           <div className="flex gap-1">
             <Button variant="ghost" size="icon" className="h-6 w-6" title="New Source" onClick={() => setCreateOpen(true)}>
               <Plus className="h-4 w-4" />
             </Button>
             <Button variant="ghost" size="icon" className="h-6 w-6" title="Refresh" onClick={refresh}>
               <RefreshCw className="h-4 w-4" />
             </Button>
           </div>
        </div>
        <div className="relative">
          <Search className="absolute left-2 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
          <Input placeholder="Search..." className="pl-8 h-9 text-sm" />
        </div>
        {error && <div className="text-xs text-red-500">{error}</div>}
        {loading && <div className="text-xs text-muted-foreground">Loading...</div>}
      </div>
      <Separator />
      <ScrollArea className="flex-1">
        <div className="p-2">
          {instances.map((instance) => (
            <div key={instance.id} className="mb-1">
              <div 
                className="flex items-center gap-2 px-2 py-1.5 hover:bg-muted/50 rounded-sm cursor-pointer group"
                onClick={() => toggleExpand(instance.id)}
              >
                <div className="text-muted-foreground">
                    {expanded[instance.id] ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                </div>
                <Server className="w-4 h-4 text-muted-foreground" />
                <span className="text-sm font-medium truncate flex-1">{instance.name}</span>
                <div className={cn("w-2 h-2 rounded-full shrink-0", instance.status === 'online' ? "bg-green-500" : "bg-red-500")} />
              </div>
              
              {expanded[instance.id] && (
                <div className="ml-4 pl-2 border-l border-muted/50 mt-1 space-y-1">
                    {instance.databases.map((db) => {
                        const dbId = `${instance.id}-${db.name}`;
                        return (
                          <div key={db.name}>
                             <div 
                                className="flex items-center gap-2 px-2 py-1 hover:bg-muted/50 rounded-sm cursor-pointer"
                                onClick={() => toggleExpand(dbId)}
                             >
                                <div className="text-muted-foreground">
                                    {expanded[dbId] ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                                </div>
                                <Database className="w-3.5 h-3.5 text-blue-500" />
                                <span className="text-sm truncate flex-1">{db.name}</span>
                                <Button variant="ghost" size="icon" className="h-6 w-6" onClick={(e) => { e.stopPropagation(); setDeleteTarget(db.name); }} title="删除">
                                  <MoreVertical className="w-3 h-3" />
                                </Button>
                             </div>
                             
                             {expanded[dbId] && (
                                 <div className="ml-4 pl-2 border-l border-muted/50 mt-1 space-y-0.5">
                                    {db.tables.map((table) => (
                                         <div key={table} className="flex items-center gap-2 px-2 py-1 text-muted-foreground hover:text-foreground hover:bg-muted/50 rounded-sm cursor-pointer transition-colors group">
                                            <Table2 className="w-3.5 h-3.5" />
                                            <span className="text-sm truncate flex-1">{table}</span>
                                            <MoreVertical className="w-3 h-3 opacity-0 group-hover:opacity-100" />
                                         </div>
                                    ))}
                                 </div>
                             )}
                          </div>
                        )
                    })}
                    {instance.databases.length === 0 && (
                        <div className="text-xs text-muted-foreground pl-4 py-1 italic">No databases</div>
                    )}
                </div>
              )}
            </div>
          ))}
        </div>
      </ScrollArea>
      <Separator />
      <div className="p-2 bg-muted/20">
          <Button variant="outline" size="sm" className="w-full justify-start text-muted-foreground font-normal">
              <Plus className="mr-2 h-4 w-4" />
              New Connection
          </Button>
      </div>

      {createOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="absolute inset-0 bg-black/40" onClick={() => !creating && setCreateOpen(false)} />
          <div className="relative z-10 w-[520px] max-w-[92vw] rounded-md border bg-background shadow-lg">
            <div className="flex items-center justify-between px-4 py-3 border-b">
              <div className="font-medium">创建 Source</div>
              <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => !creating && setCreateOpen(false)}>
                <X className="h-4 w-4" />
              </Button>
            </div>
            <div className="p-4 space-y-3">
              {formError && <div className="text-sm text-red-600">{formError}</div>}
              <div className="grid grid-cols-3 gap-3 items-center">
                <label className="text-sm text-muted-foreground">Name</label>
                <div className="col-span-2"><Input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="e.g. prod-db-01" /></div>

                <label className="text-sm text-muted-foreground">Kind</label>
                <div className="col-span-2">
                  <Select value={form.kind} onValueChange={onKindChange}>
                    <SelectTrigger className="h-9"><SelectValue placeholder="Select kind" /></SelectTrigger>
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
              <Button variant="outline" onClick={() => setCreateOpen(false)} disabled={creating}>取消</Button>
              <Button onClick={handleCreate} disabled={creating}>{creating ? "创建中..." : "创建"}</Button>
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
