<template>
  <div class="login-page">
    <header class="login-header">
      <div class="brand">
        <div class="brand-logo">
          <el-icon :size="20"><Check /></el-icon>
        </div>
        <span>待办面板</span>
      </div>
      <el-dropdown trigger="click" @command="(theme) => ui.apply(String(theme))">
        <el-button text :icon="Brush" aria-label="切换主题" title="切换主题" />
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item v-for="theme in THEMES" :key="theme.id" :command="theme.id" :disabled="ui.theme === theme.id">
              {{ theme.label }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </header>
    <main class="login-card">
      <div class="login-heading">
        <h1>登录</h1>
        <p>使用管理员分配的账号访问看板。</p>
      </div>
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" size="large" @submit.prevent="submit">
        <el-form-item label="账号" prop="username">
          <el-input v-model="form.username" placeholder="请输入账号" autocomplete="username" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            placeholder="请输入密码"
            autocomplete="current-password"
          />
        </el-form-item>
        <el-button class="login-btn" type="primary" native-type="submit" :loading="loading || autoLoading">
          {{ autoLoading ? '正在登录…' : '登录' }}
        </el-button>
      </el-form>
      <p class="tip">忘记密码请联系管理员重置。</p>
    </main>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Brush, Check } from '@element-plus/icons-vue'
import { useAuthStore } from '../stores/auth'
import { THEMES, useUiStore } from '../stores/ui'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const ui = useUiStore()

const formRef = ref<FormInstance>()
const loading = ref(false)
const autoLoading = ref(false)
const form = reactive({ username: '', password: '' })
const rules: FormRules = {
  username: [{ required: true, message: '请输入账号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

onMounted(async () => {
  const key = route.query.key
  if (!key) return
  autoLoading.value = true
  try {
    await auth.autologin(String(key))
    ElMessage.success('免登成功')
    router.replace('/')
  } catch (error) {
    ElMessage.error((error as Error).message)
    router.replace({ path: '/login' })
  } finally {
    autoLoading.value = false
  }
})

async function submit(): Promise<void> {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  loading.value = true
  try {
    await auth.login(form.username.trim(), form.password)
    ElMessage.success('欢迎回来')
    router.replace('/')
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  display: grid;
  grid-template-rows: auto 1fr;
  min-height: 100vh;
  min-height: 100svh;
  background: var(--bg);
}

.login-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 20px 32px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 9px;
  font-size: 15px;
  font-weight: 600;
}

.brand-logo {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  background: var(--accent);
  color: var(--card);
}

.login-card {
  align-self: center;
  width: min(340px, calc(100% - 48px));
  margin: 24px auto 72px;
}

.login-heading {
  margin-bottom: 28px;
}

.login-heading h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  line-height: 1.4;
}

.login-heading p {
  margin: 8px 0 0;
  color: var(--muted);
  font-size: 13px;
}

.login-card :deep(.el-form-item) {
  margin-bottom: 20px;
}

.login-card :deep(.el-input__wrapper) {
  min-height: 42px;
  padding-right: 13px;
  padding-left: 13px;
}

.login-btn {
  width: 100%;
  height: 42px;
  margin-top: 8px;
}

.tip {
  margin: 20px 0 0;
  color: var(--muted);
  font-size: 12px;
  line-height: 1.7;
}

@media (max-width: 640px) {
  .login-header {
    padding: 16px 20px;
  }

  .login-heading h1 {
    font-size: 22px;
  }
}
</style>
