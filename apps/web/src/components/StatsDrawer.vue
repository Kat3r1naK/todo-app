<template>
  <el-drawer
    :model-value="modelValue"
    title="每日统计"
    size="min(420px, 92vw)"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="stats">
      <div class="summary">
        <div class="metric">
          <span class="value">{{ todayCompleted }}</span>
          <span class="label">今日完成</span>
        </div>
        <div class="metric">
          <span class="value">{{ todayCreated }}</span>
          <span class="label">今日新增</span>
        </div>
        <div class="metric">
          <span class="value">{{ pending }}</span>
          <span class="label">未完成</span>
        </div>
      </div>

      <section class="chart-section">
        <header class="chart-head">
          <h3>近 7 天完成</h3>
          <span>合计 {{ weekTotal }} 项</span>
        </header>
        <div class="chart">
          <div v-for="day in days" :key="day.key" class="chart-col">
            <span class="bar-value">{{ day.completed || '' }}</span>
            <div class="bar-track">
              <span
                class="bar"
                :class="{ empty: day.completed === 0 }"
                :style="{ height: barHeight(day.completed) }"
              />
            </div>
            <span class="bar-label" :class="{ today: day.isToday }">{{ day.label }}</span>
          </div>
        </div>
      </section>

      <p class="hint">按本地日期归集：勾选卡片或拖入「已完成」都会计入当天。</p>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useBoardStore } from '../stores/board'

defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()

const board = useBoardStore()

const allTasks = computed(() => [
  ...board.columns.todo,
  ...board.columns.doing,
  ...board.columns.done,
  ...board.archived,
])

function dateKey(date: Date): string {
  return `${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()}`
}

const completedByDay = computed(() => {
  const map = new Map<string, number>()
  for (const task of allTasks.value) {
    if (!task.completed_at) continue
    const key = dateKey(new Date(task.completed_at))
    map.set(key, (map.get(key) ?? 0) + 1)
  }
  return map
})

const createdByDay = computed(() => {
  const map = new Map<string, number>()
  for (const task of allTasks.value) {
    if (!task.created_at) continue
    const key = dateKey(new Date(task.created_at))
    map.set(key, (map.get(key) ?? 0) + 1)
  }
  return map
})

const days = computed(() => {
  const list: { key: string; label: string; completed: number; isToday: boolean }[] = []
  for (let offset = 6; offset >= 0; offset -= 1) {
    const date = new Date()
    date.setHours(0, 0, 0, 0)
    date.setDate(date.getDate() - offset)
    const key = dateKey(date)
    list.push({
      key,
      label: offset === 0 ? '今天' : date.toLocaleDateString('zh-CN', { weekday: 'short' }),
      completed: completedByDay.value.get(key) ?? 0,
      isToday: offset === 0,
    })
  }
  return list
})

const todayKey = computed(() => dateKey(new Date()))
const todayCompleted = computed(() => completedByDay.value.get(todayKey.value) ?? 0)
const todayCreated = computed(() => createdByDay.value.get(todayKey.value) ?? 0)
const pending = computed(() => board.counts.todo + board.counts.doing)
const weekTotal = computed(() => days.value.reduce((sum, day) => sum + day.completed, 0))
const maxCount = computed(() => Math.max(1, ...days.value.map((day) => day.completed)))

function barHeight(count: number): string {
  if (count === 0) return '2px'
  return `${Math.max(10, Math.round((count / maxCount.value) * 100))}%`
}
</script>

<style scoped>
.stats {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.summary {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1px;
  overflow: hidden;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--border);
}

.metric {
  display: flex;
  flex-direction: column;
  gap: 3px;
  padding: 14px;
  background: var(--card);
}

.value {
  font-size: 22px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  line-height: 1.2;
}

.label {
  color: var(--muted);
  font-size: 12px;
}

.chart-section {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.chart-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
}

.chart-head h3 {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
}

.chart-head span {
  color: var(--muted);
  font-size: 12px;
}

.chart {
  display: flex;
  gap: 10px;
  height: 150px;
  padding: 12px 12px 0;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
}

.chart-col {
  display: flex;
  flex: 1;
  flex-direction: column;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
  height: 100%;
}

.bar-value {
  height: 14px;
  color: var(--muted);
  font-size: 11px;
  font-variant-numeric: tabular-nums;
}

.bar-track {
  display: flex;
  flex: 1;
  align-items: flex-end;
  width: 100%;
  max-width: 26px;
  overflow: hidden;
  border-radius: 3px;
  background: var(--panel);
}

.bar {
  display: block;
  width: 100%;
  border-radius: 3px 3px 0 0;
  background: var(--accent);
  opacity: 0.9;
  transition: height 0.2s;
}

.bar.empty {
  background: var(--border-strong);
  opacity: 0.5;
}

.bar-label {
  color: var(--muted);
  font-size: 11px;
}

.bar-label.today {
  color: var(--text);
  font-weight: 600;
}

.hint {
  margin: 0;
  color: var(--muted);
  font-size: 12px;
  line-height: 1.7;
}
</style>
