<template>
  <Teleport to="body">
    <Transition name="agent-window">
      <section
        v-if="modelValue"
        class="agent-window"
        :style="windowStyle"
        role="dialog"
        aria-label="看板助手"
      >
        <header class="agent-head" @pointerdown="startDrag">
          <div class="agent-title">
            <span class="agent-dot" />
            <span>看板助手</span>
          </div>
          <div class="agent-head-actions">
            <button class="agent-icon" type="button" aria-label="模型设置" title="模型设置" @click="openSettings">
              <el-icon :size="14"><Setting /></el-icon>
            </button>
            <button class="agent-close" type="button" aria-label="关闭助手" title="关闭助手" @click="close">
              <el-icon :size="14"><Close /></el-icon>
            </button>
          </div>
        </header>

        <template v-if="!settingsOpen">
        <div ref="listRef" class="agent-messages">
          <div v-if="messages.length === 0" class="agent-empty">
            <p class="agent-empty-title">可以问我：</p>
            <div class="agent-suggestions">
              <button
                v-for="suggestion in SUGGESTIONS"
                :key="suggestion"
                type="button"
                class="agent-suggestion"
                @click="sendSuggestion(suggestion)"
              >
                {{ suggestion }}
              </button>
            </div>
          </div>
          <template v-for="message in messages" :key="message.id">
            <div v-if="message.text || message.steps.length" class="agent-message" :class="message.role">
              <div class="agent-bubble">
                <p v-if="message.text" class="agent-text">{{ message.text }}</p>
                <ul v-if="message.steps.length" class="agent-steps">
                  <li v-for="step in message.steps" :key="step.toolCallId" :class="{ error: step.ok === false }">
                    {{ step.label }}
                  </li>
                </ul>
              </div>
            </div>
          </template>
          <p v-if="error" class="agent-error">{{ error }}</p>
        </div>

        <div class="agent-composer">
          <textarea
            ref="composerRef"
            v-model="draft"
            class="agent-input"
            rows="2"
            maxlength="2000"
            placeholder="问点什么…"
            aria-label="输入消息"
            @keydown.enter.exact.prevent="send"
          />
          <div class="agent-composer-bar">
            <span class="agent-hint">Enter 发送 · Shift+Enter 换行</span>
            <el-button v-if="streaming" size="small" @click="cancel">停止</el-button>
            <el-button v-else size="small" type="primary" :disabled="!draft.trim()" @click="send">发送</el-button>
          </div>
        </div>
        </template>

        <div v-else class="agent-settings">
          <p class="agent-settings-title">模型设置</p>
          <p class="agent-settings-hint">使用你自己的模型 Key；不配置则使用平台默认模型。保存后下一条消息生效。</p>
          <label class="agent-field">
            <span>服务商</span>
            <el-select
              v-model="form.provider"
              class="agent-select"
              size="small"
              filterable
              placeholder="选择服务商"
              @change="handleProviderChange"
            >
              <el-option
                v-for="provider in modelOptions.providers"
                :key="provider.id"
                :label="`${provider.name} (${provider.id})`"
                :value="provider.id"
              />
            </el-select>
          </label>
          <label class="agent-field">
            <span>模型</span>
            <el-select v-model="form.modelId" class="agent-select" size="small" filterable placeholder="选择模型">
              <el-option
                v-for="model in currentModels"
                :key="model.id"
                :label="`${model.name} (${model.id})`"
                :value="model.id"
              />
            </el-select>
          </label>
          <label class="agent-field">
            <span>API Key</span>
            <el-input
              v-model="form.apiKey"
              size="small"
              type="password"
              show-password
              :placeholder="settings.hasApiKey ? '已保存，留空保持不变' : '填写你的 API Key'"
            />
          </label>
          <p class="agent-settings-status">{{ settingsSummary }}</p>
          <p v-if="settingsLoading" class="agent-settings-hint">加载模型目录…</p>
          <p v-if="settingsError" class="agent-error">{{ settingsError }}</p>
          <div class="agent-settings-actions">
            <el-button size="small" @click="backToChat">返回对话</el-button>
            <div class="agent-settings-buttons">
              <el-button
                size="small"
                :disabled="settingsLoading || (settings.usingDefault && !settings.hasApiKey)"
                @click="clearSettings"
              >
                清除
              </el-button>
              <el-button size="small" type="primary" :loading="savingSettings" :disabled="settingsLoading" @click="saveSettings">
                保存
              </el-button>
            </div>
          </div>
        </div>
      </section>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { Close, Setting } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import type { Event } from '@ag-ui/core'
import { getToken } from '../api'

interface Anchor {
  x: number
  y: number
}

const props = defineProps<{ modelValue: boolean; anchor?: Anchor | null }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()

interface Step {
  toolCallId: string
  toolName: string
  label: string
  ok?: boolean
}

interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  text: string
  steps: Step[]
}

interface ModelOptionItem {
  id: string
  name: string
}

interface ProviderOptionItem {
  id: string
  name: string
  models: ModelOptionItem[]
}

interface ModelSettingsState {
  provider: string
  modelId: string
  hasApiKey: boolean
  maskedApiKey: string
  usingDefault: boolean
  defaultProvider: string
  defaultModel: string
}

interface SaveSettingsResponse {
  ok: boolean
  provider: string
  modelId: string
  hasApiKey: boolean
  maskedApiKey: string
  usingDefault: boolean
}

const TOOL_LABELS: Record<string, string> = {
  list_tasks: '查询任务',
  get_task: '读取任务详情',
  list_notes: '查询便签',
  get_daily_stats: '统计今日进度',
  create_task: '创建任务',
  update_task: '更新任务',
  complete_task: '完成任务',
}

const SUGGESTIONS = ['今天有哪些任务？', '便签里记了什么？', '今天完成了多少？']

const WINDOW_WIDTH = 396
const WINDOW_HEIGHT = 560

const messages = ref<ChatMessage[]>([])
const draft = ref('')
const streaming = ref(false)
const error = ref('')
const listRef = ref<HTMLElement>()
const composerRef = ref<HTMLTextAreaElement>()
const threadId = ref(crypto.randomUUID())
const position = ref<{ left: number; top: number } | null>(null)
const settingsOpen = ref(false)
const settingsLoading = ref(false)
const savingSettings = ref(false)
const settingsError = ref('')
const modelOptions = ref<{ providers: ProviderOptionItem[] }>({ providers: [] })
const settings = ref<ModelSettingsState>({
  provider: '',
  modelId: '',
  hasApiKey: false,
  maskedApiKey: '',
  usingDefault: true,
  defaultProvider: '',
  defaultModel: '',
})
const form = reactive({ provider: '', modelId: '', apiKey: '' })
const currentModels = computed(
  () => modelOptions.value.providers.find((provider) => provider.id === form.provider)?.models ?? [],
)
const settingsSummary = computed(() =>
  settings.value.usingDefault
    ? `当前使用平台默认：${settings.value.defaultProvider} / ${settings.value.defaultModel}`
    : `当前使用自定义模型：${settings.value.provider} / ${settings.value.modelId}（Key ${settings.value.maskedApiKey}）`,
)
let controller: AbortController | null = null
let drag: { startX: number; startY: number; left: number; top: number } | null = null

const windowStyle = ref<Record<string, string>>({})

function viewportSize(): { width: number; height: number } {
  return {
    width: Math.min(WINDOW_WIDTH, window.innerWidth - 24),
    height: Math.min(WINDOW_HEIGHT, window.innerHeight - 96),
  }
}

function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max)
}

function initialPosition(anchor?: Anchor | null): { left: number; top: number } {
  const { width, height } = viewportSize()
  const maxLeft = window.innerWidth - width - 12
  const maxTop = window.innerHeight - height - 12
  if (!anchor) {
    return { left: maxLeft, top: maxTop }
  }
  return {
    left: clamp(anchor.x - width / 2, 12, maxLeft),
    top: clamp(anchor.y - height - 16, 12, maxTop),
  }
}

function applyGeometry(): void {
  const { width, height } = viewportSize()
  const current = position.value ?? initialPosition(props.anchor)
  const left = clamp(current.left, 8, window.innerWidth - width - 8)
  const top = clamp(current.top, 8, window.innerHeight - height - 8)
  position.value = { left, top }
  windowStyle.value = {
    left: `${left}px`,
    top: `${top}px`,
    width: `${width}px`,
    height: `${height}px`,
  }
}

function close(): void {
  emit('update:modelValue', false)
}

function cancel(): void {
  controller?.abort()
  streaming.value = false
}

async function agentApi<T>(path: string, init: RequestInit = {}): Promise<T> {
  const res = await fetch(path, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${getToken()}`,
    },
  })
  const data = await res.json().catch(() => null)
  if (!res.ok) {
    const message =
      data && typeof data === 'object' && 'error' in data
        ? String((data as { error: unknown }).error)
        : `请求失败 (${res.status})`
    throw new Error(message)
  }
  return data as T
}

async function openSettings(): Promise<void> {
  settingsOpen.value = true
  settingsError.value = ''
  settingsLoading.value = true
  try {
    const [config, options] = await Promise.all([
      agentApi<ModelSettingsState>('/agent/model-settings'),
      agentApi<{ providers: ProviderOptionItem[] }>('/agent/model-options'),
    ])
    settings.value = config
    modelOptions.value = options
    form.provider = config.provider
    form.modelId = config.modelId
    form.apiKey = ''
  } catch (error) {
    settingsError.value = (error as Error).message
  } finally {
    settingsLoading.value = false
  }
}

function handleProviderChange(provider: string): void {
  const models = modelOptions.value.providers.find((item) => item.id === provider)?.models ?? []
  form.modelId = models.some((model) => model.id === form.modelId) ? form.modelId : models[0]?.id ?? ''
}

async function saveSettings(): Promise<void> {
  if (!form.provider || !form.modelId) {
    settingsError.value = '请选择服务商与模型'
    return
  }
  savingSettings.value = true
  settingsError.value = ''
  try {
    const body: Record<string, unknown> = { provider: form.provider, modelId: form.modelId }
    if (form.apiKey.trim()) body.apiKey = form.apiKey.trim()
    const result = await agentApi<SaveSettingsResponse>('/agent/model-settings', {
      method: 'PUT',
      body: JSON.stringify(body),
    })
    settings.value = {
      ...settings.value,
      provider: result.provider,
      modelId: result.modelId,
      hasApiKey: result.hasApiKey,
      maskedApiKey: result.maskedApiKey,
      usingDefault: result.usingDefault,
    }
    form.apiKey = ''
    settingsOpen.value = false
    ElMessage.success('模型配置已保存，下一条消息生效')
  } catch (error) {
    settingsError.value = (error as Error).message
  } finally {
    savingSettings.value = false
  }
}

async function clearSettings(): Promise<void> {
  savingSettings.value = true
  settingsError.value = ''
  try {
    await agentApi('/agent/model-settings', {
      method: 'PUT',
      body: JSON.stringify({ clearApiKey: true }),
    })
    settings.value = {
      ...settings.value,
      provider: settings.value.defaultProvider,
      modelId: settings.value.defaultModel,
      hasApiKey: false,
      maskedApiKey: '',
      usingDefault: true,
    }
    form.provider = settings.value.defaultProvider
    form.modelId = settings.value.defaultModel
    form.apiKey = ''
    ElMessage.success('已清除，使用平台默认模型')
  } catch (error) {
    settingsError.value = (error as Error).message
  } finally {
    savingSettings.value = false
  }
}

function backToChat(): void {
  settingsOpen.value = false
  settingsError.value = ''
  void nextTick(() => composerRef.value?.focus())
}

function startDrag(event: PointerEvent): void {
  if ((event.target as HTMLElement).closest('button')) return
  const current = position.value ?? initialPosition(props.anchor)
  drag = { startX: event.clientX, startY: event.clientY, left: current.left, top: current.top }
  window.addEventListener('pointermove', onDragMove)
  window.addEventListener('pointerup', stopDrag, { once: true })
}

function onDragMove(event: PointerEvent): void {
  if (!drag) return
  const { width, height } = viewportSize()
  position.value = {
    left: clamp(drag.left + event.clientX - drag.startX, 8, window.innerWidth - width - 8),
    top: clamp(drag.top + event.clientY - drag.startY, 8, window.innerHeight - height - 8),
  }
  applyGeometry()
}

function stopDrag(): void {
  drag = null
  window.removeEventListener('pointermove', onDragMove)
}

function onResize(): void {
  if (props.modelValue) applyGeometry()
}

function onKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape' && props.modelValue) close()
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      position.value = initialPosition(props.anchor)
      applyGeometry()
      void nextTick(() => {
        scrollToBottom()
        composerRef.value?.focus()
      })
    }
  },
)

watch(
  () => props.anchor,
  (anchor) => {
    if (props.modelValue) {
      position.value = initialPosition(anchor)
      applyGeometry()
    }
  },
)

watch(
  messages,
  () => {
    void nextTick(() => scrollToBottom())
  },
  { deep: true },
)

function scrollToBottom(): void {
  listRef.value?.scrollTo({ top: listRef.value.scrollHeight })
}

async function send(): Promise<void> {
  const text = draft.value.trim()
  if (!text || streaming.value) return

  draft.value = ''
  error.value = ''
  messages.value.push({ id: crypto.randomUUID(), role: 'user', text, steps: [] })
  streaming.value = true
  controller = new AbortController()

  try {
    const res = await fetch('/agent/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${getToken()}`,
      },
      body: JSON.stringify({ threadId: threadId.value, message: text }),
      signal: controller.signal,
    })
    if (!res.ok) {
      const data = (await res.json().catch(() => null)) as { error?: string } | null
      throw new Error(data?.error || `请求失败 (${res.status})`)
    }
    if (!res.body) throw new Error('服务未返回流式响应')
    await readStream(res.body)
  } catch (err) {
    if ((err as Error).name === 'AbortError') return
    error.value = (err as Error).message
  } finally {
    streaming.value = false
    controller = null
  }
}

function sendSuggestion(text: string): void {
  if (streaming.value) return
  draft.value = text
  void send()
}

async function readStream(body: ReadableStream<Uint8Array>): Promise<void> {
  const reader = body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    let boundary = buffer.indexOf('\n\n')
    while (boundary >= 0) {
      const frame = buffer.slice(0, boundary)
      buffer = buffer.slice(boundary + 2)
      applyFrame(frame)
      boundary = buffer.indexOf('\n\n')
    }
  }
  if (buffer.trim()) applyFrame(buffer)
}

function applyFrame(frame: string): void {
  const dataLines: string[] = []
  for (const line of frame.split(/\r?\n/)) {
    if (line.startsWith('data:')) dataLines.push(line.slice(5).trimStart())
  }
  if (dataLines.length === 0) return

  let payload: Event
  try {
    payload = JSON.parse(dataLines.join('\n')) as Event
  } catch {
    return
  }

  switch (payload.type) {
    case 'TEXT_MESSAGE_START':
      ensureAssistantMessage(payload.messageId)
      break
    case 'TEXT_MESSAGE_CONTENT': {
      const message = ensureAssistantMessage(payload.messageId)
      message.text += payload.delta
      break
    }
    case 'TOOL_CALL_START': {
      const parent =
        (payload.parentMessageId
          ? messages.value.find((message) => message.id === payload.parentMessageId)
          : undefined) ??
        lastAssistantMessage() ??
        ensureAssistantMessage(crypto.randomUUID())
      parent.steps.push({
        toolCallId: payload.toolCallId,
        toolName: payload.toolCallName,
        label: `${TOOL_LABELS[payload.toolCallName] ?? payload.toolCallName}…`,
      })
      break
    }
    case 'TOOL_CALL_RESULT': {
      const found = findStep(payload.toolCallId)
      if (found) markStepDone(found.step, true)
      break
    }
    case 'CUSTOM': {
      if (payload.name !== 'agent.tool_status') break
      const value = payload.value as { toolCallId?: unknown; ok?: unknown }
      if (typeof value.toolCallId !== 'string') break
      const found = findStep(value.toolCallId)
      if (found) markStepDone(found.step, value.ok !== false)
      break
    }
    case 'RUN_ERROR':
      error.value = payload.message
      break
    default:
      break
  }
}

function ensureAssistantMessage(id: string): ChatMessage {
  const existing = messages.value.find((message) => message.id === id)
  if (existing) return existing
  const message: ChatMessage = { id, role: 'assistant', text: '', steps: [] }
  messages.value.push(message)
  return message
}

function lastAssistantMessage(): ChatMessage | undefined {
  for (let index = messages.value.length - 1; index >= 0; index -= 1) {
    if (messages.value[index].role === 'assistant') return messages.value[index]
  }
  return undefined
}

function findStep(toolCallId: string): { message: ChatMessage; step: Step } | undefined {
  for (const message of messages.value) {
    const step = message.steps.find((item) => item.toolCallId === toolCallId)
    if (step) return { message, step }
  }
  return undefined
}

function markStepDone(step: Step, ok: boolean): void {
  step.ok = ok
  step.label = `${TOOL_LABELS[step.toolName] ?? step.toolName} · ${ok ? '完成' : '失败'}`
}

onMounted(() => {
  window.addEventListener('resize', onResize)
  document.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize)
  document.removeEventListener('keydown', onKeydown)
  stopDrag()
  controller?.abort()
})
</script>

<style scoped>
.agent-window {
  position: fixed;
  z-index: 1200;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 10px;
  box-shadow: 0 10px 30px rgb(0 0 0 / 12%);
}

.agent-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 9px 12px;
  border-bottom: 1px solid var(--border);
  background: var(--panel);
  cursor: grab;
  user-select: none;
}

.agent-head:active {
  cursor: grabbing;
}

.agent-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text);
  font-size: 13px;
  font-weight: 600;
}

.agent-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--accent);
}

.agent-head-actions {
  display: flex;
  align-items: center;
  gap: 2px;
}

.agent-icon,
.agent-close {
  display: flex;
  padding: 4px;
  color: var(--muted);
  background: transparent;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.agent-icon:hover,
.agent-close:hover {
  color: var(--text);
  background: var(--panel-2);
}

.agent-messages {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px;
}

.agent-messages::-webkit-scrollbar {
  width: 6px;
}

.agent-messages::-webkit-scrollbar-thumb {
  border-radius: 3px;
  background: var(--border);
}

.agent-messages::-webkit-scrollbar-track {
  background: transparent;
}

.agent-empty {
  margin: 0;
  padding: 4px 0;
}

.agent-empty-title {
  margin: 0 0 8px;
  color: var(--muted);
  font-size: 13px;
}

.agent-suggestions {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.agent-suggestion {
  padding: 5px 9px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: transparent;
  color: var(--text);
  font-family: inherit;
  font-size: 12px;
  line-height: 1.4;
  cursor: pointer;
}

.agent-suggestion:hover {
  border-color: var(--border-strong);
  background: var(--panel);
}

.agent-message {
  display: flex;
}

.agent-message.user {
  justify-content: flex-end;
}

.agent-message.assistant {
  justify-content: flex-start;
}

.agent-bubble {
  max-width: 88%;
  padding: 8px 11px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--card);
}

.agent-message.user .agent-bubble {
  border-color: var(--accent-soft);
  background: var(--accent-soft);
}

.agent-text {
  margin: 0;
  color: var(--text);
  font-size: 13px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}

.agent-steps {
  margin: 8px 0 0;
  padding: 8px 0 0;
  list-style: none;
  border-top: 1px dashed var(--border);
}

.agent-steps li {
  display: flex;
  align-items: baseline;
  gap: 6px;
  color: var(--muted);
  font-size: 12px;
  line-height: 1.8;
}

.agent-steps li::before {
  content: '';
  flex: none;
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: var(--accent);
  transform: translateY(-1px);
}

.agent-steps li.error::before {
  background: #c45656;
}

.agent-steps li.error {
  color: #c45656;
}

.agent-error {
  margin: 0;
  padding: 4px 8px;
  border-left: 2px solid #c45656;
  color: #c45656;
  font-size: 12px;
}

.agent-composer {
  padding: 10px;
  border-top: 1px solid var(--border);
}

.agent-input {
  display: block;
  width: 100%;
  box-sizing: border-box;
  padding: 8px 10px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--card);
  color: var(--text);
  font-family: inherit;
  font-size: 13px;
  line-height: 1.6;
  resize: none;
  outline: none;
}

.agent-input:focus {
  border-color: var(--accent);
}

.agent-input::placeholder {
  color: var(--muted);
}

.agent-composer-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-top: 8px;
}

.agent-hint {
  color: var(--muted);
  font-size: 11px;
}

.agent-settings {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
}

.agent-settings-title {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
}

.agent-settings-hint {
  margin: 0;
  color: var(--muted);
  font-size: 12px;
  line-height: 1.6;
}

.agent-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  color: var(--muted);
  font-size: 12px;
}

.agent-select {
  width: 100%;
}

.agent-settings-status {
  margin: 2px 0 0;
  color: var(--text);
  font-size: 12px;
  word-break: break-all;
}

.agent-settings-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-top: auto;
  padding-top: 10px;
  border-top: 1px solid var(--border);
}

.agent-settings-buttons {
  display: flex;
  gap: 8px;
}

.agent-window-enter-active,
.agent-window-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.agent-window-enter-from,
.agent-window-leave-to {
  opacity: 0;
  transform: translateY(6px);
}
</style>
