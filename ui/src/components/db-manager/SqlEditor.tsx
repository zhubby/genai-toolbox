"use client"

import * as React from "react"
import Editor from "@monaco-editor/react"
import { Play, Eraser, FileCode, Save, Loader2, Database } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { cn } from "@/lib/utils"
import { useDbManager } from "./context"
import { createTool, updateTool } from "@/lib/api"

interface SqlEditorProps {
  className?: string
}

type ParamType = "string" | "integer" | "float" | "boolean" | "array" | "map"

type ParamDef = {
  name: string
  type: ParamType
  description: string
  required?: boolean
  default?: any
}

function isCommentOrStringScope(scope: string): boolean {
  const s = String(scope).toLowerCase()
  return s.includes("comment") || s.includes("string")
}

function extractPostgresParamMaxIndex(statement: string, monaco: any): { max: number; used: number[] } {
  // Use Monaco SQL tokenizer to avoid counting placeholders inside strings/comments.
  // Falls back to a simple scan if monaco isn't ready.
  const used = new Set<number>()
  const re = /\$([1-9]\d*)/g

  if (!monaco?.editor?.tokenize) {
    let m: RegExpExecArray | null
    while ((m = re.exec(statement))) used.add(Number(m[1]))
  } else {
    const lines = statement.split(/\r?\n/)
    const tokenLines: any[][] = monaco.editor.tokenize(statement, "sql")
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i] || ""
      const tokens = tokenLines[i] || []
      if (tokens.length === 0) {
        let m: RegExpExecArray | null
        while ((m = re.exec(line))) used.add(Number(m[1]))
        continue
      }

      for (let t = 0; t < tokens.length; t++) {
        const start = tokens[t].startIndex ?? 0
        const end = t + 1 < tokens.length ? tokens[t + 1].startIndex ?? line.length : line.length
        const scopes = tokens[t].scopes ?? ""
        if (isCommentOrStringScope(scopes)) continue

        const seg = line.slice(start, end)
        let m: RegExpExecArray | null
        while ((m = re.exec(seg))) used.add(Number(m[1]))
      }
    }
  }

  const arr = Array.from(used).sort((a, b) => a - b)
  const max = arr.length ? arr[arr.length - 1] : 0
  return { max, used: arr }
}

export function SqlEditor({ className }: SqlEditorProps) {
  const [statement, setStatement] = React.useState("SELECT * FROM public.users LIMIT 10;")
  const [isRunning, setIsRunning] = React.useState(false)
  const [isSaving, setIsSaving] = React.useState(false)
  const [saveError, setSaveError] = React.useState<string | null>(null)
  const [monaco, setMonaco] = React.useState<any>(null)

  const [toolName, setToolName] = React.useState("")
  const [description, setDescription] = React.useState("")
  const [parameters, setParameters] = React.useState<ParamDef[]>([])
  const [parsedParamMax, setParsedParamMax] = React.useState(0)
  const [parseError, setParseError] = React.useState<string | null>(null)

  const { selectedSource, reloadTools } = useDbManager()

  const dbName = typeof selectedSource?.config?.database === "string" ? selectedSource?.config?.database : ""
  const connLabel = selectedSource?.name ? `${selectedSource.name}${dbName ? ` / ${dbName}` : ""}` : "未选择数据库"
  const toolKind = selectedSource?.kind === "mysql" ? "mysql-sql" : "postgres-sql"

  React.useEffect(() => {
    if (!selectedSource?.name) return
    if (!toolName.trim()) {
      const base = selectedSource.name.replace(/[^a-zA-Z0-9_-]/g, "_")
      setToolName(`${base}_sql`)
    }
    if (!description.trim()) {
      setDescription("SQL tool")
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedSource?.name])

  const handleRun = () => {
    setIsRunning(true)
    setTimeout(() => setIsRunning(false), 800)
  }

  const postgresParamInfo = React.useMemo(() => extractPostgresParamMaxIndex(statement, monaco), [statement, monaco])

  // Requirement: parameters editor is only shown after clicking Parse.
  // Any statement edits invalidate the parsed result.
  React.useEffect(() => {
    setParsedParamMax(0)
    setParseError(null)
  }, [statement])

  const handleParse = () => {
    setSaveError(null)
    setParseError(null)
    const sql = statement.trim()
    if (!sql) {
      setParsedParamMax(0)
      setParameters([])
      return
    }

    const info = extractPostgresParamMaxIndex(statement, monaco)
    const max = info.max
    if (max === 0) {
      setParsedParamMax(0)
      setParameters([])
      return
    }

    const usedSet = new Set(info.used)
    const missing: number[] = []
    for (let i = 1; i <= max; i++) {
      if (!usedSet.has(i)) missing.push(i)
    }
    if (missing.length) {
      setParsedParamMax(0)
      setParameters([])
      setParseError(`占位符必须从 $1 连续到 $${max}，缺少：${missing.map((n) => `$${n}`).join(", ")}`)
      return
    }

    setParsedParamMax(max)
    setParameters((prev) => {
      const next: ParamDef[] = []
      for (let i = 1; i <= max; i++) {
        const existing = prev[i - 1]
        next.push(
          existing || {
            name: `p${i}`,
            type: "string",
            description: `parameter $${i}`,
            required: true,
          },
        )
      }
      return next
    })
  }

  function validateBeforeSave(): string[] {
    const errs: string[] = []
    if (!selectedSource?.name) errs.push("请先选择一个数据库 source")
    if (!toolName.trim()) errs.push("Tool name 必填")
    if (!description.trim()) errs.push("Description 必填")
    if (!statement.trim()) errs.push("Statement 不能为空")

    const max = postgresParamInfo.max
    if (max > 0) {
      // Must click Parse to generate the parameter rows.
      if (parsedParamMax !== max) {
        errs.push("请先点击 Parse 解析 SQL 占位符（$n）")
      }

      // Enforce contiguous placeholders: $1..$N with no gaps.
      const usedSet = new Set(postgresParamInfo.used)
      const missing: number[] = []
      for (let i = 1; i <= max; i++) {
        if (!usedSet.has(i)) missing.push(i)
      }
      if (missing.length) {
        errs.push(`占位符必须从 $1 连续到 $${max}，缺少：${missing.map((n) => `$${n}`).join(", ")}`)
      }

      if (parameters.length !== max) {
        errs.push(`参数数量需要为 ${max}（对应 $1..$${max}）`)
      }
      const nameSeen = new Set<string>()
      for (let i = 0; i < parameters.length; i++) {
        const p = parameters[i]
        if (!p.name.trim()) errs.push(`$${i + 1} 的 name 必填`)
        if (!p.description?.trim()) errs.push(`$${i + 1} 的 description 必填`)
        const n = p.name.trim()
        if (n) {
          if (nameSeen.has(n)) errs.push(`参数名重复: ${n}`)
          nameSeen.add(n)
        }
      }
    }

    return errs
  }

  const handleSave = async () => {
    setSaveError(null)
    const errs = validateBeforeSave()
    if (errs.length) {
      setSaveError(errs.join("；"))
      return
    }
    if (!selectedSource?.name) return

    const cfg = {
      name: toolName.trim(),
      kind: toolKind,
      source: selectedSource.name,
      description: description.trim(),
      statement,
      parameters: parameters.map((p) => ({
        name: p.name.trim(),
        type: p.type,
        description: p.description.trim(),
        ...(p.required === false ? { required: false } : {}),
        ...(p.default !== undefined ? { default: p.default } : {}),
      })),
    }

    setIsSaving(true)
    try {
      try {
        await createTool({
          name: toolName.trim(),
          kind: toolKind,
          sourceName: selectedSource.name,
          config: cfg,
        })
      } catch (e: any) {
        // Upsert: if already exists (409), fall back to update.
        const msg = e?.message || String(e)
        if (String(msg).includes("409")) {
          await updateTool(toolName.trim(), {
            kind: toolKind,
            sourceName: selectedSource.name,
            config: cfg,
          })
        } else {
          throw e
        }
      }
      await reloadTools()
    } catch (e: any) {
      setSaveError(e?.message || String(e))
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <div className={cn("flex flex-col h-full bg-background min-w-0", className)}>
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
           <Button
             variant="outline"
             size="sm"
             className="gap-2 h-8 text-muted-foreground hover:text-foreground"
             onClick={handleSave}
             disabled={isSaving}
           >
             {isSaving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />} Save
           </Button>
        </div>
        
        <div className="flex items-center gap-2">
            <div className="flex items-center gap-2 mr-2 bg-muted/50 px-2 py-1 rounded border text-xs text-muted-foreground">
                <Database className="w-3 h-3" />
                <span className="truncate max-w-[220px]">{connLabel}</span>
            </div>
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
      <div className="flex-1 relative overflow-hidden min-h-0 min-w-0">
        <div className="absolute inset-0">
            <Editor
            height="100%"
            width="100%"
            defaultLanguage="sql"
            value={statement}
            onChange={(value) => setStatement(value || "")}
            onMount={(_, m) => setMonaco(m)}
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
      <div className="border-t bg-muted/5 p-3 space-y-3 overflow-auto max-h-[260px]">
        {saveError && <div className="text-xs text-red-600">{saveError}</div>}
        {parseError && <div className="text-xs text-red-600">{parseError}</div>}

        <div className="grid grid-cols-6 gap-2 items-center">
          <div className="col-span-2 text-xs text-muted-foreground">Tool name</div>
          <div className="col-span-4">
            <Input
              value={toolName}
              onChange={(e) => setToolName(e.target.value)}
              placeholder="例如：list_users"
              className="h-8 text-xs"
            />
          </div>

          <div className="col-span-2 text-xs text-muted-foreground">Description</div>
          <div className="col-span-4">
            <Input
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="例如：列出用户（示例）"
              className="h-8 text-xs"
            />
          </div>
        </div>

        <div className="text-xs text-muted-foreground">
          Kind：<span className="font-mono text-foreground">{toolKind}</span>
          {parsedParamMax > 0 ? <span className="ml-2">（已解析 $1..$${parsedParamMax}）</span> : null}
        </div>

        <Separator />

        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <div className="text-sm font-medium">Parameters（Parse 后展示，$n）</div>
            <Button
              variant="outline"
              size="sm"
              className="h-8 gap-2 text-xs"
              onClick={handleParse}
            >
              Parse
            </Button>
          </div>
          {!statement.trim() || parsedParamMax === 0 ? null : (
            <div className="space-y-2">
              {Array.from({ length: parsedParamMax }, (_, idx) => {
                const p = parameters[idx]
                return (
                  <div key={idx} className="grid grid-cols-12 gap-2 items-center">
                    <div className="col-span-2 text-xs font-mono text-muted-foreground">{`$${idx + 1}`}</div>
                    <div className="col-span-3">
                      <Input
                        value={p?.name || ""}
                        onChange={(e) =>
                          setParameters((cur) => cur.map((x, i) => (i === idx ? { ...x, name: e.target.value } : x)))
                        }
                        className="h-8 text-xs font-mono"
                      />
                    </div>
                    <div className="col-span-3">
                      <Select
                        value={p?.type || "string"}
                        onValueChange={(v) =>
                          setParameters((cur) => cur.map((x, i) => (i === idx ? { ...x, type: v as ParamType } : x)))
                        }
                      >
                        <SelectTrigger className="h-8 text-xs bg-background">
                          <SelectValue placeholder="type" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="string">string</SelectItem>
                          <SelectItem value="integer">integer</SelectItem>
                          <SelectItem value="float">float</SelectItem>
                          <SelectItem value="boolean">boolean</SelectItem>
                          <SelectItem value="array">array</SelectItem>
                          <SelectItem value="map">map</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    <div className="col-span-4">
                      <Input
                        value={p?.description || ""}
                        onChange={(e) =>
                          setParameters((cur) => cur.map((x, i) => (i === idx ? { ...x, description: e.target.value } : x)))
                        }
                        className="h-8 text-xs"
                      />
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
