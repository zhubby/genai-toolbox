"use client"

import * as React from "react"
import { GripVerticalIcon } from "lucide-react"
import { Group, Panel, Separator } from "react-resizable-panels"

import { cn } from "@/lib/utils"

// react-resizable-panels v4 API:
// - Group (旧版 PanelGroup)
// - Panel
// - Separator (旧版 PanelResizeHandle)
// - orientation prop (旧版 direction)

interface ResizablePanelGroupProps extends Omit<React.ComponentProps<typeof Group>, 'orientation'> {
  direction?: "horizontal" | "vertical"
}

function ResizablePanelGroup({
  className,
  direction = "horizontal",
  ...props
}: ResizablePanelGroupProps) {
  return (
    <Group
      data-slot="resizable-panel-group"
      data-panel-group-direction={direction}
      orientation={direction}
      className={cn(
        "flex h-full w-full",
        direction === "vertical" ? "flex-col" : "flex-row",
        className
      )}
      {...props}
    />
  )
}

function ResizablePanel({ className, ...props }: React.ComponentProps<typeof Panel>) {
  return (
    <Panel
      data-slot="resizable-panel"
      className={cn("min-w-0 min-h-0", className)}
      {...props}
    />
  )
}

interface ResizableHandleProps extends React.ComponentProps<typeof Separator> {
  withHandle?: boolean
}

function ResizableHandle({
  withHandle,
  className,
  ...props
}: ResizableHandleProps) {
  return (
    <Separator
      data-slot="resizable-handle"
      className={cn(
        // 基础样式：细线可点击区域
        "relative z-20 shrink-0 flex items-center justify-center touch-none select-none",
        "bg-transparent transition-colors",
        // 横向分隔条
        "data-[panel-group-direction=horizontal]:w-1.5 data-[panel-group-direction=horizontal]:cursor-col-resize",
        // 纵向分隔条
        "data-[panel-group-direction=vertical]:h-1.5 data-[panel-group-direction=vertical]:w-full data-[panel-group-direction=vertical]:cursor-row-resize",
        // 细线视觉（通过 after 伪元素）
        "after:absolute after:bg-border/60 after:transition-colors",
        "hover:after:bg-primary/40",
        "data-[panel-group-direction=horizontal]:after:inset-y-0 data-[panel-group-direction=horizontal]:after:left-1/2 data-[panel-group-direction=horizontal]:after:w-px data-[panel-group-direction=horizontal]:after:-translate-x-1/2",
        "data-[panel-group-direction=vertical]:after:inset-x-0 data-[panel-group-direction=vertical]:after:top-1/2 data-[panel-group-direction=vertical]:after:h-px data-[panel-group-direction=vertical]:after:-translate-y-1/2",
        // 焦点可见性
        "focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-1 focus-visible:outline-hidden",
        // 可选的抓手图标容器在纵向时旋转
        "[&[data-panel-group-direction=vertical]>div]:rotate-90",
        className
      )}
      {...props}
    >
      {withHandle && (
        <div className="bg-muted/80 hover:bg-muted z-10 flex h-5 w-1 items-center justify-center rounded-full transition-colors data-[panel-group-direction=vertical]:h-1 data-[panel-group-direction=vertical]:w-5">
          <GripVerticalIcon className="size-2 text-muted-foreground/60" />
        </div>
      )}
    </Separator>
  )
}

export { ResizablePanelGroup, ResizablePanel, ResizableHandle }
