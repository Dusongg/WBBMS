import { ref, watch, onMounted, onUnmounted, nextTick } from 'vue'
import * as THREE from 'three'

export function useThemedBookScene(options = {}) {
  const {
    themeStorageKey = 'login-theme-mode',
    isBusyRef = null,
    trackPointer = true
  } = options

  const themeMode = ref(localStorage.getItem(themeStorageKey) === 'light' ? 'light' : 'dark')
  const canvasContainer = ref(null)
  const rightColumnRef = ref(null)

  let scene
  let camera
  let renderer
  let booksGroup
  let animationId
  let ambientLight
  let dirLight
  let fillLight
  let resizeObserver = null
  const books = []
  const cursor3d = { x: 0, y: 0 }

  const getTheme3dConfig = () => {
    if (themeMode.value === 'light') {
      return {
        bgColor: 0xeef3fa,
        fogNear: 18,
        fogFar: 78,
        ambientIntensity: 1.05,
        dirIntensity: 1.0,
        fillColor: 0xffffff,
        fillIntensity: 0.65,
        palette: [0x355070, 0x588157, 0xffffff, 0xdce3ed, 0xb56576, 0x6d597a]
      }
    }
    return {
      bgColor: 0x020617,
      fogNear: 10,
      fogFar: 60,
      ambientIntensity: 0.8,
      dirIntensity: 1.2,
      fillColor: 0xe0f7fa,
      fillIntensity: 0.5,
      palette: [0x1a1a1a, 0x1a1a1a, 0xffffff, 0xdddddd, 0xf3c4c2, 0xf3c4c2]
    }
  }

  const setTheme = (mode) => {
    if (mode !== 'dark' && mode !== 'light') return
    themeMode.value = mode
  }

  const updateSceneCursor = (x, y) => {
    const width = window.innerWidth || 1
    const height = window.innerHeight || 1
    cursor3d.x = x / width - 0.5
    cursor3d.y = y / height - 0.5
  }

  const handlePointerMove = (e) => {
    updateSceneCursor(e.clientX, e.clientY)
  }

  const applyThemeToScene = () => {
    if (!scene) return
    const cfg = getTheme3dConfig()
    scene.background = new THREE.Color(cfg.bgColor)
    scene.fog = new THREE.Fog(cfg.bgColor, cfg.fogNear, cfg.fogFar)
    if (ambientLight) ambientLight.intensity = cfg.ambientIntensity
    if (dirLight) dirLight.intensity = cfg.dirIntensity
    if (fillLight) {
      fillLight.color.setHex(cfg.fillColor)
      fillLight.intensity = cfg.fillIntensity
    }
    if (renderer) renderer.toneMappingExposure = themeMode.value === 'light' ? 1.05 : 1.0
  }

  const animate = () => {
    animationId = requestAnimationFrame(animate)
    if (!scene || !camera || !renderer || !booksGroup) return

    books.forEach((book) => {
      book.position.y += Math.sin(Date.now() * 0.001 + book.userData.randomOffset) * 0.002
      book.rotation.x += book.userData.rotSpeed
      book.rotation.y += book.userData.rotSpeed
    })

    const isBusy = !!(isBusyRef && isBusyRef.value)
    if (!isBusy) {
      booksGroup.rotation.x += (cursor3d.y * 0.15 - booksGroup.rotation.x) * 0.05
      booksGroup.rotation.y += (cursor3d.x * 0.15 - booksGroup.rotation.y) * 0.05
    }

    renderer.render(scene, camera)
  }

  const initScene = () => {
    if (!canvasContainer.value) return
    const width = window.innerWidth
    const height = window.innerHeight
    if (width === 0 || height === 0) return

    scene = new THREE.Scene()
    const cfg = getTheme3dConfig()
    scene.background = new THREE.Color(cfg.bgColor)
    scene.fog = new THREE.Fog(cfg.bgColor, cfg.fogNear, cfg.fogFar)

    camera = new THREE.PerspectiveCamera(75, width / height, 0.1, 1000)
    camera.position.set(0, 0, 14)

    renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true })
    renderer.setSize(width, height)
    renderer.shadowMap.enabled = true
    renderer.shadowMap.type = THREE.PCFSoftShadowMap
    renderer.toneMapping = THREE.ACESFilmicToneMapping
    renderer.toneMappingExposure = themeMode.value === 'light' ? 1.05 : 1.0
    canvasContainer.value.appendChild(renderer.domElement)

    ambientLight = new THREE.AmbientLight(0xffffff, cfg.ambientIntensity)
    scene.add(ambientLight)

    dirLight = new THREE.DirectionalLight(0xffffff, cfg.dirIntensity)
    dirLight.position.set(15, 30, 10)
    dirLight.castShadow = true
    dirLight.shadow.mapSize.width = 2048
    dirLight.shadow.mapSize.height = 2048
    scene.add(dirLight)

    fillLight = new THREE.DirectionalLight(cfg.fillColor, cfg.fillIntensity)
    fillLight.position.set(-15, 10, -10)
    scene.add(fillLight)

    booksGroup = new THREE.Group()
    scene.add(booksGroup)

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
      const book = createBook(cfg.palette[Math.floor(Math.random() * cfg.palette.length)])
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

  const handleResize = () => {
    if (!camera || !renderer) return
    const width = window.innerWidth
    const height = window.innerHeight
    if (width === 0 || height === 0) return
    camera.aspect = width / height
    camera.updateProjectionMatrix()
    renderer.setSize(width, height)
  }

  watch(themeMode, () => {
    localStorage.setItem(themeStorageKey, themeMode.value)
    applyThemeToScene()
  })

  onMounted(() => {
    window.addEventListener('resize', handleResize)
    if (trackPointer) {
      window.addEventListener('mousemove', handlePointerMove)
    }

    nextTick(() => {
      initScene()
      if (rightColumnRef.value && typeof ResizeObserver !== 'undefined') {
        resizeObserver = new ResizeObserver(handleResize)
        resizeObserver.observe(rightColumnRef.value)
      }
    })
  })

  onUnmounted(() => {
    window.removeEventListener('resize', handleResize)
    if (trackPointer) {
      window.removeEventListener('mousemove', handlePointerMove)
    }
    if (resizeObserver) resizeObserver.disconnect()
    if (animationId) cancelAnimationFrame(animationId)
  })

  return {
    themeMode,
    setTheme,
    canvasContainer,
    rightColumnRef,
    updateSceneCursor
  }
}
