<template>
  <div class="book-gallery-page" :class="{ 'borrow-view': viewMode === 'borrow' }">
    <!-- 顶部操作栏 (仅在图书管理模式下显示) -->
    <div v-if="viewMode !== 'borrow'" class="top-navbar">
      <div class="nav-left">
        <el-popover
          v-model:visible="showCategoryDialog"
          placement="bottom-start"
          :width="300"
          trigger="click"
        >
          <template #reference>
            <el-button circle class="nav-icon-btn">
              <el-icon><Menu /></el-icon>
            </el-button>
          </template>
          <div class="category-popover">
            <div class="category-popover-header">
              <div class="category-popover-title">选择分类</div>
              <el-button 
                v-if="hasAdminOrLibrarianRole()"
                circle 
                size="small" 
                class="edit-category-btn"
                :class="{ active: isEditMode }"
                @click="handleEditModeToggle"
              >
                <el-icon><Edit /></el-icon>
              </el-button>
            </div>
            <!-- 新增分类输入框 -->
            <div v-if="showAddCategoryInput" class="add-category-input-wrapper">
              <el-input
                v-model="newCategoryName"
                placeholder="请输入分类名称，按回车确认"
                size="small"
                @keyup.enter="handleAddCategory"
                @blur="handleAddCategoryBlur"
                ref="newCategoryInputRef"
                class="add-category-input"
              />
            </div>
            <el-checkbox-group v-model="selectedCategories" @change="handleCategoryChange">
              <div
                v-for="(category, index) in availableCategories"
                :key="category"
                class="category-item"
              >
                <el-checkbox
                  :label="category"
                  class="category-checkbox"
                >
                  <span
                    v-if="!isEditMode || !hasAdminOrLibrarianRole()"
                    class="category-name"
                  >
                    {{ category }}
                  </span>
                  <el-input
                    v-else
                    :model-value="editingCategoryNames[category] || category"
                    @update:model-value="(val) => { editingCategoryNames[category] = val }"
                    size="small"
                    class="category-edit-input"
                    @blur="handleCategoryNameBlur(category)"
                    @keyup.enter="handleCategoryNameSave(category)"
                    @click.stop
                    ref="categoryEditInputRefs"
                  />
                </el-checkbox>
                <el-button
                  v-if="hasAdminOrLibrarianRole() && isEditMode"
                  circle
                  size="small"
                  type="danger"
                  class="delete-category-btn"
                  @click.stop="handleDeleteCategory(category)"
                >
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </el-checkbox-group>
            <!-- 在最后一个类别下方添加加号按钮 -->
            <div v-if="hasAdminOrLibrarianRole() && isEditMode" class="add-category-button-wrapper">
              <el-button 
                circle 
                size="small" 
                class="add-category-btn"
                @click="handleAddCategoryClick"
              >
                <el-icon><Plus /></el-icon>
              </el-button>
            </div>
            <div class="category-popover-divider"></div>
            <div class="display-mode-toggle">
              <div class="toggle-switch-container">
                <span class="toggle-label" :class="{ active: displayMode === 'merged' }">合并显示</span>
                <el-switch
                  v-model="displayMode"
                  active-value="separated"
                  inactive-value="merged"
                  class="toggle-switch"
                />
                <span class="toggle-label" :class="{ active: displayMode === 'separated' }">分类显示</span>
              </div>
            </div>
            <div class="category-popover-footer">
              <el-button size="small" @click="showCategoryDialog = false">取消</el-button>
              <el-button type="primary" size="small" @click="handleConfirmCategories">确定</el-button>
            </div>
          </div>
        </el-popover>
      </div>
      
      <div class="nav-center">
        <div class="nav-tabs">
          <button 
            class="nav-tab" 
            :class="{ active: !showListView }"
            @click="showListView = false"
          >
            图书
          </button>
          <div class="nav-divider"></div>
          <button 
            class="nav-tab"
            :class="{ active: showListView }"
            @click="showListView = true"
          >
            列表
          </button>
        </div>
      </div>

      <div class="nav-right" :class="{ 'no-add-btn': !hasAdminOrLibrarianRole() }">
        <transition name="search-input-slide">
          <el-input
            v-if="showSearchInput"
            v-model="searchKeyword"
            placeholder="搜索图书..."
            clearable
            @input="handleSearch"
            @blur="handleSearchBlur"
            class="top-search-input"
            :class="{ 'no-add-btn': !hasAdminOrLibrarianRole() }"
            ref="searchInputRef"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </transition>
        <el-button 
          circle 
          class="nav-icon-btn search-btn" 
          :class="{ 'search-active': showSearchInput }"
          @click="handleSearchClick"
        >
          <el-icon><Search /></el-icon>
        </el-button>
        
        <!-- 消息按钮 -->
        <el-popover
          v-model:visible="showMessagePopover"
          placement="bottom-end"
          :width="400"
          trigger="click"
          popper-class="message-popover"
        >
          <template #reference>
            <el-badge :value="unreadCount" :hidden="unreadCount === 0" :max="99" class="message-badge">
              <el-button 
                circle 
                class="nav-icon-btn message-btn" 
                :class="{ 'has-unread': unreadCount > 0 }"
              >
                <el-icon><Bell /></el-icon>
              </el-button>
            </el-badge>
          </template>
          
          <!-- 消息下拉列表 -->
          <div class="message-dropdown">
            <div class="message-header">
              <span class="message-title">站内消息</span>
              <el-button 
                v-if="messages.length > 0" 
                text 
                size="small" 
                @click="handleMarkAllRead"
              >
                全部已读
              </el-button>
            </div>
            
            <div class="message-list" v-loading="messageLoading">
              <div
                v-for="message in messages"
                :key="message.id || message.ID"
                class="message-item"
                :class="{ 'unread': !message.is_read }"
                @click="handleMessageClick(message)"
              >
                <div class="message-icon">
                  <el-icon v-if="message.type === 'reservation'" color="#67C23A"><Ticket /></el-icon>
                  <el-icon v-else-if="message.type === 'borrow'" color="#409EFF"><Reading /></el-icon>
                  <el-icon v-else-if="message.type === 'overdue'" color="#F56C6C"><Warning /></el-icon>
                  <el-icon v-else color="#909399"><Bell /></el-icon>
                </div>
                <div class="message-content">
                  <div class="message-title-row">
                    <span class="message-item-title">{{ message.title }}</span>
                    <span class="message-time">{{ formatMessageTime(message.created_at) }}</span>
                  </div>
                  <div class="message-text">{{ message.content }}</div>
                </div>
                <el-button
                  v-if="!message.is_read"
                  circle
                  size="small"
                  class="mark-read-btn"
                  @click.stop="handleMarkRead(message.id || message.ID)"
                  title="标记为已读"
                >
                  <el-icon><Check /></el-icon>
                </el-button>
              </div>
              
              <div v-if="messages.length === 0 && !messageLoading" class="empty-message">
                <el-icon size="48" color="#C0C4CC"><Bell /></el-icon>
                <p>暂无消息</p>
              </div>
            </div>
            
            <div v-if="messages.length > 0" class="message-footer">
              <el-button text @click="handleLoadMoreMessages">查看更多</el-button>
            </div>
          </div>
        </el-popover>
        
        <el-button
          v-if="hasAdminOrLibrarianRole()"
          circle
          class="nav-icon-btn"
          @click="handleAdd"
        >
          <el-icon><Plus /></el-icon>
        </el-button>
        <el-button circle class="nav-icon-btn">
          <el-icon><MoreFilled /></el-icon>
        </el-button>
      </div>
    </div>

    <!-- 列表视图 -->
    <div v-if="showListView" class="list-view-container">
      <div class="list-view">
        <el-table
          :data="bookList"
          style="width: 100%;"
          v-loading="loading"
          border
        >
          <el-table-column label="ID" width="80">
            <template #default="scope">
              {{ scope.row.id || scope.row.ID || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="title" label="书名" min-width="150" />
          <el-table-column prop="author" label="作者" width="120" />
          <el-table-column label="分类" width="200">
            <template #default="scope">
              <span v-if="scope.row.categories && scope.row.categories.length > 0">
                {{ scope.row.categories.map(c => c.name).join('、') }}
              </span>
              <span v-else-if="scope.row.category">
                {{ scope.row.category }}
              </span>
              <span v-else>未分类</span>
            </template>
          </el-table-column>
          <el-table-column prop="available_stock" label="可借库存" width="100" />
          <el-table-column label="操作" width="180" fixed="right" v-if="hasAdminOrLibrarianRole()">
            <template #default="scope">
              <el-button type="primary" size="small" @click="handleEdit(scope.row)">编辑</el-button>
              <el-button type="danger" size="small" @click="handleDelete(scope.row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <!-- 3D轮播视图 -->
    <div v-else class="carousel-view">
      <!-- 合并显示模式 -->
      <template v-if="displayMode === 'merged'">
        <div v-if="currentCategoryBooks && currentCategoryBooks.books && currentCategoryBooks.books.length > 0" class="category-section">
          <div
            class="carousel-container"
            @wheel="handleWheel"
            @mousemove="handleMouseMove"
            @mouseleave="handleMouseUp"
            @mousedown="handleMouseDown"
            @mouseup="handleMouseUp"
            @touchstart="handleTouchStart"
            @touchmove="handleTouchMove"
          >
            <div class="carousel">
              <div
                v-for="(book, index) in currentCategoryBooks.books"
                :key="book.id || index"
                class="book-card"
                :class="{ 'active': index === currentCategoryBooks.centerIndex }"
                :style="getBookStyle(index, currentCategoryBooks.centerIndex)"
                :data-book-id="book.id"
                @click="handleBookClick(book, index)"
              >
                <div class="book-cover-card">
                  <div class="book-cover-placeholder">
                    <img
                      v-if="book.cover_image"
                      :src="book.cover_image"
                      :alt="book.title"
                      referrerpolicy="no-referrer"
                      @error="handleImageError"
                    />
                    <div v-else class="book-cover-fallback">
                      <div class="book-cover-circle"></div>
                    </div>
                  </div>
                  <!-- 借阅/点赞/收藏操作栏 (仅在图书管理模式下显示) -->
                  <div v-if="book && book.id && viewMode !== 'borrow'" class="book-action-bar" @click.stop>
                    <el-tooltip
                      :content="getBorrowButtonTitle(book.id)"
                      placement="top"
                      :show-after="300"
                      :hide-after="0"
                      effect="dark"
                      raw-content
                    >
                      <button 
                        class="action-btn borrow-btn"
                        :class="{ 
                          active: borrowStatus[book.id]?.isBorrowed && borrowStatus[book.id]?.status === 'borrowed',
                          pending: borrowStatus[book.id]?.isBorrowed && borrowStatus[book.id]?.status === 'pending',
                          reserved: reservationStatus[book.id]?.isReserved && reservationStatus[book.id]?.status === 'pending',
                          'reserved-available': reservationStatus[book.id]?.isReserved && reservationStatus[book.id]?.status === 'available'
                        }"
                        :disabled="actionLoading[book.id]?.borrow"
                        @click="handleToggleBorrow($event, book.id)"
                      >
                        <svg class="action-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                          <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z" 
                                :fill="borrowStatus[book.id]?.isBorrowed || reservationStatus[book.id]?.isReserved ? 'currentColor' : 'none'"
                                stroke="currentColor" 
                                stroke-width="2" 
                                stroke-linecap="round" 
                                stroke-linejoin="round"/>
                          <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z" 
                                :fill="borrowStatus[book.id]?.isBorrowed || reservationStatus[book.id]?.isReserved ? 'currentColor' : 'none'"
                                stroke="currentColor" 
                                stroke-width="2" 
                                stroke-linecap="round" 
                                stroke-linejoin="round"/>
                        </svg>
                      </button>
                    </el-tooltip>
                    <button 
                      class="action-btn like-btn"
                      :class="{ active: likeStatus[book.id]?.isLiked }"
                      :disabled="actionLoading[book.id]?.like"
                      @click="handleToggleLike($event, book.id)"
                      :title="`${likeStatus[book.id]?.likeCount || 0} 个赞`"
                    >
                      <svg class="action-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" 
                              :fill="likeStatus[book.id]?.isLiked ? 'currentColor' : 'none'"
                              stroke="currentColor" 
                              stroke-width="2" 
                              stroke-linecap="round" 
                              stroke-linejoin="round"/>
                      </svg>
                    </button>
                    <button 
                      class="action-btn favorite-btn"
                      :class="{ active: favoriteStatus[book.id]?.isFavorited }"
                      :disabled="actionLoading[book.id]?.favorite"
                      @click="handleToggleFavorite($event, book.id)"
                      :title="`${favoriteStatus[book.id]?.favoriteCount || 0} 个收藏`"
                    >
                      <svg class="action-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" 
                              :fill="favoriteStatus[book.id]?.isFavorited ? 'currentColor' : 'none'"
                              stroke="currentColor" 
                              stroke-width="2" 
                              stroke-linecap="round" 
                              stroke-linejoin="round"/>
                      </svg>
                    </button>
                  </div>
                </div>
              </div>
            </div>
            <!-- 拖动提示条 -->
            <div class="drag-indicator">
              <div class="drag-handle"></div>
            </div>
          </div>
        </div>
      </template>
      
      <!-- 分类显示模式 -->
      <template v-else>
        <div
          v-for="(category, categoryIndex) in categorizedBooks"
          :key="categoryIndex"
          :id="`category-section-${categoryIndex}`"
          class="category-section"
        >
          <h2 class="category-title" :data-status="
            category.name === '已借阅' ? 'borrowed' : 
            category.name === '已逾期' ? 'overdue' : 
            category.name === '待审批' ? 'pending' : 
            category.name === '预约' ? 'reserved' : ''
          ">{{ category.name }}</h2>
          <div
            class="carousel-container"
            @wheel="handleWheel($event, categoryIndex)"
            @mousemove="handleMouseMove($event, categoryIndex)"
            @mouseleave="handleMouseUp(categoryIndex)"
            @mousedown="handleMouseDown($event, categoryIndex)"
            @mouseup="handleMouseUp(categoryIndex)"
            @touchstart="handleTouchStart($event, categoryIndex)"
            @touchmove="handleTouchMove($event, categoryIndex)"
          >
            <div class="carousel">
              <div
                v-for="(book, index) in category.books"
                :key="book.id || index"
                class="book-card"
                :class="{ 'active': index === category.centerIndex }"
                :data-book-id="book.id"
                :style="getBookStyle(index, category.centerIndex)"
                @click="handleBookClick(book, categoryIndex, index)"
              >
                <div class="book-cover-card">
                  <div class="book-cover-placeholder">
                    <img
                      v-if="book.cover_image"
                      :src="book.cover_image"
                      :alt="book.title"
                      referrerpolicy="no-referrer"
                      @error="handleImageError"
                    />
                    <div v-else class="book-cover-fallback">
                      <div class="book-cover-circle"></div>
                    </div>
                  </div>
                  <!-- 借阅/点赞/收藏操作栏 (仅在图书管理模式下显示) -->
                  <div v-if="book && book.id && viewMode !== 'borrow'" class="book-action-bar" @click.stop>
                    <el-tooltip
                      :content="getBorrowButtonTitle(book.id)"
                      placement="top"
                      :show-after="300"
                      :hide-after="0"
                      effect="dark"
                      raw-content
                    >
                      <button 
                        class="action-btn borrow-btn"
                        :class="{ 
                          active: borrowStatus[book.id]?.isBorrowed && borrowStatus[book.id]?.status === 'borrowed',
                          pending: borrowStatus[book.id]?.isBorrowed && borrowStatus[book.id]?.status === 'pending',
                          reserved: reservationStatus[book.id]?.isReserved && reservationStatus[book.id]?.status === 'pending',
                          'reserved-available': reservationStatus[book.id]?.isReserved && reservationStatus[book.id]?.status === 'available'
                        }"
                        :disabled="actionLoading[book.id]?.borrow"
                        @click="handleToggleBorrow($event, book.id)"
                      >
                        <svg class="action-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                          <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z" 
                                :fill="borrowStatus[book.id]?.isBorrowed || reservationStatus[book.id]?.isReserved ? 'currentColor' : 'none'"
                                stroke="currentColor" 
                                stroke-width="2" 
                                stroke-linecap="round" 
                                stroke-linejoin="round"/>
                          <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z" 
                                :fill="borrowStatus[book.id]?.isBorrowed || reservationStatus[book.id]?.isReserved ? 'currentColor' : 'none'"
                                stroke="currentColor" 
                                stroke-width="2" 
                                stroke-linecap="round" 
                                stroke-linejoin="round"/>
                        </svg>
                      </button>
                    </el-tooltip>
                    <button 
                      class="action-btn like-btn"
                      :class="{ active: likeStatus[book.id]?.isLiked }"
                      :disabled="actionLoading[book.id]?.like"
                      @click="handleToggleLike($event, book.id)"
                      :title="`${likeStatus[book.id]?.likeCount || 0} 个赞`"
                    >
                      <svg class="action-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" 
                              :fill="likeStatus[book.id]?.isLiked ? 'currentColor' : 'none'"
                              stroke="currentColor" 
                              stroke-width="2" 
                              stroke-linecap="round" 
                              stroke-linejoin="round"/>
                      </svg>
                    </button>
                    <button 
                      class="action-btn favorite-btn"
                      :class="{ active: favoriteStatus[book.id]?.isFavorited }"
                      :disabled="actionLoading[book.id]?.favorite"
                      @click="handleToggleFavorite($event, book.id)"
                      :title="`${favoriteStatus[book.id]?.favoriteCount || 0} 个收藏`"
                    >
                      <svg class="action-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" 
                              :fill="favoriteStatus[book.id]?.isFavorited ? 'currentColor' : 'none'"
                              stroke="currentColor" 
                              stroke-width="2" 
                              stroke-linecap="round" 
                              stroke-linejoin="round"/>
                      </svg>
                    </button>
                  </div>
                </div>
              </div>
            </div>
            <!-- 拖动提示条 -->
            <div class="drag-indicator">
              <div class="drag-handle"></div>
            </div>
          </div>
        </div>
        <!-- 分类导航按钮 -->
        <div v-if="categorizedBooks.length > 1" class="category-navigation">
          <button
            v-if="currentCategoryPage > 0"
            class="nav-button"
            @click="scrollToCategory(currentCategoryPage - 1)"
          >
            ↑
          </button>
          <button
            v-if="currentCategoryPage < categorizedBooks.length - 1"
            class="nav-button"
            @click="scrollToCategory(currentCategoryPage + 1)"
          >
            ↓
          </button>
        </div>
      </template>
    </div>



    <!-- 图书详情对话框 -->
    <el-dialog
      v-model="bookDetailVisible"
      :title="selectedBook?.title"
      width="600px"
      class="book-detail-dialog"
      :close-on-click-modal="true"
      :close-on-press-escape="true"
      :modal="true"
      :lock-scroll="false"
      @close="handleBookDetailClose"
      @closed="handleBookDetailClosed"
    >
      <div v-if="selectedBook" class="book-detail">
        <div class="detail-cover">
          <img
            :src="selectedBook.cover_image || getDefaultCover()"
            :alt="selectedBook.title"
            referrerpolicy="no-referrer"
            @error="handleImageError"
          />
        </div>
        <div class="detail-info">
          <p><strong>作者：</strong>{{ selectedBook.author }}</p>
          <p><strong>出版社：</strong>{{ selectedBook.publisher || '未知' }}</p>
          <p><strong>ISBN：</strong>{{ selectedBook.isbn }}</p>
          <p><strong>分类：</strong>
            <span v-if="selectedBook.categories && selectedBook.categories.length > 0">
              {{ selectedBook.categories.map(c => c.name).join('、') }}
            </span>
            <span v-else-if="selectedBook.category">
              {{ selectedBook.category }}
            </span>
            <span v-else>未分类</span>
          </p>
          <p><strong>价格：</strong>¥{{ selectedBook.price?.toFixed(2) || '0.00' }}</p>
          <p><strong>可借库存：</strong>{{ selectedBook.available_stock || 0 }}</p>
          <p><strong>描述：</strong>{{ selectedBook.description || '暂无描述' }}</p>
        </div>
      </div>
      <template #footer>
        <el-button @click="handleBookDetailClose">关闭</el-button>
        <el-button
          v-if="hasAdminOrLibrarianRole()"
          type="primary"
          @click="handleEditFromDetail"
        >
          编辑
        </el-button>
      </template>
    </el-dialog>

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
        <el-form-item label="书名" prop="title">
          <el-input v-model="form.title" placeholder="请输入书名" />
        </el-form-item>
        <el-form-item label="作者" prop="author">
          <el-input v-model="form.author" placeholder="请输入作者" />
        </el-form-item>
        <el-form-item label="出版社" prop="publisher">
          <el-input v-model="form.publisher" placeholder="请输入出版社" />
        </el-form-item>
        <el-form-item label="出版日期" prop="publish_date">
          <el-input v-model="form.publish_date" placeholder="请输入出版日期，如：2024-01-01" />
        </el-form-item>
        <el-form-item label="ISBN" prop="isbn">
          <el-input v-model="form.isbn" placeholder="请输入ISBN" />
        </el-form-item>
        <el-form-item label="分类" prop="category_ids">
          <el-select 
            v-model="form.category_ids" 
            placeholder="请选择分类（可多选）" 
            multiple
            clearable
            collapse-tags
            filterable
            style="width: 100%;"
            :value-key="'id'"
          >
            <el-option
              v-for="category in categoryList"
              :key="`category-${category.id}`"
              :label="category.name"
              :value="category.id"
              :disabled="false"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="封面图片" prop="cover_image">
          <el-input v-model="form.cover_image" placeholder="请输入封面图片URL" />
        </el-form-item>
        <el-form-item label="价格" prop="price">
          <el-input-number
            v-model="form.price"
            :precision="2"
            :min="0"
            placeholder="请输入价格"
            style="width: 100%;"
          />
        </el-form-item>
        <el-form-item label="总库存" prop="total_stock">
          <el-input-number
            v-model="form.total_stock"
            :min="0"
            placeholder="请输入总库存"
            style="width: 100%;"
          />
        </el-form-item>
        <el-form-item label="可借库存" prop="available_stock">
          <el-input-number
            v-model="form.available_stock"
            :min="0"
            placeholder="请输入可借库存"
            style="width: 100%;"
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="4"
            placeholder="请输入描述"
          />
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
import { ref, reactive, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  Plus,
  Menu,
  MoreFilled,
  StarFilled,
  List,
  Edit,
  Delete,
  Bell,
  Ticket,
  Reading,
  Warning,
  Check
} from '@element-plus/icons-vue'
import axios from 'axios'
import { hasAdminOrLibrarianRole } from '../utils/auth'
import { toggleLike, batchGetLikeStatus } from '@/api/like'
import { toggleFavorite, batchGetFavoriteStatus } from '@/api/favorite'

export default {
  name: 'BookList',
  components: {
    Search,
    Plus,
    Menu,
    MoreFilled,
    StarFilled,
    List,
    Edit,
    Delete,
    Bell,
    Ticket,
    Reading,
    Warning,
    Check
  },
  setup() {
    const route = useRoute()
    const router = useRouter()
    const loading = ref(false)
    const bookList = ref([])
    const searchKeyword = ref('')
    const showListView = ref(false)
    const showSearchInput = ref(false)
    const viewMode = ref('all') // 'all' 或 'borrow'
    
    // 消息相关状态
    const showMessagePopover = ref(false)
    const messages = ref([])
    const unreadCount = ref(0)
    const messageLoading = ref(false)
    const messagePage = ref(1)
    const messagePageSize = ref(10)
    
    // 点赞/收藏/借阅状态管理
    const likeStatus = ref({}) // { bookId: { isLiked: boolean, likeCount: number } }
    const favoriteStatus = ref({}) // { bookId: { isFavorited: boolean, favoriteCount: number } }
    const borrowStatus = ref({}) // { bookId: { isBorrowed: boolean, borrowRecordId: number, status: 'pending'|'borrowed' } }
    const reservationStatus = ref({}) // { bookId: { isReserved: boolean, reservationId: number, status: 'pending'|'available' } }
    const actionLoading = ref({}) // { bookId: { like: boolean, favorite: boolean, borrow: boolean } }
    const searchInputRef = ref(null)
    const bookDetailVisible = ref(false)
    const selectedBook = ref(null)
    const dialogVisible = ref(false)
    const dialogTitle = ref('新增图书')
    const formRef = ref(null)
    const isEdit = ref(false)

    // 当前选中的分类（多个）
    const selectedCategories = ref([])
    
    // 分类选择对话框显示状态
    const showCategoryDialog = ref(false)
    
    // 新增分类输入框显示状态
    const showAddCategoryInput = ref(false)
    
    // 新增分类名称
    const newCategoryName = ref('')
    
    // 新增分类输入框引用
    const newCategoryInputRef = ref(null)
    
    // 编辑模式状态
    const isEditMode = ref(false)
    
    // 正在编辑的分类名称映射
    const editingCategoryNames = ref({})
    
    // 分类编辑输入框引用
    const categoryEditInputRefs = ref([])
    
    // 显示模式：'merged' 合并显示，'separated' 分类显示
    const displayMode = ref('merged')
    
    // 监听 displayMode 变化
    watch(displayMode, () => {
      // displayMode变化时的处理逻辑
    })
    
    // 当前分类的中心索引（合并模式）
    const currentCenterIndex = ref(0)
    
    // 每个分类的中心索引（分类模式）
    const categoryCenterIndices = ref({})
    
    // 拖拽相关状态（合并模式）
    const isDragging = ref(false)
    const dragStartX = ref(0)
    
    // 拖拽相关状态（分类模式）
    const isDraggingByCategory = ref({})
    const dragStartXByCategory = ref({})
    
    // 滚轮防抖相关状态（合并模式）
    const wheelTimer = ref(null)
    
    // 滚轮防抖相关状态（分类模式）
    const wheelTimers = ref({})
    
    // 当前分类页面索引（分类模式）
    const currentCategoryPage = ref(0)
    
    // 借阅视图的分类数据（待审批、已借阅）
    const borrowCategorizedBooks = ref([])

    const form = reactive({
      id: null,
      title: '',
      author: '',
      publisher: '',
      publish_date: '',
      isbn: '',
      price: 0,
      category: '', // 保留用于兼容
      category_ids: [], // 分类ID列表
      cover_image: '',
      total_stock: 0,
      available_stock: 0,
      description: ''
    })

    // 分类列表（从后端获取）
    const categoryList = ref([])

    const rules = {
      title: [
        { required: true, message: '请输入书名', trigger: 'blur' }
      ],
      author: [
        { required: true, message: '请输入作者', trigger: 'blur' }
      ],
      isbn: [
        { required: true, message: '请输入ISBN', trigger: 'blur' }
      ]
    }

    // 获取所有可用的分类（从后端获取的分类列表）
    const availableCategories = computed(() => {
      if (categoryList.value.length > 0) {
        return categoryList.value.map(cat => cat.name)
      }
      
      // 如果没有从后端获取到分类，则从图书中提取（兼容旧数据）
      const allCategories = new Set()
      bookList.value.forEach(book => {
        if (book.categories && book.categories.length > 0) {
          book.categories.forEach(cat => {
            allCategories.add(cat.name)
          })
        } else if (book.category && book.category.trim()) {
          allCategories.add(book.category.trim())
        } else {
          allCategories.add('其他')
        }
      })
      
      // 定义标准分类顺序
      const standardCategories = ['国内', '国外', '儿童']
      const categories = [...standardCategories, ...Array.from(allCategories).filter(cat => !standardCategories.includes(cat)), '其他']
      
      // 去重并保持顺序
      return [...new Set(categories)]
    })

    // 获取当前选中分类的图书（支持多个分类，合并模式）
    const currentCategoryBooks = computed(() => {
      if (!selectedCategories.value || selectedCategories.value.length === 0) {
        return null
      }
      
      // 合并多个分类的图书
      const books = (bookList.value || []).filter(book => {
        // 优先使用新的分类系统（categories）
        if (book && book.categories && book.categories.length > 0) {
          return book.categories.some(cat => {
            return selectedCategories.value.includes(cat.name)
          })
        }
        // 兼容旧数据（category字段）
        const bookCategory = book && book.category ? book.category.trim() : ''
        return selectedCategories.value.some(cat => {
          return bookCategory === cat || (!bookCategory && cat === '其他')
        })
      })
      
      if (!books || books.length === 0) {
        return null
      }
      
      // 确保 centerIndex 在有效范围内
      let validCenterIndex = currentCenterIndex.value
      if (validCenterIndex === undefined || validCenterIndex === null || validCenterIndex >= books.length) {
        validCenterIndex = Math.floor(books.length / 2)
        currentCenterIndex.value = validCenterIndex
      } else if (validCenterIndex < 0) {
        validCenterIndex = 0
        currentCenterIndex.value = 0
      }
      
      const result = {
        books: books || [],  // 确保books永远是数组
        centerIndex: validCenterIndex
      }
      
      return result
    })

    // 获取分类显示的图书（分类模式）
    const categorizedBooks = computed(() => {
      // 如果是借阅视图模式，直接返回borrowCategorizedBooks
      if (viewMode.value === 'borrow' && borrowCategorizedBooks.value.length > 0) {
        return borrowCategorizedBooks.value
      }
      
      // 在分类显示模式下，使用选中的分类；如果没有选中，则使用所有可用分类
      let categoriesToShow = []
      if (displayMode.value === 'separated') {
        // 分类显示模式：显示所有选中的分类，如果没有选中则显示所有分类
        if (selectedCategories.value && selectedCategories.value.length > 0) {
          categoriesToShow = selectedCategories.value
        } else {
          categoriesToShow = availableCategories.value
        }
      } else {
        // 合并显示模式：只使用选中的分类
        if (!selectedCategories.value || selectedCategories.value.length === 0) {
          return []
        }
        categoriesToShow = selectedCategories.value
      }
      
      if (!categoriesToShow || categoriesToShow.length === 0) {
        return []
      }
      
      const result = categoriesToShow.map(cat => {
        const books = (bookList.value || []).filter(book => {
          // 优先使用新的分类系统（categories）
          if (book && book.categories && book.categories.length > 0) {
            return book.categories.some(c => c.name === cat)
          }
          // 兼容旧数据（category字段）
          const bookCategory = book && book.category ? book.category.trim() : ''
          return bookCategory === cat || (!bookCategory && cat === '其他')
        })
        
        // 确保中心索引在有效范围内
        let validCenterIndex = categoryCenterIndices.value[cat]
        if (validCenterIndex === undefined || validCenterIndex === null || validCenterIndex >= books.length) {
          validCenterIndex = books.length > 0 ? Math.floor(books.length / 2) : 0
          categoryCenterIndices.value[cat] = validCenterIndex
        } else if (validCenterIndex < 0) {
          validCenterIndex = 0
          categoryCenterIndices.value[cat] = 0
        }
        
        return {
          name: cat,
          books: books || [],
          centerIndex: validCenterIndex
        }
      }).filter(cat => cat.books && cat.books.length > 0)

      return result
    })

    const getDefaultCover = () => {
      // 返回一个透明的 1x1 像素图片的 data URI，避免网络请求
      return 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMSIgaGVpZ2h0PSIxIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjxyZWN0IHdpZHRoPSIxMDAlIiBoZWlnaHQ9IjEwMCUiIGZpbGw9InRyYW5zcGFyZW50Ii8+PC9zdmc+'
    }

    const handleImageError = (e) => {
      // 避免无限循环：如果图片已经加载失败，隐藏图片元素
      if (e.target.dataset.errorHandled) {
        e.target.style.display = 'none'
        return
      }
      // 标记已处理，避免重复处理
      e.target.dataset.errorHandled = 'true'
      // 隐藏图片，显示占位符
      e.target.style.display = 'none'
    }

    // 根据索引获取卡片渐变颜色类
    const getBookCardClass = (index, centerIndex) => {
      const colors = [
        'gradient-pink',
        'gradient-red',
        'gradient-teal',
        'gradient-brown',
        'gradient-gray',
        'gradient-rose',
        'gradient-blue',
        'gradient-green',
        'gradient-dark',
        'gradient-yellow',
        'gradient-black'
      ]
      return colors[index % colors.length]
    }

    const getBookStyle = (index, centerIndex) => {
      // 安全检查：确保 index 和 centerIndex 都是有效数字
      if (index === undefined || index === null || centerIndex === undefined || centerIndex === null) {
        return {
          transform: 'translateX(0%) rotateY(0deg) scale(0.8)',
          zIndex: 1,
          opacity: 0.5
        }
      }
      
      const diff = index - centerIndex
      const absDistance = Math.abs(diff)
      
      // 中心卡片
      if (diff === 0) {
        return {
          transform: 'translateX(0%) rotateY(0deg) scale(1.1)',
          zIndex: 50,
          opacity: 1,
          filter: 'brightness(1.2)'
        }
      }
      
      // 左侧卡片
      if (diff < 0) {
        const offset = -30 * absDistance - 10
        const rotation = 50 + (absDistance * 5)
        const scale = Math.max(0.6, 1 - absDistance * 0.15)
        return {
          transform: `translateX(${offset}%) rotateY(${rotation}deg) scale(${scale})`,
          zIndex: 50 - absDistance,
          opacity: Math.max(0.3, 1 - absDistance * 0.2)
        }
      }
      
      // 右侧卡片 - 和左侧卡片对称，z-index递减，让右侧第一个卡片在中心卡片下方，但能漏出右边一角
      // 调整offset和旋转角度，让右边角更明显
      const offset = 30 * absDistance + 15 // 增加offset让右边角更明显
      // 右侧第一个卡片旋转角度稍微小一点，让右边角更明显
      const rotation = absDistance === 1 ? -45 : -50 - (absDistance * 5)
      const scale = Math.max(0.6, 1 - absDistance * 0.15)
      // 右侧第一个卡片zIndex设为49（在中心卡片50下方），确保能漏出右边一角
      return {
        transform: `translateX(${offset}%) rotateY(${rotation}deg) scale(${scale})`,
        zIndex: 50 - absDistance, // 右侧第一个卡片zIndex是49，确保在中心卡片下方但能漏出右边角
        opacity: Math.max(0.3, 1 - absDistance * 0.2)
      }
    }

    // 合并模式的滚动函数
    const scrollLeft = () => {
      // 如果 currentCenterIndex 是 undefined，初始化为中间位置
      if (currentCenterIndex.value === undefined || currentCenterIndex.value === null) {
        if (currentCategoryBooks.value && currentCategoryBooks.value.books && currentCategoryBooks.value.books.length > 0) {
          currentCenterIndex.value = Math.floor(currentCategoryBooks.value.books.length / 2)
        } else {
          currentCenterIndex.value = 0
        }
      }
      
      if (currentCategoryBooks.value && currentCenterIndex.value > 0) {
        currentCenterIndex.value = currentCenterIndex.value - 1
      }
    }

    const scrollRight = () => {
      // 如果 currentCenterIndex 是 undefined，初始化为中间位置
      if (currentCenterIndex.value === undefined || currentCenterIndex.value === null) {
        if (currentCategoryBooks.value && currentCategoryBooks.value.books && currentCategoryBooks.value.books.length > 0) {
          currentCenterIndex.value = Math.floor(currentCategoryBooks.value.books.length / 2)
        } else {
          currentCenterIndex.value = 0
        }
      }
      
      if (currentCategoryBooks.value && currentCenterIndex.value < currentCategoryBooks.value.books.length - 1) {
        currentCenterIndex.value = currentCenterIndex.value + 1
      }
    }

    // 分类模式的滚动函数
    const scrollLeftByCategory = (categoryIndex) => {
      const category = categorizedBooks.value[categoryIndex]
      if (category && category.centerIndex > 0) {
        categoryCenterIndices.value[category.name] = category.centerIndex - 1
      }
    }

    const scrollRightByCategory = (categoryIndex) => {
      const category = categorizedBooks.value[categoryIndex]
      if (category && category.centerIndex < category.books.length - 1) {
        categoryCenterIndices.value[category.name] = category.centerIndex + 1
      }
    }
    
    // 鼠标拖拽功能（合并模式）
    const handleMouseDown = (event, categoryIndex) => {
      if (displayMode.value === 'merged') {
        isDragging.value = true
        dragStartX.value = event.clientX
      } else {
        isDraggingByCategory.value[categoryIndex] = true
        dragStartXByCategory.value[categoryIndex] = event.clientX
      }
    }
    
    const handleMouseUp = (categoryIndex) => {
      if (displayMode.value === 'merged') {
        isDragging.value = false
      } else if (categoryIndex !== undefined) {
        isDraggingByCategory.value[categoryIndex] = false
      }
    }
    
    const handleTouchStart = (event, categoryIndex) => {
      if (event.touches && event.touches.length > 0) {
        if (displayMode.value === 'merged') {
          dragStartX.value = event.touches[0].clientX
        } else {
          dragStartXByCategory.value[categoryIndex] = event.touches[0].clientX
        }
      }
    }
    
    const handleTouchMove = (event, categoryIndex) => {
      if (!event.touches || event.touches.length === 0) return
      
      if (displayMode.value === 'merged') {
        if (!currentCategoryBooks.value) return
        
        const diff = event.touches[0].clientX - dragStartX.value
        
        if (Math.abs(diff) > 50) {
          if (diff > 0 && currentCenterIndex.value > 0) {
            scrollLeft()
          } else if (diff < 0 && currentCenterIndex.value < currentCategoryBooks.value.books.length - 1) {
            scrollRight()
          }
          dragStartX.value = event.touches[0].clientX
        }
      } else {
        const category = categorizedBooks.value[categoryIndex]
        if (!category) return
        
        const diff = event.touches[0].clientX - dragStartXByCategory.value[categoryIndex]
        
        if (Math.abs(diff) > 50) {
          if (diff > 0 && category.centerIndex > 0) {
            scrollLeftByCategory(categoryIndex)
          } else if (diff < 0 && category.centerIndex < category.books.length - 1) {
            scrollRightByCategory(categoryIndex)
          }
          dragStartXByCategory.value[categoryIndex] = event.touches[0].clientX
        }
      }
    }

    const handleWheel = (event, categoryIndex) => {
      // 如果对话框打开，不处理滚动视图的滚动
      if (bookDetailVisible.value || dialogVisible.value) {
        return
      }
      
      // 只有在对话框关闭时才阻止默认滚动行为
      event.preventDefault()
      
      if (displayMode.value === 'merged') {
        // 如果已经有滚轮事件在处理，忽略新的滚轮事件
        if (wheelTimer.value) {
          return
        }
        
        // 检查是否有图书数据
        if (!currentCategoryBooks.value || !currentCategoryBooks.value.books || currentCategoryBooks.value.books.length === 0) {
          return
        }
        
        // 根据滚轮方向滚动
        if (event.deltaY > 0) {
          scrollRight()
        } else {
          scrollLeft()
        }
        
        // 设置防抖定时器，500ms内不允许再次滚动
        wheelTimer.value = setTimeout(() => {
          wheelTimer.value = null
        }, 500)
      } else {
        // 如果已经有滚轮事件在处理，忽略新的滚轮事件
        if (wheelTimers.value[categoryIndex]) {
          return
        }
        
        // 根据滚轮方向滚动
        if (event.deltaY > 0) {
          scrollRightByCategory(categoryIndex)
        } else {
          scrollLeftByCategory(categoryIndex)
        }
        
        // 设置防抖定时器，500ms内不允许再次滚动
        wheelTimers.value[categoryIndex] = setTimeout(() => {
          wheelTimers.value[categoryIndex] = null
        }, 500)
      }
    }

    const handleMouseMove = (event, categoryIndex) => {
      // 如果对话框打开，不处理拖拽
      if (bookDetailVisible.value || dialogVisible.value) {
        return
      }
      
      // 只处理拖拽逻辑
      if (displayMode.value === 'merged') {
        if (!isDragging.value) return
        
        if (!currentCategoryBooks.value) return
        
        const diff = event.clientX - dragStartX.value
        
        if (Math.abs(diff) > 50) {
          if (diff > 0 && currentCenterIndex.value > 0) {
            scrollLeft()
          } else if (diff < 0 && currentCenterIndex.value < currentCategoryBooks.value.books.length - 1) {
            scrollRight()
          }
          dragStartX.value = event.clientX
        }
      } else {
        if (!isDraggingByCategory.value[categoryIndex]) return
        
        const category = categorizedBooks.value[categoryIndex]
        if (!category) return
        
        const diff = event.clientX - dragStartXByCategory.value[categoryIndex]
        
        if (Math.abs(diff) > 50) {
          if (diff > 0 && category.centerIndex > 0) {
            scrollLeftByCategory(categoryIndex)
          } else if (diff < 0 && category.centerIndex < category.books.length - 1) {
            scrollRightByCategory(categoryIndex)
          }
          dragStartXByCategory.value[categoryIndex] = event.clientX
        }
      }
    }

    const handleBookClick = (book, categoryIndexOrBookIndex, bookIndex) => {
      // 点击图书时，清除所有定时器和拖拽状态，确保点击后立即可以使用滚轮和拖拽
      if (wheelTimer.value) {
        clearTimeout(wheelTimer.value)
        wheelTimer.value = null
      }
      
      Object.keys(wheelTimers.value).forEach(key => {
        if (wheelTimers.value[key]) {
          clearTimeout(wheelTimers.value[key])
          wheelTimers.value[key] = null
        }
      })
      
      isDragging.value = false
      Object.keys(isDraggingByCategory.value).forEach(key => {
        isDraggingByCategory.value[key] = false
      })
      
      if (displayMode.value === 'merged') {
        // 合并模式：handleBookClick(book, index)
        // categoryIndexOrBookIndex 实际上是 bookIndex
        const index = categoryIndexOrBookIndex
        if (!currentCategoryBooks.value) return
        
        // 如果点击的是中心图书，显示详情
        if (index === currentCenterIndex.value) {
          selectedBook.value = book
          bookDetailVisible.value = true
        } else {
          // 如果点击的不是中心图书，滑动到该图书
          currentCenterIndex.value = index
        }
      } else {
        // 分类模式：handleBookClick(book, categoryIndex, index)
        const categoryIndex = categoryIndexOrBookIndex
        const index = bookIndex
        
        const category = categorizedBooks.value[categoryIndex]
        if (!category) return
        
        // 如果点击的是中心图书，显示详情
        if (index === category.centerIndex) {
          selectedBook.value = book
          bookDetailVisible.value = true
        } else {
          // 如果点击的不是中心图书，滑动到该图书
          categoryCenterIndices.value[category.name] = index
        }
      }
    }

    // 处理分类切换
    const handleCategoryChange = () => {
      // 切换分类时，重置中心索引到中间位置
      if (currentCategoryBooks.value && currentCategoryBooks.value.books.length > 0) {
        currentCenterIndex.value = Math.floor(currentCategoryBooks.value.books.length / 2)
      } else {
        currentCenterIndex.value = 0
      }
    }

    // 确认分类选择
    const handleConfirmCategories = () => {
      handleCategoryChange()
      showCategoryDialog.value = false
    }

    // 添加分类
    const handleAddCategory = async () => {
      const categoryName = newCategoryName.value.trim()
      if (!categoryName) {
        ElMessage.warning('请输入分类名称')
        return
      }
      
      // 检查分类是否已存在
      if (categoryList.value.some(cat => cat.name === categoryName)) {
        ElMessage.warning('该分类已存在')
        newCategoryName.value = ''
        return
      }
      
      try {
        const response = await axios.post('/category/createCategory', {
          name: categoryName,
          description: '',
          sort: categoryList.value.length + 1
        })
        
        if (response.code === 200) {
          ElMessage.success('分类添加成功')
          newCategoryName.value = ''
          // 刷新分类列表
          await fetchCategoryList()
          // 等待DOM更新，确保 availableCategories 已更新
          await nextTick()
          // 如果处于编辑模式，更新编辑状态（确保新添加的分类在编辑状态中）
          if (isEditMode.value) {
            // 确保新添加的分类在编辑状态中
            if (!editingCategoryNames.value.hasOwnProperty(categoryName)) {
              editingCategoryNames.value[categoryName] = categoryName
            }
            // 同步所有分类到编辑状态（防止遗漏）
            availableCategories.value.forEach(cat => {
              if (!editingCategoryNames.value.hasOwnProperty(cat)) {
                editingCategoryNames.value[cat] = cat
              }
            })
          }
          // 自动选中新添加的分类
          if (!selectedCategories.value.includes(categoryName)) {
            selectedCategories.value.push(categoryName)
          }
          // 保持输入框显示，方便继续添加
          // showAddCategoryInput.value = false
        } else {
          ElMessage.error(response.msg || '添加分类失败')
        }
      } catch (error) {
        console.error('添加分类失败:', error)
        ElMessage.error('添加分类失败，请检查后端服务是否正常运行')
      }
    }

    // 处理新增分类按钮点击
    const handleAddCategoryClick = async () => {
      showAddCategoryInput.value = !showAddCategoryInput.value
      if (showAddCategoryInput.value) {
        // 等待DOM更新后聚焦输入框
        await nextTick()
        if (newCategoryInputRef.value) {
          newCategoryInputRef.value.focus()
        }
      } else {
        // 如果隐藏输入框，清空内容
        newCategoryName.value = ''
      }
    }

    // 处理新增分类输入框失焦
    const handleAddCategoryBlur = () => {
      // 如果输入框为空，隐藏输入框
      if (!newCategoryName.value.trim()) {
        showAddCategoryInput.value = false
      }
    }

    // 处理编辑模式切换
    const handleEditModeToggle = () => {
      isEditMode.value = !isEditMode.value
      if (!isEditMode.value) {
        // 退出编辑模式时，清空编辑状态
        editingCategoryNames.value = {}
        showAddCategoryInput.value = false
      } else {
        // 进入编辑模式时，初始化编辑状态
        availableCategories.value.forEach(cat => {
          editingCategoryNames.value[cat] = cat
        })
      }
    }

    // 处理分类名称保存
    const handleCategoryNameSave = async (oldName) => {
      // 如果 oldName 不在 editingCategoryNames 中，说明可能是新添加的分类，直接返回
      if (!editingCategoryNames.value.hasOwnProperty(oldName)) {
        return
      }
      
      const newName = editingCategoryNames.value[oldName]?.trim()
      if (!newName) {
        ElMessage.warning('分类名称不能为空')
        // 如果 oldName 仍然存在于 availableCategories 中，恢复原值
        if (availableCategories.value.includes(oldName)) {
          editingCategoryNames.value[oldName] = oldName
        } else {
          // 如果 oldName 不存在了（可能是新添加的），删除编辑状态
          delete editingCategoryNames.value[oldName]
        }
        return
      }
      
      if (newName === oldName) {
        // 名称未改变，不需要更新
        return
      }
      
      // 检查新名称是否已存在
      if (availableCategories.value.includes(newName) && newName !== oldName) {
        ElMessage.warning('该分类名称已存在')
        editingCategoryNames.value[oldName] = oldName
        return
      }
      
      // 找到对应的分类ID
      const category = categoryList.value.find(cat => cat.name === oldName)
      if (!category) {
        ElMessage.error('找不到该分类')
        return
      }
      
      try {
        const response = await axios.put('/category/updateCategory', {
          id: category.id,
          name: newName,
          description: category.description || '',
          sort: category.sort || 0
        })
        
        if (response.code === 200) {
          ElMessage.success('分类名称更新成功')
          // 刷新分类列表
          await fetchCategoryList()
          // 更新选中状态
          const index = selectedCategories.value.indexOf(oldName)
          if (index !== -1) {
            selectedCategories.value[index] = newName
          }
          // 更新编辑状态
          delete editingCategoryNames.value[oldName]
          editingCategoryNames.value[newName] = newName
        } else {
          ElMessage.error(response.msg || '更新分类失败')
          editingCategoryNames.value[oldName] = oldName
        }
      } catch (error) {
        console.error('更新分类失败:', error)
        ElMessage.error('更新分类失败，请检查后端服务是否正常运行')
        editingCategoryNames.value[oldName] = oldName
      }
    }

    // 处理分类名称失焦
    const handleCategoryNameBlur = (oldName) => {
      // 如果 oldName 不在 editingCategoryNames 中，说明可能是新添加的分类，直接返回
      if (!editingCategoryNames.value.hasOwnProperty(oldName)) {
        return
      }
      // 如果输入框值为空，但 oldName 仍然存在于 availableCategories 中，恢复原值
      if (!editingCategoryNames.value[oldName] || !editingCategoryNames.value[oldName].trim()) {
        if (availableCategories.value.includes(oldName)) {
          editingCategoryNames.value[oldName] = oldName
        }
        return
      }
      handleCategoryNameSave(oldName)
    }

    // 处理删除分类
    const handleDeleteCategory = async (categoryName) => {
      // 找到对应的分类ID
      const category = categoryList.value.find(cat => cat.name === categoryName)
      if (!category) {
        ElMessage.error('找不到该分类')
        return
      }
      
      try {
        await ElMessageBox.confirm(
          `确定要删除分类"${categoryName}"吗？`,
          '提示',
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )
        
        const response = await axios.delete('/category/deleteCategory', {
          data: { id: category.id }
        })
        
        if (response.code === 200) {
          ElMessage.success('分类删除成功')
          // 刷新分类列表
          await fetchCategoryList()
          // 从选中状态中移除
          const index = selectedCategories.value.indexOf(categoryName)
          if (index !== -1) {
            selectedCategories.value.splice(index, 1)
          }
          // 从编辑状态中移除
          delete editingCategoryNames.value[categoryName]
        } else {
          ElMessage.error(response.msg || '删除分类失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('删除分类失败:', error)
          ElMessage.error('删除分类失败，请检查后端服务是否正常运行')
        }
      }
    }

    // 获取分类列表
    const fetchCategoryList = async () => {
      try {
        const response = await axios.get('/category/getCategoryList')
        if (response.code === 200) {
          const data = response.data || []
          // 确保ID是数字类型，避免响应式更新循环
          // 注意：后端返回的字段可能是 ID（大写）或 id（小写）
          const categories = data
            .filter(cat => cat && ((cat.id !== undefined && cat.id !== null) || (cat.ID !== undefined && cat.ID !== null))) // 过滤掉无效的分类
            .map(cat => {
              // 处理 ID 字段（可能是大写或小写）
              const categoryId = cat.id !== undefined ? cat.id : (cat.ID !== undefined ? cat.ID : null)
              if (categoryId === null || categoryId === undefined) {
                return null
              }
              const id = typeof categoryId === 'number' ? categoryId : Number(categoryId)
              // 确保ID是有效的数字
              if (isNaN(id) || id <= 0) {
                return null
              }
              return {
                id,
                name: cat.name || '',
                description: cat.description || '',
                sort: cat.sort || 0
              }
            })
            .filter(cat => cat !== null) // 过滤掉无效的分类
          categoryList.value = categories
        } else {
          console.error('获取分类列表失败:', response.msg)
          ElMessage.warning('获取分类列表失败: ' + (response.msg || '未知错误'))
        }
      } catch (error) {
        console.error('获取分类列表失败:', error)
        ElMessage.error('获取分类列表失败，请检查后端服务是否正常运行')
      }
    }

    // 批量查询点赞/收藏/借阅状态
    const fetchLikeAndFavoriteStatus = async () => {
      if (!bookList.value || bookList.value.length === 0) {
        return
      }
      
      // 先初始化所有图书的状态为默认值，防止undefined错误
      bookList.value.forEach((book, idx) => {
        if (book && book.id) {
          if (!likeStatus.value[book.id]) {
            likeStatus.value[book.id] = { isLiked: false, likeCount: 0 }
          }
          if (!favoriteStatus.value[book.id]) {
            favoriteStatus.value[book.id] = { isFavorited: false, favoriteCount: 0 }
          }
          if (!borrowStatus.value[book.id]) {
            borrowStatus.value[book.id] = { isBorrowed: false, borrowRecordId: null, status: null }
          }
          if (!reservationStatus.value[book.id]) {
            reservationStatus.value[book.id] = { isReserved: false, reservationId: null }
          }
        }
      })
      
      try {
        const bookIds = bookList.value.map(book => book?.id).filter(id => id)
        
        if (bookIds.length === 0) {
          return
        }
        
        // 批量查询点赞状态
        const likeResponse = await batchGetLikeStatus(bookIds)
        if (likeResponse && likeResponse.code === 200 && likeResponse.data) {
          likeResponse.data.forEach(item => {
            if (item && item.book_id) {
              likeStatus.value[item.book_id] = {
                isLiked: item.is_liked || false,
                likeCount: item.like_count || 0
              }
            }
          })
        }
        
        // 批量查询收藏状态
        const favoriteResponse = await batchGetFavoriteStatus(bookIds)
        if (favoriteResponse && favoriteResponse.code === 200 && favoriteResponse.data) {
          favoriteResponse.data.forEach(item => {
            if (item && item.book_id) {
              favoriteStatus.value[item.book_id] = {
                isFavorited: item.is_favorited || false,
                favoriteCount: item.favorite_count || 0
              }
            }
          })
        }
        
        // 批量查询借阅状态
        const borrowResponse = await axios.get('/borrow/getMyBorrowList', {
          params: { page: 1, pageSize: 1000 }
        })
        if (borrowResponse && borrowResponse.code === 200 && borrowResponse.data) {
          const borrowRecords = borrowResponse.data.list || []
          // 保留status为"borrowed"或"pending"的记录
          borrowRecords
            .filter(record => record.status === 'borrowed' || record.status === 'pending')
            .forEach(record => {
              // 兼容book_id和BookID两种字段名
              const bookId = record.book_id || record.BookID
              if (bookId) {
                // 兼容id和ID两种字段名
                const recordId = record.id || record.ID
                if (recordId) {
                  borrowStatus.value[bookId] = {
                    isBorrowed: true,
                    borrowRecordId: recordId,
                    status: record.status // 'borrowed' 或 'pending'
                  }
                }
              }
            })
        }
        
        // 批量查询预约状态
        try {
          const reservationResponse = await axios.get('/reservation/getMyReservations', {
            params: { page: 1, pageSize: 1000 }
          })
          if (reservationResponse && reservationResponse.code === 200 && reservationResponse.data) {
            const reservations = reservationResponse.data.list || []
            // 保留status为"pending"或"available"的预约
            reservations
              .filter(reservation => reservation.status === 'pending' || reservation.status === 'available')
              .forEach(reservation => {
                // 兼容book_id和BookID两种字段名
                const bookId = reservation.book_id || reservation.BookID
                if (bookId) {
                  // 兼容ID和id两种字段名
                  const reservationId = reservation.id || reservation.ID
                  if (reservationId) {
                    reservationStatus.value[bookId] = {
                      isReserved: true,
                      reservationId: reservationId,
                      status: reservation.status // 保存预约状态
                    }
                  }
                }
              })
          }
        } catch (reservationError) {
          console.error('查询预约状态失败:', reservationError)
          // 静默失败，不影响其他状态查询
        }
      } catch (error) {
        console.error('批量查询点赞/收藏/借阅状态失败:', error)
        // 静默失败，已初始化默认值
      }
    }
    
    // 切换借阅/还书
    const handleToggleBorrow = async (event, bookId) => {
      event.stopPropagation() // 阻止事件冒泡
      
      if (!bookId) return // 防止bookId为undefined
      
      if (!actionLoading.value[bookId]) {
        actionLoading.value[bookId] = {}
      }
      if (actionLoading.value[bookId].borrow) return // 防止重复点击
      
      const currentStatus = borrowStatus.value[bookId]?.status
      const reservationInfo = reservationStatus.value[bookId]
      const isReserved = reservationInfo?.isReserved
      const reservationStatusType = reservationInfo?.status
      
      // 如果是预约状态且为pending，提示取消预约
      // 如果是available状态，直接借阅（不显示取消对话框）
      if (isReserved && reservationStatusType === 'pending') {
        try {
          await ElMessageBox.confirm(
            '您已预约该图书，是否取消预约？',
            '取消预约',
            {
              confirmButtonText: '取消预约',
              cancelButtonText: '保留',
              type: 'warning',
              center: true
            }
          )
          
          // 用户确认取消预约
          actionLoading.value[bookId].borrow = true
          const reservationId = reservationStatus.value[bookId]?.reservationId
          
          if (!reservationId) {
            // 如果本地没有reservationId，尝试重新获取预约信息
            try {
              const reservationResponse = await axios.get('/reservation/getMyReservations', {
                params: { page: 1, pageSize: 1000 }
              })
              if (reservationResponse.code === 200) {
                const reservations = reservationResponse.data.list || []
                const reservation = reservations.find(r => 
                  (r.book_id === bookId || r.BookID === bookId) && 
                  (r.status === 'pending' || r.status === 'available')
                )
                if (reservation) {
                  const foundReservationId = reservation.id || reservation.ID
                  if (foundReservationId) {
                    // 更新本地状态
                    reservationStatus.value[bookId] = {
                      isReserved: true,
                      reservationId: foundReservationId,
                      status: reservation.status
                    }
                    // 使用找到的ID继续取消操作
                    const response = await axios.delete(`/reservation/cancel/${foundReservationId}`)
                    if (response.code === 200) {
                      reservationStatus.value[bookId] = {
                        isReserved: false,
                        reservationId: null,
                        status: null
                      }
                      ElMessage.success('已取消预约')
                      await fetchLikeAndFavoriteStatus()
                    } else {
                      ElMessage.error(response.msg || '取消预约失败')
                    }
                    actionLoading.value[bookId].borrow = false
                    return
                  }
                }
              }
            } catch (err) {
              console.error('重新获取预约信息失败:', err)
            }
            
            ElMessage.error('预约ID不存在，请刷新页面后重试')
            actionLoading.value[bookId].borrow = false
            return
          }
          
          try {
            const response = await axios.delete(`/reservation/cancel/${reservationId}`)
            if (response.code === 200) {
              // 更新状态
              reservationStatus.value[bookId] = {
                isReserved: false,
                reservationId: null
              }
              ElMessage.success('已取消预约')
              // 重新获取状态，确保UI更新
              await fetchLikeAndFavoriteStatus()
            } else {
              ElMessage.error(response.msg || '取消预约失败')
            }
          } catch (err) {
            console.error('取消预约失败:', err)
            ElMessage.error(err.response?.data?.msg || '取消预约失败')
          }
        } catch (error) {
          if (error !== 'cancel') {
            console.error('取消预约失败:', error)
            ElMessage.error(error.response?.data?.msg || '取消预约失败')
          }
        } finally {
          actionLoading.value[bookId].borrow = false
        }
        return
      }
      
      // 如果是待批准状态，提示取消借阅申请
      if (currentStatus === 'pending') {
        try {
          await ElMessageBox.confirm(
            '您的借阅申请正在等待管理员审批，是否取消申请？',
            '取消借阅申请',
            {
              confirmButtonText: '取消申请',
              cancelButtonText: '保留',
              type: 'warning',
              center: true
            }
          )
          
          // 用户确认取消借阅申请
          actionLoading.value[bookId].borrow = true
          let borrowRecordId = borrowStatus.value[bookId]?.borrowRecordId
          
          if (!borrowRecordId) {
            // 如果本地没有borrowRecordId，尝试重新获取借阅记录
            try {
              const borrowResponse = await axios.get('/borrow/getMyBorrowList', {
                params: { page: 1, pageSize: 1000 }
              })
              if (borrowResponse.code === 200) {
                const borrowRecords = borrowResponse.data.list || []
                const record = borrowRecords.find(r => 
                  (r.book_id === bookId || r.BookID === bookId) && 
                  r.status === 'pending'
                )
                if (record) {
                  borrowRecordId = record.id || record.ID
                  // 更新本地状态
                  if (borrowRecordId) {
                    borrowStatus.value[bookId] = {
                      isBorrowed: true,
                      borrowRecordId: borrowRecordId,
                      status: 'pending'
                    }
                  }
                }
              }
            } catch (err) {
              console.error('重新获取借阅记录失败:', err)
            }
            
            if (!borrowRecordId) {
              ElMessage.error('借阅记录ID不存在，请刷新页面后重试')
              actionLoading.value[bookId].borrow = false
              return
            }
          }
          
          try {
            const response = await axios.post('/borrow/cancelBorrowRequest', {
              record_id: borrowRecordId
            })
            if (response.code === 200) {
              // 更新状态
              borrowStatus.value[bookId] = {
                isBorrowed: false,
                borrowRecordId: null,
                status: null
              }
              ElMessage.success('已取消借阅申请')
              // 重新获取状态，确保UI更新
              await fetchLikeAndFavoriteStatus()
            } else {
              ElMessage.error(response.msg || '取消申请失败')
            }
          } catch (err) {
            console.error('取消借阅申请失败:', err)
            ElMessage.error(err.response?.data?.msg || '取消申请失败')
          }
        } catch (error) {
          if (error !== 'cancel') {
            console.error('取消借阅申请失败:', error)
            ElMessage.error(error.response?.data?.msg || '取消申请失败')
          }
        } finally {
          actionLoading.value[bookId].borrow = false
        }
        return
      }
      
      actionLoading.value[bookId].borrow = true
      
      try {
        const isBorrowed = borrowStatus.value[bookId]?.isBorrowed
        const borrowRecordId = borrowStatus.value[bookId]?.borrowRecordId
        
        if (isBorrowed && borrowRecordId && currentStatus === 'borrowed') {
          // 还书
          const response = await axios.post('/borrow/returnBook', {
            id: borrowRecordId
          })
          if (response.code === 200) {
            // 更新状态
            borrowStatus.value[bookId] = {
              isBorrowed: false,
              borrowRecordId: null,
              status: null
            }
            ElMessage.success('还书成功 📚')
          } else {
            ElMessage.error(response.msg || '还书失败')
          }
        } else {
          // 借书 - 检查是否有可用的预约
          const reservationInfo = reservationStatus.value[bookId]
          const reservationId = reservationInfo?.status === 'available' ? reservationInfo.reservationId : null
          
          const borrowRequest = {
            book_id: bookId
          }
          if (reservationId) {
            borrowRequest.reservation_id = reservationId
          }
          
          const response = await axios.post('/borrow/borrowBook', borrowRequest)
          if (response.code === 200 && response.data) {
            // 更新状态
            borrowStatus.value[bookId] = {
              isBorrowed: true,
              borrowRecordId: response.data.id,
              status: 'pending' // 新借阅申请状态为待批准
            }
            // 如果使用了预约，清除预约状态
            if (reservationId) {
              reservationStatus.value[bookId] = {
                isReserved: false,
                reservationId: null,
                status: null
              }
            }
            ElMessage.success('借阅申请已提交，等待管理员审批 📝')
          } else if (response.code === 4001) {
            // 库存不足，提示用户预约
            await handleStockInsufficient(bookId)
          } else {
            ElMessage.error(response.msg || '借阅失败')
          }
        }
      } catch (error) {
        console.error('切换借阅状态失败:', error)
        ElMessage.error(error.response?.data?.msg || '操作失败，请稍后再试')
      } finally {
        actionLoading.value[bookId].borrow = false
      }
    }

    // 处理库存不足的情况
    const handleStockInsufficient = async (bookId) => {
      try {
        await ElMessageBox.confirm(
          '该图书暂无库存，是否加入预约队列？图书归还后将按预约顺序通知您。',
          '库存不足 📚',
          {
            confirmButtonText: '加入预约',
            cancelButtonText: '取消',
            type: 'warning',
            center: true
          }
        )
        
        // 用户确认预约，调用预约API
        const response = await axios.post('/reservation/create', {
          book_id: bookId
        })
        
        if (response.code === 200) {
          // 获取预约队列信息
          const queueInfo = response.data?.queue_position || ''
          const queueText = queueInfo ? `，当前队列位置：第 ${queueInfo} 位` : ''
          
          // 更新预约状态
          reservationStatus.value[bookId] = {
            isReserved: true,
            reservationId: response.data.reservation_id || null,
            status: 'pending' // 新预约状态为pending
          }
          
          ElMessage.success({
            message: `预约成功！🔖${queueText}`,
            duration: 3000
          })
          
          // 刷新借阅状态（预约也会显示在"我的借阅"页面）
          if (viewMode.value === 'borrow') {
            await fetchMyBorrowBooks()
          }
        } else {
          ElMessage.error(response.msg || '预约失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('预约失败:', error)
          ElMessage.error(error.response?.data?.msg || '预约失败，请稍后再试')
        }
      }
    }
    
    // 消息相关方法
    const fetchMessages = async () => {
      messageLoading.value = true
      try {
        const response = await axios.get('/message/getMessages', {
          params: {
            page: messagePage.value,
            pageSize: messagePageSize.value
          }
        })
        if (response.code === 200) {
          messages.value = response.data.list || []
        }
      } catch (error) {
        console.error('获取消息失败:', error)
      } finally {
        messageLoading.value = false
      }
    }
    
    const fetchUnreadCount = async () => {
      try {
        const response = await axios.get('/message/getUnreadCount')
        if (response.code === 200) {
          unreadCount.value = response.data.count || 0
        }
      } catch (error) {
        console.error('获取未读数量失败:', error)
      }
    }
    
    const handleMessageClick = async (message) => {
      if (!message.is_read) {
        await handleMarkRead(message.id || message.ID)
      }
      
      // 根据消息类型和关联信息跳转到相应页面
      if (message.type === 'reservation' && message.related_id) {
        // 预约消息：跳转到图书管理页面，并尝试定位到相关图书
        // 首先需要获取预约信息，找到book_id
        try {
          const reservationResponse = await axios.get(`/reservation/getMyReservations`, {
            params: { page: 1, pageSize: 1000 }
          })
          if (reservationResponse.code === 200) {
            const reservations = reservationResponse.data.list || []
            const reservation = reservations.find(r => r.id === message.related_id)
            if (reservation && reservation.book_id) {
              // 跳转到图书管理页面，并传递book_id参数用于定位
              router.push(`/books?bookId=${reservation.book_id}`)
              showMessagePopover.value = false
              // 等待路由跳转后，滚动到对应图书
              await nextTick()
              setTimeout(() => {
                const bookElement = document.querySelector(`[data-book-id="${reservation.book_id}"]`)
                if (bookElement) {
                  bookElement.scrollIntoView({ behavior: 'smooth', block: 'center' })
                  // 高亮显示
                  bookElement.classList.add('message-highlight')
                  setTimeout(() => {
                    bookElement.classList.remove('message-highlight')
                  }, 3000)
                }
              }, 500)
              return
            }
          }
        } catch (error) {
          console.error('获取预约信息失败:', error)
        }
        // 如果无法获取预约信息，跳转到我的借阅页面
        router.push('/books?view=borrow')
        showMessagePopover.value = false
      } else {
        // 其他消息类型，跳转到我的借阅页面
        router.push('/books?view=borrow')
        showMessagePopover.value = false
      }
    }
    
    const handleMarkRead = async (messageId) => {
      if (!messageId) {
        console.error('消息ID为空')
        return
      }
      
      try {
        const response = await axios.put(`/message/read/${messageId}`)
        if (response.code === 200) {
          // 更新本地消息状态 - 兼容id和ID两种字段名
          const message = messages.value.find(m => (m.id === messageId || m.ID === messageId))
          if (message) {
            message.is_read = true
            message.read_at = new Date().toISOString()
          }
          // 更新未读数量
          unreadCount.value = Math.max(0, unreadCount.value - 1)
          ElMessage.success('已标记为已读')
        } else {
          ElMessage.error(response.msg || '标记已读失败')
        }
      } catch (error) {
        console.error('标记已读失败:', error)
        ElMessage.error(error.response?.data?.msg || '标记已读失败')
      }
    }
    
    const handleMarkAllRead = async () => {
      try {
        const response = await axios.put('/message/readAll')
        if (response.code === 200) {
          messages.value.forEach(m => {
            m.is_read = true
          })
          unreadCount.value = 0
          ElMessage.success('全部标记为已读')
        }
      } catch (error) {
        console.error('标记全部已读失败:', error)
        ElMessage.error('操作失败')
      }
    }
    
    const handleLoadMoreMessages = () => {
      // 可以实现加载更多消息的逻辑
      ElMessage.info('功能开发中...')
    }
    
    const formatMessageTime = (dateString) => {
      if (!dateString) return ''
      const date = new Date(dateString)
      const now = new Date()
      const diff = now - date
      const seconds = Math.floor(diff / 1000)
      const minutes = Math.floor(seconds / 60)
      const hours = Math.floor(minutes / 60)
      const days = Math.floor(hours / 24)
      
      if (days > 7) {
        return date.toLocaleDateString()
      } else if (days > 0) {
        return `${days}天前`
      } else if (hours > 0) {
        return `${hours}小时前`
      } else if (minutes > 0) {
        return `${minutes}分钟前`
      } else {
        return '刚刚'
      }
    }
    
    // 获取借阅按钮的详细状态提示
    const getBorrowButtonTitle = (bookId) => {
      const borrowInfo = borrowStatus.value[bookId]
      const reservationInfo = reservationStatus.value[bookId]
      
      // 已借阅状态
      if (borrowInfo?.isBorrowed && borrowInfo?.status === 'borrowed') {
        return '<div style="text-align: left; line-height: 1.6;">📖 <strong>已借阅</strong><br/></div>'
      }
      
      // 待审批状态
      if (borrowInfo?.isBorrowed && borrowInfo?.status === 'pending') {
        return '<div style="text-align: left; line-height: 1.6;">⏳ <strong>待管理员审批</strong><br/>点击可取消借阅申请</div>'
      }
      
      // 预约状态 - 可借阅
      if (reservationInfo?.isReserved && reservationInfo?.status === 'available') {
        return '<div style="text-align: left; line-height: 1.6;">🔖 <strong>预约图书已可借阅</strong><br/>点击立即提交借阅申请</div>'
      }
      
      // 预约状态 - 等待中
      if (reservationInfo?.isReserved && reservationInfo?.status === 'pending') {
        return '<div style="text-align: left; line-height: 1.6;">🔖 <strong>已预约（等待库存）</strong><br/>点击可取消预约</div>'
      }
      
      // 默认状态 - 可借阅
      return '<div style="text-align: left; line-height: 1.6;">📚 <strong>点击借阅此书</strong></div>'
    }
    
    // 切换点赞
    const handleToggleLike = async (event, bookId) => {
      event.stopPropagation() // 阻止事件冒泡
      
      if (!bookId) return // 防止bookId为undefined
      
      if (!actionLoading.value[bookId]) {
        actionLoading.value[bookId] = {}
      }
      if (actionLoading.value[bookId].like) return // 防止重复点击
      
      actionLoading.value[bookId].like = true
      
      try {
        const response = await toggleLike(bookId)
        if (response.code === 200 && response.data) {
          // 更新状态
          likeStatus.value[bookId] = {
            isLiked: response.data.is_liked,
            likeCount: response.data.like_count || 0
          }
          
          // 显示提示
          if (response.data.is_liked) {
            ElMessage.success('点赞成功 ❤️')
          } else {
            ElMessage.info('已取消点赞')
          }
        } else {
          ElMessage.error(response.msg || '操作失败')
        }
      } catch (error) {
        console.error('切换点赞状态失败:', error)
        ElMessage.error('操作失败，请稍后再试')
      } finally {
        actionLoading.value[bookId].like = false
      }
    }
    
    // 切换收藏
    const handleToggleFavorite = async (event, bookId) => {
      event.stopPropagation() // 阻止事件冒泡
      
      if (!bookId) return // 防止bookId为undefined
      
      if (!actionLoading.value[bookId]) {
        actionLoading.value[bookId] = {}
      }
      if (actionLoading.value[bookId].favorite) return // 防止重复点击
      
      actionLoading.value[bookId].favorite = true
      
      try {
        const response = await toggleFavorite(bookId)
        if (response.code === 200 && response.data) {
          // 更新状态
          favoriteStatus.value[bookId] = {
            isFavorited: response.data.is_favorited,
            favoriteCount: response.data.favorite_count || 0
          }
          
          // 显示提示
          if (response.data.is_favorited) {
            ElMessage.success('收藏成功 ⭐')
          } else {
            ElMessage.info('已取消收藏')
          }
        } else {
          ElMessage.error(response.msg || '操作失败')
        }
      } catch (error) {
        console.error('切换收藏状态失败:', error)
        ElMessage.error('操作失败，请稍后再试')
      } finally {
        actionLoading.value[bookId].favorite = false
      }
    }

    const fetchBookList = async () => {
      loading.value = true
      try {
        const params = {
          page: 1,
          pageSize: 1000 // 获取所有图书用于展示
        }
        if (searchKeyword.value) {
          params.keyword = searchKeyword.value
        }

        const response = await axios.get('/book/getBookList', { params })
        if (response.code === 200) {
          bookList.value = response.data.list || []
          // 如果还没有选中分类，默认选择全部分类
          if (selectedCategories.value.length === 0 && availableCategories.value.length > 0) {
            selectedCategories.value = [...availableCategories.value]
            handleCategoryChange()
          }
          // 批量查询点赞/收藏状态
          await fetchLikeAndFavoriteStatus()
        } else {
          ElMessage.error(response.msg || '获取数据失败')
        }
      } catch (error) {
        console.error('获取图书列表失败:', error)
        ElMessage.error('获取图书列表失败')
      } finally {
        loading.value = false
      }
    }

    // 获取我的借阅图书
    const fetchMyBorrowBooks = async () => {
      loading.value = true
      try {
        // 并行获取借阅记录和预约记录
        const [borrowResponse, reservationResponse] = await Promise.all([
          axios.get('/borrow/getMyBorrowList', { params: { page: 1, pageSize: 1000 } }),
          axios.get('/reservation/getMyReservations', { params: { page: 1, pageSize: 1000 } }).catch(() => ({ code: 200, data: { list: [] } }))
        ])

        if (borrowResponse.code === 200) {
          const borrowRecords = borrowResponse.data.list || []
          const reservationRecords = reservationResponse.code === 200 ? (reservationResponse.data.list || []) : []
          
          // 按状态分组：已借阅、已逾期、待审批、预约
          const borrowedBooks = []
          const overdueBooks = []
          const pendingBooks = []
          const reservedBooks = []
          
          // 处理借阅记录
          borrowRecords.forEach(record => {
            if (record.book) {
              if (record.status === 'borrowed') {
                borrowedBooks.push(record.book)
              } else if (record.status === 'overdue') {
                overdueBooks.push(record.book)
              } else if (record.status === 'pending') {
                pendingBooks.push(record.book)
              }
            }
          })
          
          // 处理预约记录（状态为pending或available的预约）
          reservationRecords.forEach(record => {
            if (record.book && (record.status === 'pending' || record.status === 'available')) {
              reservedBooks.push(record.book)
            }
          })
          
          // 合并所有图书用于状态查询
          bookList.value = [...borrowedBooks, ...overdueBooks, ...pendingBooks, ...reservedBooks]
          
          // 强制使用分类显示模式
          displayMode.value = 'separated'
          
          // 设置虚拟分类：按顺序 已借阅、已逾期、待审批、预约
          selectedCategories.value = []
          const categoriesOrder = [
            { name: '已借阅', books: borrowedBooks },
            { name: '已逾期', books: overdueBooks },
            { name: '待审批', books: pendingBooks },
            { name: '预约', books: reservedBooks }
          ]
          
          // 将图书按虚拟分类组织
          const borrowCategories = []
          categoriesOrder.forEach(category => {
            if (category.books.length > 0) {
              selectedCategories.value.push(category.name)
              borrowCategories.push({
                name: category.name,
                books: category.books,
                centerIndex: Math.floor(category.books.length / 2)
              })
            }
          })
          
          // 保存到borrowCategorizedBooks，用于覆盖categorizedBooks
          borrowCategorizedBooks.value = borrowCategories
          
          // 批量查询点赞/收藏状态
          await fetchLikeAndFavoriteStatus()
        } else {
          ElMessage.error(borrowResponse.msg || '获取借阅记录失败')
        }
      } catch (error) {
        console.error('获取借阅记录失败:', error)
        ElMessage.error('获取借阅记录失败')
      } finally {
        loading.value = false
      }
    }

    // 根据viewMode加载数据
    const loadData = async () => {
      if (viewMode.value === 'borrow') {
        await fetchMyBorrowBooks()
      } else {
        // 清除借阅视图的分类数据
        borrowCategorizedBooks.value = []
        // 恢复合并显示模式
        displayMode.value = 'merged'
        // 重置分类选择（清除借阅视图的虚拟分类）
        selectedCategories.value = []
        await fetchBookList()
      }
    }

    const handleSearch = () => {
      fetchBookList()
    }

    const handleSearchClick = () => {
      showSearchInput.value = !showSearchInput.value
      
      // 如果打开搜索框，聚焦到输入框
      if (showSearchInput.value) {
        nextTick(() => {
          if (searchInputRef.value) {
            searchInputRef.value.focus()
          }
        })
      }
    }

    const handleSearchBlur = () => {
      // 如果搜索框为空，延迟关闭搜索框（给用户时间点击清除按钮）
      if (!searchKeyword.value) {
        setTimeout(() => {
          if (!searchKeyword.value) {
            showSearchInput.value = false
          }
        }, 200)
      }
    }

    const handleAdd = async () => {
      isEdit.value = false
      dialogTitle.value = '新增图书'
      resetForm()
      // 确保分类列表已加载（无论是否为空都重新加载，确保数据最新）
      await fetchCategoryList()
      
      dialogVisible.value = true
    }

    const handleEdit = async (row) => {
      isEdit.value = true
      dialogTitle.value = '编辑图书'
      bookDetailVisible.value = false
      
      // 确保分类列表已加载（无论是否为空都重新加载，确保数据最新）
      await fetchCategoryList()
      
      // 确保id是有效的数字
      let bookId = row.id
      
      if (bookId === null || bookId === undefined) {
        if (row.ID) {
          bookId = row.ID
        } else {
          ElMessage.error('图书ID无效')
          return
        }
      }
      
      // 转换为数字
      if (typeof bookId === 'string') {
        bookId = parseInt(bookId, 10)
      } else if (typeof bookId !== 'number') {
        bookId = Number(bookId)
      }
      
      if (isNaN(bookId) || bookId <= 0) {
        ElMessage.error('图书ID无效，无法编辑')
        return
      }
      
      // 提取分类ID，确保是数字类型且有效
      // 注意：后端返回的字段可能是 ID（大写）或 id（小写）
      const categoryIds = row.categories && row.categories.length > 0
        ? row.categories
            .filter(cat => cat && ((cat.id !== undefined && cat.id !== null) || (cat.ID !== undefined && cat.ID !== null)))
            .map(cat => {
              // 处理 ID 字段（可能是大写或小写）
              const categoryId = cat.id !== undefined ? cat.id : (cat.ID !== undefined ? cat.ID : null)
              if (categoryId === null || categoryId === undefined) {
                return null
              }
              const id = Number(categoryId)
              return isNaN(id) || id <= 0 ? null : id
            })
            .filter(id => id !== null)
        : []
      
      // 先重置表单，避免响应式更新循环
      resetForm()
      
      // 使用 nextTick 确保数据更新在下一个事件循环中完成，避免递归更新
      await nextTick()
      
      form.id = bookId
      form.title = row.title
      form.author = row.author
      form.publisher = row.publisher || ''
      form.publish_date = row.publish_date || ''
      form.isbn = row.isbn
      form.price = row.price || 0
      form.category = row.category || '' // 保留用于兼容
      form.cover_image = row.cover_image || ''
      form.total_stock = row.total_stock || 0
      form.available_stock = row.available_stock || 0
      form.description = row.description || ''
      
      // 最后设置 category_ids，使用新数组避免响应式更新循环
      await nextTick()
      form.category_ids = categoryIds.slice()
      
      await nextTick()
      dialogVisible.value = true
    }

    const handleEditFromDetail = () => {
      if (!selectedBook.value) {
        ElMessage.error('请先选择图书')
        return
      }
      handleEdit(selectedBook.value)
    }

    // 处理图书详情对话框关闭
    const handleBookDetailClose = async () => {
      // 对话框关闭时，确保滚动功能正常
      bookDetailVisible.value = false
      selectedBook.value = null
      
      // 清除所有滚动定时器，确保滚动功能恢复正常
      if (wheelTimer.value) {
        clearTimeout(wheelTimer.value)
        wheelTimer.value = null
      }
      
      // 清除所有分类的滚动定时器
      Object.keys(wheelTimers.value).forEach(key => {
        if (wheelTimers.value[key]) {
          clearTimeout(wheelTimers.value[key])
          wheelTimers.value[key] = null
        }
      })
      
      // 强制清除所有可能阻止滚动的状态
      isDragging.value = false
      Object.keys(isDraggingByCategory.value).forEach(key => {
        isDraggingByCategory.value[key] = false
      })
      
      // // 确保 currentCenterIndex 不是 undefined
      // if (currentCenterIndex.value === undefined || currentCenterIndex.value === null) {
      //   if (currentCategoryBooks.value && currentCategoryBooks.value.books && currentCategoryBooks.value.books.length > 0) {
      //     currentCenterIndex.value = Math.floor(currentCategoryBooks.value.books.length / 2)
      //   } else {
      //     currentCenterIndex.value = 0
      //   }
      // }
      
      // 使用 nextTick 确保状态更新完成后再恢复滚动功能
      await nextTick()
    }

    // 处理图书详情对话框完全关闭（动画完成后）
    const handleBookDetailClosed = () => {
      // 确保状态已更新
      bookDetailVisible.value = false
      selectedBook.value = null
      
      // 清除所有滚动定时器
      if (wheelTimer.value) {
        clearTimeout(wheelTimer.value)
        wheelTimer.value = null
      }
      
      Object.keys(wheelTimers.value).forEach(key => {
        if (wheelTimers.value[key]) {
          clearTimeout(wheelTimers.value[key])
          wheelTimers.value[key] = null
        }
      })
      
      // 清除拖拽状态
      isDragging.value = false
      Object.keys(isDraggingByCategory.value).forEach(key => {
        isDraggingByCategory.value[key] = false
      })
      
      // // 确保 currentCenterIndex 不是 undefined
      // if (currentCenterIndex.value === undefined || currentCenterIndex.value === null) {
      //   if (currentCategoryBooks.value && currentCategoryBooks.value.books && currentCategoryBooks.value.books.length > 0) {
      //     currentCenterIndex.value = Math.floor(currentCategoryBooks.value.books.length / 2)
      //   } else {
      //     currentCenterIndex.value = 0
      //   }
      // }
    }

    const handleDelete = (row) => {
      ElMessageBox.confirm(
        `确定要删除图书《${row.title}》吗？`,
        '提示',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      ).then(async () => {
        try {
          const response = await axios.delete(`/book/deleteBook/${row.id}`)
          if (response.code === 200) {
            ElMessage.success('删除成功')
            fetchBookList()
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
            updateBook()
          } else {
            createBook()
          }
        }
      })
    }

    const createBook = async () => {
      try {
        const response = await axios.post('/book/createBook', form)
        if (response.code === 200) {
          ElMessage.success('创建成功')
          dialogVisible.value = false
          fetchBookList()
        } else {
          ElMessage.error(response.msg || '创建失败')
        }
      } catch (error) {
        console.error('创建失败:', error)
        ElMessage.error('创建失败')
      }
    }

    const updateBook = async () => {
      try {
        // 确保id字段存在且有效
        if (!form.id || form.id === 0 || form.id === null || form.id === undefined) {
          ElMessage.error('图书ID无效，无法更新')
          return
        }
        
        // 确保id是数字类型
        const bookId = typeof form.id === 'string' ? parseInt(form.id, 10) : form.id
        if (isNaN(bookId) || bookId === 0) {
          ElMessage.error('图书ID无效，无法更新')
          return
        }
        
        // 创建更新数据，确保id是数字
        const updateData = {
          ...form,
          id: bookId
        }
        
        const response = await axios.put('/book/updateBook', updateData)
        if (response.code === 200) {
          ElMessage.success('更新成功')
          dialogVisible.value = false
          fetchBookList()
        } else {
          ElMessage.error(response.msg || '更新失败')
        }
      } catch (error) {
        console.error('更新失败:', error)
        ElMessage.error('更新失败')
      }
    }

    const resetForm = () => {
      // 先清空 category_ids，避免响应式更新问题
      form.category_ids = []
      
      Object.assign(form, {
        id: null,
        title: '',
        author: '',
        publisher: '',
        publish_date: '',
        isbn: '',
        price: 0,
        category: '',
        category_ids: [], // 确保是空数组
        cover_image: '',
        total_stock: 0,
        available_stock: 0,
        description: ''
      })
      
      // 再次确保 category_ids 是空数组
      form.category_ids = []
      
      if (formRef.value) {
        formRef.value.clearValidate()
      }
    }

    const handleDialogClose = () => {
      resetForm()
    }

    // 滚动到指定分类
    const scrollToCategory = (categoryIndex) => {
      const element = document.getElementById(`category-section-${categoryIndex}`)
      if (element) {
        // 找到滚动容器（.main）
        const scrollContainer = document.querySelector('.main')
        if (scrollContainer) {
          const containerRect = scrollContainer.getBoundingClientRect()
          const elementRect = element.getBoundingClientRect()
          const scrollTop = scrollContainer.scrollTop
          const targetY = elementRect.top - containerRect.top + scrollTop
          scrollContainer.scrollTo({ top: targetY, behavior: 'smooth' })
        } else {
          // 如果没有找到滚动容器，使用 window 滚动
          element.scrollIntoView({ behavior: 'smooth', block: 'start' })
        }
        currentCategoryPage.value = categoryIndex
      }
    }
    
    // 监听滚动，更新当前分类页面
    const handleScroll = () => {
      if (displayMode.value !== 'separated') return
      
      const scrollContainer = document.querySelector('.main')
      if (!scrollContainer) return
      
      const sections = categorizedBooks.value.map((_, index) => 
        document.getElementById(`category-section-${index}`)
      ).filter(Boolean)
      
      const containerRect = scrollContainer.getBoundingClientRect()
      const containerHeight = containerRect.height
      
      sections.forEach((section, index) => {
        const sectionRect = section.getBoundingClientRect()
        const relativeTop = sectionRect.top - containerRect.top
        if (relativeTop <= containerHeight / 2 && relativeTop + sectionRect.height >= containerHeight / 2) {
          currentCategoryPage.value = index
        }
      })
    }

    // 监听路由查询参数变化
    watch(() => route.query.view, (newView) => {
      if (newView === 'borrow') {
        viewMode.value = 'borrow'
      } else {
        viewMode.value = 'all'
      }
      loadData()
    }, { immediate: true })
    
    // 监听消息弹窗的打开，自动加载消息
    watch(showMessagePopover, (newVal) => {
      if (newVal) {
        fetchMessages()
      }
    })
    
    // 监听bookId参数，用于从消息跳转定位图书
    watch(() => route.query.bookId, (bookId) => {
      if (bookId) {
        nextTick(() => {
          setTimeout(() => {
            const bookElement = document.querySelector(`[data-book-id="${bookId}"]`)
            if (bookElement) {
              bookElement.scrollIntoView({ behavior: 'smooth', block: 'center' })
              // 高亮显示
              bookElement.classList.add('message-highlight')
              setTimeout(() => {
                bookElement.classList.remove('message-highlight')
                // 清除URL参数
                router.replace({ query: { ...route.query, bookId: undefined } })
              }, 3000)
            }
          }, 500)
        })
      }
    })

    onMounted(() => {
      fetchCategoryList()
      fetchUnreadCount() // 获取未读消息数量
      // loadData() 会在 watch 中自动调用
      // 监听滚动事件（监听 .main 容器的滚动）
      nextTick(() => {
        const scrollContainer = document.querySelector('.main')
        if (scrollContainer) {
          scrollContainer.addEventListener('scroll', handleScroll)
        } else {
          // 如果没有找到滚动容器，监听 window 滚动
          window.addEventListener('scroll', handleScroll)
        }
      })
    })

    onUnmounted(() => {
      // 清理滚轮定时器（合并模式）
      if (wheelTimer.value) {
        clearTimeout(wheelTimer.value)
      }
      // 清理滚轮定时器（分类模式）
      Object.keys(wheelTimers.value).forEach(categoryIndex => {
        if (wheelTimers.value[categoryIndex]) {
          clearTimeout(wheelTimers.value[categoryIndex])
        }
      })
      // 移除滚动监听
      const scrollContainer = document.querySelector('.main')
      if (scrollContainer) {
        scrollContainer.removeEventListener('scroll', handleScroll)
      } else {
        window.removeEventListener('scroll', handleScroll)
      }
    })

    return {
      router,
      viewMode,
      loading,
      bookList,
      searchKeyword,
      showListView,
      showSearchInput,
      searchInputRef,
      // 消息相关
      showMessagePopover,
      messages,
      unreadCount,
      messageLoading,
      fetchMessages,
      handleMessageClick,
      handleMarkRead,
      handleMarkAllRead,
      handleLoadMoreMessages,
      formatMessageTime,
      getBorrowButtonTitle,
      bookDetailVisible,
      selectedBook,
      dialogVisible,
      dialogTitle,
      form,
      rules,
      formRef,
      hasAdminOrLibrarianRole,
      selectedCategories,
      availableCategories,
      categoryList,
      currentCategoryBooks,
      categorizedBooks,
      borrowCategorizedBooks,
      displayMode,
      showCategoryDialog,
      handleConfirmCategories,
      fetchCategoryList,
      showAddCategoryInput,
      newCategoryName,
      handleAddCategory,
      handleAddCategoryClick,
      handleAddCategoryBlur,
      newCategoryInputRef,
      isEditMode,
      editingCategoryNames,
      handleEditModeToggle,
      handleCategoryNameSave,
      handleCategoryNameBlur,
      handleDeleteCategory,
      categoryEditInputRefs,
      handleSearchBlur,
      getDefaultCover,
      handleImageError,
      getBookStyle,
      getBookCardClass,
      scrollLeft,
      scrollRight,
      handleWheel,
      handleMouseMove,
      handleMouseDown,
      handleMouseUp,
      handleTouchStart,
      handleTouchMove,
      handleBookClick,
      handleCategoryChange,
      currentCategoryPage,
      scrollToCategory,
      handleSearch,
      handleSearchClick,
      handleAdd,
      handleEdit,
      handleEditFromDetail,
      handleDelete,
      handleSubmit,
      handleDialogClose,
      handleBookDetailClose,
      handleBookDetailClosed,
      // 点赞/收藏/借阅相关
      likeStatus,
      favoriteStatus,
      borrowStatus,
      reservationStatus,
      actionLoading,
      handleToggleLike,
      handleToggleFavorite,
      handleToggleBorrow
    }
  }
}
</script>

<style scoped>
.book-gallery-page {
  position: relative;
  width: 100%;
  min-height: 100vh;
  background: linear-gradient(to bottom right, #f3f4f6, #e5e7eb);
  overflow: hidden;
}

/* 借阅视图模式 - 移除顶部间距 */
.book-gallery-page.borrow-view .carousel-container {
  height: calc(100vh - 80px); /* 只减去底部导航的高度 */
  margin-top: 0;
  padding-top: 40px;
}

/* 顶部导航栏 */
.top-navbar {
  position: fixed;
  top: 20px;
  left: 20px;
  right: 20px;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0;
  pointer-events: none;
}

.nav-left,
.nav-right,
.nav-center {
  display: flex;
  align-items: center;
  gap: 16px;
  pointer-events: auto;
}

/* 右侧按钮整体左移 */
.nav-right {
  margin-right: 40px;
}

/* 普通用户模式：右侧按钮额外左移 */
.nav-right.no-add-btn {
  margin-right: 88px; /* 40px + 48px */
}

.category-select {
  width: 150px;
}

.category-select :deep(.el-input__inner) {
  background: white;
  border: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border-radius: 20px;
  padding: 8px 16px;
  font-size: 14px;
}

.nav-icon-btn {
  width: 40px;
  height: 40px;
  background: white;
  border: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
}

.nav-icon-btn.search-btn.search-active {
  transform: translateX(-240px);
}

.top-search-input {
  width: 220px;
  position: absolute;
  right: 192px; /* 管理员模式：152px + 40px（按钮左移的距离） */
}

/* 普通用户模式：2个按钮（搜索 + 更多），搜索框更靠左避免重叠 */
.top-search-input.no-add-btn {
  right: 168px; /* 120px + 48px（额外左移的距离） */
}

.top-search-input :deep(.el-input__wrapper) {
  background: white;
  border: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border-radius: 20px;
  padding: 4px 12px;
}

.search-input-slide-enter-active,
.search-input-slide-leave-active {
  transition: all 0.3s ease;
}

.search-input-slide-enter-from {
  opacity: 0;
  transform: translateX(20px);
}

.search-input-slide-leave-to {
  opacity: 0;
  transform: translateX(20px);
}

.nav-center {
  display: flex;
  align-items: center;
}

.nav-tabs {
  display: flex;
  align-items: center;
  gap: 12px;
  background: white;
  border-radius: 9999px;
  padding: 8px 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.nav-tab {
  font-size: 14px;
  font-weight: 500;
  color: #6b7280;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
}

.nav-tab.active {
  color: #111827;
}

.nav-divider {
  width: 1px;
  height: 16px;
  background: #d1d5db;
}

/* 列表视图容器 */
.list-view-container {
  position: absolute;
  top: 80px;
  left: 0;
  right: 0;
  bottom: 100px;
  padding: 20px;
  overflow: hidden;
}

.list-view {
  background: white;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}

/* 3D轮播视图 */
.carousel-view {
  position: relative;
  width: 100%;
  min-height: 100vh;
  padding: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
}

.category-section {
  width: 100%;
  height: 100vh;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  margin-bottom: 0;
  padding: 20px 0;
  box-sizing: border-box;
}

.category-section:last-child {
  margin-bottom: 0;
}

.category-title {
  font-size: 24px;
  font-weight: bold;
  margin-bottom: 30px;
  padding-left: 40px;
  color: #111827;
  width: 100%;
  text-align: left;
  position: relative;
}

/* 借阅视图的特殊分类标题样式 */
.category-section:has(.category-title:contains('待审批')) .category-title::before,
.category-section .category-title:first-child::before {
  content: '';
  position: absolute;
  left: 10px;
  top: 50%;
  transform: translateY(-50%);
  width: 20px;
  height: 20px;
  border-radius: 50%;
}

/* 待审批标题标记 */
h2.category-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

/* 借阅视图分类标题图标 */
h2.category-title[data-status="borrowed"]::before {
  content: '📖';
  font-size: 28px;
}

h2.category-title[data-status="overdue"]::before {
  content: '⚠️';
  font-size: 28px;
}

h2.category-title[data-status="overdue"] {
  color: #ef4444; /* 红色警告色 */
}

h2.category-title[data-status="pending"]::before {
  content: '⏳';
  font-size: 28px;
}

h2.category-title[data-status="reserved"]::before {
  content: '🔖';
  font-size: 28px;
}

h2.category-title[data-status="reserved"] {
  color: #8b5cf6; /* 紫色 */
}

/* 分类选择弹出框样式 */
.category-popover {
  padding: 8px 0;
}

.category-popover-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  padding: 0 12px;
}

.category-popover-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  flex: 1;
}

.edit-category-btn {
  width: 24px;
  height: 24px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #dcdfe6;
  background: #fff;
  color: #409eff;
  transition: all 0.3s;
}

.edit-category-btn:hover {
  background: #409eff;
  color: #fff;
  border-color: #409eff;
}

.edit-category-btn.active {
  background: #409eff;
  color: #fff;
  border-color: #409eff;
}

.add-category-btn {
  width: 24px;
  height: 24px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #dcdfe6;
  background: #fff;
  color: #409eff;
  transition: all 0.3s;
}

.add-category-btn:hover {
  background: #409eff;
  color: #fff;
  border-color: #409eff;
}

.add-category-input-wrapper {
  padding: 0 12px;
  margin-bottom: 12px;
}

.add-category-input {
  width: 100%;
}

.category-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 8px 12px;
  gap: 8px;
}

.category-checkbox {
  flex: 1;
  display: flex;
  align-items: center;
}

.category-name {
  flex: 1;
  cursor: pointer;
}

.category-edit-input {
  flex: 1;
  min-width: 0;
}

.delete-category-btn {
  width: 20px;
  height: 20px;
  padding: 0;
  flex-shrink: 0;
}

.add-category-button-wrapper {
  display: flex;
  justify-content: center;
  padding: 8px 12px;
  margin-top: 8px;
}

.category-popover-divider {
  height: 1px;
  background: #ebeef5;
  margin: 16px 0;
}

.display-mode-toggle {
  padding: 0 12px;
  margin-bottom: 12px;
}

.display-mode-label {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 12px;
}

.toggle-switch-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.toggle-label {
  font-size: 13px;
  color: #909399;
  transition: color 0.3s;
  flex: 1;
  text-align: center;
}

.toggle-label.active {
  color: #409eff;
  font-weight: 500;
}

.toggle-switch {
  flex-shrink: 0;
}

.category-popover-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #ebeef5;
}

.carousel-container {
  position: relative;
  height: calc(100vh - 200px); /* 根据视口高度自适应，减去顶部导航和底部导航的高度 */
  width: 100%;
  max-width: 1600px;
  overflow: hidden;
  padding: 20px 0;
  display: flex;
  align-items: center;
  justify-content: center;
  perspective: 2500px;
  cursor: grab;
}

.carousel-container:active {
  cursor: grabbing;
}

.carousel {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  width: 100%;
  position: relative;
}

.book-card {
  position: absolute;
  width: 320px;
  height: 480px;
  cursor: pointer;
  transition: all 0.5s ease-out;
  transform-style: preserve-3d;
}

.book-cover-card {
  width: 100%;
  height: 100%;
  border-radius: 16px;
  /* 使用多层阴影创建更柔和的渐变效果 */
  box-shadow: 
    0 10px 30px rgba(0, 0, 0, 0.15),
    0 20px 60px rgba(0, 0, 0, 0.1),
    0 30px 80px rgba(0, 0, 0, 0.08);
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background: transparent;
}

/* 渐变背景颜色 */
.gradient-pink {
  background: linear-gradient(to bottom right, #ec4899, #9333ea);
}

.gradient-red {
  background: linear-gradient(to bottom right, #ef4444, #f97316);
}

.gradient-teal {
  background: linear-gradient(to bottom right, #14b8a6, #06b6d4);
}

.gradient-brown {
  background: linear-gradient(to bottom right, #a16207, #d97706);
}

.gradient-gray {
  background: linear-gradient(to bottom right, #e5e7eb, #9ca3af);
}

.gradient-rose {
  background: linear-gradient(to bottom right, #fb7185, #ec4899);
}

.gradient-blue {
  background: linear-gradient(to bottom right, #3b82f6, #4f46e5);
}

.gradient-green {
  background: linear-gradient(to bottom right, #10b981, #14b8a6);
}

.gradient-dark {
  background: linear-gradient(to bottom right, #4b5563, #1f2937);
}

.gradient-yellow {
  background: linear-gradient(to bottom right, #ca8a04, #b45309);
}

.gradient-black {
  background: linear-gradient(to bottom right, #1f2937, #000000);
}

.book-header {
  margin-bottom: 16px;
}

.book-title {
  font-size: 29px;
  font-weight: bold;
  margin: 0 0 8px 0;
  line-height: 1.3;
  color: white;
}

.book-author {
  font-size: 16px;
  opacity: 0.9;
  margin: 0;
  color: white;
}

.book-cover-image {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0;
}

.book-cover-placeholder {
  width: 100%;
  height: 100%;
  aspect-ratio: 3/4;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 16px;
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.book-cover-placeholder img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.book-cover-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.book-cover-circle {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(8px);
}


.book-card.active {
  transform: scale(1.1) !important;
  z-index: 50 !important;
}

/* 底部功能栏 */
.bottom-bar {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 100;
}

.bottom-bar-content {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  border-radius: 9999px;
  padding: 12px 20px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
  display: flex;
  align-items: center;
  gap: 16px;
}

.bottom-btn {
  min-width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  cursor: pointer;
  color: #6b7280;
  font-size: 20px;
  padding: 0 12px;
  border-radius: 20px;
  transition: all 0.2s;
}

.bottom-btn:hover {
  background: rgba(0, 0, 0, 0.05);
  color: #111827;
}

.bottom-btn.active {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
}

.bottom-btn-icon {
  width: 24px;
  height: 24px;
  background: #1f2937;
  border-radius: 4px;
}

.bottom-divider {
  width: 1px;
  height: 24px;
  background: #e5e7eb;
  margin: 0 8px;
}

.user-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 16px;
}

.user-name {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  white-space: nowrap;
}

/* 图书详情对话框 */
.book-detail-dialog :deep(.el-dialog) {
  border-radius: 24px;
}

.book-detail {
  display: flex;
  gap: 24px;
}

.detail-cover {
  flex-shrink: 0;
}

.detail-cover img {
  width: 200px;
  height: 300px;
  object-fit: cover;
  border-radius: 12px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.2);
}

.detail-info {
  flex: 1;
}

.detail-info p {
  margin: 12px 0;
  line-height: 1.6;
  color: #374151;
}

/* 分类导航按钮 */
.category-navigation {
  position: fixed;
  right: 30px;
  top: 50%;
  transform: translateY(-50%);
  z-index: 100;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.nav-button {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.9);
  border: 2px solid #409eff;
  color: #409eff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  font-size: 20px;
}

.nav-button:hover {
  background: #409eff;
  color: white;
  transform: scale(1.1);
  box-shadow: 0 6px 16px rgba(64, 158, 255, 0.4);
}

.nav-button:active {
  transform: scale(0.95);
}

/* 拖动提示条 */
.drag-indicator {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  width: 200px;
  height: 4px;
  /* 使用渐变背景，从透明到半透明再到透明，消除明显的横切感 */
  background: linear-gradient(
    to right,
    rgba(255, 255, 255, 0) 0%,
    rgba(255, 255, 255, 0.2) 20%,
    rgba(255, 255, 255, 0.4) 50%,
    rgba(255, 255, 255, 0.2) 80%,
    rgba(255, 255, 255, 0) 100%
  );
  backdrop-filter: blur(8px);
  border-radius: 2px;
  z-index: 100;
  pointer-events: none;
  transition: opacity 0.3s ease;
  /* 移除明显的阴影边界 */
  box-shadow: none;
}

.carousel-container:hover .drag-indicator {
  opacity: 1;
}

.drag-indicator:not(:hover) {
  opacity: 0.6;
}

.drag-handle {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  width: 60px;
  height: 4px;
  /* 使用渐变背景，从透明到半透明再到透明 */
  background: linear-gradient(
    to right,
    rgba(255, 255, 255, 0.4) 0%,
    rgba(255, 255, 255, 0.9) 50%,
    rgba(255, 255, 255, 0.4) 100%
  );
  border-radius: 2px;
  /* 使用柔和的阴影，避免明显的边界 */
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
  animation: drag-pulse 2s ease-in-out infinite;
}

@keyframes drag-pulse {
  0%, 100% {
    opacity: 0.8;
    transform: translate(-50%, -50%) scaleX(1);
  }
  50% {
    opacity: 1;
    transform: translate(-50%, -50%) scaleX(1.2);
  }
}

/* 点赞/收藏操作栏 - 右下角显示 */
.book-action-bar {
  position: absolute;
  bottom: 4px;
  right: 4px;
  display: flex;
  flex-direction: row;
  gap: 6px;
  opacity: 0;
  transform: scale(0.8);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  z-index: 10;
}

.book-card:hover .book-action-bar,
.book-card.active .book-action-bar {
  opacity: 1;
  transform: scale(1);
}

.action-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  background: transparent;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  transition: all 0.3s ease;
}

.action-btn:hover {
  transform: scale(1.2);
}

.action-btn:active {
  transform: scale(0.9);
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.borrow-btn {
  color: #3b82f6;
}

.borrow-btn .action-icon {
  filter: drop-shadow(0 0 2px rgba(255, 255, 255, 0.8));
}

.borrow-btn.active {
  animation: borrow-shake 0.5s cubic-bezier(0.68, -0.55, 0.265, 1.55);
}

.borrow-btn.pending {
  color: #f59e0b;
  animation: pending-pulse 1.5s ease-in-out infinite;
}

.borrow-btn.reserved {
  color: #67C23A;
}

.borrow-btn.reserved-available {
  color: #67C23A;
  animation: reserved-available-pulse 1.5s ease-in-out infinite;
}

@keyframes reserved-available-pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
    box-shadow: 0 0 0 0 rgba(103, 194, 58, 0.4);
  }
  50% {
    opacity: 0.9;
    transform: scale(1.1);
    box-shadow: 0 0 0 8px rgba(103, 194, 58, 0);
  }
}

.like-btn {
  color: #ef4444;
}

.like-btn .action-icon {
  filter: drop-shadow(0 0 2px rgba(255, 255, 255, 0.8));
}

.like-btn.active {
  animation: like-bounce 0.5s cubic-bezier(0.68, -0.55, 0.265, 1.55);
}

.favorite-btn {
  color: #f59e0b;
}

.favorite-btn .action-icon {
  filter: drop-shadow(0 0 2px rgba(255, 255, 255, 0.8));
}

.favorite-btn.active {
  animation: favorite-rotate 0.6s cubic-bezier(0.68, -0.55, 0.265, 1.55);
}

.action-icon {
  width: 24px;
  height: 24px;
  transition: all 0.3s ease;
}

.action-btn:hover .action-icon {
  transform: scale(1.1);
}

@keyframes borrow-shake {
  0% {
    transform: translateX(0) scale(1);
  }
  15% {
    transform: translateX(-5px) scale(1.1);
  }
  30% {
    transform: translateX(5px) scale(1.2);
  }
  45% {
    transform: translateX(-5px) scale(1.1);
  }
  60% {
    transform: translateX(5px) scale(1.1);
  }
  75% {
    transform: translateX(-3px) scale(1.05);
  }
  85% {
    transform: translateX(3px) scale(1.05);
  }
  100% {
    transform: translateX(0) scale(1);
  }
}

@keyframes pending-pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.6;
    transform: scale(1.1);
  }
}

@keyframes like-bounce {
  0% {
    transform: scale(1);
  }
  25% {
    transform: scale(1.3);
  }
  50% {
    transform: scale(0.9);
  }
  75% {
    transform: scale(1.1);
  }
  100% {
    transform: scale(1);
  }
}

@keyframes favorite-rotate {
  0% {
    transform: rotate(0deg) scale(1);
  }
  25% {
    transform: rotate(72deg) scale(1.2);
  }
  50% {
    transform: rotate(144deg) scale(1);
  }
  75% {
    transform: rotate(216deg) scale(1.1);
  }
  100% {
    transform: rotate(360deg) scale(1);
  }
}

/* 消息按钮和下拉框样式 */
.message-badge {
  display: inline-block;
}

.message-btn.has-unread {
  animation: message-breath 2s ease-in-out infinite;
}

@keyframes message-breath {
  0%, 100% {
    transform: scale(1);
    box-shadow: 0 0 0 0 rgba(64, 158, 255, 0.4);
  }
  50% {
    transform: scale(1.05);
    box-shadow: 0 0 0 8px rgba(64, 158, 255, 0);
  }
}

.message-dropdown {
  max-height: 600px;
  display: flex;
  flex-direction: column;
}

.message-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
}

.message-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.message-list {
  max-height: 400px;
  overflow-y: auto;
}

.message-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 16px;
  cursor: pointer;
  transition: background-color 0.3s;
  border-bottom: 1px solid #f5f5f5;
}

.message-item:hover {
  background-color: #f9fafb;
}

.message-item.unread {
  background-color: #f0f9ff;
}

.message-icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background-color: #f3f4f6;
}

.message-content {
  flex: 1;
  min-width: 0;
}

.message-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.message-item-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.message-time {
  font-size: 12px;
  color: #909399;
  flex-shrink: 0;
  margin-left: 8px;
}

.message-text {
  font-size: 13px;
  color: #606266;
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
}

.mark-read-btn {
  flex-shrink: 0;
  opacity: 0;
  transition: opacity 0.3s;
}

.message-item:hover .mark-read-btn {
  opacity: 1;
}

.empty-message {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #909399;
}

.empty-message p {
  margin-top: 12px;
  font-size: 14px;
}

.message-footer {
  padding: 12px 16px;
  text-align: center;
  border-top: 1px solid #f0f0f0;
}

/* 消息跳转高亮样式 */
.book-card.message-highlight {
  animation: message-highlight-pulse 1s ease-in-out 3;
  box-shadow: 0 0 20px rgba(64, 158, 255, 0.6) !important;
}

@keyframes message-highlight-pulse {
  0%, 100% {
    box-shadow: 0 0 20px rgba(64, 158, 255, 0.6);
    transform: scale(1);
  }
  50% {
    box-shadow: 0 0 30px rgba(64, 158, 255, 0.8);
    transform: scale(1.02);
  }
}
</style>
