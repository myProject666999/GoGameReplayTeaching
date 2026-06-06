<template>
  <div class="page">
    <div class="page-header">
      <h1>死活题出题</h1>
      <p class="page-desc">创建一道新的死活题，分享给大家练习</p>
    </div>

    <div class="create-layout">
      <div class="boards-section">
        <div class="board-card">
          <div class="board-card-header">
            <h3 class="board-card-title">初始局面</h3>
            <div class="board-card-controls">
              <div class="color-toggle">
                <span
                  :class="['color-dot', 'B', setupColor === 'B' ? 'active' : '']"
                  @click="setupColor = 'B'"
                ></span>
                <span
                  :class="['color-dot', 'W', setupColor === 'W' ? 'active' : '']"
                  @click="setupColor = 'W'"
                ></span>
              </div>
              <button class="btn btn-ghost" @click="clearSetup">
                清除
              </button>
            </div>
          </div>
          <div class="board-card-body">
            <GoBoard
              :board-size="form.board_size"
              :board="setupBoard"
              :interactive="true"
              :current-player="setupColor"
              @click-intersection="handleSetupClick"
              @right-click-intersection="handleSetupRightClick"
            />
          </div>
          <p class="board-hint">点击放置棋子，右键移除棋子</p>
        </div>

        <div class="board-card">
          <div class="board-card-header">
            <h3 class="board-card-title">正解变化</h3>
            <div class="board-card-controls">
              <button class="btn btn-small" @click="solutionUndo" :disabled="solutionMoves.length === 0">
                悔棋
              </button>
              <button class="btn btn-ghost" @click="resetSolution">
                重置
              </button>
            </div>
          </div>
          <div class="board-card-body">
            <GoBoard
              :board-size="form.board_size"
              :board="solutionBoard"
              :last-move="solutionLastMove"
              :interactive="true"
              :current-player="solutionCurrentPlayer"
              @click-intersection="handleSolutionClick"
            />
          </div>
          <div class="solution-moves-display">
            <span class="moves-label">已录 {{ solutionMoves.length }} 手</span>
            <div class="moves-list">
              <span
                v-for="(m, idx) in solutionMoves"
                :key="idx"
                :class="['move-chip', m.color]"
              >
                {{ idx + 1 }}. {{ m.move }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <div class="form-section">
        <div class="form-card">
          <div class="form-group">
            <label class="form-label">标题 <span class="required">*</span></label>
            <input
              v-model="form.title"
              class="form-input"
              type="text"
              placeholder="请输入题目名称"
            />
          </div>

          <div class="form-row">
            <div class="form-group">
              <label class="form-label">棋盘大小</label>
              <select v-model.number="form.board_size" class="form-select" @change="handleBoardSizeChange">
                <option :value="9">9路</option>
                <option :value="13">13路</option>
                <option :value="19">19路</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">目标</label>
              <select v-model="form.goal" class="form-select">
                <option value="black_kill">黑先杀</option>
                <option value="black_live">黑先活</option>
                <option value="white_kill">白先杀</option>
                <option value="white_live">白先活</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">难度</label>
              <select v-model="form.difficulty" class="form-select">
                <option value="easy">入门</option>
                <option value="medium">初级</option>
                <option value="hard">中级</option>
                <option value="expert">高级</option>
              </select>
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">题目描述</label>
            <textarea
              v-model="form.description"
              class="form-textarea"
              rows="4"
              placeholder="题目说明、提示、出处等..."
            ></textarea>
          </div>

          <div class="form-group checkbox-group">
            <label class="checkbox-label">
              <input type="checkbox" v-model="form.is_public" class="checkbox-input" />
              <span class="checkbox-custom"></span>
              <span class="checkbox-text">公开此题（其他用户可见）</span>
            </label>
          </div>

          <div class="sgf-preview">
            <div class="sgf-preview-header">
              <span class="sgf-label">SGF 预览</span>
            </div>
            <div class="sgf-content">
              <div class="sgf-row">
                <span class="sgf-key">initial_sgf:</span>
                <code class="sgf-value">{{ initialSgfPreview }}</code>
              </div>
              <div class="sgf-row">
                <span class="sgf-key">solution_sgf:</span>
                <code class="sgf-value">{{ solutionSgfPreview }}</code>
              </div>
            </div>
          </div>

          <div v-if="error" class="error-message">
            {{ error }}
          </div>

          <div class="form-actions">
            <button class="btn btn-secondary" @click="$router.back()" :disabled="submitting">
              取消
            </button>
            <button class="btn btn-primary" @click="handleSubmit" :disabled="submitting">
              {{ submitting ? '创建中...' : '创建题目' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { createProblem } from '@/api/problem'
import { useUserStore } from '@/stores/user'
import { CoordToMove } from '@/utils/sgf'
import GoBoard from '@/components/GoBoard.vue'

const userStore = useUserStore()
const router = useRouter()

const submitting = ref(false)
const error = ref('')

const form = reactive({
  title: '',
  board_size: 19,
  goal: 'black_kill',
  difficulty: 'medium',
  description: '',
  is_public: true
})

const setupColor = ref('B')
const setupBoard = ref([])
const solutionBoard = ref([])
const solutionMoves = ref([])
const solutionLastMove = ref(null)
const solutionCurrentPlayer = ref('B')

function createEmptyBoard(size) {
  const b = []
  for (let i = 0; i < size; i++) {
    b.push(new Array(size).fill(''))
  }
  return b
}

function cloneBoard(b) {
  return b.map(row => [...row])
}

function initBoards() {
  setupBoard.value = createEmptyBoard(form.board_size)
  resetSolution()
}

function resetSolution() {
  solutionBoard.value = cloneBoard(setupBoard.value)
  solutionMoves.value = []
  solutionLastMove.value = null
  solutionCurrentPlayer.value = form.goal.startsWith('black') ? 'B' : 'W'
}

watch(() => form.goal, () => {
  if (solutionMoves.value.length === 0) {
    solutionCurrentPlayer.value = form.goal.startsWith('black') ? 'B' : 'W'
  }
})

function handleBoardSizeChange() {
  initBoards()
}

function handleSetupClick(x, y) {
  setupBoard.value[y][x] = setupColor.value
}

function handleSetupRightClick(x, y) {
  setupBoard.value[y][x] = ''
}

function clearSetup() {
  setupBoard.value = createEmptyBoard(form.board_size)
  resetSolution()
}

function handleSolutionClick(x, y) {
  if (solutionBoard.value[y]?.[x]) return
  const color = solutionCurrentPlayer.value
  solutionBoard.value[y][x] = color
  solutionLastMove.value = { x, y }
  solutionMoves.value.push({ x, y, color, move: CoordToMove(x, y) })
  solutionCurrentPlayer.value = color === 'B' ? 'W' : 'B'
}

function solutionUndo() {
  if (solutionMoves.value.length === 0) return
  const last = solutionMoves.value.pop()
  solutionBoard.value[last.y][last.x] = ''
  solutionCurrentPlayer.value = last.color
  if (solutionMoves.value.length > 0) {
    const prev = solutionMoves.value[solutionMoves.value.length - 1]
    solutionLastMove.value = { x: prev.x, y: prev.y }
  } else {
    solutionLastMove.value = null
  }
}

const initialSgfPreview = computed(() => {
  const ab = []
  const aw = []
  for (let y = 0; y < form.board_size; y++) {
    for (let x = 0; x < form.board_size; x++) {
      const c = setupBoard.value[y]?.[x]
      if (c === 'B') ab.push(CoordToMove(x, y))
      else if (c === 'W') aw.push(CoordToMove(x, y))
    }
  }
  let sgf = `(;SZ[${form.board_size}]`
  if (ab.length > 0) sgf += 'AB' + ab.map(m => `[${m}]`).join('')
  if (aw.length > 0) sgf += 'AW' + aw.map(m => `[${m}]`).join('')
  sgf += ')'
  return sgf
})

const solutionSgfPreview = computed(() => {
  const ab = []
  const aw = []
  for (let y = 0; y < form.board_size; y++) {
    for (let x = 0; x < form.board_size; x++) {
      const c = setupBoard.value[y]?.[x]
      if (c === 'B') ab.push(CoordToMove(x, y))
      else if (c === 'W') aw.push(CoordToMove(x, y))
    }
  }
  let sgf = `(;SZ[${form.board_size}]`
  if (ab.length > 0) sgf += 'AB' + ab.map(m => `[${m}]`).join('')
  if (aw.length > 0) sgf += 'AW' + aw.map(m => `[${m}]`).join('')
  for (const m of solutionMoves.value) {
    sgf += `;${m.color}[${m.move}]`
  }
  sgf += ')'
  return sgf
})

function validate() {
  if (!form.title.trim()) {
    error.value = '请输入标题'
    return false
  }
  const hasStones = setupBoard.value.some(row => row.some(c => c))
  if (!hasStones) {
    error.value = '请设置初始局面'
    return false
  }
  if (solutionMoves.value.length === 0) {
    error.value = '请录入正解变化'
    return false
  }
  return true
}

async function handleSubmit() {
  error.value = ''
  if (!validate()) return
  submitting.value = true
  try {
    const data = {
      ...form,
      initial_sgf: initialSgfPreview.value,
      solution_sgf: solutionSgfPreview.value
    }
    const res = await createProblem(data)
    router.push(`/problems/${res.id || res}`)
  } catch (e) {
    error.value = e.message || '创建失败'
  } finally {
    submitting.value = false
  }
}

initBoards()
</script>

<style scoped>
.page {
  max-width: 1600px;
  margin: 0 auto;
  padding: 32px 24px;
}

.page-header {
  margin-bottom: 28px;
}

.page-header h1 {
  font-size: 30px;
  color: #1a1a2e;
  margin: 0 0 8px;
}

.page-desc {
  color: #6b7280;
  margin: 0;
}

.create-layout {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 28px;
  align-items: start;
}

@media (max-width: 1200px) {
  .create-layout {
    grid-template-columns: 1fr;
  }
}

.boards-section {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.board-card {
  background: white;
  border-radius: 16px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
}

.board-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}

.board-card-title {
  margin: 0;
  font-size: 17px;
  color: #1a1a2e;
  font-weight: 600;
}

.board-card-controls {
  display: flex;
  align-items: center;
  gap: 10px;
}

.color-toggle {
  display: flex;
  gap: 6px;
  padding: 4px;
  background: #f3f4f6;
  border-radius: 10px;
}

.color-dot {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 1px 3px rgba(0,0,0,0.15);
  border: 3px solid transparent;
}

.color-dot.B {
  background: radial-gradient(circle at 30% 30%, #555, #000);
}

.color-dot.W {
  background: radial-gradient(circle at 30% 30%, #fff, #ccc);
}

.color-dot.active {
  border-color: #f5576c;
  transform: scale(1.08);
}

.board-card-body {
  display: flex;
  justify-content: center;
}

.board-hint {
  text-align: center;
  margin: 12px 0 0;
  font-size: 13px;
  color: #9ca3af;
}

.solution-moves-display {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px solid #f3f4f6;
}

.moves-label {
  display: block;
  font-size: 13px;
  color: #6b7280;
  margin-bottom: 8px;
  font-weight: 500;
}

.moves-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.move-chip {
  padding: 3px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  font-family: monospace;
}

.move-chip.B {
  background: #1f2937;
  color: white;
}

.move-chip.W {
  background: #f3f4f6;
  color: #374151;
  border: 1px solid #e5e7eb;
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-card {
  background: white;
  border-radius: 16px;
  padding: 28px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
}

.form-group {
  margin-bottom: 18px;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
  margin-bottom: 18px;
}

.form-label {
  display: block;
  font-size: 14px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 8px;
}

.required {
  color: #ef4444;
}

.form-input,
.form-select,
.form-textarea {
  width: 100%;
  padding: 10px 14px;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  font-size: 14px;
  background: white;
  transition: border-color 0.2s, box-shadow 0.2s;
  box-sizing: border-box;
  color: #1f2937;
  font-family: inherit;
}

.form-textarea {
  resize: vertical;
  min-height: 80px;
  line-height: 1.6;
}

.form-input:focus,
.form-select:focus,
.form-textarea:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 4px rgba(102,126,234,0.08);
}

.form-select {
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%236b7280' d='M6 8.5L1.5 4h9z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 36px;
}

.checkbox-group {
  margin-bottom: 22px;
}

.checkbox-label {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  user-select: none;
}

.checkbox-input {
  display: none;
}

.checkbox-custom {
  width: 20px;
  height: 20px;
  border: 2px solid #d1d5db;
  border-radius: 5px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  flex-shrink: 0;
}

.checkbox-input:checked + .checkbox-custom {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-color: transparent;
}

.checkbox-input:checked + .checkbox-custom::after {
  content: '';
  width: 6px;
  height: 10px;
  border: solid white;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg) translate(-1px, -1px);
}

.checkbox-text {
  font-size: 14px;
  color: #374151;
}

.sgf-preview {
  background: #f9fafb;
  border-radius: 10px;
  padding: 14px;
  margin-bottom: 20px;
}

.sgf-preview-header {
  margin-bottom: 10px;
}

.sgf-label {
  font-size: 13px;
  font-weight: 600;
  color: #6b7280;
}

.sgf-content {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.sgf-row {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  font-size: 12px;
}

.sgf-key {
  color: #6b7280;
  font-weight: 500;
  flex-shrink: 0;
  min-width: 90px;
}

.sgf-value {
  color: #374151;
  word-break: break-all;
  background: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 11px;
  line-height: 1.5;
}

.error-message {
  padding: 12px 16px;
  background: #fef2f2;
  color: #b91c1c;
  border-radius: 10px;
  font-size: 14px;
  margin-bottom: 18px;
  border: 1px solid #fecaca;
}

.form-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.btn {
  padding: 11px 26px;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.2s;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(102,126,234,0.35);
}

.btn-secondary {
  background: #f3f4f6;
  color: #374151;
}

.btn-secondary:hover:not(:disabled) {
  background: #e5e7eb;
}

.btn-ghost {
  background: transparent;
  color: #6b7280;
  padding: 6px 14px;
  font-size: 13px;
}

.btn-ghost:hover:not(:disabled) {
  background: #f3f4f6;
  color: #374151;
}

.btn-small {
  padding: 6px 14px;
  font-size: 13px;
  background: #f3f4f6;
  color: #374151;
}

.btn-small:hover:not(:disabled) {
  background: #e5e7eb;
}
</style>
