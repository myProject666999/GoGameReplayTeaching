<template>
  <div class="game-player">
    <div class="player-main">
      <div class="move-info">
        <span class="move-text">第 {{ moveNumber }} / {{ totalMoves }} 手</span>
      </div>
      <div class="control-buttons">
        <button class="ctrl-btn" @click="$emit('first')" :disabled="moveNumber <= 0" title="跳到开始">
          <span class="btn-icon">|&lt;</span>
        </button>
        <button class="ctrl-btn" @click="$emit('prev')" :disabled="moveNumber <= 0" title="上一手">
          <span class="btn-icon">&lt;</span>
        </button>
        <button class="ctrl-btn" @click="$emit('next', 0)" :disabled="!hasBranch || moveNumber >= totalMoves" title="下一手">
          <span class="btn-icon">&gt;</span>
        </button>
        <button class="ctrl-btn" @click="$emit('last')" :disabled="moveNumber >= totalMoves" title="跳到末尾">
          <span class="btn-icon">&gt;|</span>
        </button>
      </div>
      <div class="slider-container">
        <input
          type="range"
          :min="0"
          :max="totalMoves"
          :value="moveNumber"
          @input="handleSliderChange"
          class="move-slider"
        />
      </div>
    </div>
    <div v-if="branchCount > 1" class="branch-container">
      <span class="branch-label">分支:</span>
      <div class="branch-buttons">
        <button
          v-for="i in branchCount"
          :key="i - 1"
          class="branch-btn"
          :class="{ active: (i - 1) === currentBranch }"
          @click="$emit('next', i - 1)"
        >
          {{ i }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  moveNumber: {
    type: Number,
    default: 0
  },
  totalMoves: {
    type: Number,
    default: 0
  },
  hasBranch: {
    type: Boolean,
    default: false
  },
  branchCount: {
    type: Number,
    default: 0
  },
  currentBranch: {
    type: Number,
    default: 0
  }
})

const emit = defineEmits(['prev', 'next', 'first', 'last', 'jump'])

function handleSliderChange(e) {
  const val = parseInt(e.target.value, 10)
  emit('jump', val)
}
</script>

<style scoped>
.game-player {
  background: white;
  border-radius: 12px;
  padding: 16px 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  border: 1px solid #f3f4f6;
}

.player-main {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.move-info {
  flex-shrink: 0;
}

.move-text {
  font-size: 15px;
  font-weight: 600;
  color: #1f2937;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.control-buttons {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.ctrl-btn {
  width: 40px;
  height: 40px;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  background: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  color: #374151;
  font-weight: 600;
  font-size: 13px;
}

.ctrl-btn:hover:not(:disabled) {
  border-color: #667eea;
  background: #f5f3ff;
  color: #667eea;
  transform: translateY(-1px);
}

.ctrl-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.slider-container {
  flex: 1;
  min-width: 160px;
}

.move-slider {
  width: 100%;
  height: 6px;
  border-radius: 3px;
  background: #e5e7eb;
  outline: none;
  -webkit-appearance: none;
  appearance: none;
  cursor: pointer;
}

.move-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  cursor: pointer;
  border: 3px solid white;
  box-shadow: 0 2px 6px rgba(102, 126, 234, 0.4);
  transition: transform 0.15s;
}

.move-slider::-webkit-slider-thumb:hover {
  transform: scale(1.15);
}

.move-slider::-moz-range-thumb {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  cursor: pointer;
  border: 3px solid white;
  box-shadow: 0 2px 6px rgba(102, 126, 234, 0.4);
}

.branch-container {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px solid #f3f4f6;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.branch-label {
  font-size: 13px;
  color: #6b7280;
  font-weight: 500;
}

.branch-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.branch-btn {
  min-width: 36px;
  height: 34px;
  padding: 0 12px;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  background: white;
  cursor: pointer;
  font-weight: 600;
  font-size: 13px;
  color: #374151;
  transition: all 0.2s;
}

.branch-btn:hover {
  border-color: #667eea;
  color: #667eea;
}

.branch-btn.active {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-color: transparent;
  color: white;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.35);
}
</style>
