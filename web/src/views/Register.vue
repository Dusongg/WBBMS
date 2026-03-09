<template>
  <div class="animated-login-root" :class="`theme-${themeMode}`" ref="rightColumnRef">
    <div ref="canvasContainer" class="login-three-container"></div>

    <div class="theme-switch" role="group" aria-label="Theme mode">
      <button
        class="theme-switch-btn"
        :class="{ active: themeMode === 'dark' }"
        aria-label="切换暗色模式"
        title="暗色模式"
        @click="setTheme('dark')"
      >
        <svg class="theme-icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
          <path
            d="M20 14.5a8 8 0 1 1-10.5-10 7 7 0 1 0 10.5 10z"
            stroke="currentColor"
            stroke-width="1.8"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
      </button>
      <button
        class="theme-switch-btn"
        :class="{ active: themeMode === 'light' }"
        aria-label="切换亮色模式"
        title="亮色模式"
        @click="setTheme('light')"
      >
        <svg class="theme-icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
          <circle cx="12" cy="12" r="4.2" stroke="currentColor" stroke-width="1.8" />
          <path d="M12 2.5v2.2M12 19.3v2.2M4.7 4.7l1.6 1.6M17.7 17.7l1.6 1.6M2.5 12h2.2M19.3 12h2.2M4.7 19.3l1.6-1.6M17.7 6.3l1.6-1.6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
        </svg>
      </button>
    </div>

    <div class="animated-login-center register-center">
      <div class="brand-logo">
        <span class="brand-text">LIBRARY OS</span>
      </div>

      <div class="login-panel register-panel">
        <el-card class="login-card glass-card">
          <div class="panel-title">Create Account</div>
          <el-form
            :model="registerForm"
            :rules="rules"
            ref="registerFormRef"
            label-width="0"
            class="design-form"
          >
            <el-form-item prop="username">
              <el-input
                v-model="registerForm.username"
                placeholder="Username"
                clearable
                class="design-input"
              />
            </el-form-item>
            <el-form-item prop="password">
              <el-input
                v-model="registerForm.password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="Password (at least 6 characters)"
                class="design-input"
              >
                <template #suffix>
                  <span
                    class="password-toggle-icon"
                    role="button"
                    tabindex="0"
                    :aria-label="showPassword ? '隐藏密码' : '显示密码'"
                    @click="showPassword = !showPassword"
                    @keydown.enter.prevent="showPassword = !showPassword"
                  >
                    <svg
                      v-show="!showPassword"
                      class="toggle-svg"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="1.5"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    >
                      <path d="M12 4.5C7 4.5 3 8 2 12s4 7.5 10 7.5 10-3.5 11-7.5-4-7.5-11-7.5z" />
                      <circle cx="12" cy="12" r="3" />
                    </svg>
                    <svg
                      v-show="showPassword"
                      class="toggle-svg"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="1.5"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    >
                      <path d="M12 4.5C7 4.5 3 8 2 12s4 7.5 10 7.5 10-3.5 11-7.5-4-7.5-11-7.5z" />
                      <circle cx="12" cy="12" r="3" />
                      <line x1="2" y1="2" x2="22" y2="22" />
                    </svg>
                  </span>
                </template>
              </el-input>
            </el-form-item>
            <el-form-item prop="confirmPassword">
              <el-input
                v-model="registerForm.confirmPassword"
                :type="showConfirmPassword ? 'text' : 'password'"
                placeholder="Confirm password"
                class="design-input"
              >
                <template #suffix>
                  <span
                    class="password-toggle-icon"
                    role="button"
                    tabindex="0"
                    :aria-label="showConfirmPassword ? '隐藏确认密码' : '显示确认密码'"
                    @click="showConfirmPassword = !showConfirmPassword"
                    @keydown.enter.prevent="showConfirmPassword = !showConfirmPassword"
                  >
                    <svg
                      v-show="!showConfirmPassword"
                      class="toggle-svg"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="1.5"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    >
                      <path d="M12 4.5C7 4.5 3 8 2 12s4 7.5 10 7.5 10-3.5 11-7.5-4-7.5-11-7.5z" />
                      <circle cx="12" cy="12" r="3" />
                    </svg>
                    <svg
                      v-show="showConfirmPassword"
                      class="toggle-svg"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="1.5"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    >
                      <path d="M12 4.5C7 4.5 3 8 2 12s4 7.5 10 7.5 10-3.5 11-7.5-4-7.5-11-7.5z" />
                      <circle cx="12" cy="12" r="3" />
                      <line x1="2" y1="2" x2="22" y2="22" />
                    </svg>
                  </span>
                </template>
              </el-input>
            </el-form-item>
            <el-form-item prop="real_name">
              <el-input
                v-model="registerForm.real_name"
                placeholder="Full name"
                clearable
                class="design-input"
              />
            </el-form-item>
            <el-form-item prop="id_card">
              <el-input
                v-model="registerForm.id_card"
                placeholder="ID card number"
                clearable
                class="design-input"
              />
            </el-form-item>
            <el-form-item prop="email">
              <el-input
                v-model="registerForm.email"
                placeholder="Email (optional)"
                clearable
                class="design-input"
              />
            </el-form-item>
            <el-form-item prop="phone">
              <el-input
                v-model="registerForm.phone"
                placeholder="Phone number (optional)"
                clearable
                class="design-input"
              />
            </el-form-item>
            <el-form-item prop="address">
              <el-input
                v-model="registerForm.address"
                placeholder="Address (optional)"
                clearable
                class="design-input"
              />
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                :loading="loading"
                @click="handleRegister"
                class="login-btn-design"
              >
                {{ loading ? 'Submitting...' : 'Create Account' }}
              </el-button>
            </el-form-item>

            <div class="signup-row">
              <span>Already have an account?</span>
              <a @click.prevent="$router.push('/login')" href="#">Sign in</a>
            </div>
          </el-form>
        </el-card>
      </div>

      <div class="footer-links">
        <a href="#">Privacy Policy</a>
        <a href="#">Terms of Service</a>
        <a href="#">Contact</a>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import axios from 'axios'
import { useThemedBookScene } from '../composables/useThemedBookScene'

export default {
  name: 'Register',
  setup() {
    const router = useRouter()
    const registerFormRef = ref(null)
    const loading = ref(false)
    const showPassword = ref(false)
    const showConfirmPassword = ref(false)
    const { themeMode, setTheme, canvasContainer, rightColumnRef } = useThemedBookScene({
      isBusyRef: loading,
      trackPointer: true
    })

    const registerForm = reactive({
      username: '',
      password: '',
      confirmPassword: '',
      real_name: '',
      id_card: '',
      email: '',
      phone: '',
      address: ''
    })

    const validateConfirmPassword = (rule, value, callback) => {
      if (value !== registerForm.password) {
        callback(new Error('两次输入密码不一致'))
      } else {
        callback()
      }
    }

    const rules = {
      username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
      password: [
        { required: true, message: '请输入密码', trigger: 'blur' },
        { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
      ],
      confirmPassword: [
        { required: true, message: '请再次输入密码', trigger: 'blur' },
        { validator: validateConfirmPassword, trigger: 'blur' }
      ],
      real_name: [{ required: true, message: '请输入真实姓名', trigger: 'blur' }],
      id_card: [{ required: true, message: '请输入身份证号', trigger: 'blur' }]
    }

    const handleRegister = async () => {
      if (!registerFormRef.value) return

      await registerFormRef.value.validate(async (valid) => {
        if (valid) {
          loading.value = true
          try {
            const response = await axios.post('/auth/register', registerForm)
            if (response.code === 200) {
              ElMessage.success('注册成功，等待审核')
              router.push('/login')
            } else {
              ElMessage.error(response.msg || '注册失败')
            }
          } catch (error) {
            console.error('注册失败:', error)
            ElMessage.error('注册失败，请检查网络连接')
          } finally {
            loading.value = false
          }
        }
      })
    }

    return {
      registerForm,
      rules,
      registerFormRef,
      loading,
      showPassword,
      showConfirmPassword,
      themeMode,
      setTheme,
      canvasContainer,
      rightColumnRef,
      handleRegister
    }
  }
}
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;500;700&family=IBM+Plex+Sans:wght@300;400;600&display=swap');

.animated-login-root {
  --page-bg: #020617;
  --page-fg: #f9fafb;
  --brand-fg: rgba(249, 250, 251, 0.92);
  --footer-muted: rgba(249, 250, 251, 0.6);
  --footer-hover: #f9fafb;
  --login-fg: rgba(250, 250, 250, 0.98);
  --login-muted: rgba(148, 163, 184, 0.85);
  --login-border: rgba(255, 255, 255, 0.08);
  --login-focus: rgba(255, 255, 255, 0.2);
  --glass-bg: rgba(15, 23, 42, 0.25);
  --input-bg: rgba(255, 255, 255, 0.06);
  --input-bg-hover: rgba(255, 255, 255, 0.08);
  --input-bg-focus: rgba(255, 255, 255, 0.08);
  --button-bg: rgba(255, 255, 255, 0.1);
  --button-bg-hover: rgba(255, 255, 255, 0.14);
  --button-bg-active: rgba(255, 255, 255, 0.08);
  position: relative;
  min-height: 100vh;
  background: var(--page-bg);
}

.animated-login-root.theme-light {
  --page-bg: #eef3fa;
  --page-fg: #0f172a;
  --brand-fg: rgba(15, 23, 42, 0.92);
  --footer-muted: rgba(51, 65, 85, 0.74);
  --footer-hover: #0f172a;
  --login-fg: rgba(15, 23, 42, 0.96);
  --login-muted: rgba(71, 85, 105, 0.88);
  --login-border: rgba(15, 23, 42, 0.16);
  --login-focus: rgba(37, 99, 235, 0.38);
  --glass-bg: rgba(255, 255, 255, 0.64);
  --input-bg: rgba(255, 255, 255, 0.74);
  --input-bg-hover: rgba(255, 255, 255, 0.9);
  --input-bg-focus: rgba(255, 255, 255, 0.98);
  --button-bg: rgba(37, 99, 235, 0.14);
  --button-bg-hover: rgba(37, 99, 235, 0.22);
  --button-bg-active: rgba(37, 99, 235, 0.16);
}

.login-three-container {
  position: fixed;
  inset: 0;
  z-index: 0;
  width: 100%;
  height: 100%;
}

.login-three-container canvas {
  display: block;
  width: 100% !important;
  height: 100% !important;
}

.theme-switch {
  position: fixed;
  top: 1rem;
  right: 1rem;
  z-index: 30;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.35rem;
  border-radius: 999px;
  background: rgba(15, 23, 42, 0.18);
  border: 1px solid var(--login-border);
  backdrop-filter: blur(12px);
}

.theme-switch-btn {
  border: none;
  background: transparent;
  color: var(--login-muted);
  width: 34px;
  height: 34px;
  padding: 0;
  border-radius: 999px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.22s ease;
}

.theme-switch-btn:hover {
  color: var(--login-fg);
}

.theme-switch-btn.active {
  background: var(--button-bg-hover);
  color: var(--login-fg);
  box-shadow: inset 0 0 0 1px var(--login-border);
}

.theme-icon {
  width: 17px;
  height: 17px;
}

.animated-login-center {
  position: relative;
  z-index: 1;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem 1.5rem 3rem;
  color: var(--page-fg);
}

.register-center {
  padding-top: 4.4rem;
}

.brand-logo {
  display: flex;
  align-items: center;
  margin-bottom: 2rem;
  font-family: 'DM Sans', system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  font-weight: 600;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.brand-text {
  font-size: clamp(36px, 6vw, 72px);
  color: var(--brand-fg);
  opacity: 0.92;
}

.login-panel {
  --login-radius: 12px;
  --login-space: 1.1rem;
  --login-font: 'DM Sans', system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  width: 100%;
  max-width: 440px;
  flex-shrink: 0;
}

.register-panel {
  max-width: 460px;
}

.glass-card {
  background: var(--glass-bg) !important;
  border-radius: var(--login-radius);
  border: 1px solid var(--login-border);
  box-shadow: none;
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
}

.glass-card :deep(.el-card__body),
.glass-card :deep(.el-card) {
  background: transparent !important;
}

.glass-card :deep(.el-card__body) {
  padding: var(--login-space) 1.3rem 1.3rem;
}

.panel-title {
  font-family: var(--login-font);
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--login-fg);
  margin-bottom: var(--login-space);
  text-align: center;
  letter-spacing: 0.04em;
}

.design-form {
  font-family: var(--login-font);
}

.design-form :deep(.el-form-item) {
  margin-bottom: var(--login-space);
}

::deep(.el-input__wrapper) {
  background-color: var(--input-bg) !important;
  border-radius: var(--login-radius) !important;
  box-shadow: inset 0 0 0 1px var(--login-border) !important;
  padding: 0.56rem 0.95rem !important;
  transition: box-shadow 0.2s ease, background-color 0.2s ease;
}

::deep(.el-input__wrapper:hover) {
  background-color: var(--input-bg-hover) !important;
}

::deep(.el-input__wrapper.is-focus) {
  background-color: var(--input-bg-focus) !important;
  box-shadow: inset 0 0 0 1px var(--login-focus) !important;
}

::deep(.el-input__inner) {
  font-size: 0.875rem;
  color: var(--login-fg) !important;
}

::deep(.el-input__inner::placeholder) {
  color: var(--login-muted);
}

.password-toggle-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  cursor: pointer;
  color: var(--login-muted);
  transition: color 0.2s ease;
}

.password-toggle-icon:hover {
  color: var(--login-fg);
}

.password-toggle-icon:focus-visible {
  outline: 1px solid var(--login-focus);
  outline-offset: 2px;
}

.password-toggle-icon .toggle-svg {
  width: 100%;
  height: 100%;
  flex-shrink: 0;
}

.login-btn-design {
  width: 100%;
  height: 44px;
  border-radius: var(--login-radius) !important;
  border: 1px solid var(--login-border) !important;
  background: var(--button-bg) !important;
  font-family: var(--login-font);
  font-size: 0.8125rem;
  font-weight: 500;
  letter-spacing: 0.06em;
  color: var(--login-fg) !important;
  transition: background 0.2s ease, border-color 0.2s ease;
}

.login-btn-design:hover {
  background: var(--button-bg-hover) !important;
  border-color: var(--login-focus) !important;
}

.login-btn-design:active {
  background: var(--button-bg-active) !important;
}

.signup-row {
  display: flex;
  justify-content: center;
  gap: 0.35rem;
  font-size: 0.75rem;
  color: var(--login-muted);
  margin-top: var(--login-space);
}

.signup-row a {
  color: var(--login-fg);
  text-decoration: none;
}

.signup-row a:hover {
  text-decoration: underline;
}

.footer-links {
  display: flex;
  gap: 2rem;
  font-size: 0.78rem;
  color: var(--footer-muted);
  margin-top: 1.4rem;
}

.footer-links a {
  text-decoration: none;
  cursor: pointer;
  transition: color 0.2s ease;
}

.footer-links a:hover {
  color: var(--footer-hover);
}

@media (max-width: 640px) {
  .animated-login-center {
    padding: 1.4rem 1rem 2rem;
  }

  .register-center {
    padding-top: 4rem;
  }

  .login-panel,
  .register-panel {
    max-width: 100%;
  }

  .glass-card :deep(.el-card__body) {
    padding: 1rem 1.05rem 1.1rem;
  }

  .footer-links {
    gap: 1rem;
    flex-wrap: wrap;
    justify-content: center;
  }
}
</style>

