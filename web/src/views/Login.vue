<template>
  <div id="loader" v-if="isLoadingScene">
    <div class="loader-content">
      <div class="loader-text">LIBRARY OS</div>
      <div class="loader-bar"></div>
    </div>
  </div>

  <div ref="canvasContainer" id="canvas-container"></div>

  <div class="login-section" ref="loginSectionRef">
    <el-card class="login-card glass-card" ref="loginCardRef">
      <template #header>
        <div class="card-header">
          <h2 class="brand-title">LIBRARY<span class="dot">.</span>OS</h2>
          <p class="brand-subtitle">EST. 2025 — CONCEPT TO REALITY</p>
        </div>
      </template>

      <el-form
        :model="loginForm"
        :rules="rules"
        ref="loginFormRef"
        label-width="0"
        class="design-form"
      >
        <el-form-item prop="username">
          <div class="input-label">IDENTITY</div>
          <el-input
            v-model="loginForm.username"
            placeholder="USERNAME OR ID"
            clearable
            class="design-input"
          />
        </el-form-item>
        <el-form-item prop="password">
          <div class="input-label">ACCESS KEY</div>
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="PASSWORD"
            show-password
            @keyup.enter="handleLogin"
            class="design-input"
          />
        </el-form-item>
        
        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            @click="handleLogin"
            class="login-btn-design"
          >
            {{ loading ? 'VERIFYING...' : 'ENTER SYSTEM' }}
          </el-button>
        </el-form-item>

        <el-form-item>
          <el-button
            type="text"
            @click="$router.push('/register')"
            class="register-link"
          >
            APPLY FOR MEMBERSHIP &rarr;
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>

  <div class="scroll-content"></div>
</template>

<script>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import axios from 'axios'
import * as THREE from 'three'
import gsap from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'
import { setToken, setUserInfo } from '../utils/auth'

export default {
  name: 'Login',
  setup() {
    gsap.registerPlugin(ScrollTrigger)
    const router = useRouter()
    
    const loginFormRef = ref(null)
    const canvasContainer = ref(null)
    const loginCardRef = ref(null)
    const loginSectionRef = ref(null)
    const loading = ref(false)
    const isLoadingScene = ref(true)
    
    const loginForm = reactive({
      username: '',
      password: ''
    })

    const rules = {
      username: [{ required: true, message: 'PLEASE ENTER IDENTITY', trigger: 'blur' }],
      password: [{ required: true, message: 'PLEASE ENTER PASSWORD', trigger: 'blur' }]
    }

    // --- Three.js Variables ---
    let scene, camera, renderer, booksGroup, animationId
    let books = []
    const cursor = { x: 0, y: 0 }

    // --- 3D Scene Init ---
    const initThreeJS = () => {
      scene = new THREE.Scene()
      // 背景保持高亮灰白，配合玻璃效果
      const bgColor = 0xF2F2F2 
      scene.background = new THREE.Color(bgColor)
      scene.fog = new THREE.Fog(bgColor, 10, 60)

      camera = new THREE.PerspectiveCamera(75, window.innerWidth / window.innerHeight, 0.1, 1000)
      camera.position.set(0, 0, 14)

      renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true })
      renderer.setSize(window.innerWidth, window.innerHeight)
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

      const fillLight = new THREE.DirectionalLight(0xE0F7FA, 0.5)
      fillLight.position.set(-15, 10, -10)
      scene.add(fillLight)

      booksGroup = new THREE.Group()
      scene.add(booksGroup)

      // HAL2 Palette
      const palette = [
        0x1A1A1A, 0x1A1A1A, // Black
        0xFFFFFF, // White
        0xDDDDDD, // Grey
        0xF3C4C2, 0xF3C4C2 // Pink
      ]

      const createBook = (color) => {
        const group = new THREE.Group()
        const coverMat = new THREE.MeshStandardMaterial({ color: color, roughness: 0.8, metalness: 0.1 })
        const cover = new THREE.Mesh(new THREE.BoxGeometry(1.1, 1.5, 0.25), coverMat)
        cover.castShadow = true
        cover.receiveShadow = true
        group.add(cover)

        const pagesMat = new THREE.MeshStandardMaterial({ color: 0xFFFFFF, roughness: 0.9 })
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

      // Scroll Animation
      gsap.timeline({
        scrollTrigger: {
          trigger: ".scroll-content",
          start: "top top",
          end: "bottom bottom",
          scrub: 1.2
        }
      })
      .to(camera.position, { x: 8, y: 4, z: 4, duration: 2 })
      .to(camera.rotation, { x: 0, y: 0.6, z: 0, duration: 2 }, "<")
      .to(camera.position, { x: -6, y: -3, z: -8, duration: 3 })
      .to(camera.rotation, { x: 0, y: -0.3, z: 0, duration: 3 }, "<")

      animate()
      setTimeout(() => { isLoadingScene.value = false }, 800)
    }

    const animate = () => {
      animationId = requestAnimationFrame(animate)
      const time = new THREE.Clock().getElapsedTime()

      books.forEach(book => {
        book.position.y += Math.sin(Date.now() * 0.001 + book.userData.randomOffset) * 0.002
        book.rotation.x += book.userData.rotSpeed
        book.rotation.y += book.userData.rotSpeed
      })

      if (!loading.value) {
        booksGroup.rotation.x += (cursor.y * 0.15 - booksGroup.rotation.x) * 0.05
        booksGroup.rotation.y += (cursor.x * 0.15 - booksGroup.rotation.y) * 0.05
      }

      renderer.render(scene, camera)
    }

    const onMouseMove = (event) => {
      cursor.x = event.clientX / window.innerWidth - 0.5
      cursor.y = event.clientY / window.innerHeight - 0.5
    }

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
               playSuccessAnimation(() => router.push('/'))
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

    // --- 修复后的成功动画 ---
    const playSuccessAnimation = (onComplete) => {
      ScrollTrigger.getAll().forEach(t => t.kill())
      
      if (loginCardRef.value) {
        gsap.to(loginCardRef.value.$el, {
          y: 50, opacity: 0, duration: 0.6, ease: "power3.in",
          onComplete: () => { if (loginSectionRef.value) loginSectionRef.value.style.display = 'none' }
        })
      }

      gsap.to(camera.position, { z: camera.position.z + 5, duration: 0.5, ease: "power2.out" })
      gsap.to(camera.position, { 
        z: -60, duration: 1.8, delay: 0.5, ease: "expo.in", onComplete: onComplete 
      })
      
      // ✅ 修复点：分开处理雾效属性和颜色对象
      gsap.to(scene.fog, { near: 0.1, far: 10, duration: 2, delay: 0.5 })
      gsap.to(scene.fog.color, { r: 1, g: 1, b: 1, duration: 2, delay: 0.5 })
      gsap.to(scene.background, { r: 1, g: 1, b: 1, duration: 2, delay: 0.5 })
    }

    onMounted(() => {
      initThreeJS()
      window.addEventListener('mousemove', onMouseMove)
      window.addEventListener('resize', () => {
        if(camera) {
          camera.aspect = window.innerWidth / window.innerHeight
          camera.updateProjectionMatrix()
          renderer.setSize(window.innerWidth, window.innerHeight)
        }
      })
    })

    onUnmounted(() => {
      cancelAnimationFrame(animationId)
      ScrollTrigger.getAll().forEach(t => t.kill())
    })

    return {
      loginForm, rules, loginFormRef, loading, isLoadingScene,
      handleLogin, canvasContainer, loginCardRef, loginSectionRef
    }
  }
}
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Oswald:wght@400;500;700&family=Inter:wght@300;400;600&display=swap');

#canvas-container { position: fixed; top: 0; left: 0; width: 100%; height: 100%; z-index: 1; }
.scroll-content { height: 400vh; position: relative; z-index: 2; pointer-events: none; }
.login-section {
  position: fixed; top: 0; left: 0; width: 100%; height: 100%;
  display: flex; justify-content: center; align-items: center;
  z-index: 10; pointer-events: none;
}

/* --- 磨砂玻璃核心样式 (Glassmorphism) --- */
.glass-card {
  width: 380px;
  pointer-events: auto;
  
  /* 背景：半透明白 + 高斯模糊 */
  background: rgba(255, 255, 255, 0.45) !important;
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  
  /* 边框：微弱的白色内描边，模拟玻璃厚度 */
  border: 1px solid rgba(255, 255, 255, 0.8) !important;
  border-radius: 24px !important; /* 大圆角 */
  
  /* 阴影：柔和的散射光 */
  box-shadow: 0 8px 32px 0 rgba(31, 38, 135, 0.15) !important;
  
  padding: 40px 30px;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.glass-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 12px 40px 0 rgba(31, 38, 135, 0.2) !important;
}

/* 标题区域 */
.card-header {
  text-align: left;
  margin-bottom: 30px;
  /* 分割线改为淡入淡出 */
  border-bottom: 1px solid rgba(0,0,0,0.1);
  padding-bottom: 20px;
}

.brand-title {
  font-family: 'Oswald', sans-serif;
  font-size: 32px; font-weight: 700; color: #1A1A1A;
  margin: 0; letter-spacing: 1px; line-height: 1;
}

.dot { color: #F3C4C2; }

.brand-subtitle {
  font-family: 'Inter', sans-serif;
  font-size: 10px; font-weight: 600; color: #666;
  margin-top: 8px; letter-spacing: 2px;
}

/* 表单区域 */
.design-form { margin-top: 20px; }

.input-label {
  font-family: 'Inter', sans-serif;
  font-size: 10px; font-weight: 700; color: #1A1A1A;
  margin-bottom: 6px; letter-spacing: 1px; opacity: 0.8;
}

/* Input 修改：更通透 */
:deep(.el-input__wrapper) {
  background-color: rgba(255, 255, 255, 0.3) !important; /* 极淡的背景 */
  box-shadow: none !important;
  border-bottom: 1px solid rgba(0,0,0,0.2) !important;
  border-radius: 0 !important;
  padding-left: 0 !important;
  transition: all 0.3s;
}

:deep(.el-input__wrapper.is-focus) {
  border-bottom: 1px solid #1A1A1A !important;
  background-color: rgba(255, 255, 255, 0.6) !important;
}

:deep(.el-input__inner) {
  font-family: 'Inter', sans-serif; font-weight: 500;
  color: #1A1A1A !important; font-size: 16px; letter-spacing: 0.5px;
}

/* 按钮设计：保留黑色形成对比，但增加圆角呼应玻璃 */
.login-btn-design {
  width: 100%; height: 50px; margin-top: 20px;
  background: #1A1A1A !important;
  border: none !important;
  border-radius: 12px !important; /* 适度圆角 */
  
  font-family: 'Oswald', sans-serif;
  font-size: 16px; font-weight: 500; letter-spacing: 2px; color: #fff;
  transition: all 0.3s;
  box-shadow: 0 4px 15px rgba(0,0,0,0.2);
}

.login-btn-design:hover {
  background: #F3C4C2 !important;
  color: #1A1A1A !important;
  transform: scale(1.02);
}

.register-link {
  width: 100%; margin-top: 10px;
  font-family: 'Inter', sans-serif; font-size: 12px; font-weight: 600;
  color: #666 !important; letter-spacing: 1px; text-align: center;
}
.register-link:hover { color: #1A1A1A !important; text-decoration: underline; }

/* Loader */
#loader {
  position: fixed; top: 0; left: 0; width: 100%; height: 100%;
  background: #fff; z-index: 999;
  display: flex; justify-content: center; align-items: center;
}
.loader-text { font-family: 'Oswald', sans-serif; font-size: 24px; font-weight: 700; color: #1A1A1A; margin-bottom: 10px; letter-spacing: 4px; }
.loader-bar { width: 100px; height: 2px; background: #1A1A1A; animation: expand 1.5s infinite ease-in-out; }
@keyframes expand { 0% { width: 0; } 50% { width: 100px; } 100% { width: 0; margin-left: 100px; } }
</style>