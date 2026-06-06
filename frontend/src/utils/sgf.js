export class Node {
  constructor() {
    this.properties = {}
    this.children = []
    this.parent = null
  }
}

export class GameTree {
  constructor() {
    this.root = null
    this.boardSize = 19
  }
}

class Parser {
  constructor(input) {
    this.input = input
    this.pos = 0
  }

  skipWhitespace() {
    while (this.pos < this.input.length) {
      const ch = this.input[this.pos]
      if (ch === ' ' || ch === '\t' || ch === '\n' || ch === '\r') {
        this.pos++
      } else {
        break
      }
    }
  }

  parsePropIdent() {
    const start = this.pos
    while (this.pos < this.input.length) {
      const ch = this.input[this.pos]
      if (ch >= 'A' && ch <= 'Z') {
        this.pos++
      } else {
        break
      }
    }
    return this.input.substring(start, this.pos)
  }

  parsePropValue() {
    if (this.pos >= this.input.length || this.input[this.pos] !== '[') {
      throw new Error(`expected '[' at position ${this.pos}`)
    }
    this.pos++

    let sb = ''
    while (this.pos < this.input.length) {
      if (this.input[this.pos] === '\\' && this.pos + 1 < this.input.length) {
        this.pos++
        sb += this.input[this.pos]
        this.pos++
        continue
      }
      if (this.input[this.pos] === ']') {
        this.pos++
        return sb
      }
      sb += this.input[this.pos]
      this.pos++
    }
    throw new Error('unclosed property value')
  }

  parseNode() {
    if (this.pos >= this.input.length || this.input[this.pos] !== ';') {
      throw new Error(`expected ';' at position ${this.pos}`)
    }
    this.pos++

    const node = new Node()

    while (this.pos < this.input.length) {
      this.skipWhitespace()
      if (this.pos >= this.input.length) {
        break
      }
      const ch = this.input[this.pos]
      if (ch === ';' || ch === '(' || ch === ')') {
        break
      }

      const ident = this.parsePropIdent()
      if (ident === '') {
        break
      }

      const values = []
      while (this.pos < this.input.length && this.input[this.pos] === '[') {
        const val = this.parsePropValue()
        values.push(val)
      }
      if (values.length > 0) {
        node.properties[ident] = values
      }
    }

    return node
  }

  parseGameTree(parent) {
    if (this.pos >= this.input.length || this.input[this.pos] !== '(') {
      throw new Error(`expected '(' at position ${this.pos}`)
    }
    this.pos++
    this.skipWhitespace()

    let root = null
    let prev = null

    while (this.pos < this.input.length && this.input[this.pos] !== ')') {
      if (this.input[this.pos] === '(') {
        if (prev === null) {
          throw new Error(`unexpected '(' at position ${this.pos}`)
        }
        const child = this.parseGameTree(prev)
        prev.children.push(child)
      } else if (this.input[this.pos] === ';') {
        const node = this.parseNode()
        node.parent = prev
        if (prev !== null) {
          prev.children.push(node)
        } else {
          root = node
        }
        prev = node
      } else {
        this.skipWhitespace()
        if (this.pos >= this.input.length) {
          break
        }
        if (this.input[this.pos] !== ')' && this.input[this.pos] !== '(' && this.input[this.pos] !== ';') {
          throw new Error(`unexpected character '${this.input[this.pos]}' at position ${this.pos}`)
        }
      }
      this.skipWhitespace()
    }

    if (this.pos >= this.input.length || this.input[this.pos] !== ')') {
      throw new Error(`expected ')' at position ${this.pos}`)
    }
    this.pos++

    return root
  }
}

export function Parse(input) {
  const p = new Parser(input)
  p.skipWhitespace()

  const tree = new GameTree()
  if (p.pos >= input.length || input[p.pos] !== '(') {
    throw new Error("SGF must start with '('")
  }

  const root = p.parseGameTree(null)
  tree.root = root

  if (root.properties['SZ'] && root.properties['SZ'].length > 0) {
    const size = parseInt(root.properties['SZ'][0], 10)
    if (!isNaN(size)) {
      tree.boardSize = size
    }
  }
  if (tree.boardSize === 0) {
    tree.boardSize = 19
  }

  return tree
}

export function MoveToCoord(move, boardSize) {
  if (move.length < 2) {
    return { x: -1, y: -1, valid: false }
  }
  if (move === '' || move === 'tt' || move === 'pass') {
    return { x: -1, y: -1, valid: true }
  }
  const x = move.charCodeAt(0) - 'a'.charCodeAt(0)
  const y = move.charCodeAt(1) - 'a'.charCodeAt(0)
  if (x < 0 || x >= boardSize || y < 0 || y >= boardSize) {
    return { x: -1, y: -1, valid: false }
  }
  return { x, y, valid: true }
}

export function CoordToMove(x, y) {
  return String.fromCharCode('a'.charCodeAt(0) + x) + String.fromCharCode('a'.charCodeAt(0) + y)
}

export class GameState {
  constructor(tree) {
    this.boardSize = tree.boardSize
    this.board = []
    for (let i = 0; i < tree.boardSize; i++) {
      this.board[i] = []
      for (let j = 0; j < tree.boardSize; j++) {
        this.board[i][j] = ''
      }
    }
    this.moveNumber = 0
    this.current = tree.root
    this.path = [0]
    this.komi = 0
    this.captures = { B: 0, W: 0 }
    this.applyNode(tree.root, true)
  }

  applyNode(node, isRoot) {
    if (!isRoot) {
      if (node.properties['B'] && node.properties['B'].length > 0) {
        const b = node.properties['B'][0]
        const { x, y, valid } = MoveToCoord(b, this.boardSize)
        if (valid && x >= 0) {
          this.board[y][x] = 'B'
          this.moveNumber++
          this.removeCaptures(x, y, 'W')
        } else if (b === '' || b === 'tt') {
          this.moveNumber++
        }
      }
      if (node.properties['W'] && node.properties['W'].length > 0) {
        const w = node.properties['W'][0]
        const { x, y, valid } = MoveToCoord(w, this.boardSize)
        if (valid && x >= 0) {
          this.board[y][x] = 'W'
          this.moveNumber++
          this.removeCaptures(x, y, 'B')
        } else if (w === '' || w === 'tt') {
          this.moveNumber++
        }
      }
    }

    if (node.properties['AB']) {
      for (const m of node.properties['AB']) {
        const { x, y, valid } = MoveToCoord(m, this.boardSize)
        if (valid && x >= 0) {
          this.board[y][x] = 'B'
        }
      }
    }
    if (node.properties['AW']) {
      for (const m of node.properties['AW']) {
        const { x, y, valid } = MoveToCoord(m, this.boardSize)
        if (valid && x >= 0) {
          this.board[y][x] = 'W'
        }
      }
    }

    if (node.properties['KM'] && node.properties['KM'].length > 0) {
      const k = parseFloat(node.properties['KM'][0])
      if (!isNaN(k)) {
        this.komi = k
      }
    }
  }

  removeCaptures(x, y, opponent) {
    const visited = {}
    const directions = [[-1, 0], [1, 0], [0, -1], [0, 1]]

    for (const d of directions) {
      const nx = x + d[0]
      const ny = y + d[1]
      if (nx < 0 || nx >= this.boardSize || ny < 0 || ny >= this.boardSize) {
        continue
      }
      const key = `${nx},${ny}`
      if (visited[key]) {
        continue
      }
      if (this.board[ny][nx] === opponent) {
        const group = this.findGroup(nx, ny, visited)
        if (!this.hasLiberty(group)) {
          for (const pos of group) {
            this.board[pos[1]][pos[0]] = ''
            this.captures[opponent]++
          }
        }
      }
    }
  }

  findGroup(x, y, visited) {
    const color = this.board[y][x]
    const directions = [[-1, 0], [1, 0], [0, -1], [0, 1]]
    const group = []
    const stack = [[x, y]]

    while (stack.length > 0) {
      const pos = stack.pop()
      const key = `${pos[0]},${pos[1]}`
      if (visited[key]) {
        continue
      }
      visited[key] = true
      if (pos[0] < 0 || pos[0] >= this.boardSize || pos[1] < 0 || pos[1] >= this.boardSize) {
        continue
      }
      if (this.board[pos[1]][pos[0]] !== color) {
        continue
      }
      group.push(pos)
      for (const d of directions) {
        const nx = pos[0] + d[0]
        const ny = pos[1] + d[1]
        const nkey = `${nx},${ny}`
        if (!visited[nkey]) {
          stack.push([nx, ny])
        }
      }
    }
    return group
  }

  hasLiberty(group) {
    const directions = [[-1, 0], [1, 0], [0, -1], [0, 1]]
    for (const pos of group) {
      for (const d of directions) {
        const nx = pos[0] + d[0]
        const ny = pos[1] + d[1]
        if (nx < 0 || nx >= this.boardSize || ny < 0 || ny >= this.boardSize) {
          continue
        }
        if (this.board[ny][nx] === '') {
          return true
        }
      }
    }
    return false
  }

  Next(childIndex) {
    if (this.current === null || childIndex >= this.current.children.length) {
      return false
    }
    const next = this.current.children[childIndex]
    this.applyNode(next, false)
    this.path.push(childIndex)
    this.current = next
    return true
  }

  Previous() {
    if (this.current === null || this.current.parent === null) {
      return false
    }
    this.current = this.current.parent
    this.path.pop()
    this.rebuild()
    return true
  }

  rebuild() {
    this.board = []
    for (let i = 0; i < this.boardSize; i++) {
      this.board[i] = []
      for (let j = 0; j < this.boardSize; j++) {
        this.board[i][j] = ''
      }
    }
    this.moveNumber = 0
    this.captures = { B: 0, W: 0 }

    const path = [...this.path]
    let node = this.current
    while (node.parent !== null) {
      node = node.parent
    }

    this.applyNode(node, true)
    let current = node
    for (let i = 1; i < path.length; i++) {
      if (path[i] < current.children.length) {
        current = current.children[path[i]]
        this.applyNode(current, false)
      }
    }
  }

  JumpTo(targetPath) {
    if (targetPath.length === 0) {
      return false
    }
    const path = [...targetPath]
    let node = this.current
    while (node.parent !== null) {
      node = node.parent
    }
    for (let i = 1; i < path.length; i++) {
      if (path[i] >= node.children.length) {
        return false
      }
      node = node.children[path[i]]
    }
    this.current = node
    this.path = path
    this.rebuild()
    return true
  }
}

export function CollectAllPaths(node, currentPath, result) {
  const path = [...currentPath]
  result.push(path)
  for (let i = 0; i < node.children.length; i++) {
    const childPath = [...path, i]
    CollectAllPaths(node.children[i], childPath, result)
  }
}

export function GetPathFromRoot(node) {
  const path = []
  let current = node
  while (current.parent !== null) {
    const parent = current.parent
    for (let i = 0; i < parent.children.length; i++) {
      if (parent.children[i] === current) {
        path.unshift(i)
        break
      }
    }
    current = parent
  }
  path.unshift(0)
  return path
}

export function NewGameState(tree) {
  return new GameState(tree)
}
