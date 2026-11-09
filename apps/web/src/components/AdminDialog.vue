<template>
  <el-dialog
    :model-value="modelValue"
    title="账号管理"
    width="760px"
    @open="loadUsers"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="create-row">
      <el-input v-model="form.username" placeholder="新账号（字母/数字/_/-）" class="grow" />
      <el-input v-model="form.password" type="password" show-password placeholder="初始密码" class="grow" />
      <el-checkbox v-model="form.is_admin">管理员</el-checkbox>
      <el-button type="primary" :loading="creating" @click="createUser">创建</el-button>
    </div>

    <el-table v-loading="loading" :data="users" size="default" max-height="400">
      <el-table-column label="账号" min-width="140">
        <template #default="{ row }">
          <div class="cell-user">
            <span class="name">{{ row.username }}</span>
            <el-tag v-if="row.is_admin" size="small" effect="plain" type="warning">管理员</el-tag>
            <el-tag v-if="row.id === auth.user?.id" size="small" effect="plain" type="info">当前</el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="task_count" label="任务数" width="80" align="center" />
      <el-table-column label="免登链接" width="110" align="center">
        <template #default="{ row }">
          <el-button text type="primary" size="small" :icon="CopyDocument" @click="copyLink(row as AdminUser)">复制</el-button>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" align="right">
        <template #default="{ row }">
          <el-button text size="small" :icon="Key" @click="resetPassword(row as AdminUser)">重置密码</el-button>
          <el-button
            text
            type="danger"
            size="small"
            :icon="Delete"
            :disabled="row.id === auth.user?.id"
            @click="removeUser(row as AdminUser)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <p class="hint">免登链接可直接登录对应账号，请妥善保管；删除账号会同时删除该账号的全部任务和便签。</p>
  </el-dialog>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CopyDocument, Delete, Key } from '@element-plus/icons-vue'
import { api } from '../api'
import { useAuthStore } from '../stores/auth'
import type { AdminUser, AdminUsersResponse, OkResponse } from '../types'

defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()

const auth = useAuthStore()
const users = ref<AdminUser[]>([])
const loading = ref(false)
const creating = ref(false)
const form = reactive<{ username: string; password: string; is_admin: boolean }>({
  username: '',
  password: '',
  is_admin: false,
})

async function loadUsers(): Promise<void> {
  loading.value = true
  try {
    const data = await api<AdminUsersResponse>('/admin/users')
    users.value = data.users ?? []
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    loading.value = false
  }
}

async function createUser(): Promise<void> {
  const username = form.username.trim()
  if (!/^[A-Za-z0-9_-]{1,32}$/.test(username)) {
    ElMessage.warning('账号需为 1-32 位字母、数字、下划线或短横线')
    return
  }
  if (!form.password) {
    ElMessage.warning('请输入初始密码')
    return
  }
  creating.value = true
  try {
    await api<OkResponse>('/admin/users', {
      method: 'POST',
      body: { username, password: form.password, is_admin: form.is_admin },
    })
    ElMessage.success('账号已创建')
    form.username = ''
    form.password = ''
    form.is_admin = false
    await loadUsers()
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    creating.value = false
  }
}

async function resetPassword(row: AdminUser): Promise<void> {
  let value = ''
  try {
    const result = await ElMessageBox.prompt(`为「${row.username}」设置新密码`, '重置密码', {
      confirmButtonText: '保存',
      cancelButtonText: '取消',
      inputType: 'password',
      inputValidator: (input: string) => {
        if (!input) return '密码不能为空'
        if (input.length > 72) return '密码长度需为 1-72 位'
        return true
      },
    })
    value = result.value
  } catch {
    return
  }
  try {
    await api<OkResponse>(`/admin/users/${row.id}/password`, {
      method: 'POST',
      body: { password: value },
    })
    ElMessage.success('密码已重置')
  } catch (error) {
    ElMessage.error((error as Error).message)
  }
}

async function removeUser(row: AdminUser): Promise<void> {
  try {
    await ElMessageBox.confirm(`确定删除账号「${row.username}」及其全部数据吗？`, '删除账号', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await api<OkResponse>(`/admin/users/${row.id}`, { method: 'DELETE' })
    ElMessage.success('账号已删除')
    await loadUsers()
  } catch (error) {
    ElMessage.error((error as Error).message)
  }
}

async function copyLink(row: AdminUser): Promise<void> {
  const url = `${location.origin}/login?key=${encodeURIComponent(row.autologin_key)}`
  try {
    await navigator.clipboard.writeText(url)
    ElMessage.success('免登链接已复制')
  } catch {
    const textarea = document.createElement('textarea')
    textarea.value = url
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    textarea.remove()
    ElMessage.success('免登链接已复制')
  }
}
</script>

<style scoped>
.create-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
}

.grow {
  flex: 1;
}

.cell-user {
  display: flex;
  align-items: center;
  gap: 6px;
}

.cell-user .name {
  font-weight: 500;
}

.hint {
  margin: 14px 0 0;
  color: var(--muted);
  font-size: 12px;
  line-height: 1.6;
}

@media (max-width: 640px) {
  .create-row {
    flex-wrap: wrap;
  }

  .grow {
    flex: 1 1 100%;
  }
}
</style>
