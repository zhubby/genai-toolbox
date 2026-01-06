"use client"

import * as React from "react"
import { GripVerticalIcon } from "lucide-react"
import * as ResizablePrimitive from "react-resizable-panels"

import { cn } from "@/lib/utils"

const PanelGroup = (ResizablePrimitive as any).PanelGroup || (ResizablePrimitive as any).Group
const Panel = (ResizablePrimitive as any).Panel
const PanelResizeHandle = (ResizablePrimitive as any).PanelResizeHandle || (ResizablePrimitive as any).Separator

function ResizablePanelGroup({
  className,
  ...props
}: React.ComponentProps<typeof PanelGroup>) {
  const isVertical = (props as any)?.direction === "vertical";
  return (
    <PanelGroup
      data-slot="resizable-panel-group"
      className={cn(
        "flex h-full w-full",
        isVertical ? "flex-col" : undefined,
        // 兼容依赖 data 属性的样式（Tailwind v4 data 变体）
        "data-[panel-group-direction=vertical]:flex-col",
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

function ResizableHandle({
  withHandle,
  className,
  ...props
}: React.ComponentProps<typeof PanelResizeHandle> & {
  withHandle?: boolean
}) {
  return (
    <PanelResizeHandle
      data-slot="resizable-handle"
      className={cn(
        // 可拖拽分割条：更大的可点击区域 + 细线视觉 + 更高层级
        // 横向：提供更宽的点击区域 w-3，细线通过 after 伪元素呈现
        // 纵向：提供更高的点击高度 h-3，同样用 after 呈现细线
        "relative z-20 shrink-0 flex items-center justify-center touch-none select-none",
        "data-[panel-group-direction=horizontal]:w-3 data-[panel-group-direction=horizontal]:cursor-col-resize",
        "data-[panel-group-direction=vertical]:h-3 data-[panel-group-direction=vertical]:w-full data-[panel-group-direction=vertical]:cursor-row-resize",
        // 细线视觉
        "after:absolute after:bg-border",
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
        <div className="bg-border z-10 flex h-4 w-3 items-center justify-center rounded-xs border">
          <GripVerticalIcon className="size-2.5" />
        </div>
      )}
    </PanelResizeHandle>
  )
}

export { ResizablePanelGroup, ResizablePanel, ResizableHandle }
