<template>
  <div class="game-detail-page">
    <div v-if="loading" class="loading-wrap">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else-if="game" class="detail-container">
      <div class="detail-header">
        <div class="header-info">
          <h1 class="game-title">{{ game.title || '未命名对局' }}</h1>
          <div class="game-meta">
            <div class="meta-player">
              <span class="stone-dot black"></span>
              <span class="player-name">{{ game.black_player || '黑方' }}</span>
              <span class="vs">VS</span>
              <span class="stone-dot white"></span>
              <span class="player-name">{{ game.white_player || '白方' }}</span>
            </div>
            <div class="meta-items">
              <span v-if="game.komi != null" class="meta-item">贴目: {{ game.komi }}</span>
              <span v-if="game.result" class="meta-item">结果: {{ game.result }}</span>
              <span v-if="game.game_date" class="meta-item">日期: {{ formatDate(game.game_date) }}</span>
              <span v-if="game.creator" class="meta-item">上传者: {{ game.creator.nickname || game.creator.username }}</span>
              <span class="meta-item">浏览: {{ game.view_count || 0 }}</span>
            </div>
          </div>
        </div>
        <div v-if="canEdit" class="header-actions">
          <button class="btn-edit" @click="handleEdit">编辑</button>
          <button class="btn-delete" @click="handleDelete">删除</button>
        </div>
      </div>

      <div class="detail-body">
        <div class="left-panel">
          <div class="board-wrapper" @contextmenu.prevent="showMarkerMenu = false">
            <GoBoard
              :board-size="boardSize"
              :board="board"
              :last-move="lastMove"
              :markers="currentMarkers"
              :interactive="true"
              @click-intersection="handleBoardClick"
              @right-click-intersection="handleBoardRightClick"
            />
            <div v-if="showMarkerMenu" class="marker-menu" :style="markerMenuStyle">
              <div class="menu-title">添加标记</div>
              <button class="menu-item" @click="addMarker('black_adv')">
                <span class="mi-icon black-adv-icon"></span>黑优
              </button>
              <button class="menu-item" @click="addMarker('white_adv')">
                <span class="mi-icon white-adv-icon"></span>白优
              </button>
              <button class="menu-item" @click="addMarker('key')">
                <span class="mi-icon key-icon"></span>关键手
              </button>
              <button class="menu-item" @click="addMarker('question')">
                <span class="mi-icon question-icon">?</span>疑问手
              </button>
              <button class="menu-item" @click="addMarker('good')">
                <span class="mi-icon good-icon"></span>好棋
              </button>
            </div>
          </div>

          <GamePlayer
            :move-number="moveNumber"
            :total-moves="totalMoves"
            :has-branch="hasBranch"
            :branch-count="branchCount"
            :current-branch="currentBranch"
            @prev="goPrev"
            @next="goNext"
            @first="goFirst"
            @last="goLast"
            @jump="jumpToMove"
          />

          <div class="marker-add-bar">
            <span class="hint-text">提示：右键棋盘上的位置可添加标记</span>
          </div>
        </div>

        <div class="right-panel">
          <div class="panel-section comments-section">
            <div class="section-header">
              <h3>点评</h3>
              <span class="badge">{{ comments.length }}</span>
            </div>

            <div v-if="comments.length === 0" class="empty-comments">
              <p>暂无点评，来发表第一条吧</p>
            </div>

            <div v-else class="comments-list">
              <div v-for="(group, moveNum) in groupedComments" :key="moveNum" class="comment-group">
                <div class="group-header">
                  <span class="move-badge">第 {{ moveNum }} 手</span>
                </div>
                <div v-for="comment in group" :key="comment.id" class="comment-item">
                  <div class="comment-head">
                    <span class="comment-user">{{ comment.user?.nickname || comment.user?.username || '匿名' }}</span>
                    <span class="comment-time">{{ formatDateTime(comment.created_at) }}</span>
                  </div>
                  <div v-if="comment.content" class="comment-content">{{ comment.content }}</div>
                  <button
                    v-if="comment.variation_sgf"
                    class="variation-btn"
                    @click="toggleVariation(comment)"
                  >
                    {{ showingVariation === comment.id ? '隐藏变化图' : '查看变化图' }}
                  </button>
                </div>
              </div>
            </div>

            <div v-if="userStore.isLoggedIn" class="comment-input-area">
              <textarea
                v-model="newCommentContent"
                placeholder="写下你的点评..."
                class="comment-textarea"
                rows="3"
              ></textarea>
              <div class="comment-actions">
                <span class="current-move-hint">当前手: 第 {{ moveNumber }} 手</span>
                <button class="btn-submit" @click="submitComment" :disabled="submitting">
                  {{ submitting ? '提交中...' : '发表点评' }}
                </button>
              </div>
            </div>
            <div v-else class="login-hint">
              登录后可发表点评
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getGame, deleteGame } from '@/api/game'
import { listCommentsByGame, createComment } from '@/api/comment'
import { listMarkersByGame, createMarker, deleteMarker } from '@/api/marker'
import { Parse, NewGameState, CollectAllPaths, MoveToCoord } from '@/utils/sgf'
import { useUserStore } from '@/stores/user'
import GoBoard from '@/components/GoBoard.vue'
import GamePlayer from '@/components/GamePlayer.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const game = ref(null)
const loading = ref(false)
const gameState = ref(null)
const allPaths = ref([])
const moveNumberToPath = ref({})
const comments = ref([])
const markers = ref([])
const boardSize = ref(19)
const newCommentContent = ref('')
const submitting = ref(false)
const showingVariation = ref(null)
const showMarkerMenu = ref(false)
const markerMenuPos = reactive({ x: 0, y: 0 })
const markerMenuTarget = reactive({ x: -1, y: -1 })

const canEdit = computed(() => {
  if (!userStore.isLoggedIn || !game.value) return false
  if (userStore.user?.role === 'admin') return true
  if (game.value.creator && game.value.creator.id === userStore.user?.id) return true
  return false
})

const board = computed(() => {
  return gameState.value ? gameState.value.board : []
})

const moveNumber = computed(() => {
  return gameState.value ? gameState.value.moveNumber : 0
})

const lastMove = computed(() => {
  if (!gameState.value || !gameState.value.current) return null
  const node = gameState.value.current
  let move = null
  if (node.properties['B'] && node.properties['B'].length > 0) {
    move = node.properties['B'][0]
  } else if (node.properties['W'] && node.properties['W'].length > 0) {
    move = node.properties['W'][0]
  }
  if (!move) return null
  const { x, y, valid } = MoveToCoord(move, boardSize.value)
  if (!valid || x < 0) return null
  return { x, y }
})

const totalMoves = computed(() => {
  let max = 0
  for (const path of allPaths.value) {
    if (path.length - 1 > max) {
      max = path.length - 1
    }
  }
  return max
})

const hasBranch = computed(() => {
  if (!gameState.value || !gameState.value.current) return false
  return gameState.value.current.children.length > 0
})

const branchCount = computed(() => {
  if (!gameState.value || !gameState.value.current) return 0
  return gameState.value.current.children.length
})

const currentBranch = computed(() => {
  if (!gameState.value) return 0
  const path = gameState.value.path
  return path.length > 1 ? path[path.length - 1] : 0
})

const currentMarkers = computed(() => {
  if (!gameState.value || !markers.value.length) return []
  const pathKey = JSON.stringify(gameState.value.path)
  const pathMoveNum = gameState.value.path.length - 1
  return markers.value
    .filter(m => {
      if (m.move_number == null) return false
      if (m.move_number > pathMoveNum) return false
      const targetPath = moveNumberToPath.value[m.move_number]
      if (!targetPath) return false
      const targetKey = JSON.stringify(targetPath)
      return targetKey === pathKey
    })
    .map(m => ({
      x: m.x,
      y: m.y,
      type: m.marker_type || m.type
    }))
})

const groupedComments = computed(() => {
  const groups = {}
  for (const c of comments.value) {
    const key = c.move_number || 0
    if (!groups[key]) groups[key] = []
    groups[key].push(c)
  }
  const sorted = {}
  Object.keys(groups).sort((a, b) => Number(a) - Number(b)).forEach(k => {
    sorted[k] = groups[k]
  })
  return sorted
})

const markerMenuStyle = computed(() => ({
  left: markerMenuPos.x + 'px',
  top: markerMenuPos.y + 'px'
}))

function formatDate(str) {
  if (!str) return ''
  const d = new Date(str)
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function formatDateTime(str) {
  if (!str) return ''
  const d = new Date(str)
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${y}-${m}-${day} ${hh}:${mm}`
}

function buildMoveNumberMap() {
  const map = {}
  if (!gameState.value) return
  for (const path of allPaths.value) {
    const moveNum = path.length - 1
    if (map[moveNum] == null) {
      map[moveNum] = path
    }
  }
  moveNumberToPath.value = map
}

async function loadGame() {
  loading.value = true
  try {
    game.value = await getGame(route.params.id)
    const tree = Parse(game.value.sgf_content)
    boardSize.value = tree.boardSize
    gameState.value = NewGameState(tree)

    const paths = []
    CollectAllPaths(tree.root, [0], paths)
    allPaths.value = paths
    buildMoveNumberMap()

    await loadComments()
    await loadMarkers()
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function loadComments() {
  try {
    const res = await listCommentsByGame(route.params.id)
    comments.value = res.comments || res || []
  } catch (e) {
    comments.value = []
  }
}

async function loadMarkers() {
  try {
    const res = await listMarkersByGame(route.params.id)
    markers.value = res.markers || res || []
  } catch (e) {
    markers.value = []
  }
}

function goPrev() {
  if (gameState.value && gameState.value.Previous()) {
    showMarkerMenu.value = false
  }
}

function goNext(branchIndex = 0) {
  if (gameState.value && gameState.value.Next(branchIndex)) {
    showMarkerMenu.value = false
  }
}

function goFirst() {
  if (!gameState.value) return
  gameState.value.JumpTo([0])
  showMarkerMenu.value = false
}

function goLast() {
  if (!gameState.value || allPaths.value.length === 0) return
  let longest = allPaths.value[0]
  for (const p of allPaths.value) {
    if (p.length > longest.length) longest = p
  }
  gameState.value.JumpTo(longest)
  showMarkerMenu.value = false
}

function jumpToMove(targetMoveNum) {
  if (!gameState.value) return
  if (targetMoveNum === 0) {
    gameState.value.JumpTo([0])
    return
  }
  const current = gameState.value.path
  let bestPath = null
  let bestScore = -1
  for (const p of allPaths.value) {
    if (p.length - 1 < targetMoveNum) continue
    const subPath = p.slice(0, targetMoveNum + 1)
    let score = 0
    const minLen = Math.min(current.length, subPath.length)
    for (let i = 0; i < minLen; i++) {
      if (current[i] === subPath[i]) score++
      else break
    }
    if (score > bestScore) {
      bestScore = score
      bestPath = subPath
    }
  }
  if (bestPath) {
    gameState.value.JumpTo(bestPath)
  }
  showMarkerMenu.value = false
}

function handleBoardClick(x, y) {
  if (!gameState.value || !gameState.value.current) return
  const children = gameState.value.current.children
  for (let i = 0; i < children.length; i++) {
    const child = children[i]
    let moveVal = null
    if (child.properties['B'] && child.properties['B'].length > 0) {
      moveVal = child.properties['B'][0]
    } else if (child.properties['W'] && child.properties['W'].length > 0) {
      moveVal = child.properties['W'][0]
    }
    if (!moveVal) continue
    const { x: mx, y: my, valid } = MoveToCoord(moveVal, boardSize.value)
    if (valid && mx === x && my === y) {
      gameState.value.Next(i)
      showMarkerMenu.value = false
      return
    }
  }
}

function handleBoardRightClick(x, y) {
  if (!userStore.isLoggedIn) return
  markerMenuTarget.x = x
  markerMenuTarget.y = y
  showMarkerMenu.value = true
  const container = document.querySelector('.board-wrapper')
  if (container) {
    const rect = container.getBoundingClientRect()
    const padding = 35
    const size = 600
    const cellSize = (size - padding * 2) / (boardSize.value - 1)
    markerMenuPos.x = padding + x * cellSize + 15
    markerMenuPos.y = padding + y * cellSize + 15
  }
}

async function addMarker(type) {
  if (!game.value || !gameState.value) return
  showMarkerMenu.value = false
  try {
    const data = {
      game_id: game.value.id,
      move_number: moveNumber.value,
      x: markerMenuTarget.x,
      y: markerMenuTarget.y,
      marker_type: type
    }
    await createMarker(data)
    await loadMarkers()
  } catch (e) {
    console.error(e)
  }
}

async function submitComment() {
  if (!newCommentContent.value.trim()) return
  submitting.value = true
  try {
    await createComment({
      game_id: game.value.id,
      move_number: moveNumber.value,
      content: newCommentContent.value.trim()
    })
    newCommentContent.value = ''
    await loadComments()
  } catch (e) {
    console.error(e)
  } finally {
    submitting.value = false
  }
}

function toggleVariation(comment) {
  if (showingVariation.value === comment.id) {
    showingVariation.value = null
  } else {
    showingVariation.value = comment.id
  }
}

function handleEdit() {
  alert('编辑功能待实现')
}

async function handleDelete() {
  if (!confirm('确定删除此棋谱？')) return
  try {
    await deleteGame(route.params.id)
    router.push('/')
  } catch (e) {
    console.error(e)
  }
}

function handleDocumentClick(e) {
  if (showMarkerMenu.value) {
    const menu = document.querySelector('.marker-menu')
    if (menu && !menu.contains(e.target)) {
      showMarkerMenu.value = false
    }
  }
}

onMounted(() => {
  loadGame()
  document.addEventListener('click', handleDocumentClick)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleDocumentClick)
})

watch(() => route.params.id, () => {
  loadGame()
})
</script>

<style scoped>
.game-detail-page {
  min-height: 100vh;
  background: #f5f7fa;
}

.loading-wrap {
  text-align: center;
  padding: 120px 20px;
  color: #6b7280;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #e5e7eb;
  border-top-color: #667eea;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.detail-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 24px;
}

.detail-header {
  background: white;
  border-radius: 16px;
  padding: 24px 28px;
  margin-bottom: 20px;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  border: 1px solid #f3f4f6;
}

.header-info {
  flex: 1;
  min-width: 0;
}

.game-title {
  margin: 0 0 14px;
  font-size: 24px;
  color: #1a1a2e;
  font-weight: 700;
}

.game-meta {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.meta-player {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.stone-dot {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  flex-shrink: 0;
}

.stone-dot.black {
  background: radial-gradient(circle at 30% 30%, #4b5563, #111827);
  box-shadow: inset -1px -1px 2px rgba(0, 0, 0, 0.3);
}

.stone-dot.white {
  background: radial-gradient(circle at 30% 30%, #ffffff, #e5e7eb);
  border: 1px solid #d1d5db;
}

.player-name {
  font-size: 15px;
  font-weight: 600;
  color: #1f2937;
}

.vs {
  font-size: 13px;
  font-weight: 700;
  color: #9ca3af;
  padding: 2px 10px;
  background: #f9fafb;
  border-radius: 6px;
}

.meta-items {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.meta-item {
  font-size: 13px;
  color: #6b7280;
}

.header-actions {
  display: flex;
  gap: 10px;
  flex-shrink: 0;
}

.btn-edit,
.btn-delete {
  padding: 10px 20px;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-edit {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-edit:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 16px rgba(102, 126, 234, 0.35);
}

.btn-delete {
  background: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
}

.btn-delete:hover {
  background: #fee2e2;
}

.detail-body {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 380px;
  gap: 20px;
}

.left-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

.board-wrapper {
  position: relative;
  background: white;
  border-radius: 16px;
  padding: 24px;
  display: flex;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  border: 1px solid #f3f4f6;
}

.marker-menu {
  position: absolute;
  z-index: 100;
  background: white;
  border-radius: 12px;
  padding: 8px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  border: 1px solid #e5e7eb;
  min-width: 140px;
}

.menu-title {
  font-size: 12px;
  color: #9ca3af;
  padding: 6px 10px 8px;
  border-bottom: 1px solid #f3f4f6;
  font-weight: 600;
}

.menu-item {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border: none;
  background: transparent;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  color: #374151;
  transition: background 0.15s;
  text-align: left;
}

.menu-item:hover {
  background: #f5f3ff;
  color: #667eea;
}

.mi-icon {
  width: 20px;
  height: 20px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-weight: bold;
  font-size: 14px;
}

.black-adv-icon {
  width: 0;
  height: 0;
  border-left: 8px solid transparent;
  border-right: 8px solid transparent;
  border-bottom: 14px solid #000;
}

.white-adv-icon {
  width: 0;
  height: 0;
  border-left: 8px solid transparent;
  border-right: 8px solid transparent;
  border-bottom: 14px solid #fff;
  filter: drop-shadow(0 0 1px #000);
}

.key-icon {
  width: 14px;
  height: 14px;
  border: 2px solid #d32f2f;
  border-radius: 50%;
}

.question-icon {
  color: #ff9800;
  font-size: 16px;
  font-weight: bold;
}

.good-icon {
  width: 16px;
  height: 10px;
  border: 3px solid #4caf50;
  border-left: none;
  border-top: none;
  transform: rotate(45deg);
  margin-left: -4px;
  margin-right: -4px;
}

.marker-add-bar {
  padding: 4px 8px;
  text-align: center;
}

.hint-text {
  font-size: 12px;
  color: #9ca3af;
}

.right-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

.panel-section {
  background: white;
  border-radius: 16px;
  padding: 20px 22px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  border: 1px solid #f3f4f6;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  padding-bottom: 14px;
  border-bottom: 1px solid #f3f4f6;
}

.section-header h3 {
  margin: 0;
  font-size: 17px;
  color: #1a1a2e;
  font-weight: 700;
}

.badge {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-size: 12px;
  padding: 2px 10px;
  border-radius: 10px;
  font-weight: 600;
}

.empty-comments {
  text-align: center;
  padding: 32px 16px;
  color: #9ca3af;
  font-size: 14px;
}

.comments-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
  max-height: 500px;
  overflow-y: auto;
  padding-right: 4px;
}

.comments-list::-webkit-scrollbar {
  width: 6px;
}

.comments-list::-webkit-scrollbar-thumb {
  background: #e5e7eb;
  border-radius: 3px;
}

.comment-group {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.group-header {
  margin-bottom: 2px;
}

.move-badge {
  display: inline-block;
  padding: 4px 12px;
  background: #f5f3ff;
  color: #667eea;
  font-size: 12px;
  font-weight: 600;
  border-radius: 6px;
}

.comment-item {
  background: #fafafa;
  border-radius: 10px;
  padding: 12px 14px;
  border: 1px solid #f3f4f6;
}

.comment-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}

.comment-user {
  font-size: 13px;
  font-weight: 600;
  color: #1f2937;
}

.comment-time {
  font-size: 12px;
  color: #9ca3af;
}

.comment-content {
  font-size: 14px;
  color: #374151;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}

.variation-btn {
  margin-top: 8px;
  padding: 4px 12px;
  border: 1px solid #c7d2fe;
  background: #eef2ff;
  color: #4f46e5;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.variation-btn:hover {
  background: #e0e7ff;
}

.comment-input-area {
  margin-top: 18px;
  padding-top: 16px;
  border-top: 1px solid #f3f4f6;
}

.comment-textarea {
  width: 100%;
  padding: 12px 14px;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  font-size: 14px;
  font-family: inherit;
  resize: vertical;
  box-sizing: border-box;
  transition: border-color 0.2s;
}

.comment-textarea:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 4px rgba(102, 126, 234, 0.08);
}

.comment-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 10px;
  gap: 12px;
}

.current-move-hint {
  font-size: 12px;
  color: #9ca3af;
}

.btn-submit {
  padding: 10px 22px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-submit:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 6px 16px rgba(102, 126, 234, 0.35);
}

.btn-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.login-hint {
  margin-top: 18px;
  padding-top: 16px;
  border-top: 1px solid #f3f4f6;
  text-align: center;
  font-size: 13px;
  color: #9ca3af;
}

@media (max-width: 1024px) {
  .detail-body {
    grid-template-columns: 1fr;
  }

  .right-panel {
    order: 2;
  }
}

@media (max-width: 640px) {
  .detail-container {
    padding: 16px;
  }

  .detail-header {
    padding: 20px;
    flex-direction: column;
    align-items: stretch;
  }

  .header-actions {
    justify-content: flex-end;
  }

  .board-wrapper {
    padding: 12px;
    overflow-x: auto;
  }

  .panel-section {
    padding: 16px;
  }
}
</style>
