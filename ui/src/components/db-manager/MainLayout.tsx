"use client"

import * as React from "react"
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable"
import { Sidebar } from "./Sidebar"
import { SqlEditor } from "./SqlEditor"
import { ResultTable } from "./ResultTable"
import { SchemaViewer } from "./SchemaViewer"

import { StatusBar } from "./StatusBar"

export function MainLayout() {
  return (
    <div className="h-screen w-full bg-background overflow-hidden flex flex-col">
       <header className="h-12 border-b flex items-center px-4 shrink-0 bg-background z-10">
          <div className="flex items-center gap-2 font-semibold select-none">
             <div className="w-6 h-6 bg-primary rounded flex items-center justify-center text-primary-foreground text-xs font-bold shadow-sm">DB</div>
             GenAI Toolbox DB Manager
          </div>
       </header>
       <div className="flex-1 overflow-hidden">
        <ResizablePanelGroup direction="horizontal">
            {/* Left Sidebar: Instances */}
            <ResizablePanel defaultSize={25} minSize={15} maxSize={40}>
                <Sidebar className="h-full" />
            </ResizablePanel>
            
            <ResizableHandle />

            {/* Middle: SQL Editor & Results */}
            <ResizablePanel defaultSize={50} minSize={30} className="min-w-0 overflow-hidden">
                <ResizablePanelGroup direction="vertical" className="!flex-col h-full">
                    <ResizablePanel defaultSize={50} minSize={20}>
                        <SqlEditor />
                    </ResizablePanel>
                    
                    <ResizableHandle className="h-px w-full bg-border" />
                    
                    <ResizablePanel defaultSize={50} minSize={20}>
                        <ResultTable />
                    </ResizablePanel>
                </ResizablePanelGroup>
            </ResizablePanel>

            <ResizableHandle />

            {/* Right Sidebar: Schema */}
            <ResizablePanel defaultSize={25} minSize={15} maxSize={40}>
                <SchemaViewer />
            </ResizablePanel>
        </ResizablePanelGroup>
       </div>
       <StatusBar />
    </div>
  )
}
