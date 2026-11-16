import { randomUUID } from "node:crypto"
import { EventType, type Event } from "@ag-ui/core"
import { summarizeJson } from "../http/events.js"

export type PiMapperEvent =
  | { type: "message_start"; message: { role: string } }
  | { type: "message_end"; message: { role: string } }
  | { type: "message_update"; assistantMessageEvent: { type: string; delta?: string } }
  | { type: "tool_execution_start"; toolCallId: string; toolName: string; args: unknown }
  | { type: "tool_execution_end"; toolCallId: string; result: unknown; isError: boolean; durationMs?: number }

export class AguiMapper {
  private currentAssistantMessageId: string | null = null
  private lastAssistantMessageId: string | null = null

  constructor(private readonly emit: (event: Event) => void) {}

  handle(event: PiMapperEvent): void {
    switch (event.type) {
      case "message_start":
        if (event.message.role === "assistant") {
          this.openAssistantMessage()
        }
        return
      case "message_end":
        if (event.message.role === "assistant" && this.currentAssistantMessageId) {
          this.emit({ type: EventType.TEXT_MESSAGE_END, messageId: this.currentAssistantMessageId })
          this.currentAssistantMessageId = null
        }
        return
      case "message_update":
        if (event.assistantMessageEvent.type !== "text_delta") return
        this.emit({
          type: EventType.TEXT_MESSAGE_CONTENT,
          messageId: this.currentAssistantMessageId ?? this.openAssistantMessage(),
          delta: event.assistantMessageEvent.delta ?? "",
        })
        return
      case "tool_execution_start":
        this.emit({
          type: EventType.TOOL_CALL_START,
          toolCallId: event.toolCallId,
          toolCallName: event.toolName,
          parentMessageId: this.lastAssistantMessageId ?? undefined,
        })
        this.emit({
          type: EventType.TOOL_CALL_ARGS,
          toolCallId: event.toolCallId,
          delta: stringifyArgs(event.args),
        })
        this.emit({ type: EventType.TOOL_CALL_END, toolCallId: event.toolCallId })
        return
      case "tool_execution_end":
        this.emit({
          type: EventType.TOOL_CALL_RESULT,
          messageId: randomUUID(),
          toolCallId: event.toolCallId,
          content: summarizeJson(event.result, 4000),
          role: "tool",
        })
        this.emit({
          type: EventType.CUSTOM,
          name: "agent.tool_status",
          value: {
            toolCallId: event.toolCallId,
            ok: !event.isError,
            durationMs: event.durationMs ?? null,
          },
        })
    }
  }

  private openAssistantMessage(): string {
    const messageId = randomUUID()
    this.currentAssistantMessageId = messageId
    this.lastAssistantMessageId = messageId
    this.emit({ type: EventType.TEXT_MESSAGE_START, messageId, role: "assistant" })
    return messageId
  }
}

function stringifyArgs(value: unknown): string {
  try {
    return JSON.stringify(value ?? {})
  } catch {
    return "{}"
  }
}
