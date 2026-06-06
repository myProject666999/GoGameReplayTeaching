<template>
  <div class="page">
    <div class="page-header">
      <div class="header-left">
        <h1>死活题训练</h1>
        <p class="page-desc">提升棋力的核心练习</p>
      </div>
      <router-link v-if="userStore.isLoggedIn" to="/problems/create" class="btn-create">
        <span class="plus">+</span> 出题
      </router-link>
    </div>

    <div class="filters">
      <div class="filter-group">
        <label class="filter-label">目标</label>
        <select v-model="goal" @change="loadProblems">
          <option value="">全部</option>
          <option value="black_kill">黑先杀</option>
          <option value="black_live">黑先活</option>
          <option value="white_kill">白先杀</option>
          <option value="white_live">白先活</option>
        </select>
      </div>
      <div class="filter-group">
        <label class="filter-label">难度</label>
        <select v-model="difficulty" @change="loadProblems">
          <option value="">全部</option>
          <option value="easy">入门</option>
          <option value="medium">初级</option>
          <option value="hard">中级</option>
          <option value="expert">高级</option>
        </select>
      </div>
      <div class="filter-group search-group">
        <label class="filter-label">搜索</label>
        <div class="search-input">
          <span class="search-icon">🔍</span>
          <input
            v-model="keyword"
            placeholder="搜索题目..."
            @keyup.enter="loadProblems"
          />
        </div>
      </div>
      <button @click="loadProblems" class="btn-filter">
        {{ loading ? '加载中...' : '搜索' }}
      </button>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else-if="problems.length === 0" class="empty">
      <div class="empty-icon">🧩</div>
      <p class="empty-title">暂无题目</p>
      <p class="empty-desc">{{ hasFilter ? '没有找到匹配的题目，试试调整筛选条件' : '还没有题目，快来出题吧' }}</p>
      <router-link v-if="userStore.isLoggedIn && !hasFilter" to="/problems/create" class="btn-empty">
        我要出题
      </router-link>
    </div>

    <div v-else class="problem-grid">
      <div
        v-for="problem in problems"
        :key="problem.id"
        class="problem-card"
        @click="$router.push(`/problems/${problem.id}`)"
      >
        <div class="card-top">
          <span :class="['tag goal', problem.goal]">
            {{ formatGoal(problem.goal) }}
          </span>
          <span :class="['tag diff', problem.difficulty]">
            {{ formatDifficulty(problem.difficulty) }}
          </span>
        </div>

        <h3 class="problem-title">{{ problem.title || '未命名题目' }}</h3>

        <p v-if="problem.description" class="problem-desc">
          {{ problem.description.length > 60 ? problem.description.slice(0, 60) + '...' : problem.description }}
        </p>

        <div class="card-footer">
          <div class="card-info">
            <span class="info-item">
              <span class="info-icon">⬛</span>
              {{ problem.board_size }}路
            </span>
            <span v-if="problem.creator" class="info-item">
              <span class="info-icon">👤</span>
              {{ problem.creator.nickname || problem.creator.username }}
            </span>
          </div>
          <span class="arrow">→</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { listProblems } from '@/api/problem'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const problems = ref([])
const keyword = ref('')
const goal = ref('')
const difficulty = ref('')
const loading = ref(false)

const hasFilter = computed(() => keyword.value || goal.value || difficulty.value)

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

async function loadProblems() {
  loading.value = true
  try {
    const res = await listProblems({
      keyword: keyword.value,
      goal: goal.value,
      difficulty: difficulty.value
    })
    problems.value = res.problems || []
  } catch (e) {
    problems.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadProblems()
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
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
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
  box-shadow: 0 8px 20px rgba(245,87,108,0.3);
}
.plus {
  font-size: 18px;
  line-height: 1;
}
.filters {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  align-items: flex-end;
  margin-bottom: 28px;
  padding: 24px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
}
.filter-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.search-group {
  flex: 1;
  min-width: 240px;
}
.filter-label {
  font-size: 13px;
  font-weight: 600;
  color: #6b7280;
}
.filter-group select {
  padding: 10px 36px 10px 14px;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  font-size: 14px;
  background: white;
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s;
  min-width: 120px;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%236b7280' d='M6 8.5L1.5 4h9z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
}
.filter-group select:focus {
  outline: none;
  border-color: #f5576c;
  box-shadow: 0 0 0 4px rgba(245,87,108,0.08);
}
.search-input {
  position: relative;
  display: flex;
  align-items: center;
}
.search-icon {
  position: absolute;
  left: 14px;
  font-size: 15px;
  opacity: 0.5;
}
.search-input input {
  width: 100%;
  padding: 10px 14px 10px 40px;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  font-size: 14px;
  background: white;
  transition: border-color 0.2s, box-shadow 0.2s;
  box-sizing: border-box;
}
.search-input input:focus {
  outline: none;
  border-color: #f5576c;
  box-shadow: 0 0 0 4px rgba(245,87,108,0.08);
}
.btn-filter {
  padding: 10px 28px;
  background: #1f2937;
  color: white;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: background 0.2s, transform 0.2s;
  height: 42px;
}
.btn-filter:hover:not(:disabled) {
  background: #374151;
  transform: translateY(-1px);
}
.btn-filter:disabled {
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
  border-top-color: #f5576c;
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
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
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
  box-shadow: 0 8px 20px rgba(245,87,108,0.3);
}
.problem-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
}
.problem-card {
  background: white;
  border-radius: 16px;
  padding: 22px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
  border: 1px solid #f3f4f6;
  display: flex;
  flex-direction: column;
}
.problem-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 30px rgba(0,0,0,0.08);
}
.card-top {
  display: flex;
  gap: 8px;
  margin-bottom: 14px;
}
.tag {
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
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
.problem-title {
  font-size: 18px;
  color: #1a1a2e;
  margin: 0 0 10px;
  font-weight: 600;
  line-height: 1.4;
}
.problem-desc {
  color: #6b7280;
  font-size: 13px;
  margin: 0 0 16px;
  line-height: 1.5;
  flex: 1;
}
.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 14px;
  border-top: 1px solid #f3f4f6;
}
.card-info {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
.info-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #6b7280;
}
.info-icon {
  font-size: 12px;
  opacity: 0.7;
}
.arrow {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f3f4f6;
  border-radius: 50%;
  color: #6b7280;
  font-size: 14px;
  transition: all 0.2s;
}
.problem-card:hover .arrow {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  color: white;
  transform: translateX(3px);
}
</style>
