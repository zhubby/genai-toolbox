"use client"

import * as React from "react"
import { Sidebar } from "./Sidebar"
import { SchemaViewer } from "./SchemaViewer"
import { StatusBar } from "./StatusBar"
import { DbManagerProvider } from "./context"
import { Button } from "@/components/ui/button"
import { PanelLeft, PanelRight, Settings, Package, MessageSquareText } from "lucide-react"
import { SqlEditor } from "./SqlEditor"
import { ResultTable } from "./ResultTable"
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable"
import { ToolsetsDialog } from "./ToolsetsDialog"
import { PromptsetsDialog } from "./PromptsetsDialog"

export function MainLayout() {
  const [showLeft, setShowLeft] = React.useState(true)
  const [showRight, setShowRight] = React.useState(true)
  const [gearOpen, setGearOpen] = React.useState(false)
  const [toolsetsOpen, setToolsetsOpen] = React.useState(false)
  const [promptsetsOpen, setPromptsetsOpen] = React.useState(false)

  React.useEffect(() => {
    if (!gearOpen) return
    const onDocClick = () => setGearOpen(false)
    document.addEventListener("click", onDocClick)
    return () => document.removeEventListener("click", onDocClick)
  }, [gearOpen])

  return (
    <DbManagerProvider>
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

            <div className="relative">
              <Button
                variant="outline"
                size="sm"
                className="h-8"
                title="设置"
                onClick={(e) => {
                  e.stopPropagation()
                  setGearOpen((v) => !v)
                }}
              >
                <Settings className="w-4 h-4" />
              </Button>
              {gearOpen && (
                <div
                  className="absolute right-0 top-10 z-50 w-48 rounded-md border bg-background shadow-md overflow-hidden"
                  onClick={(e) => e.stopPropagation()}
                >
                  <button
                    className="w-full flex items-center gap-2 px-3 py-2 text-sm hover:bg-muted"
                    onClick={() => {
                      setGearOpen(false)
                      setToolsetsOpen(true)
                    }}
                  >
                    <Package className="w-4 h-4 text-muted-foreground" />
                    工具集合
                  </button>
                  <button
                    className="w-full flex items-center gap-2 px-3 py-2 text-sm hover:bg-muted"
                    onClick={() => {
                      setGearOpen(false)
                      setPromptsetsOpen(true)
                    }}
                  >
                    <MessageSquareText className="w-4 h-4 text-muted-foreground" />
                    Prompt 集合
                  </button>
                  <button
                    className="w-full flex items-center gap-2 px-3 py-2 text-sm text-muted-foreground hover:bg-muted"
                    onClick={() => {
                      setGearOpen(false)
                      alert("设置暂未实现")
                    }}
                  >
                    <Settings className="w-4 h-4 text-muted-foreground" />
                    设置
                  </button>
                </div>
              )}
            </div>
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

      <ToolsetsDialog open={toolsetsOpen} onOpenChange={setToolsetsOpen} />
      <PromptsetsDialog open={promptsetsOpen} onOpenChange={setPromptsetsOpen} />
    </DbManagerProvider>
  )
}
