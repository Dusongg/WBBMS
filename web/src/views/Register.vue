<template>
  <div class="register-container">
    <el-card class="register-card">
      <template #header>
        <div class="card-header">
          <h2>用户注册</h2>
        </div>
      </template>

      <el-form
        :model="registerForm"
        :rules="rules"
        ref="registerFormRef"
        label-width="100px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="registerForm.username"
            placeholder="请输入用户名"
            clearable
          />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="registerForm.password"
            type="password"
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="registerForm.confirmPassword"
            type="password"
            placeholder="请再次输入密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="真实姓名" prop="real_name">
          <el-input
            v-model="registerForm.real_name"
            placeholder="请输入真实姓名"
            clearable
          />
        </el-form-item>
        <el-form-item label="身份证号" prop="id_card">
          <el-input
            v-model="registerForm.id_card"
            placeholder="请输入身份证号"
            clearable
          />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input
            v-model="registerForm.email"
            placeholder="请输入邮箱（可选）"
            clearable
          />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input
            v-model="registerForm.phone"
            placeholder="请输入手机号（可选）"
            clearable
          />
        </el-form-item>
        <el-form-item label="地址" prop="address">
          <el-input
            v-model="registerForm.address"
            placeholder="请输入地址（可选）"
            clearable
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            @click="handleRegister"
            style="width: 100%;"
          >
            注册
          </el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            type="text"
            @click="$router.push('/login')"
            style="width: 100%;"
          >
            已有账号？立即登录
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import axios from 'axios'

export default {
  name: 'Register',
  setup() {
    const router = useRouter()
    const registerFormRef = ref(null)
    const loading = ref(false)

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
      username: [
        { required: true, message: '请输入用户名', trigger: 'blur' }
      ],
      password: [
        { required: true, message: '请输入密码', trigger: 'blur' },
        { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
      ],
      confirmPassword: [
        { required: true, message: '请再次输入密码', trigger: 'blur' },
        { validator: validateConfirmPassword, trigger: 'blur' }
      ],
      real_name: [
        { required: true, message: '请输入真实姓名', trigger: 'blur' }
      ],
      id_card: [
        { required: true, message: '请输入身份证号', trigger: 'blur' }
      ]
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
      handleRegister
    }
  }
}
</script>

<style scoped>
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.register-card {
  width: 500px;
  max-width: 100%;
}

.card-header {
  text-align: center;
}

.card-header h2 {
  margin: 0;
  color: #409EFF;
}
</style>

