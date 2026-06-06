<template>
  <div class="page">
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else-if="problem" class="detail-layout">
      <div class="board-section">
        <div class="board-wrapper">
          <GoBoard
            :board-size="boardSize"
            :board="displayBoard"
            :last-move="lastMove"
            :interactive="!showingSolution"
            :current-player="currentPlayer"
            @click-intersection="handleClick"
          />
        </div>

        <div class="controls">
          <button
            class="btn btn-primary"
            @click="submitAnswer"
            :disabled="userMoves.length === 0 || submitting"
          >
            {{ submitting ? '提交中...' : '提交答案' }}
          </button>
          <button class="btn btn-warning" @click="undoMove" :disabled="userMoves.length === 0">
            悔棋
          </button>
          <button class="btn btn-secondary" @click="resetBoard">
            重置
          </button>
          <button
            :class="['btn', showingSolution ? 'btn-info-active' : 'btn-info']"
            @click="toggleSolution"
          >
            {{ showingSolution ? '隐藏正解' : '查看正解' }}
          </button>
        </div>

        <div v-if="resultMessage" :class="['result-banner', resultSuccess ? 'success' : 'error']">
          {{ resultMessage }}
        </div>
      </div>

      <div class="info-section">
        <div class="info-card">
          <h1 class="problem-title">{{ problem.title }}</h1>
          <div class="tags">
            <span :class="['tag goal', problem.goal]">{{ formatGoal(problem.goal) }}</span>
            <span :class="['tag diff', problem.difficulty]">{{ formatDifficulty(problem.difficulty) }}</span>
          </div>
          <div class="meta">
            <div class="meta-item">
              <span class="meta-label">棋盘</span>
              <span class="meta-value">{{ boardSize }}路</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">当前</span>
              <span class="meta-value">
                <span :class="['stone-dot', currentPlayer]"></span>
                {{ currentPlayer === 'B' ? '黑方' : '白方' }}落子
              </span>
            </div>
            <div v-if="problem.creator" class="meta-item">
              <span class="meta-label">出题者</span>
              <span class="meta-value">{{ problem.creator.nickname || problem.creator.username }}</span>
            </div>
          </div>
          <p v-if="problem.description" class="description">{{ problem.description }}</p>
        </div>

        <div v-if="showingSolution" class="info-card">
          <h3 class="card-title">正解变化</h3>
          <div class="solution-controls">
            <button class="btn btn-small" @click="solutionPrev" :disabled="solutionIndex === 0">
              上一步
            </button>
            <span class="solution-progress">
              {{ solutionIndex }} / {{ solutionMoves.length }}
            </span>
            <button
              class="btn btn-small"
              @click="solutionNext"
              :disabled="solutionIndex >= solutionMoves.length"
            >
              下一步
            </button>
          </div>
        </div>

        <div class="info-card">
          <h3 class="card-title">作答历史</h3>
          <div v-if="attemptsLoading" class="mini-loading">加载中...</div>
          <div v-else-if="attempts.length === 0" class="empty-history">
            暂无作答记录
          </div>
          <div v-else class="attempt-list">
            <div
              v-for="(attempt, idx) in attempts"
              :key="attempt.id || idx"
              :class="['attempt-item', attempt.is_correct ? 'correct' : 'wrong']"
            >
              <div class="attempt-icon">{{ attempt.is_correct ? '✓' : '✗' }}</div>
              <div class="attempt-info">
                <div class="attempt-moves">{{ attempt.move_count || 0 }} 手</div>
                <div class="attempt-time" v-if="attempt.time_spent">
                  {{ formatTime(attempt.time_spent) }}
                </div>
              </div>
              <div class="attempt-date" v-if="attempt.created_at">
                {{ formatDate(attempt.created_at) }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getProblem, attemptProblem, listProblemAttempts } from '@/api/problem'
import { Parse, NewGameState, GameState, Node, GameTree, MoveToCoord, CoordToMove } from '@/utils/sgf'
import GoBoard from '@/components/GoBoard.vue'

const route = useRoute()
const problem = ref(null)
const loading = ref(false)
const submitting = ref(false)
const attemptsLoading = ref(false)
const attempts = ref([])

const boardSize = ref(19)
let initialGameState = null
const board = ref([])
const userMoves = ref([])
const currentPlayer = ref('B')
const lastMove = ref(null)

const showingSolution = ref(false)
const solutionIndex = ref(0)
const solutionMoves = ref([])
let solutionGameState = null

const resultMessage = ref('')
const resultSuccess = ref(false)
const startTime = ref(Date.now())

const displayBoard = computed(() => {
  if (showingSolution.value && solutionGameState) {
    return solutionGameState.board
  }
  return board.value
})

function formatGoal(g) {
  const map = {
    black_kill: '黑先杀',
    black_live: '黑先活',
    white_kill: '白先杀',
    white_live: '白先活'
  }
  return map[g] || g
}

function formatDifficulty(d) {
  const map = {
    easy: '入门',
    medium: '初级',
    hard: '中级',
    expert: '高级'
  }
  return map[d] || d
}

function formatTime(seconds) {
  if (!seconds) return '0秒'
  if (seconds < 60) return `${seconds}秒`
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${m}分${s}秒`
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  const diff = now - d
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  return d.toLocaleDateString()
}

function cloneBoard(b) {
  return b.map(row => [...row])
}

function resetToInitial() {
  if (initialGameState) {
    board.value = cloneBoard(initialGameState.board)
    userMoves.value = []
    lastMove.value = null
    if (problem.value) {
      currentPlayer.value = problem.value.goal.startsWith('black') ? 'B' : 'W'
    }
    resultMessage.value = ''
    resultSuccess.value = false
    startTime.value = Date.now()
  }
}

function handleClick(x, y) {
  if (showingSolution.value) return
  if (board.value[y]?.[x]) return

  const color = currentPlayer.value
  board.value[y][x] = color
  lastMove.value = { x, y }
  userMoves.value.push({ x, y, color, move: CoordToMove(x, y) })
  currentPlayer.value = color === 'B' ? 'W' : 'B'
}

function undoMove() {
  if (userMoves.value.length === 0) return
  const last = userMoves.value.pop()
  board.value[last.y][last.x] = ''
  currentPlayer.value = last.color
  if (userMoves.value.length > 0) {
    const prev = userMoves.value[userMoves.value.length - 1]
    lastMove.value = { x: prev.x, y: prev.y }
  } else {
    lastMove.value = null
  }
  resultMessage.value = ''
}

function resetBoard() {
  resetToInitial()
  if (showingSolution.value) {
    showingSolution.value = false
  }
}

async function submitAnswer() {
  if (userMoves.value.length === 0 || submitting.value) return
  submitting.value = true
  resultMessage.value = ''
  try {
    const timeSpent = Math.floor((Date.now() - startTime.value) / 1000)
    const res = await attemptProblem(route.params.id, {
      user_moves: userMoves.value.map(m => m.move),
      time_spent: timeSpent
    })
    if (res.is_correct) {
      resultSuccess.value = true
      resultMessage.value = '🎉 回答正确！太棒了！'
    } else {
      resultSuccess.value = false
      resultMessage.value = '✗ 答案不对，再试试吧！'
    }
    loadAttempts()
  } catch (e) {
    resultSuccess.value = false
    resultMessage.value = e.message || '提交失败'
  } finally {
    submitting.value = false
  }
}

function parseSolutionMoves() {
  if (!problem.value?.solution_sgf) return []
  try {
    const tree = Parse(problem.value.solution_sgf)
    const moves = []
    let node = tree.root
    while (node && node.children && node.children.length > 0) {
      node = node.children[0]
      if (node.properties['B']?.[0]) {
        moves.push({ color: 'B', move: node.properties['B'][0] })
      } else if (node.properties['W']?.[0]) {
        moves.push({ color: 'W', move: node.properties['W'][0] })
      }
    }
    return moves
  } catch (e) {
    return []
  }
}

function initSolutionState() {
  if (!problem.value?.solution_sgf) return
  try {
    const tree = Parse(problem.value.solution_sgf)
    solutionGameState = NewGameState(tree)
  } catch (e) {
    solutionGameState = null
  }
}

function resetSolution() {
  initSolutionState()
  solutionIndex.value = 0
}

function toggleSolution() {
  if (showingSolution.value) {
    showingSolution.value = false
  } else {
    showingSolution.value = true
    resetSolution()
  }
}

function solutionNext() {
  if (!solutionGameState || solutionIndex.value >= solutionMoves.value.length) return
  if (solutionGameState.Next(0)) {
    solutionIndex.value++
  }
}

function solutionPrev() {
  if (!solutionGameState || solutionIndex.value === 0) return
  if (solutionGameState.Previous()) {
    solutionIndex.value--
  }
}

async function loadAttempts() {
  attemptsLoading.value = true
  try {
    const res = await listProblemAttempts(route.params.id)
    attempts.value = res.attempts || res || []
  } catch (e) {
    attempts.value = []
  } finally {
    attemptsLoading.value = false
  }
}

async function loadProblem() {
  loading.value = true
  try {
    problem.value = await getProblem(route.params.id)
    const tree = Parse(problem.value.initial_sgf)
    boardSize.value = tree.boardSize
    initialGameState = NewGameState(tree)
    solutionMoves.value = parseSolutionMoves()
    resetToInitial()
    loadAttempts()
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadProblem()
})

watch(() => route.params.id, () => {
  loadProblem()
})
</script>

<style scoped>
.page {
  max-width: 1400px;
  margin: 0 auto;
  padding: 32px 24px;
}

.loading {
  text-align: center;
  padding: 80px 20px;
  color: #6b7280;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #e5e7eb;
  border-top-color: #f5576c;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.detail-layout {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 28px;
  align-items: start;
}

@media (max-width: 960px) {
  .detail-layout {
    grid-template-columns: 1fr;
  }
}

.board-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
}

.board-wrapper {
  background: white;
  padding: 20px;
  border-radius: 16px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
}

.controls {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: center;
}

.btn {
  padding: 10px 22px;
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

.btn-warning {
  background: linear-gradient(135deg, #f6d365 0%, #fda085 100%);
  color: #7c2d12;
}

.btn-warning:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(253,160,133,0.35);
}

.btn-secondary {
  background: #f3f4f6;
  color: #374151;
}

.btn-secondary:hover:not(:disabled) {
  background: #e5e7eb;
}

.btn-info {
  background: #dbeafe;
  color: #1e40af;
}

.btn-info:hover:not(:disabled) {
  background: #bfdbfe;
}

.btn-info-active {
  background: linear-gradient(135deg, #60a5fa 0%, #3b82f6 100%);
  color: white;
}

.btn-small {
  padding: 6px 14px;
  font-size: 13px;
  background: #f3f4f6;
  color: #374151;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-small:hover:not(:disabled) {
  background: #e5e7eb;
}

.btn-small:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.result-banner {
  padding: 14px 24px;
  border-radius: 12px;
  font-weight: 600;
  font-size: 15px;
  text-align: center;
  min-width: 300px;
}

.result-banner.success {
  background: linear-gradient(135deg, #84fab0 0%, #8fd3f4 100%);
  color: #065f46;
}

.result-banner.error {
  background: linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%);
  color: #991b1b;
}

.info-section {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.info-card {
  background: white;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
}

.problem-title {
  margin: 0 0 14px;
  font-size: 24px;
  color: #1a1a2e;
  font-weight: 700;
}

.tags {
  display: flex;
  gap: 8px;
  margin-bottom: 18px;
}

.tag {
  padding: 5px 14px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 600;
  white-space: nowrap;
}

.tag.goal.black_kill,
.tag.goal.black_live {
  background: #1f2937;
  color: white;
}

.tag.goal.white_kill,
.tag.goal.white_live {
  background: #f3f4f6;
  color: #374151;
  border: 1px solid #e5e7eb;
}

.tag.diff.easy {
  background: linear-gradient(135deg, #84fab0 0%, #8fd3f4 100%);
  color: #065f46;
}

.tag.diff.medium {
  background: linear-gradient(135deg, #fccb90 0%, #d57eeb 100%);
  color: #7c2d12;
}

.tag.diff.hard {
  background: linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%);
  color: #991b1b;
}

.tag.diff.expert {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  padding: 14px 0;
  margin-bottom: 14px;
  border-top: 1px solid #f3f4f6;
  border-bottom: 1px solid #f3f4f6;
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.meta-label {
  font-size: 12px;
  color: #9ca3af;
  font-weight: 500;
}

.meta-value {
  font-size: 14px;
  color: #1f2937;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 6px;
}

.stone-dot {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  display: inline-block;
  box-shadow: 0 1px 3px rgba(0,0,0,0.2);
}

.stone-dot.B {
  background: radial-gradient(circle at 30% 30%, #555, #000);
}

.stone-dot.W {
  background: radial-gradient(circle at 30% 30%, #fff, #ccc);
  border: 1px solid #bbb;
}

.description {
  margin: 0;
  color: #4b5563;
  line-height: 1.7;
  font-size: 14px;
}

.card-title {
  margin: 0 0 16px;
  font-size: 17px;
  color: #1a1a2e;
  font-weight: 600;
}

.solution-controls {
  display: flex;
  align-items: center;
  gap: 12px;
  justify-content: center;
}

.solution-progress {
  font-size: 14px;
  color: #6b7280;
  font-weight: 500;
  min-width: 60px;
  text-align: center;
}

.mini-loading {
  text-align: center;
  color: #9ca3af;
  padding: 20px;
  font-size: 14px;
}

.empty-history {
  text-align: center;
  color: #9ca3af;
  padding: 24px;
  font-size: 14px;
}

.attempt-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-height: 320px;
  overflow-y: auto;
}

.attempt-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border-radius: 10px;
  border: 1px solid #f3f4f6;
  transition: background 0.2s;
}

.attempt-item:hover {
  background: #fafafa;
}

.attempt-item.correct {
  background: linear-gradient(135deg, #ecfdf5 0%, #f0fdfa 100%);
  border-color: #a7f3d0;
}

.attempt-item.wrong {
  background: linear-gradient(135deg, #fef2f2 0%, #fff1f2 100%);
  border-color: #fecaca;
}

.attempt-icon {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 14px;
  flex-shrink: 0;
}

.attempt-item.correct .attempt-icon {
  background: #10b981;
  color: white;
}

.attempt-item.wrong .attempt-icon {
  background: #ef4444;
  color: white;
}

.attempt-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.attempt-moves {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
}

.attempt-time {
  font-size: 12px;
  color: #9ca3af;
}

.attempt-date {
  font-size: 12px;
  color: #9ca3af;
}
</style>
