"use client"

import * as React from "react"
import { Database, Search, Server, Table2, Plus, RefreshCw, ChevronRight, ChevronDown, MoreVertical } from "lucide-react"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { cn } from "@/lib/utils"

interface SidebarProps extends React.HTMLAttributes<HTMLDivElement> {}

const mockInstances = [
  {
    id: "prod-db-01",
    name: "Production Primary",
    type: "PostgreSQL",
    status: "online",
    databases: [
      { 
        name: "users_db", 
        tables: ["public.users", "public.profiles", "public.sessions", "auth.audit_logs"] 
      },
      { 
        name: "orders_db", 
        tables: ["public.orders", "public.order_items", "public.products", "inventory.stock"] 
      },
    ],
  },
  {
    id: "dev-db-01",
    name: "Development",
    type: "MySQL",
    status: "online",
    databases: [
      { name: "dev_main", tables: ["test_table_1", "test_table_2"] },
    ],
  },
  {
    id: "analytics-01",
    name: "Analytics Warehouse",
    type: "ClickHouse",
    status: "offline",
    databases: [],
  },
]

export function Sidebar({ className, ...props }: SidebarProps) {
  const [expanded, setExpanded] = React.useState<Record<string, boolean>>({
    "prod-db-01": true,
    "prod-db-01-users_db": true,
  });

  const toggleExpand = (id: string) => {
    setExpanded(prev => ({ ...prev, [id]: !prev[id] }));
  };

  return (
    <div className={cn("flex flex-col h-full border-r bg-background min-w-0", className)} {...props}>
      <div className="p-4 space-y-4">
        <div className="flex items-center justify-between">
           <h2 className="text-sm font-semibold tracking-tight uppercase text-muted-foreground">Explorer</h2>
           <div className="flex gap-1">
             <Button variant="ghost" size="icon" className="h-6 w-6">
                <Plus className="h-4 w-4" />
             </Button>
             <Button variant="ghost" size="icon" className="h-6 w-6">
                <RefreshCw className="h-4 w-4" />
             </Button>
           </div>
        </div>
        <div className="relative">
          <Search className="absolute left-2 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
          <Input placeholder="Search..." className="pl-8 h-9 text-sm" />
        </div>
      </div>
      <Separator />
      <ScrollArea className="flex-1">
        <div className="p-2">
          {mockInstances.map((instance) => (
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
    </div>
  )
}
