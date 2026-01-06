"use client"

import * as React from "react"
import { Sidebar } from "./Sidebar"
import { SchemaViewer } from "./SchemaViewer"
import { StatusBar } from "./StatusBar"
import { Button } from "@/components/ui/button"
import { PanelLeft, PanelRight } from "lucide-react"
import { SqlEditor } from "./SqlEditor"
import { ResultTable } from "./ResultTable"
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable"

export function MainLayout() {
  const [showLeft, setShowLeft] = React.useState(true)
  const [showRight, setShowRight] = React.useState(true)

  return (
    <div className="h-screen w-full bg-background overflow-hidden flex flex-col">
      <header className="h-12 border-b flex items-center justify-between px-4 shrink-0 bg-background z-10">
        <div className="flex items-center gap-2 font-semibold select-none">
          <div className="w-6 h-6 bg-primary rounded flex items-center justify-center text-primary-foreground text-xs font-bold shadow-sm">DB</div>
          GenAI Toolbox DB Manager
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant={showLeft ? "secondary" : "outline"}
            size="sm"
            onClick={() => setShowLeft((v) => !v)}
            className="h-8"
            title={showLeft ? "隐藏左侧" : "显示左侧"}
          >
            <PanelLeft className="w-4 h-4" />
          </Button>
          <Button
            variant={showRight ? "secondary" : "outline"}
            size="sm"
            onClick={() => setShowRight((v) => !v)}
            className="h-8"
            title={showRight ? "隐藏右侧" : "显示右侧"}
          >
            <PanelRight className="w-4 h-4" />
          </Button>
        </div>
      </header>
      <div className="flex-1 overflow-hidden min-w-0">
        <div className="h-full w-full flex">
          {showLeft && (
            <div className="h-full border-r min-w-0" style={{ width: "25%" }}>
              <Sidebar className="h-full" />
            </div>
          )}
          <div className="flex-1 min-w-0">
            <ResizablePanelGroup direction="vertical" className="flex h-full w-full">
              <ResizablePanel defaultSize={55} minSize={25} className="min-h-0">
                <SqlEditor />
              </ResizablePanel>
              <ResizableHandle withHandle />
              <ResizablePanel defaultSize={45} minSize={25} className="min-h-0">
                <ResultTable />
              </ResizablePanel>
            </ResizablePanelGroup>
          </div>
          {showRight && (
            <div className="h-full border-l min-w-0" style={{ width: "25%" }}>
              <SchemaViewer />
            </div>
          )}
        </div>
      </div>
      <StatusBar />
    </div>
  )
}
