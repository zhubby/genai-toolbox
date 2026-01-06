"use client"

import { Wifi, GitBranch, Check } from "lucide-react"
import { useDbManager } from "./context"

export function StatusBar() {
  const { selectedSource } = useDbManager()
  const name = selectedSource?.name
  return (
    <div className="h-7 border-t bg-primary text-primary-foreground text-xs flex items-center justify-between px-3 select-none">
       <div className="flex items-center gap-4">
          <div className="flex items-center gap-1.5 hover:bg-primary-foreground/10 px-1.5 py-0.5 rounded cursor-pointer transition-colors">
              <Wifi className="w-3 h-3" />
              <span>
                {name ? (
                  <>
                    Connected to <strong>{name}</strong>
                  </>
                ) : (
                  <span className="opacity-80">Not connected</span>
                )}
              </span>
          </div>
          <div className="flex items-center gap-1.5 hover:bg-primary-foreground/10 px-1.5 py-0.5 rounded cursor-pointer transition-colors">
              <GitBranch className="w-3 h-3" />
              <span>main</span>
          </div>
       </div>

       <div className="flex items-center gap-4">
          <div className="flex items-center gap-1.5">
             <span className="opacity-70">Ln 12, Col 45</span>
          </div>
          <div className="flex items-center gap-1.5">
             <span className="opacity-70">UTF-8</span>
          </div>
          <div className="flex items-center gap-1.5 hover:bg-primary-foreground/10 px-1.5 py-0.5 rounded cursor-pointer transition-colors">
             <Check className="w-3 h-3" />
             <span>Ready</span>
          </div>
       </div>
    </div>
  )
}
