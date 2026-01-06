"use client"

import * as React from "react"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import Editor from "@monaco-editor/react"

export function SchemaViewer() {
  const ddl = `CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
`

  return (
    <div className="flex flex-col h-full bg-muted/5 min-w-0">
       <div className="p-3 border-b font-medium text-sm flex items-center justify-between">
          <span>Table: public.users</span>
       </div>
       <Tabs defaultValue="ddl" className="flex-1 flex flex-col">
          <div className="px-3 pt-2">
            <TabsList className="w-full">
                <TabsTrigger value="ddl" className="flex-1">DDL</TabsTrigger>
                <TabsTrigger value="info" className="flex-1">Info</TabsTrigger>
                <TabsTrigger value="indexes" className="flex-1">Indexes</TabsTrigger>
            </TabsList>
          </div>
          <TabsContent value="ddl" className="flex-1 min-h-0 min-w-0 mt-2 p-0 border-t relative overflow-hidden">
             <div className="absolute inset-0">
                <Editor
                    height="100%"
                    width="100%"
                    defaultLanguage="sql"
                    value={ddl}
                    theme="vs-light"
                    options={{
                        minimap: { enabled: false },
                        fontSize: 12,
                        lineNumbers: "off",
                        readOnly: true,
                        scrollBeyondLastLine: false,
                        automaticLayout: true,
                    }}
                />
             </div>
          </TabsContent>
          <TabsContent value="info" className="p-4 text-sm text-muted-foreground">
             Table information placeholder.
          </TabsContent>
           <TabsContent value="indexes" className="p-4 text-sm text-muted-foreground">
             Indexes information placeholder.
          </TabsContent>
       </Tabs>
    </div>
  )
}
