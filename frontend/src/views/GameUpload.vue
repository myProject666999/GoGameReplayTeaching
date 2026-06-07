<template>
  <div class="page">
    <div class="page-header">
      <h1>上传棋谱</h1>
      <p class="page-desc">分享您的精彩对局</p>
    </div>
    <div class="form-container">
      <div class="section">
        <div class="section-title">基本信息</div>
        <div class="form-group">
          <label>标题 <span class="required">*</span></label>
          <input v-model="form.title" placeholder="例如：第23届LG杯决赛 申真谞 vs 柯洁" />
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>黑方</label>
            <input v-model="form.black_player" placeholder="黑方棋手姓名" />
          </div>
          <div class="form-group">
            <label>白方</label>
            <input v-model="form.white_player" placeholder="白方棋手姓名" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>棋盘大小</label>
            <select v-model.number="form.board_size">
              <option :value="9">9路</option>
              <option :value="13">13路</option>
              <option :value="19">19路</option>
            </select>
          </div>
          <div class="form-group">
            <label>贴目</label>
            <input v-model.number="form.komi" type="number" step="0.5" min="0" />
          </div>
          <div class="form-group">
            <label>结果</label>
            <input v-model="form.result" placeholder="例如：黑中盘胜、白2.5目胜" />
          </div>
          <div class="form-group">
            <label>对局日期</label>
            <input v-model="form.date_played" type="date" />
          </div>
        </div>
      </div>

      <div class="section">
        <div class="section-title">棋谱数据 <span class="required">*</span></div>
        <div class="sgf-tabs">
          <button
            :class="['tab-btn', { active: sgfMode === 'text' }]"
            @click="sgfMode = 'text'"
          >
            粘贴SGF
          </button>
          <button
            :class="['tab-btn', { active: sgfMode === 'file' }]"
            @click="sgfMode = 'file'"
          >
            上传文件
          </button>
        </div>
        <div v-show="sgfMode === 'text'" class="form-group">
          <textarea
            v-model="form.sgf_content" rows="12" placeholder="在此粘贴SGF格式的棋谱内容..."></textarea>
        </div>
        <div v-show="sgfMode === 'file'" class="form-group">
          <div
            :class="['file-drop', { dragging: isDragging }]"
            @dragover.prevent="isDragging = true"
            @dragleave="isDragging = false"
            @drop.prevent="handleDrop"
            @click="$refs.fileInput.click()"
          >
            <input
              ref="fileInput"
              type="file"
              accept=".sgf"
              style="display:none"
              @change="handleFileSelect"
            />
            <div class="drop-icon">📄</div>
            <p class="drop-text">
              <strong>{{ fileName || '点击或拖拽 .sgf 文件到此处' }}</strong>
            </p>
            <p class="drop-hint">支持 .sgf 格式文件</p>
          </div>
        </div>
      </div>

      <div class="section">
        <div class="section-title">其他设置</div>
        <div class="form-group">
          <label>描述</label>
          <textarea v-model="form.description" rows="3" placeholder="对局背景、解说等..."></textarea>
        </div>
        <label class="checkbox-group">
          <input type="checkbox" v-model="form.is_public" />
          <span class="checkbox-custom"></span>
          <span class="checkbox-label">公开此棋谱（其他用户可见）</span>
        </label>
      </div>

      <div class="actions">
        <button @click="submit" :disabled="loading" class="btn-submit">
          <span v-if="loading" class="spinner"></span>
          {{ loading ? '上传中...' : '上传棋谱' }}
        </button>
      </div>
      <transition name="fade">
        <div v-if="error" class="error">{{ error }}</div>
      </transition>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { createGame } from '@/api/game'

const router = useRouter()
const loading = ref(false)
const error = ref('')
const sgfMode = ref('text')
const isDragging = ref(false)
const fileName = ref('')

const form = reactive({
  title: '',
  black_player: '',
  white_player: '',
  board_size: 19,
  komi: 6.5,
  result: '',
  date_played: '',
  sgf_content: '',
  description: '',
  is_public: false
})

function handleFileSelect(e) {
  const file = e.target.files[0]
  if (file) {
    readFile(file)
  }
}

function handleDrop(e) {
  isDragging.value = false
  const file = e.dataTransfer.files[0]
  if (file) {
    readFile(file)
  }
}

function readFile(file) {
  if (!file.name.endsWith('.sgf')) {
    error.value = '请上传 .sgf 格式文件'
    return
  }
  fileName.value = file.name
  const reader = new FileReader()
  reader.onload = (e) => {
    let content = e.target.result
    if (content.charCodeAt(0) === 0xFEFF) {
      content = content.slice(1)
    }
    form.sgf_content = content.replace(/\r/g, '').trim()
    error.value = ''
  }
  reader.onerror = () => {
    error.value = '文件读取失败'
  }
  reader.readAsText(file, 'UTF-8')
}

async function submit() {
  if (!form.sgf_content.trim()) {
    error.value = '请提供SGF棋谱内容'
    return
  }
  loading.value = true
  error.value = ''
  try {
    const payload = {
      ...form,
      komi: Number(form.komi) || 6.5,
      board_size: Number(form.board_size) || 19,
      sgf_content: form.sgf_content.replace(/\r/g, '').trim()
    }
    if (!payload.date_played) {
      delete payload.date_played
    }
    const game = await createGame(payload)
    router.push(`/games/${game.id}`)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.page {
  max-width: 900px;
  margin: 0 auto;
  padding: 32px 24px;
}
.page-header {
  margin-bottom: 28px;
}
.page-header h1 {
  font-size: 28px;
  color: #1a1a2e;
  margin: 0 0 8px;
}
.page-desc {
  color: #6b7280;
  margin: 0;
}
.form-container {
  background: white;
  border-radius: 16px;
  padding: 32px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.06);
}
.section {
  margin-bottom: 28px;
}
.section:last-of-type:not(:last-child) {
  padding-bottom: 24px;
  border-bottom: 1px solid #f3f4f6;
}
.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 16px;
  padding-left: 12px;
  border-left: 4px solid #667eea;
}
.form-row {
  display: flex;
  gap: 16px;
}
.form-group {
  flex: 1;
  margin-bottom: 18px;
}
.form-group:last-child {
  margin-bottom: 0;
}
.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #374151;
  font-size: 14px;
}
.required {
  color: #ef4444;
}
.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 10px 14px;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  font-size: 14px;
  transition: border-color 0.2s, box-shadow 0.2s;
  box-sizing: border-box;
  background: #fafafa;
  color: #1f2937;
}
.form-group textarea {
  font-family: 'Consolas', 'Monaco', monospace;
  resize: vertical;
  min-height: 100px;
}
.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #667eea;
  background: white;
  box-shadow: 0 0 0 4px rgba(102,126,234,0.08);
}
.sgf-tabs {
  display: flex;
  gap: 4px;
  margin-bottom: 16px;
  background: #f3f4f6;
  padding: 4px;
  border-radius: 10px;
  width: fit-content;
}
.tab-btn {
  padding: 8px 20px;
  border: none;
  background: transparent;
  cursor: pointer;
  border-radius: 8px;
  font-size: 14px;
  color: #6b7280;
  transition: all 0.2s;
}
.tab-btn.active {
  background: white;
  color: #667eea;
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}
.file-drop {
  border: 2px dashed #d1d5db;
  border-radius: 12px;
  padding: 40px 24px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
  background: #fafafa;
}
.file-drop.dragging {
  border-color: #667eea;
  background: rgba(102,126,234,0.05);
}
.file-drop:hover {
  border-color: #667eea;
  background: rgba(102,126,234,0.03);
}
.drop-icon {
  font-size: 40px;
  margin-bottom: 12px;
}
.drop-text {
  color: #374151;
  margin: 0 0 6px;
  font-size: 14px;
}
.drop-text strong {
  color: #667eea;
}
.drop-hint {
  color: #9ca3af;
  margin: 0;
  font-size: 13px;
}
.checkbox-group {
  display: flex;
  align-items: center;
  cursor: pointer;
  user-select: none;
  gap: 10px;
  margin-top: 8px;
}
.checkbox-group input {
  display: none;
}
.checkbox-custom {
  width: 20px;
  height: 20px;
  border: 2px solid #d1d5db;
  border-radius: 6px;
  position: relative;
  transition: all 0.2s;
  background: white;
}
.checkbox-group input:checked + .checkbox-custom {
  background: #667eea;
  border-color: #667eea;
}
.checkbox-group input:checked + .checkbox-custom::after {
  content: '';
  position: absolute;
  left: 5px;
  top: 1px;
  width: 6px;
  height: 10px;
  border: solid white;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
}
.checkbox-label {
  color: #374151;
  font-size: 14px;
}
.actions {
  margin-top: 8px;
}
.btn-submit {
  width: 100%;
  padding: 14px 32px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 12px;
  cursor: pointer;
  font-size: 16px;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  transition: transform 0.2s, box-shadow 0.2s;
}
.btn-submit:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(102,126,234,0.3);
}
.btn-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}
.error {
  margin-top: 16px;
  background: #fef2f2;
  color: #dc2626;
  padding: 12px 16px;
  border-radius: 10px;
  font-size: 14px;
}
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
