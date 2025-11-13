<template>
  <div class="system">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>用户管理</span>
          <el-button type="primary" @click="handleAdd">新增用户</el-button>
        </div>
      </template>

      <!-- 搜索栏 -->
      <div class="search-bar">
        <el-input
          v-model="searchKeyword"
          placeholder="请输入用户名、姓名、邮箱进行搜索"
          clearable
          @input="handleSearch"
          style="width: 400px; margin-right: 10px;"
        >
          <template #prefix>
            <el-icon><SearchIcon /></el-icon>
          </template>
        </el-input>
      </div>

      <!-- 用户表格 -->
      <el-table
        :data="userList"
        style="width: 100%; margin-top: 20px;"
        v-loading="loading"
        border
      >
        <el-table-column label="ID" width="80">
          <template #default="scope">
            {{ scope.row.id || scope.row.ID || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="real_name" label="真实姓名" width="120" />
        <el-table-column prop="email" label="邮箱" width="180" />
        <el-table-column prop="phone" label="手机号" width="120" />
        <el-table-column prop="role" label="角色" width="120">
          <template #default="scope">
            <el-tag :type="getRoleType(scope.row.role)">
              {{ getRoleText(scope.row.role) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
              {{ scope.row.status === 'active' ? '正常' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <el-button
              type="primary"
              size="small"
              @click="handleEdit(scope.row)"
            >
              编辑
            </el-button>
            <el-button
              type="danger"
              size="small"
              @click="handleDelete(scope.row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form
        :model="form"
        :rules="rules"
        ref="formRef"
        label-width="100px"
      >
        <el-form-item label="用户名" prop="username" v-if="!isEdit">
          <el-input v-model="form.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!isEdit">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="真实姓名" prop="real_name">
          <el-input v-model="form.real_name" placeholder="请输入真实姓名" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="form.phone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="form.role" placeholder="请选择角色" style="width: 100%;">
            <el-option label="系统管理员" value="admin" />
            <el-option label="图书管理员" value="librarian" />
            <el-option label="普通读者" value="reader" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-select v-model="form.status" placeholder="请选择状态" style="width: 100%;">
            <el-option label="正常" value="active" />
            <el-option label="停用" value="inactive" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import axios from 'axios'

export default {
  name: 'System',
  components: {
    SearchIcon: Search
  },
  setup() {
    const loading = ref(false)
    const userList = ref([])
    const total = ref(0)
    const currentPage = ref(1)
    const pageSize = ref(10)
    const searchKeyword = ref('')
    const dialogVisible = ref(false)
    const dialogTitle = ref('新增用户')
    const formRef = ref(null)
    const isEdit = ref(false)

    const form = reactive({
      id: null,
      username: '',
      password: '',
      real_name: '',
      email: '',
      phone: '',
      role: 'reader',
      status: 'active'
    })

    const rules = {
      username: [
        { required: true, message: '请输入用户名', trigger: 'blur' }
      ],
      password: [
        { required: true, message: '请输入密码', trigger: 'blur' },
        { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
      ],
      role: [
        { required: true, message: '请选择角色', trigger: 'change' }
      ],
      status: [
        { required: true, message: '请选择状态', trigger: 'change' }
      ]
    }

    const getRoleType = (role) => {
      const map = {
        admin: 'danger',
        librarian: 'warning',
        reader: 'success'
      }
      return map[role] || 'info'
    }

    const getRoleText = (role) => {
      const map = {
        admin: '系统管理员',
        librarian: '图书管理员',
        reader: '普通读者'
      }
      return map[role] || role
    }

    const fetchUserList = async () => {
      loading.value = true
      try {
        const params = {
          page: currentPage.value,
          pageSize: pageSize.value
        }
        if (searchKeyword.value) {
          params.keyword = searchKeyword.value
        }

        const response = await axios.get('/system/getUserList', { params })
        if (response.code === 200) {
          userList.value = response.data.list || []
          total.value = response.data.total || 0
        } else {
          ElMessage.error(response.msg || '获取数据失败')
        }
      } catch (error) {
        console.error('获取用户列表失败:', error)
        ElMessage.error('获取用户列表失败')
      } finally {
        loading.value = false
      }
    }

    const handleSearch = () => {
      currentPage.value = 1
      fetchUserList()
    }

    const handleSizeChange = (val) => {
      pageSize.value = val
      currentPage.value = 1
      fetchUserList()
    }

    const handlePageChange = (val) => {
      currentPage.value = val
      fetchUserList()
    }

    const handleAdd = () => {
      isEdit.value = false
      dialogTitle.value = '新增用户'
      resetForm()
      dialogVisible.value = true
    }

    const handleEdit = (row) => {
      isEdit.value = true
      dialogTitle.value = '编辑用户'
      Object.assign(form, {
        id: row.id,
        username: row.username,
        password: '',
        real_name: row.real_name || '',
        email: row.email || '',
        phone: row.phone || '',
        role: row.role,
        status: row.status
      })
      dialogVisible.value = true
    }

    const handleDelete = (row) => {
      ElMessageBox.confirm(`确定要删除用户 ${row.username} 吗？`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          const response = await axios.delete('/system/deleteUser', {
            data: { id: row.id }
          })
          if (response.code === 200) {
            ElMessage.success('删除成功')
            fetchUserList()
          } else {
            ElMessage.error(response.msg || '删除失败')
          }
        } catch (error) {
          console.error('删除失败:', error)
          ElMessage.error('删除失败')
        }
      }).catch(() => {})
    }

    const handleSubmit = async () => {
      if (!formRef.value) return

      await formRef.value.validate(async (valid) => {
        if (valid) {
          if (isEdit.value) {
            updateUser()
          } else {
            createUser()
          }
        }
      })
    }

    const createUser = async () => {
      try {
        const response = await axios.post('/system/createUser', form)
        if (response.code === 200) {
          ElMessage.success('创建成功')
          dialogVisible.value = false
          fetchUserList()
        } else {
          ElMessage.error(response.msg || '创建失败')
        }
      } catch (error) {
        console.error('创建失败:', error)
        ElMessage.error('创建失败')
      }
    }

    const updateUser = async () => {
      try {
        const response = await axios.put('/system/updateUser', form)
        if (response.code === 200) {
          ElMessage.success('更新成功')
          dialogVisible.value = false
          fetchUserList()
        } else {
          ElMessage.error(response.msg || '更新失败')
        }
      } catch (error) {
        console.error('更新失败:', error)
        ElMessage.error('更新失败')
      }
    }

    const resetForm = () => {
      Object.assign(form, {
        id: null,
        username: '',
        password: '',
        real_name: '',
        email: '',
        phone: '',
        role: 'reader',
        status: 'active'
      })
      if (formRef.value) {
        formRef.value.clearValidate()
      }
    }

    const handleDialogClose = () => {
      resetForm()
    }

    onMounted(() => {
      fetchUserList()
    })

    return {
      loading,
      userList,
      total,
      currentPage,
      pageSize,
      searchKeyword,
      dialogVisible,
      dialogTitle,
      form,
      rules,
      formRef,
      isEdit,
      getRoleType,
      getRoleText,
      handleSearch,
      handleSizeChange,
      handlePageChange,
      handleAdd,
      handleEdit,
      handleDelete,
      handleSubmit,
      handleDialogClose
    }
  }
}
</script>

<style scoped>
.system {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-bar {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>

