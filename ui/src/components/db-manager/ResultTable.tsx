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
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Clock, CheckCircle2, ChevronLeft, ChevronRight, Download, Filter } from "lucide-react"

export function ResultTable() {
  const mockData = Array.from({ length: 50 }).map((_, i) => ({
    id: i + 1,
    username: `user_${i + 1}`,
    email: `user${i + 1}@example.com`,
    role: i % 3 === 0 ? "admin" : "user",
    status: i % 4 === 0 ? "inactive" : "active",
    created_at: new Date(Date.now() - Math.random() * 10000000000).toISOString(),
    last_login: new Date(Date.now() - Math.random() * 100000000).toISOString(),
  }))

  return (
    <div className="flex flex-col h-full w-full max-w-full min-w-0 overflow-hidden bg-background">
      <div className="flex items-center justify-between px-4 py-2 border-b bg-muted/10 h-10 shrink-0">
         <div className="flex items-center gap-4 text-xs text-muted-foreground">
            <span className="flex items-center gap-1.5 text-green-600 font-medium">
                <CheckCircle2 className="w-3.5 h-3.5" />
                Success
            </span>
            <span className="flex items-center gap-1.5">
                <Clock className="w-3.5 h-3.5" />
                14ms
            </span>
            <span className="text-foreground font-medium">50 rows</span>
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
            <TableRow className="hover:bg-transparent border-b-muted/50">
              <TableHead className="w-[60px] h-9">ID</TableHead>
              <TableHead className="h-9">Username</TableHead>
              <TableHead className="h-9">Email</TableHead>
              <TableHead className="h-9">Role</TableHead>
              <TableHead className="h-9">Status</TableHead>
              <TableHead className="h-9">Created At</TableHead>
              <TableHead className="h-9">Last Login</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {mockData.map((row) => (
              <TableRow key={row.id} className="border-b-muted/40 hover:bg-muted/20">
                <TableCell className="py-2 font-mono text-xs text-muted-foreground">{row.id}</TableCell>
                <TableCell className="py-2 font-medium">{row.username}</TableCell>
                <TableCell className="py-2 text-muted-foreground">{row.email}</TableCell>
                <TableCell className="py-2">
                    <Badge variant={row.role === 'admin' ? "default" : "secondary"} className="text-[10px] px-1.5 h-5 font-normal tracking-wide uppercase">{row.role}</Badge>
                </TableCell>
                <TableCell className="py-2">
                    <div className="flex items-center gap-2">
                        <span className={`w-1.5 h-1.5 rounded-full ${row.status === 'active' ? 'bg-green-500' : 'bg-gray-300'}`} />
                        <span className="capitalize text-sm">{row.status}</span>
                    </div>
                </TableCell>
                <TableCell className="py-2 text-xs text-muted-foreground font-mono">{row.created_at}</TableCell>
                <TableCell className="py-2 text-xs text-muted-foreground font-mono">{row.last_login}</TableCell>
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
