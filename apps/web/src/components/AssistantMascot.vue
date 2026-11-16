<template>
  <div
    v-if="!failed"
    ref="assistantEl"
    class="assistant"
    :class="{ positioned, dragging }"
    :style="rootStyle"
  >
    <button
      ref="stageEl"
      class="assistant-stage"
      type="button"
      aria-label="看板助手，可拖动，点击打开对话"
      @pointerdown="handlePointerDown"
      @click.stop="handleClick"
    >
      <canvas ref="canvasEl" class="assistant-canvas" />
    </button>
  </div>
  <button
    v-else
    class="assistant-fallback"
    type="button"
    aria-label="打开看板助手"
    title="打开看板助手"
    @click="openFallback"
  >
    助手
  </button>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  AmbientLight,
  CylinderGeometry,
  DirectionalLight,
  Group,
  Mesh,
  MeshBasicMaterial,
  MeshStandardMaterial,
  PerspectiveCamera,
  Scene,
  SphereGeometry,
  Timer,
  TorusGeometry,
  Vector2,
  WebGLRenderer,
} from 'three'
import { useUiStore } from '../stores/ui'

const POSITION_KEY = 'todo.assistant.position'

const ui = useUiStore()
const emit = defineEmits<{ open: [anchor: { x: number; y: number }] }>()

const assistantEl = ref<HTMLDivElement | null>(null)
const stageEl = ref<HTMLButtonElement | null>(null)
const canvasEl = ref<HTMLCanvasElement | null>(null)
const failed = ref(false)
const positioned = ref(false)
const dragging = ref(false)
const position = ref<{ left: number; top: number } | null>(null)

const rootStyle = computed(() =>
  positioned.value && position.value
    ? { left: `${position.value.left}px`, top: `${position.value.top}px` }
    : undefined,
)

const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches

let renderer: WebGLRenderer | null = null
let scene: Scene | null = null
let camera: PerspectiveCamera | null = null
let character: Group | null = null
let antenna: Group | null = null
let eyeGroup: Group | null = null
let eyes: Group[] = []
let headMaterial: MeshStandardMaterial | null = null
let earMaterial: MeshStandardMaterial | null = null
let antennaMaterial: MeshStandardMaterial | null = null
let eyeMaterial: MeshBasicMaterial | null = null
let pupilMaterial: MeshBasicMaterial | null = null
let timer: Timer | null = null
let rafId = 0
let blinkAt = 0
let blinkUntil = 0
let bounceAt = -1
let leanZ = 0
let resizeObserver: ResizeObserver | null = null
let moved = false
let dragEndedAt = 0
let dragOrigin = { x: 0, y: 0, left: 0, top: 0 }

const pointer = new Vector2()

function themeColor(name: string, fallback: string): string {
  const value = getComputedStyle(document.documentElement).getPropertyValue(name).trim()
  return value || fallback
}

function applyThemeColors(): void {
  if (!headMaterial || !earMaterial || !antennaMaterial || !eyeMaterial || !pupilMaterial) return
  headMaterial.color.set(themeColor('--accent', '#45634c'))
  earMaterial.color.set(themeColor('--accent', '#45634c'))
  antennaMaterial.color.set(themeColor('--col-doing', '#a77c32'))
  eyeMaterial.color.set(themeColor('--card', '#ffffff'))
  pupilMaterial.color.set(themeColor('--text', '#292e28'))
  renderFrame()
}

function buildScene(): void {
  scene = new Scene()

  camera = new PerspectiveCamera(32, 1, 0.1, 50)
  camera.position.set(0, 0.55, 4.9)
  camera.lookAt(0, 0.05, 0)

  scene.add(new AmbientLight(0xffffff, 1.15))

  const keyLight = new DirectionalLight(0xffffff, 2.4)
  keyLight.position.set(2.4, 3.6, 4)
  scene.add(keyLight)

  const fillLight = new DirectionalLight(0xffffff, 0.9)
  fillLight.position.set(-3, -1.5, 2.2)
  scene.add(fillLight)

  const rimLight = new DirectionalLight(0xffffff, 1.1)
  rimLight.position.set(0, 2.2, -4)
  scene.add(rimLight)

  character = new Group()
  scene.add(character)

  headMaterial = new MeshStandardMaterial({ roughness: 0.42, metalness: 0 })
  const headGeometry = new SphereGeometry(1, 32, 24)
  const positions = headGeometry.attributes.position
  for (let i = 0; i < positions.count; i += 1) {
    const widen = 1 + Math.max(0, -positions.getY(i)) * 0.12
    positions.setX(i, positions.getX(i) * widen)
    positions.setZ(i, positions.getZ(i) * widen)
  }
  headGeometry.computeVertexNormals()
  const head = new Mesh(headGeometry, headMaterial)
  head.scale.set(1, 0.92, 0.95)
  character.add(head)

  earMaterial = new MeshStandardMaterial({ roughness: 0.55, metalness: 0 })
  const earGeometry = new SphereGeometry(0.2, 16, 12)
  for (const side of [-1, 1]) {
    const ear = new Mesh(earGeometry, earMaterial)
    ear.position.set(side * 0.97, 0.02, 0)
    ear.scale.set(0.55, 1, 0.9)
    character.add(ear)
  }

  eyeMaterial = new MeshBasicMaterial()
  pupilMaterial = new MeshBasicMaterial()
  const blushMaterial = new MeshBasicMaterial({ transparent: true, opacity: 0.55 })
  blushMaterial.color.set('#d49393')

  eyeGroup = new Group()
  character.add(eyeGroup)
  const scleraGeometry = new SphereGeometry(0.12, 16, 12)
  const pupilGeometry = new SphereGeometry(0.055, 12, 10)
  const highlightGeometry = new SphereGeometry(0.02, 8, 6)
  for (const side of [-1, 1]) {
    const eye = new Group()
    eye.position.set(side * 0.29, 0.03, 0.885)

    const sclera = new Mesh(scleraGeometry, eyeMaterial)
    sclera.scale.set(1, 1.15, 0.6)
    eye.add(sclera)

    const pupil = new Mesh(pupilGeometry, pupilMaterial)
    pupil.position.set(0, 0.005, 0.055)
    pupil.scale.set(1, 1.1, 0.55)
    eye.add(pupil)

    const highlight = new Mesh(highlightGeometry, eyeMaterial)
    highlight.position.set(0.035, 0.04, 0.072)
    eye.add(highlight)

    eyeGroup.add(eye)
    eyes.push(eye)
  }

  const mouth = new Mesh(new TorusGeometry(0.11, 0.016, 8, 24, Math.PI), eyeMaterial)
  mouth.position.set(0, -0.2, 0.915)
  mouth.rotation.z = Math.PI
  character.add(mouth)

  const blushGeometry = new SphereGeometry(0.1, 12, 8)
  for (const side of [-1, 1]) {
    const blush = new Mesh(blushGeometry, blushMaterial)
    blush.position.set(side * 0.57, -0.22, 0.755)
    blush.scale.set(1, 0.6, 0.16)
    character.add(blush)
  }

  antenna = new Group()
  antenna.position.set(0, 0.88, 0)
  character.add(antenna)
  antennaMaterial = new MeshStandardMaterial({ roughness: 0.7, metalness: 0, flatShading: true })
  const stem = new Mesh(new CylinderGeometry(0.025, 0.035, 0.34, 8), antennaMaterial)
  stem.position.y = 0.17
  antenna.add(stem)
  const tip = new Mesh(new SphereGeometry(0.09, 12, 8), antennaMaterial)
  tip.position.y = 0.36
  antenna.add(tip)
}

function renderFrame(): void {
  if (renderer && scene && camera) renderer.render(scene, camera)
}

function handleResize(): void {
  if (!renderer || !camera || !stageEl.value) return
  const width = stageEl.value.clientWidth
  const height = stageEl.value.clientHeight
  if (!width || !height) return
  renderer.setSize(width, height, false)
  camera.aspect = width / height
  camera.updateProjectionMatrix()
  renderFrame()
}

function tick(): void {
  rafId = requestAnimationFrame(tick)
  if (!renderer || !scene || !camera || !timer) return
  timer.update()
  const time = timer.getElapsed()

  if (character) {
    character.position.y = Math.sin(time * 1.6) * 0.045
    character.rotation.y += (pointer.x * 0.32 - character.rotation.y) * 0.07
    character.rotation.x += (pointer.y * 0.16 - character.rotation.x) * 0.07
    character.rotation.z += (leanZ - character.rotation.z) * 0.09

    let squash = 0
    if (bounceAt >= 0) {
      const elapsed = time - bounceAt
      if (elapsed > 1.1) {
        bounceAt = -1
      } else {
        squash = Math.sin(elapsed * 20) * 0.16 * Math.exp(-elapsed * 3.2)
      }
    }
    const breath = Math.sin(time * 2) * 0.014
    character.scale.set(1 + squash * 0.7 + breath, 1 - squash - breath, 1 + squash * 0.7 + breath)
  }

  if (antenna) antenna.rotation.z = Math.sin(time * 2.1) * 0.07 + pointer.x * 0.18

  if (time > blinkAt) {
    blinkUntil = time + 0.12
    blinkAt = time + 2.6 + Math.random() * 3
  }
  const eyeScale = time < blinkUntil ? 0.14 : 1
  for (const eye of eyes) eye.scale.y += (eyeScale - eye.scale.y) * 0.55

  if (eyeGroup) {
    eyeGroup.position.x += (pointer.x * 0.06 - eyeGroup.position.x) * 0.1
    eyeGroup.position.y += (-pointer.y * 0.045 - eyeGroup.position.y) * 0.1
  }

  renderer.render(scene, camera)
}

function handlePointerMove(event: PointerEvent): void {
  pointer.x = (event.clientX / (window.innerWidth || 1)) * 2 - 1
  pointer.y = (event.clientY / (window.innerHeight || 1)) * 2 - 1
}

function clampPosition(left: number, top: number): { left: number; top: number } {
  const width = assistantEl.value?.offsetWidth ?? 132
  const height = assistantEl.value?.offsetHeight ?? 132
  const margin = 8
  return {
    left: Math.min(Math.max(left, margin), Math.max(window.innerWidth - width - margin, margin)),
    top: Math.min(Math.max(top, margin), Math.max(window.innerHeight - height - margin, margin)),
  }
}

function handlePointerDown(event: PointerEvent): void {
  if (event.button !== 0 || !assistantEl.value) return
  const rect = assistantEl.value.getBoundingClientRect()
  position.value = { left: rect.left, top: rect.top }
  positioned.value = true
  dragOrigin = { x: event.clientX, y: event.clientY, left: rect.left, top: rect.top }
  moved = false
  dragging.value = true
  try {
    stageEl.value?.setPointerCapture(event.pointerId)
  } catch {
    /* synthetic events have no active pointer to capture */
  }
}

function handleDragMove(event: PointerEvent): void {
  if (!dragging.value) return
  const dx = event.clientX - dragOrigin.x
  const dy = event.clientY - dragOrigin.y
  if (!moved && Math.abs(dx) + Math.abs(dy) > 4) moved = true
  if (!moved) return
  position.value = clampPosition(dragOrigin.left + dx, dragOrigin.top + dy)
  leanZ = Math.max(-0.28, Math.min(0.28, -dx * 0.0015))
}

function handleDragEnd(event: PointerEvent): void {
  if (!dragging.value) return
  dragging.value = false
  leanZ = 0
  if (stageEl.value?.hasPointerCapture(event.pointerId)) {
    stageEl.value.releasePointerCapture(event.pointerId)
  }
  if (moved && position.value) {
    localStorage.setItem(POSITION_KEY, JSON.stringify(position.value))
  }
  dragEndedAt = moved ? performance.now() : 0
}

function restorePosition(): void {
  try {
    const saved = JSON.parse(localStorage.getItem(POSITION_KEY) || 'null') as {
      left?: number
      top?: number
    } | null
    if (typeof saved?.left !== 'number' || typeof saved?.top !== 'number') return
    position.value = clampPosition(saved.left, saved.top)
    positioned.value = true
  } catch {
    /* ignore malformed value */
  }
}

function repositionOnResize(): void {
  if (positioned.value && position.value) {
    position.value = clampPosition(position.value.left, position.value.top)
  }
}

function handleClick(): void {
  if (performance.now() - dragEndedAt < 350) return
  const rect = stageEl.value?.getBoundingClientRect()
  emit('open', {
    x: rect ? rect.left + rect.width / 2 : window.innerWidth - 80,
    y: rect ? rect.top : window.innerHeight - 80,
  })
}

function openFallback(): void {
  emit('open', { x: window.innerWidth - 80, y: window.innerHeight - 80 })
}

function handleContextLost(event: Event): void {
  event.preventDefault()
  console.warn('[assistant] WebGL context lost，使用备用入口')
  failed.value = true
}

function initScene(): void {
  if (!canvasEl.value) return
  try {
    renderer = new WebGLRenderer({ canvas: canvasEl.value, alpha: true, antialias: true })
    renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2))
    canvasEl.value.addEventListener('webglcontextlost', handleContextLost)
    buildScene()
    applyThemeColors()
    handleResize()
    resizeObserver = new ResizeObserver(handleResize)
    if (stageEl.value) resizeObserver.observe(stageEl.value)

    if (!reduced) {
      timer = new Timer()
      timer.connect(document)
      blinkAt = 1.6 + Math.random() * 2
      rafId = requestAnimationFrame(tick)
    } else {
      renderFrame()
    }
  } catch (error) {
    console.warn('[assistant] 3D 初始化失败，使用备用入口', error)
    failed.value = true
    renderer?.dispose()
    renderer = null
  }
}

watch(
  () => ui.theme,
  () => applyThemeColors(),
  { flush: 'post' },
)

onMounted(() => {
  restorePosition()
  initScene()
  if (!reduced) window.addEventListener('pointermove', handlePointerMove, { passive: true })
  window.addEventListener('pointermove', handleDragMove)
  window.addEventListener('pointerup', handleDragEnd)
  window.addEventListener('pointercancel', handleDragEnd)
  window.addEventListener('resize', repositionOnResize)
})

onBeforeUnmount(() => {
  cancelAnimationFrame(rafId)
  canvasEl.value?.removeEventListener('webglcontextlost', handleContextLost)
  window.removeEventListener('pointermove', handlePointerMove)
  window.removeEventListener('pointermove', handleDragMove)
  window.removeEventListener('pointerup', handleDragEnd)
  window.removeEventListener('pointercancel', handleDragEnd)
  window.removeEventListener('resize', repositionOnResize)
  timer?.disconnect()
  resizeObserver?.disconnect()
  scene?.traverse((object) => {
    if (!(object instanceof Mesh)) return
    object.geometry.dispose()
    const materials = Array.isArray(object.material) ? object.material : [object.material]
    for (const material of materials) material.dispose()
  })
  renderer?.dispose()
  renderer = null
})
</script>

<style scoped>
.assistant {
  position: fixed;
  right: 26px;
  bottom: 24px;
  z-index: 30;
  width: 132px;
  height: 132px;
  pointer-events: none;
}

.assistant.positioned {
  right: auto;
  bottom: auto;
}

.assistant-stage {
  position: relative;
  width: 100%;
  height: 100%;
  padding: 0;
  border: 0;
  background: transparent;
  cursor: grab;
  pointer-events: auto;
  border-radius: var(--radius-lg);
  touch-action: none;
  user-select: none;
  -webkit-user-select: none;
}

.assistant.dragging .assistant-stage {
  cursor: grabbing;
}

.assistant-canvas {
  display: block;
  width: 100%;
  height: 100%;
  transition: transform 0.18s ease;
}

.assistant-stage:hover .assistant-canvas {
  transform: translateY(-2px);
}

.assistant.dragging .assistant-canvas {
  transform: none;
}

.assistant-fallback {
  position: fixed;
  right: 26px;
  bottom: 24px;
  z-index: 30;
  width: 46px;
  height: 46px;
  border: 1px solid var(--border-strong);
  border-radius: 50%;
  background: var(--card);
  color: var(--text);
  font-size: 12px;
  cursor: pointer;
}

.assistant-fallback:hover {
  border-color: var(--accent);
  color: var(--accent);
}

@media (max-width: 760px) {
  .assistant {
    right: 16px;
    bottom: 14px;
    width: 104px;
    height: 104px;
  }
}
</style>
