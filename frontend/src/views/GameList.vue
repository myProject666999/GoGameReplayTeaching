<template>
  <div class="page">
    <div class="page-header">
      <div class="header-left">
        <h1>棋谱库</h1>
        <p class="page-desc">探索精彩的围棋对局</p>
      </div>
      <router-link v-if="userStore.isLoggedIn" to="/games/upload" class="btn-create">
        <span class="plus">+</span> 上传棋谱
      </router-link>
    </div>

    <div class="search-bar">
      <div class="search-input">
        <span class="search-icon">🔍</span>
        <input
          v-model="keyword"
          type="text"
          placeholder="搜索棋谱标题、黑白方棋手..."
          @keyup.enter="loadGames"
        />
        <button v-if="keyword" @click="keyword = ''; loadGames()" class="clear-btn">✕</button>
      </div>
      <button @click="loadGames" class="btn-search" :disabled="loading">
        {{ loading ? '加载中...' : '搜索' }}
      </button>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else-if="games.length === 0" class="empty">
      <div class="empty-icon">📂</div>
      <p class="empty-title">暂无棋谱</p>
      <p class="empty-desc">{{ keyword ? '没有找到匹配的棋谱，试试其他关键词' : '还没有棋谱，快来上传第一局吧' }}</p>
      <router-link v-if="userStore.isLoggedIn && !keyword" to="/games/upload" class="btn-empty">
        上传棋谱
      </router-link>
    </div>

    <div v-else class="game-grid">
      <div
        v-for="game in games"
        :key="game.id"
        class="game-card"
        @click="$router.push(`/games/${game.id}`)"
      >
        <div class="card-header">
          <h3 class="game-title">{{ game.title || '未命名对局' }}</h3>
          <span v-if="game.board_size" class="board-size">{{ game.board_size }}路</span>
        </div>

        <div class="players">
          <div class="player black">
            <span class="player-dot black-dot"></span>
            <span class="player-name">{{ game.black_player || '黑方' }}</span>
          </div>
          <span class="vs">VS</span>
          <div class="player white">
            <span class="player-dot white-dot"></span>
            <span class="player-name">{{ game.white_player || '白方' }}</span>
          </div>
        </div>

        <div v-if="game.result" class="result">
          <span class="result-icon">🏆</span>
          {{ game.result }}
        </div>

        <div class="card-footer">
          <div class="meta-left">
            <span v-if="game.creator" class="meta-item">
              <span class="meta-icon">👤</span>
              {{ game.creator.nickname || game.creator.username }}
            </span>
          </div>
          <div class="meta-right">
            <span v-if="game.created_at" class="meta-item">
              <span class="meta-icon">📅</span>
              {{ formatDate(game.created_at) }}
            </span>
            <span class="meta-item">
              <span class="meta-icon">👁</span>
              {{ game.view_count || 0 }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { listGames } from '@/api/game'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const games = ref([])
const keyword = ref('')
const loading = ref(false)

function formatDate(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

async function loadGames() {
  loading.value = true
  try {
    const res = await listGames({ keyword: keyword.value })
    games.value = res.games || []
  } catch (e) {
    games.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadGames()
})
</script>

<style scoped>
.page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 32px 24px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
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
.btn-create {
  padding: 12px 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 12px;
  cursor: pointer;
  font-size: 15px;
  font-weight: 600;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: transform 0.2s, box-shadow 0.2s;
}
.btn-create:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(102,126,234,0.3);
}
.plus {
  font-size: 18px;
  line-height: 1;
}
.search-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 28px;
}
.search-input {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
  max-width: 520px;
}
.search-icon {
  position: absolute;
  left: 16px;
  font-size: 16px;
  opacity: 0.5;
}
.search-input input {
  width: 100%;
  padding: 14px 44px 14px 44px;
  border: 2px solid #e5e7eb;
  border-radius: 12px;
  font-size: 15px;
  background: white;
  transition: border-color 0.2s, box-shadow 0.2s;
  box-sizing: border-box;
}
.search-input input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 4px rgba(102,126,234,0.08);
}
.clear-btn {
  position: absolute;
  right: 12px;
  width: 24px;
  height: 24px;
  border: none;
  background: #e5e7eb;
  border-radius: 50%;
  cursor: pointer;
  font-size: 12px;
  color: #6b7280;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}
.clear-btn:hover {
  background: #d1d5db;
}
.btn-search {
  padding: 14px 32px;
  background: #1f2937;
  color: white;
  border: none;
  border-radius: 12px;
  cursor: pointer;
  font-size: 15px;
  font-weight: 600;
  transition: background 0.2s, transform 0.2s;
}
.btn-search:hover:not(:disabled) {
  background: #374151;
  transform: translateY(-1px);
}
.btn-search:disabled {
  opacity: 0.6;
  cursor: not-allowed;
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
  border-top-color: #667eea;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 16px;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}
.empty {
  text-align: center;
  padding: 80px 20px;
  background: white;
  border-radius: 16px;
}
.empty-icon {
  font-size: 64px;
  margin-bottom: 16px;
  opacity: 0.6;
}
.empty-title {
  font-size: 20px;
  color: #1f2937;
  margin: 0 0 8px;
  font-weight: 600;
}
.empty-desc {
  color: #6b7280;
  margin: 0 0 24px;
}
.btn-empty {
  padding: 12px 28px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  font-size: 15px;
  font-weight: 600;
  text-decoration: none;
  display: inline-block;
  transition: transform 0.2s, box-shadow 0.2s;
}
.btn-empty:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(102,126,234,0.3);
}
.game-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}
.game-card {
  background: white;
  border-radius: 16px;
  padding: 22px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
  border: 1px solid #f3f4f6;
}
.game-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 30px rgba(0,0,0,0.08);
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 16px;
}
.game-title {
  font-size: 17px;
  color: #1a1a2e;
  margin: 0;
  font-weight: 600;
  line-height: 1.4;
  flex: 1;
}
.board-size {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}
.players {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 14px;
  padding: 14px 16px;
  background: #f9fafb;
  border-radius: 10px;
}
.player {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}
.player.white {
  justify-content: flex-end;
}
.player-dot {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  flex-shrink: 0;
}
.black-dot {
  background: radial-gradient(circle at 30% 30%, #4b5563, #111827);
  box-shadow: inset -1px -1px 2px rgba(0,0,0,0.3);
}
.white-dot {
  background: radial-gradient(circle at 30% 30%, #ffffff, #e5e7eb);
  border: 1px solid #d1d5db;
  box-shadow: inset -1px -1px 2px rgba(0,0,0,0.05);
}
.player-name {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 120px;
}
.vs {
  font-size: 12px;
  font-weight: 700;
  color: #9ca3af;
  padding: 2px 8px;
  background: white;
  border-radius: 4px;
}
.result {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: linear-gradient(135deg, #fef3c7, #fde68a);
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  color: #92400e;
  margin-bottom: 16px;
}
.result-icon {
  font-size: 14px;
}
.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 14px;
  border-top: 1px solid #f3f4f6;
  font-size: 12px;
  color: #6b7280;
}
.meta-left,
.meta-right {
  display: flex;
  align-items: center;
  gap: 12px;
}
.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
}
.meta-icon {
  font-size: 13px;
  opacity: 0.7;
}
</style>
