<template>
  <div class="login-page">
    <div class="glow glow-a" />
    <div class="glow glow-b" />
    <div class="login-card">
      <div class="brand">
        <div class="brand-logo">
          <el-icon :size="26"><Checked /></el-icon>
        </div>
        <div>
          <h1>待办面板</h1>
          <p>VUE 3 · GIN · MYSQL</p>
        </div>
      </div>
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" size="large" @submit.prevent="submit">
        <el-form-item label="账号" prop="username">
          <el-input v-model="form.username" placeholder="请输入账号" :prefix-icon="User" autocomplete="username" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            placeholder="请输入密码"
            :prefix-icon="Lock"
            autocomplete="current-password"
            @keyup.enter="submit"
          />
        </el-form-item>
        <el-button class="login-btn" type="primary" native-type="submit" :loading="loading || autoLoading">
          登 录
        </el-button>
      </el-form>
      <p class="tip">默认管理员 admin / admin，登录后可在账号管理中修改密码</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Checked, Lock, User } from '@element-plus/icons-vue'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

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
  position: relative;
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 24px;
  overflow: hidden;
  background: var(--bg-glow), var(--bg);
}

.glow {
  position: absolute;
  border-radius: 50%;
  filter: blur(110px);
  opacity: 0.35;
  pointer-events: none;
}

.glow-a {
  width: 480px;
  height: 480px;
  left: -120px;
  top: -140px;
  background: var(--el-color-primary);
}

.glow-b {
  width: 420px;
  height: 420px;
  right: -100px;
  bottom: -120px;
  background: var(--col-doing);
  opacity: 0.22;
}

.login-card {
  position: relative;
  z-index: 1;
  width: min(400px, 100%);
  padding: 34px 32px 26px;
  border-radius: 22px;
  border: 1px solid var(--border);
  background: var(--login-card-bg);
  box-shadow: var(--shadow);
  backdrop-filter: blur(16px);
}

.brand {
  display: flex;
  align-items: center;
  gap: 13px;
  margin-bottom: 28px;
}

.brand-logo {
  width: 46px;
  height: 46px;
  border-radius: 14px;
  display: grid;
  place-items: center;
  color: #fff;
  background: linear-gradient(135deg, var(--el-color-primary), var(--col-doing));
  box-shadow: 0 8px 20px rgba(74, 158, 255, 0.32);
}

.brand h1 {
  margin: 0;
  font-size: 21px;
  letter-spacing: 0.5px;
}

.brand p {
  margin: 4px 0 0;
  font-size: 10px;
  letter-spacing: 2.5px;
  color: var(--muted);
}

.login-btn {
  width: 100%;
  margin-top: 6px;
}

.tip {
  margin: 18px 0 0;
  font-size: 12px;
  line-height: 1.6;
  color: var(--muted);
  text-align: center;
}
</style>
