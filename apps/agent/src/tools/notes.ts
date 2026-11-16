import { Type } from "typebox"
import { callApi } from "../goapi/client.js"
import { getRunContext } from "./context.js"
import { definePlatformTool } from "./types.js"

interface NoteDto {
  id: number
  content: string
  color: string
  created_at: string
  updated_at: string
}

const listNotesParams = Type.Object({
  limit: Type.Optional(Type.Integer({ minimum: 1, maximum: 50, description: "返回条数，默认 20" })),
})

export const listNotesTool = definePlatformTool<typeof listNotesParams>({
  name: "list_notes",
  label: "查询便签",
  description: "查询当前用户的便签（独立便签实体），返回内容摘要与颜色。",
  sideEffect: "read",
  parameters: listNotesParams,
  async execute(_toolCallId, params) {
    const { token } = getRunContext()
    const data = await callApi<{ notes: NoteDto[] }>("/notes", { token })
    const limit = params.limit ?? 20
    const notes = data.notes.slice(0, limit).map((note) => ({
      id: note.id,
      content: note.content.length > 300 ? `${note.content.slice(0, 300)}…` : note.content,
      color: note.color,
      updated_at: note.updated_at,
    }))

    return {
      content: [
        {
          type: "text" as const,
          text: JSON.stringify({ total: data.notes.length, returned: notes.length, notes }),
        },
      ],
      details: { total: data.notes.length },
    }
  },
})
