<template>
  <div class="animated-login-root" ref="rightColumnRef">
    <!-- 全屏 3D 书本背景 -->
    <div ref="canvasContainer" class="login-three-container"></div>

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
import { ref, reactive, onMounted, onUnmounted, computed, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import axios from 'axios'
import * as THREE from 'three'
import { setToken, setUserInfo } from '../utils/auth'

export default {
  name: 'Login',
  setup() {
    const router = useRouter()

    const loginFormRef = ref(null)
    const canvasContainer = ref(null)
    const rightColumnRef = ref(null)
    const loading = ref(false)
    const rememberMe = ref(false)
    const showPassword = ref(false)

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

    // --- Three.js 右侧 3D 书本背景 ---
    let scene, camera, renderer, booksGroup, animationId
    let books = []
    const cursor3d = { x: 0, y: 0 }

    const initThreeJS = () => {
      if (!canvasContainer.value) return

      const width = window.innerWidth
      const height = window.innerHeight
      if (width === 0 || height === 0) return

      scene = new THREE.Scene()
      const bgColor = 0x020617
      scene.background = new THREE.Color(bgColor)
      scene.fog = new THREE.Fog(bgColor, 10, 60)

      camera = new THREE.PerspectiveCamera(75, width / height, 0.1, 1000)
      camera.position.set(0, 0, 14)

      renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true })
      renderer.setSize(width, height)
      renderer.shadowMap.enabled = true
      renderer.shadowMap.type = THREE.PCFSoftShadowMap
      renderer.toneMapping = THREE.ACESFilmicToneMapping
      renderer.toneMappingExposure = 1.0
      canvasContainer.value.appendChild(renderer.domElement)

      const ambientLight = new THREE.AmbientLight(0xffffff, 0.8)
      scene.add(ambientLight)

      const dirLight = new THREE.DirectionalLight(0xffffff, 1.2)
      dirLight.position.set(15, 30, 10)
      dirLight.castShadow = true
      dirLight.shadow.mapSize.width = 2048
      dirLight.shadow.mapSize.height = 2048
      scene.add(dirLight)

      const fillLight = new THREE.DirectionalLight(0xe0f7fa, 0.5)
      fillLight.position.set(-15, 10, -10)
      scene.add(fillLight)

      booksGroup = new THREE.Group()
      scene.add(booksGroup)

      const palette = [
        0x1a1a1a, 0x1a1a1a,
        0xffffff,
        0xdddddd,
        0xf3c4c2, 0xf3c4c2
      ]

      const createBook = (color) => {
        const group = new THREE.Group()
        const coverMat = new THREE.MeshStandardMaterial({ color, roughness: 0.8, metalness: 0.1 })
        const cover = new THREE.Mesh(new THREE.BoxGeometry(1.1, 1.5, 0.25), coverMat)
        cover.castShadow = true
        cover.receiveShadow = true
        group.add(cover)
        const pagesMat = new THREE.MeshStandardMaterial({ color: 0xffffff, roughness: 0.9 })
        const pages = new THREE.Mesh(new THREE.BoxGeometry(1.05, 1.45, 0.22), pagesMat)
        pages.position.x = 0.05
        group.add(pages)
        return group
      }

      for (let i = 0; i < 90; i++) {
        const book = createBook(palette[Math.floor(Math.random() * palette.length)])
        book.position.set(
          (Math.random() - 0.5) * 60,
          (Math.random() - 0.5) * 40,
          (Math.random() - 0.5) * 70 - 10
        )
        book.rotation.set(Math.random() * Math.PI, Math.random() * Math.PI, 0)
        book.userData = {
          rotSpeed: 0.001 + Math.random() * 0.002,
          randomOffset: Math.random() * 100
        }
        booksGroup.add(book)
        books.push(book)
      }

      animate()
    }

    const animate = () => {
      animationId = requestAnimationFrame(animate)
      if (!scene || !camera || !renderer) return

      books.forEach((book) => {
        book.position.y += Math.sin(Date.now() * 0.001 + book.userData.randomOffset) * 0.002
        book.rotation.x += book.userData.rotSpeed
        book.rotation.y += book.userData.rotSpeed
      })

      if (!loading.value) {
        booksGroup.rotation.x += (cursor3d.y * 0.15 - booksGroup.rotation.x) * 0.05
        booksGroup.rotation.y += (cursor3d.x * 0.15 - booksGroup.rotation.y) * 0.05
      }

      renderer.render(scene, camera)
    }

    const handleResize = () => {
      if (!camera || !renderer) return
      const width = window.innerWidth
      const height = window.innerHeight
      if (width === 0 || height === 0) return
      camera.aspect = width / height
      camera.updateProjectionMatrix()
      renderer.setSize(width, height)
    }

    const handleMouseMove = (e) => {
      mouseX.value = e.clientX
      mouseY.value = e.clientY
      cursor3d.x = e.clientX / window.innerWidth - 0.5
      cursor3d.y = e.clientY / window.innerHeight - 0.5
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

    let resizeObserver = null

    onMounted(() => {
      window.addEventListener('mousemove', handleMouseMove)
      window.addEventListener('resize', handleResize)

      nextTick(() => {
        initThreeJS()
        if (rightColumnRef.value && typeof ResizeObserver !== 'undefined') {
          resizeObserver = new ResizeObserver(handleResize)
          resizeObserver.observe(rightColumnRef.value)
        }
      })

      const purpleTimer = setupBlink(isPurpleBlinking)
      const blackTimer = setupBlink(isBlackBlinking)

      onUnmounted(() => {
        window.removeEventListener('mousemove', handleMouseMove)
        window.removeEventListener('resize', handleResize)
        if (resizeObserver && rightColumnRef.value) {
          resizeObserver.disconnect()
        }
        if (animationId) cancelAnimationFrame(animationId)
        clearTimeout(purpleTimer)
        clearTimeout(blackTimer)
      })
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
  position: relative;
  min-height: 100vh;
  background: #020617;
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
  color: #f9fafb;
  transform: scale(1.12);
  transform-origin: center center;
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
  color: rgba(249, 250, 251, 0.6);
  margin-top: 2rem;
}

.footer-links a {
  text-decoration: none;
  cursor: pointer;
  transition: color 0.2s ease;
}

.footer-links a:hover {
  color: #f9fafb;
}

/* 登录框：简约设计 token + 仅保留透明高斯模糊 */
.login-panel {
  --login-fg: rgba(250, 250, 250, 0.98);
  --login-muted: rgba(148, 163, 184, 0.85);
  --login-border: rgba(255, 255, 255, 0.08);
  --login-focus: rgba(255, 255, 255, 0.2);
  --login-radius: 12px;
  --login-space: 1.25rem;
  --login-font: 'DM Sans', system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  width: 100%;
  max-width: 380px;
  flex-shrink: 0;
}

/* 仅保留透明 + 高斯模糊（毛玻璃），其余极简 */
.glass-card {
  background: rgba(15, 23, 42, 0.25) !important;
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
  background-color: rgba(255, 255, 255, 0.06) !important;
  border-radius: var(--login-radius) !important;
  box-shadow: inset 0 0 0 1px var(--login-border) !important;
  padding: 0.6rem 1rem !important;
  transition: box-shadow 0.2s ease, background-color 0.2s ease;
}

::deep(.el-input__wrapper:hover) {
  background-color: rgba(255, 255, 255, 0.08) !important;
}

::deep(.el-input__wrapper.is-focus) {
  background-color: rgba(255, 255, 255, 0.08) !important;
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
  background: rgba(255, 255, 255, 0.1) !important;
  font-family: var(--login-font);
  font-size: 0.8125rem;
  font-weight: 500;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--login-fg) !important;
  transition: background 0.2s ease, border-color 0.2s ease;
}

.login-btn-design:hover {
  background: rgba(255, 255, 255, 0.14) !important;
  border-color: var(--login-focus) !important;
}

.login-btn-design:active {
  background: rgba(255, 255, 255, 0.08) !important;
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
