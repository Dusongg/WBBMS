<template>
  <div class="animated-login-root" :class="`theme-${themeMode}`" ref="rightColumnRef">
    <!-- 全屏 3D 书本背景 -->
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

    <!-- 居中内容：角色区在上 + 登录框在下 -->
    <div class="animated-login-center">
      <div class="brand-logo">
        <span class="brand-text">LIBRARY OS</span>
      </div>

      <div class="character-stage">
        <!-- Purple tall rectangle -->
        <div ref="purpleRef" class="char char-purple" :style="purpleStyle">
          <div class="char-eyes" :style="purpleEyesStyle">
            <div
              class="eye eye-white"
              :class="{ 'eye-blink': isPurpleBlinking }"
            ></div>
            <div
              class="eye eye-white"
              :class="{ 'eye-blink': isPurpleBlinking }"
            ></div>
          </div>
        </div>

        <!-- Black tall rectangle -->
        <div ref="blackRef" class="char char-black" :style="blackStyle">
          <div class="char-eyes" :style="blackEyesStyle">
            <div
              class="eye eye-white"
              :class="{ 'eye-blink': isBlackBlinking }"
            ></div>
            <div
              class="eye eye-white"
              :class="{ 'eye-blink': isBlackBlinking }"
            ></div>
          </div>
        </div>

        <!-- Orange semi-circle -->
        <div ref="orangeRef" class="char char-orange" :style="orangeStyle">
          <div class="char-eyes" :style="orangeEyesStyle">
            <div class="pupil"></div>
            <div class="pupil"></div>
          </div>
        </div>

        <!-- Yellow character -->
        <div ref="yellowRef" class="char char-yellow" :style="yellowStyle">
          <div class="char-eyes" :style="yellowEyesStyle">
            <div class="pupil"></div>
            <div class="pupil"></div>
          </div>
          <div class="char-mouth" :style="yellowMouthStyle"></div>
        </div>
      </div>

      <div class="login-panel">
        <el-card class="login-card glass-card">
          <el-form
            :model="loginForm"
            :rules="rules"
            ref="loginFormRef"
            label-width="0"
            class="design-form"
          >
            <el-form-item prop="username">
              <el-input
                v-model="loginForm.username"
                placeholder="USERNAME OR ID"
                clearable
                class="design-input"
                @focus="isTyping = true"
                @blur="isTyping = false"
              />
            </el-form-item>

            <el-form-item prop="password">
              <el-input
                v-model="loginForm.password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="••••••••"
                class="design-input"
                @keyup.enter="handleLogin"
                @focus="isTyping = true"
                @blur="isTyping = false"
              >
                <template #suffix>
                  <span
                    class="password-toggle-icon"
                    role="button"
                    tabindex="0"
                    :aria-label="showPassword ? 'Hide password' : 'Show password'"
                    @click="showPassword = !showPassword"
                    @keydown.enter.prevent="showPassword = !showPassword"
                  >
                    <!-- Eye open: show password -->
                    <svg
                      v-show="!showPassword"
                      class="toggle-svg"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="1.5"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      aria-hidden="true"
                    >
                      <path d="M12 4.5C7 4.5 3 8 2 12s4 7.5 10 7.5 10-3.5 11-7.5-4-7.5-11-7.5z" />
                      <circle cx="12" cy="12" r="3" />
                    </svg>
                    <!-- Eye off: hide password -->
                    <svg
                      v-show="showPassword"
                      class="toggle-svg"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="1.5"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      aria-hidden="true"
                    >
                      <path d="M12 4.5C7 4.5 3 8 2 12s4 7.5 10 7.5 10-3.5 11-7.5-4-7.5-11-7.5z" />
                      <circle cx="12" cy="12" r="3" />
                      <line x1="2" y1="2" x2="22" y2="22" />
                    </svg>
                  </span>
                </template>
              </el-input>
            </el-form-item>

            <el-form-item>
              <div class="remember-row">
                <el-checkbox v-model="rememberMe">Remember for 30 days</el-checkbox>
                <a href="#" class="forget-link">Forgot password?</a>
              </div>
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                :loading="loading"
                @click="handleLogin"
                class="login-btn-design"
              >
                {{ loading ? 'SIGNING IN...' : 'ENTER SYSTEM' }}
              </el-button>
            </el-form-item>

            <div class="signup-row">
              <span>Don't have an account?</span>
              <a @click.prevent="$router.push('/register')" href="#">Sign up</a>
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
import { ref, reactive, onMounted, onUnmounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import axios from 'axios'
import { setToken, setUserInfo } from '../utils/auth'
import { useThemedBookScene } from '../composables/useThemedBookScene'

export default {
  name: 'Login',
  setup() {
    const router = useRouter()

    const loginFormRef = ref(null)
    const loading = ref(false)
    const rememberMe = ref(false)
    const showPassword = ref(false)
    const { themeMode, setTheme, canvasContainer, rightColumnRef, updateSceneCursor } = useThemedBookScene({
      isBusyRef: loading,
      trackPointer: false
    })

    const loginForm = reactive({
      username: '',
      password: ''
    })

    const rules = {
      username: [{ required: true, message: 'PLEASE ENTER IDENTITY', trigger: 'blur' }],
      password: [{ required: true, message: 'PLEASE ENTER PASSWORD', trigger: 'blur' }]
    }

    // --- 角色动画相关状态 ---
    const mouseX = ref(0)
    const mouseY = ref(0)

    const isPurpleBlinking = ref(false)
    const isBlackBlinking = ref(false)
    const isTyping = ref(false)
    const isLookingAtEachOther = ref(false)
    const isPurplePeeking = ref(false)

    const purpleRef = ref(null)
    const blackRef = ref(null)
    const yellowRef = ref(null)
    const orangeRef = ref(null)

    const handleMouseMove = (e) => {
      mouseX.value = e.clientX
      mouseY.value = e.clientY
      updateSceneCursor(e.clientX, e.clientY)
    }

    const setupBlink = (blinkRef) => {
      const getRandomBlinkInterval = () => Math.random() * 4000 + 3000
      const scheduleBlink = () => {
        const timeout = setTimeout(() => {
          blinkRef.value = true
          setTimeout(() => {
            blinkRef.value = false
            scheduleBlink()
          }, 150)
        }, getRandomBlinkInterval())
        return timeout
      }
      return scheduleBlink()
    }

    const calculatePosition = (elRef) => {
      if (!elRef.value) return { faceX: 0, faceY: 0, bodySkew: 0 }

      const rect = elRef.value.getBoundingClientRect()
      const centerX = rect.left + rect.width / 2
      const centerY = rect.top + rect.height / 3

      const deltaX = mouseX.value - centerX
      const deltaY = mouseY.value - centerY

      const faceX = Math.max(-15, Math.min(15, deltaX / 20))
      const faceY = Math.max(-10, Math.min(10, deltaY / 30))
      const bodySkew = Math.max(-6, Math.min(6, -deltaX / 120))

      return { faceX, faceY, bodySkew }
    }

    const purplePos = computed(() => calculatePosition(purpleRef))
    const blackPos = computed(() => calculatePosition(blackRef))
    const yellowPos = computed(() => calculatePosition(yellowRef))
    const orangePos = computed(() => calculatePosition(orangeRef))

    // 互相对视
    watch(isTyping, (val) => {
      if (val) {
        isLookingAtEachOther.value = true
        const t = setTimeout(() => {
          isLookingAtEachOther.value = false
        }, 800)
        return () => clearTimeout(t)
      } else {
        isLookingAtEachOther.value = false
      }
    })

    // 偷窥密码
    watch(
      () => ({ pwd: loginForm.password, visible: showPassword.value }),
      ({ pwd, visible }) => {
        if (pwd && visible) {
          isPurplePeeking.value = true
          const t = setTimeout(() => {
            isPurplePeeking.value = false
          }, 800)
          return () => clearTimeout(t)
        } else {
          isPurplePeeking.value = false
        }
      },
      { deep: true }
    )

    // --- 角色样式计算（大致对应 React 中的行为） ---
    const purpleStyle = computed(() => {
      const p = purplePos.value
      const baseHeight = 400
      const isTypingOrHiding = isTyping.value || (loginForm.password && !showPassword.value)

      return {
        left: '70px',
        width: '180px',
        height: isTypingOrHiding ? baseHeight + 40 + 'px' : baseHeight + 'px',
        backgroundColor: '#6C3FF5',
        borderRadius: '10px 10px 0 0',
        zIndex: 1,
        transformOrigin: 'bottom center',
        transform: isTypingOrHiding
          ? `skewX(${(p.bodySkew || 0) - 12}deg) translateX(40px)`
          : `skewX(${p.bodySkew || 0}deg)`,
        transition: 'transform 0.7s ease-in-out, height 0.7s ease-in-out'
      }
    })

    const purpleEyesStyle = computed(() => {
      const p = purplePos.value
      let left = 45 + p.faceX
      let top = 40 + p.faceY

      if (loginForm.password && showPassword.value) {
        left = isPurplePeeking.value ? 20 : 20
        top = isPurplePeeking.value ? 35 : 35
      } else if (isLookingAtEachOther.value) {
        left = 55
        top = 65
      }

      return {
        left: `${left}px`,
        top: `${top}px`,
        gap: '32px'
      }
    })

    const blackStyle = computed(() => {
      const p = blackPos.value
      let transform = `skewX(${p.bodySkew || 0}deg)`

      if (loginForm.password && showPassword.value) {
        transform = 'skewX(0deg)'
      } else if (isLookingAtEachOther.value) {
        transform = `skewX(${(p.bodySkew || 0) * 1.5 + 10}deg) translateX(20px)`
      } else if (isTyping.value || (loginForm.password && !showPassword.value)) {
        transform = `skewX(${(p.bodySkew || 0) * 1.5}deg)`
      }

      return {
        left: '240px',
        width: '120px',
        height: '310px',
        backgroundColor: '#2D2D2D',
        borderRadius: '8px 8px 0 0',
        zIndex: 2,
        transformOrigin: 'bottom center',
        transform,
        transition: 'transform 0.7s ease-in-out'
      }
    })

    const blackEyesStyle = computed(() => {
      const p = blackPos.value
      let left = 26 + p.faceX
      let top = 32 + p.faceY

      if (loginForm.password && showPassword.value) {
        left = 10
        top = 28
      } else if (isLookingAtEachOther.value) {
        left = 32
        top = 12
      }

      return {
        left: `${left}px`,
        top: `${top}px`,
        gap: '24px'
      }
    })

    const orangeStyle = computed(() => {
      const p = orangePos.value
      let transform = `skewX(${p.bodySkew || 0}deg)`
      if (loginForm.password && showPassword.value) {
        transform = 'skewX(0deg)'
      }
      return {
        left: '0px',
        width: '240px',
        height: '200px',
        backgroundColor: '#FF9B6B',
        borderRadius: '120px 120px 0 0',
        zIndex: 3,
        transformOrigin: 'bottom center',
        transform,
        transition: 'transform 0.7s ease-in-out'
      }
    })

    const orangeEyesStyle = computed(() => {
      const p = orangePos.value
      let left = 82 + (p.faceX || 0)
      let top = 90 + (p.faceY || 0)

      if (loginForm.password && showPassword.value) {
        left = 50
        top = 85
      }

      return {
        left: `${left}px`,
        top: `${top}px`,
        gap: '32px'
      }
    })

    const yellowStyle = computed(() => {
      const p = yellowPos.value
      let transform = `skewX(${p.bodySkew || 0}deg)`
      if (loginForm.password && showPassword.value) {
        transform = 'skewX(0deg)'
      }
      return {
        left: '310px',
        width: '140px',
        height: '230px',
        backgroundColor: '#E8D754',
        borderRadius: '70px 70px 0 0',
        zIndex: 4,
        transformOrigin: 'bottom center',
        transform,
        transition: 'transform 0.7s ease-in-out'
      }
    })

    const yellowEyesStyle = computed(() => {
      const p = yellowPos.value
      let left = 52 + (p.faceX || 0)
      let top = 40 + (p.faceY || 0)

      if (loginForm.password && showPassword.value) {
        left = 20
        top = 35
      }

      return {
        left: `${left}px`,
        top: `${top}px`,
        gap: '24px'
      }
    })

    const yellowMouthStyle = computed(() => {
      const p = yellowPos.value
      let left = 40 + (p.faceX || 0)
      let top = 88 + (p.faceY || 0)

      if (loginForm.password && showPassword.value) {
        left = 10
        top = 88
      }

      return {
        left: `${left}px`,
        top: `${top}px`
      }
    })

    const handleLogin = async () => {
      if (!loginFormRef.value) return
      await loginFormRef.value.validate(async (valid) => {
        if (valid) {
          loading.value = true
          try {
            const response = await axios.post('/auth/login', loginForm)
            if (response.code === 200) {
              setToken(response.data.token)
              setUserInfo(response.data)
              router.push('/')
            } else {
              ElMessage.error(response.msg || '验证失败')
              loading.value = false
            }
          } catch (error) {
            ElMessage.error('网络连接错误')
            loading.value = false
          }
        }
      })
    }

    onMounted(() => {
      window.addEventListener('mousemove', handleMouseMove)
      purpleTimer = setupBlink(isPurpleBlinking)
      blackTimer = setupBlink(isBlackBlinking)
    })

    let purpleTimer = null
    let blackTimer = null
    onUnmounted(() => {
      window.removeEventListener('mousemove', handleMouseMove)
      if (purpleTimer) clearTimeout(purpleTimer)
      if (blackTimer) clearTimeout(blackTimer)
    })

    return {
      loginForm,
      rules,
      loginFormRef,
      canvasContainer,
      rightColumnRef,
      loading,
      rememberMe,
      showPassword,
      themeMode,
      setTheme,
      // 动画状态
      isPurpleBlinking,
      isBlackBlinking,
      isTyping,
      isLookingAtEachOther,
      isPurplePeeking,
      purpleRef,
      blackRef,
      yellowRef,
      orangeRef,
      purpleStyle,
      purpleEyesStyle,
      blackStyle,
      blackEyesStyle,
      orangeStyle,
      orangeEyesStyle,
      yellowStyle,
      yellowEyesStyle,
      yellowMouthStyle,
      handleLogin
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

/* 全屏 3D 书本背景 */
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

/* 居中内容：角色区 + 登录框，单列垂直排列 */
.animated-login-center {
  position: relative;
  z-index: 1;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem 1.5rem 3rem;
  padding-top: 4vh;
  color: var(--page-fg);
  transform: scale(1.12);
  transform-origin: center center;
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
  font-family: 'DM Sans', system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
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

.brand-logo {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 5rem;
  font-family: 'DM Sans', system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  font-weight: 600;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.brand-icon {
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 0.9rem;
  border: 1px solid rgba(248, 250, 252, 0.24);
  background: rgba(15, 23, 42, 0.75);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.1rem;
}

.brand-text {
  font-size: 80px;
  color: var(--brand-fg);
  opacity: 0.9;
}

.character-stage {
  position: relative;
  width: 550px;
  height: 400px;
  margin: 0 auto 1.5rem;
  flex-shrink: 0;
}

.char {
  position: absolute;
  bottom: 0;
}

.char-eyes {
  position: absolute;
  display: flex;
  align-items: center;
}

.eye {
  border-radius: 999px;
  transition: height 0.15s ease-out;
}

.eye-white {
  width: 18px;
  height: 18px;
  background-color: #ffffff;
}

.eye-blink {
  height: 2px !important;
}

.pupil {
  width: 12px;
  height: 12px;
  border-radius: 999px;
  background-color: #2d2d2d;
}

.char-mouth {
  position: absolute;
  width: 80px;
  height: 4px;
  background-color: #2d2d2d;
  border-radius: 999px;
}

.footer-links {
  display: flex;
  gap: 2rem;
  font-size: 0.78rem;
  color: var(--footer-muted);
  margin-top: 2rem;
}

.footer-links a {
  text-decoration: none;
  cursor: pointer;
  transition: color 0.2s ease;
}

.footer-links a:hover {
  color: var(--footer-hover);
}

/* 登录框：简约设计 token + 仅保留透明高斯模糊 */
.login-panel {
  --login-radius: 12px;
  --login-space: 1.25rem;
  --login-font: 'DM Sans', system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  width: 100%;
  max-width: 380px;
  flex-shrink: 0;
}

/* 仅保留透明 + 高斯模糊（毛玻璃），其余极简 */
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
  padding: var(--login-space) 1.5rem 1.5rem;
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
  padding: 0.6rem 1rem !important;
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

.password-toggle-icon:focus {
  outline: none;
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

.remember-row {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.75rem;
  color: var(--login-muted);
}

.forget-link {
  color: var(--login-fg);
  text-decoration: none;
  font-size: 0.75rem;
}

.forget-link:hover {
  text-decoration: underline;
  color: var(--login-fg);
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
  text-transform: uppercase;
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

@media (max-width: 640px) {
  .animated-login-center {
    padding: 1.5rem 1rem 2rem;
  }

  .login-panel {
    max-width: 100%;
  }

  .glass-card :deep(.el-card__body) {
    padding: 1rem 1.25rem 1.25rem;
  }

  .character-stage {
    width: 100%;
    max-width: 550px;
    height: 320px;
  }
}
</style>
