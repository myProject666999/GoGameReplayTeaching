<template>
  <div ref="containerRef" class="go-board-container">
    <canvas
      ref="canvasRef"
      @click="handleClick"
      @contextmenu.prevent="handleRightClick"
    ></canvas>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'

const props = defineProps({
  boardSize: {
    type: Number,
    default: 19,
    validator: (v) => [9, 13, 19].includes(v)
  },
  board: {
    type: Array,
    default: () => []
  },
  lastMove: {
    type: Object,
    default: null
  },
  markers: {
    type: Array,
    default: () => []
  },
  showCoords: {
    type: Boolean,
    default: true
  },
  interactive: {
    type: Boolean,
    default: false
  },
  currentPlayer: {
    type: String,
    default: 'B'
  }
})

const emit = defineEmits(['click-intersection', 'right-click-intersection'])

const containerRef = ref(null)
const canvasRef = ref(null)
let resizeObserver = null
let ctx = null

const padding = computed(() => {
  const base = props.showCoords ? 35 : 20
  return base
})

const canvasSize = computed(() => {
  return 600
})

const cellSize = computed(() => {
  return (canvasSize.value - padding.value * 2) / (props.boardSize - 1)
})

const starPoints = computed(() => {
  const size = props.boardSize
  if (size === 19) {
    return [
      [3, 3], [3, 9], [3, 15],
      [9, 3], [9, 9], [9, 15],
      [15, 3], [15, 9], [15, 15]
    ]
  } else if (size === 13) {
    return [
      [3, 3], [3, 9],
      [6, 6],
      [9, 3], [9, 9]
    ]
  } else {
    return [
      [2, 2], [2, 6],
      [4, 4],
      [6, 2], [6, 6]
    ]
  }
})

const colLabels = computed(() => {
  const labels = []
  for (let i = 0; i < props.boardSize; i++) {
    let charCode = 65 + i
    if (charCode >= 73) charCode++
    labels.push(String.fromCharCode(charCode))
  }
  return labels
})

const getPixelFromCoord = (x, y) => {
  return {
    px: padding.value + x * cellSize.value,
    py: padding.value + y * cellSize.value
  }
}

const getCoordFromPixel = (px, py) => {
  const x = Math.round((px - padding.value) / cellSize.value)
  const y = Math.round((py - padding.value) / cellSize.value)
  if (x < 0 || x >= props.boardSize || y < 0 || y >= props.boardSize) {
    return null
  }
  return { x, y }
}

const drawBoard = () => {
  if (!ctx) return

  const size = canvasSize.value
  ctx.clearRect(0, 0, size, size)

  ctx.fillStyle = '#DEB887'
  ctx.fillRect(0, 0, size, size)

  ctx.strokeStyle = '#000'
  ctx.lineWidth = 1

  for (let i = 0; i < props.boardSize; i++) {
    const start = getPixelFromCoord(0, i)
    const end = getPixelFromCoord(props.boardSize - 1, i)
    ctx.beginPath()
    ctx.moveTo(start.px, start.py)
    ctx.lineTo(end.px, end.py)
    ctx.stroke()

    const vStart = getPixelFromCoord(i, 0)
    const vEnd = getPixelFromCoord(i, props.boardSize - 1)
    ctx.beginPath()
    ctx.moveTo(vStart.px, vStart.py)
    ctx.lineTo(vEnd.px, vEnd.py)
    ctx.stroke()
  }

  ctx.fillStyle = '#000'
  starPoints.value.forEach(([x, y]) => {
    const { px, py } = getPixelFromCoord(x, y)
    ctx.beginPath()
    ctx.arc(px, py, 3.5, 0, Math.PI * 2)
    ctx.fill()
  })

  if (props.showCoords) {
    ctx.fillStyle = '#333'
    ctx.font = '12px sans-serif'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'

    colLabels.value.forEach((label, i) => {
      const { px } = getPixelFromCoord(i, 0)
      ctx.fillText(label, px, padding.value / 2)
      ctx.fillText(label, px, size - padding.value / 2)
    })

    for (let i = 0; i < props.boardSize; i++) {
      const label = String(props.boardSize - i)
      const { py } = getPixelFromCoord(0, i)
      ctx.textAlign = 'center'
      ctx.fillText(label, padding.value / 2, py)
      ctx.fillText(label, size - padding.value / 2, py)
    }
  }
}

const drawStone = (x, y, color) => {
  if (!ctx) return

  const { px, py } = getPixelFromCoord(x, y)
  const radius = cellSize.value * 0.42

  ctx.save()
  ctx.shadowColor = 'rgba(0, 0, 0, 0.4)'
  ctx.shadowBlur = 4
  ctx.shadowOffsetX = 2
  ctx.shadowOffsetY = 2

  const gradient = ctx.createRadialGradient(
    px - radius * 0.3, py - radius * 0.3, radius * 0.1,
    px, py, radius
  )

  if (color === 'B') {
    gradient.addColorStop(0, '#555')
    gradient.addColorStop(1, '#000')
  } else {
    gradient.addColorStop(0, '#fff')
    gradient.addColorStop(1, '#ccc')
  }

  ctx.fillStyle = gradient
  ctx.beginPath()
  ctx.arc(px, py, radius, 0, Math.PI * 2)
  ctx.fill()

  ctx.restore()
}

const drawLastMove = () => {
  if (!ctx || !props.lastMove) return

  const { x, y } = props.lastMove
  const { px, py } = getPixelFromCoord(x, y)
  const radius = cellSize.value * 0.15

  const stoneColor = props.board && props.board[y] && props.board[y][x]
  ctx.strokeStyle = stoneColor === 'W' ? '#d32f2f' : '#1976d2'
  ctx.lineWidth = 2

  ctx.beginPath()
  ctx.arc(px, py, radius, 0, Math.PI * 2)
  ctx.stroke()
}

const drawMarker = (marker) => {
  if (!ctx) return

  const { x, y, type } = marker
  const { px, py } = getPixelFromCoord(x, y)
  const size = cellSize.value * 0.3

  ctx.save()

  if (type === 'black_adv') {
    ctx.fillStyle = '#000'
    ctx.beginPath()
    ctx.moveTo(px, py - size)
    ctx.lineTo(px + size * 0.866, py + size * 0.5)
    ctx.lineTo(px - size * 0.866, py + size * 0.5)
    ctx.closePath()
    ctx.fill()
  } else if (type === 'white_adv') {
    ctx.fillStyle = '#fff'
    ctx.strokeStyle = '#000'
    ctx.lineWidth = 1.5
    ctx.beginPath()
    ctx.moveTo(px, py - size)
    ctx.lineTo(px + size * 0.866, py + size * 0.5)
    ctx.lineTo(px - size * 0.866, py + size * 0.5)
    ctx.closePath()
    ctx.fill()
    ctx.stroke()
  } else if (type === 'key') {
    ctx.strokeStyle = '#d32f2f'
    ctx.lineWidth = 2.5
    ctx.beginPath()
    ctx.arc(px, py, size, 0, Math.PI * 2)
    ctx.stroke()
  } else if (type === 'question') {
    ctx.fillStyle = '#ff9800'
    ctx.font = `bold ${size * 1.6}px sans-serif`
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText('?', px, py)
  } else if (type === 'good') {
    ctx.strokeStyle = '#4caf50'
    ctx.lineWidth = 3
    ctx.beginPath()
    ctx.moveTo(px - size * 0.6, py)
    ctx.lineTo(px - size * 0.1, py + size * 0.5)
    ctx.lineTo(px + size * 0.7, py - size * 0.5)
    ctx.stroke()
  }

  ctx.restore()
}

const draw = () => {
  if (!ctx) return

  drawBoard()

  for (let y = 0; y < props.boardSize; y++) {
    for (let x = 0; x < props.boardSize; x++) {
      const cell = props.board[y]?.[x]
      if (cell === 'B' || cell === 'W') {
        drawStone(x, y, cell)
      }
    }
  }

  drawLastMove()

  props.markers.forEach(marker => drawMarker(marker))
}

const setupCanvas = () => {
  const canvas = canvasRef.value
  if (!canvas) return

  const size = canvasSize.value
  const dpr = window.devicePixelRatio || 1

  canvas.width = size * dpr
  canvas.height = size * dpr
  canvas.style.width = size + 'px'
  canvas.style.height = size + 'px'

  ctx = canvas.getContext('2d')
  ctx.scale(dpr, dpr)

  draw()
}

const handleClick = (event) => {
  if (!props.interactive || !canvasRef.value) return

  const rect = canvasRef.value.getBoundingClientRect()
  const px = event.clientX - rect.left
  const py = event.clientY - rect.top
  const coord = getCoordFromPixel(px, py)

  if (coord) {
    emit('click-intersection', coord.x, coord.y)
  }
}

const handleRightClick = (event) => {
  if (!canvasRef.value) return

  const rect = canvasRef.value.getBoundingClientRect()
  const px = event.clientX - rect.left
  const py = event.clientY - rect.top
  const coord = getCoordFromPixel(px, py)

  if (coord) {
    emit('right-click-intersection', coord.x, coord.y)
  }
}

watch(
  () => [props.board, props.boardSize, props.lastMove, props.markers, props.showCoords],
  () => {
    draw()
  },
  { deep: true }
)

onMounted(() => {
  setupCanvas()

  if (containerRef.value && 'ResizeObserver' in window) {
    resizeObserver = new ResizeObserver(() => {
      setupCanvas()
    })
    resizeObserver.observe(containerRef.value)
  }
})

onUnmounted(() => {
  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
  }
})
</script>

<style scoped>
.go-board-container {
  display: inline-block;
  line-height: 0;
}

canvas {
  display: block;
  cursor: default;
  user-select: none;
}

.go-board-container :deep(canvas) {
  cursor: pointer;
}
</style>
