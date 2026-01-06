"use client"

import * as React from "react"
import Editor from "@monaco-editor/react"
import { Play, Eraser, FileCode, Save, Loader2, Database } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"

interface SqlEditorProps {
  className?: string
}

export function SqlEditor({ className }: SqlEditorProps) {
  const [code, setCode] = React.useState("SELECT * FROM public.users LIMIT 10;")
  const [isRunning, setIsRunning] = React.useState(false);

  const handleRun = () => {
    setIsRunning(true);
    setTimeout(() => setIsRunning(false), 800);
  };

  return (
    <div className="flex flex-col h-full bg-background">
      <div className="flex items-center justify-between p-2 border-b bg-muted/10 h-12">
        <div className="flex items-center gap-2">
           <Button 
                size="sm" 
                className="bg-green-600 hover:bg-green-700 text-white gap-2 h-8 px-3 font-semibold shadow-sm transition-all"
                onClick={handleRun}
                disabled={isRunning}
            >
             {isRunning ? <Loader2 className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4 fill-current" />}
             {isRunning ? 'Running' : 'Run'}
           </Button>
           <Separator orientation="vertical" className="h-6 mx-1" />
           <Button variant="outline" size="sm" className="gap-2 h-8 text-muted-foreground hover:text-foreground">
             <Save className="w-4 h-4" /> Save
           </Button>
        </div>
        
        <div className="flex items-center gap-2">
            <div className="flex items-center gap-2 mr-2 bg-muted/50 px-2 py-1 rounded border text-xs text-muted-foreground">
                <Database className="w-3 h-3" />
                <span>prod-db-01 / users_db</span>
            </div>
            <Separator orientation="vertical" className="h-6 mx-1" />
            <Select defaultValue="postgres">
                <SelectTrigger className="w-[130px] h-8 text-xs bg-background">
                    <SelectValue placeholder="Dialect" />
                </SelectTrigger>
                <SelectContent>
                    <SelectItem value="postgres">PostgreSQL</SelectItem>
                    <SelectItem value="mysql">MySQL</SelectItem>
                    <SelectItem value="sql">Standard SQL</SelectItem>
                </SelectContent>
            </Select>
            <div className="flex gap-1 ml-1">
                <Button variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground hover:text-foreground" title="Format SQL">
                    <FileCode className="w-4 h-4" />
                </Button>
                <Button variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground hover:text-foreground" title="Clear">
                    <Eraser className="w-4 h-4" />
                </Button>
            </div>
        </div>
      </div>
      <div className="flex-1 min-h-0 relative overflow-hidden">
        <Editor
          height="100%"
          defaultLanguage="sql"
          value={code}
          onChange={(value) => setCode(value || "")}
          theme="vs-light" 
          options={{
            minimap: { enabled: false },
            fontSize: 14,
            fontFamily: "'Geist Mono', 'Menlo', 'Monaco', 'Courier New', monospace",
            lineNumbers: "on",
            roundedSelection: true,
            scrollBeyondLastLine: false,
            readOnly: false,
            automaticLayout: true,
            padding: { top: 16, bottom: 16 },
            renderLineHighlight: 'all',
          }}
        />
      </div>
    </div>
  )
}
