package sgf

import (
	"fmt"
	"strconv"
	"strings"
)

type Node struct {
	Properties map[string][]string `json:"properties"`
	Children   []*Node             `json:"children,omitempty"`
	Parent     *Node               `json:"-"`
}

type GameTree struct {
	Root      *Node `json:"root"`
	BoardSize int   `json:"board_size"`
}

type Parser struct {
	input string
	pos   int
}

func Parse(input string) (*GameTree, error) {
	p := &Parser{input: input, pos: 0}
	p.skipWhitespace()

	tree := &GameTree{}
	if p.pos >= len(p.input) || p.input[p.pos] != '(' {
		return nil, fmt.Errorf("SGF must start with '('")
	}

	root, err := p.parseGameTree(nil)
	if err != nil {
		return nil, err
	}
	tree.Root = root

	if sz, ok := root.Properties["SZ"]; ok && len(sz) > 0 {
		if size, err := strconv.Atoi(sz[0]); err == nil {
			tree.BoardSize = size
		}
	}
	if tree.BoardSize == 0 {
		tree.BoardSize = 19
	}

	return tree, nil
}

func (p *Parser) parseGameTree(parent *Node) (*Node, error) {
	if p.pos >= len(p.input) || p.input[p.pos] != '(' {
		return nil, fmt.Errorf("expected '(' at position %d", p.pos)
	}
	p.pos++
	p.skipWhitespace()

	var root *Node
	var prev *Node

	for p.pos < len(p.input) && p.input[p.pos] != ')' {
		if p.input[p.pos] == '(' {
			if prev == nil {
				return nil, fmt.Errorf("unexpected '(' at position %d", p.pos)
			}
			child, err := p.parseGameTree(prev)
			if err != nil {
				return nil, err
			}
			prev.Children = append(prev.Children, child)
		} else if p.input[p.pos] == ';' {
			node, err := p.parseNode()
			if err != nil {
				return nil, err
			}
			node.Parent = prev
			if prev != nil {
				prev.Children = append(prev.Children, node)
			} else {
				root = node
			}
			prev = node
		} else {
			p.skipWhitespace()
			if p.pos >= len(p.input) {
				break
			}
			if p.input[p.pos] != ')' && p.input[p.pos] != '(' && p.input[p.pos] != ';' {
				return nil, fmt.Errorf("unexpected character '%c' at position %d", p.input[p.pos], p.pos)
			}
		}
		p.skipWhitespace()
	}

	if p.pos >= len(p.input) || p.input[p.pos] != ')' {
		return nil, fmt.Errorf("expected ')' at position %d", p.pos)
	}
	p.pos++

	return root, nil
}

func (p *Parser) parseNode() (*Node, error) {
	if p.pos >= len(p.input) || p.input[p.pos] != ';' {
		return nil, fmt.Errorf("expected ';' at position %d", p.pos)
	}
	p.pos++

	node := &Node{
		Properties: make(map[string][]string),
	}

	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			break
		}
		ch := p.input[p.pos]
		if ch == ';' || ch == '(' || ch == ')' {
			break
		}

		ident, err := p.parsePropIdent()
		if err != nil {
			return nil, err
		}
		if ident == "" {
			break
		}

		var values []string
		for p.pos < len(p.input) && p.input[p.pos] == '[' {
			val, err := p.parsePropValue()
			if err != nil {
				return nil, err
			}
			values = append(values, val)
		}
		if len(values) > 0 {
			node.Properties[ident] = values
		}
	}

	return node, nil
}

func (p *Parser) parsePropIdent() (string, error) {
	start := p.pos
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch >= 'A' && ch <= 'Z' {
			p.pos++
		} else {
			break
		}
	}
	return p.input[start:p.pos], nil
}

func (p *Parser) parsePropValue() (string, error) {
	if p.pos >= len(p.input) || p.input[p.pos] != '[' {
		return "", fmt.Errorf("expected '[' at position %d", p.pos)
	}
	p.pos++

	var sb strings.Builder
	for p.pos < len(p.input) {
		if p.input[p.pos] == '\\' && p.pos+1 < len(p.input) {
			p.pos++
			sb.WriteByte(p.input[p.pos])
			p.pos++
			continue
		}
		if p.input[p.pos] == ']' {
			p.pos++
			return sb.String(), nil
		}
		sb.WriteByte(p.input[p.pos])
		p.pos++
	}
	return "", fmt.Errorf("unclosed property value")
}

func (p *Parser) skipWhitespace() {
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			p.pos++
		} else {
			break
		}
	}
}

func MoveToCoord(move string, boardSize int) (int, int, bool) {
	if len(move) < 2 {
		return -1, -1, false
	}
	if move == "" || move == "tt" || move == "pass" {
		return -1, -1, true
	}
	x := int(move[0] - 'a')
	y := int(move[1] - 'a')
	if x < 0 || x >= boardSize || y < 0 || y >= boardSize {
		return -1, -1, false
	}
	return x, y, true
}

func CoordToMove(x, y int) string {
	return string(rune('a'+x)) + string(rune('a'+y))
}

type GameState struct {
	BoardSize  int
	Board      [][]string
	MoveNumber int
	Current    *Node
	Path       []int
	Komi       float64
	Captures   map[string]int
}

func NewGameState(tree *GameTree) *GameState {
	gs := &GameState{
		BoardSize: tree.BoardSize,
		Board:     make([][]string, tree.BoardSize),
		Current:   tree.Root,
		Captures:  map[string]int{"B": 0, "W": 0},
		Path:      []int{0},
	}
	for i := range gs.Board {
		gs.Board[i] = make([]string, tree.BoardSize)
	}
	gs.applyNode(tree.Root, true)
	return gs
}

func (gs *GameState) applyNode(node *Node, isRoot bool) {
	if !isRoot {
		if b, ok := node.Properties["B"]; ok && len(b) > 0 {
			x, y, valid := MoveToCoord(b[0], gs.BoardSize)
			if valid && x >= 0 {
				gs.Board[y][x] = "B"
				gs.MoveNumber++
				gs.removeCaptures(x, y, "W")
			} else if b[0] == "" || b[0] == "tt" {
				gs.MoveNumber++
			}
		}
		if w, ok := node.Properties["W"]; ok && len(w) > 0 {
			x, y, valid := MoveToCoord(w[0], gs.BoardSize)
			if valid && x >= 0 {
				gs.Board[y][x] = "W"
				gs.MoveNumber++
				gs.removeCaptures(x, y, "B")
			} else if w[0] == "" || w[0] == "tt" {
				gs.MoveNumber++
			}
		}
	}

	if ab, ok := node.Properties["AB"]; ok {
		for _, m := range ab {
			x, y, valid := MoveToCoord(m, gs.BoardSize)
			if valid && x >= 0 {
				gs.Board[y][x] = "B"
			}
		}
	}
	if aw, ok := node.Properties["AW"]; ok {
		for _, m := range aw {
			x, y, valid := MoveToCoord(m, gs.BoardSize)
			if valid && x >= 0 {
				gs.Board[y][x] = "W"
			}
		}
	}

	if km, ok := node.Properties["KM"]; ok && len(km) > 0 {
		if k, err := strconv.ParseFloat(km[0], 64); err == nil {
			gs.Komi = k
		}
	}
}

func (gs *GameState) removeCaptures(x, y int, opponent string) {
	visited := make(map[string]bool)
	directions := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	for _, d := range directions {
		nx, ny := x+d[0], y+d[1]
		if nx < 0 || nx >= gs.BoardSize || ny < 0 || ny >= gs.BoardSize {
			continue
		}
		key := fmt.Sprintf("%d,%d", nx, ny)
		if visited[key] {
			continue
		}
		if gs.Board[ny][nx] == opponent {
			group := gs.findGroup(nx, ny, visited)
			if !gs.hasLiberty(group) {
				for _, pos := range group {
					gs.Board[pos[1]][pos[0]] = ""
					gs.Captures[opponent]++
				}
			}
		}
	}
}

func (gs *GameState) findGroup(x, y int, visited map[string]bool) [][2]int {
	color := gs.Board[y][x]
	directions := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	var group [][2]int
	stack := [][2]int{{x, y}}

	for len(stack) > 0 {
		pos := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		key := fmt.Sprintf("%d,%d", pos[0], pos[1])
		if visited[key] {
			continue
		}
		visited[key] = true
		if pos[0] < 0 || pos[0] >= gs.BoardSize || pos[1] < 0 || pos[1] >= gs.BoardSize {
			continue
		}
		if gs.Board[pos[1]][pos[0]] != color {
			continue
		}
		group = append(group, pos)
		for _, d := range directions {
			nx, ny := pos[0]+d[0], pos[1]+d[1]
			nkey := fmt.Sprintf("%d,%d", nx, ny)
			if !visited[nkey] {
				stack = append(stack, [2]int{nx, ny})
			}
		}
	}
	return group
}

func (gs *GameState) hasLiberty(group [][2]int) bool {
	directions := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for _, pos := range group {
		for _, d := range directions {
			nx, ny := pos[0]+d[0], pos[1]+d[1]
			if nx < 0 || nx >= gs.BoardSize || ny < 0 || ny >= gs.BoardSize {
				continue
			}
			if gs.Board[ny][nx] == "" {
				return true
			}
		}
	}
	return false
}

func (gs *GameState) Next(childIndex int) bool {
	if gs.Current == nil || childIndex >= len(gs.Current.Children) {
		return false
	}
	next := gs.Current.Children[childIndex]
	gs.applyNode(next, false)
	gs.Path = append(gs.Path, childIndex)
	gs.Current = next
	return true
}

func (gs *GameState) Previous() bool {
	if gs.Current == nil || gs.Current.Parent == nil {
		return false
	}
	gs.Current = gs.Current.Parent
	gs.Path = gs.Path[:len(gs.Path)-1]
	gs.rebuild()
	return true
}

func (gs *GameState) rebuild() {
	gs.Board = make([][]string, gs.BoardSize)
	for i := range gs.Board {
		gs.Board[i] = make([]string, gs.BoardSize)
	}
	gs.MoveNumber = 0
	gs.Captures = map[string]int{"B": 0, "W": 0}

	path := make([]int, len(gs.Path))
	copy(path, gs.Path)
	node := gs.Current
	for node.Parent != nil {
		node = node.Parent
	}

	gs.applyNode(node, true)
	current := node
	for i := 1; i < len(path); i++ {
		if path[i] < len(current.Children) {
			current = current.Children[path[i]]
			gs.applyNode(current, false)
		}
	}
}

func (gs *GameState) JumpTo(targetPath []int) bool {
	if len(targetPath) == 0 {
		return false
	}
	path := make([]int, len(targetPath))
	copy(path, targetPath)
	node := gs.Current
	for node.Parent != nil {
		node = node.Parent
	}
	for i := 1; i < len(path); i++ {
		if path[i] >= len(node.Children) {
			return false
		}
		node = node.Children[path[i]]
	}
	gs.Current = node
	gs.Path = path
	gs.rebuild()
	return true
}

func CollectAllPaths(node *Node, currentPath []int, result *[][]int) {
	path := make([]int, len(currentPath))
	copy(path, currentPath)
	*result = append(*result, path)
	for i, child := range node.Children {
		childPath := append([]int{}, path...)
		childPath = append(childPath, i)
		CollectAllPaths(child, childPath, result)
	}
}

func GetPathFromRoot(node *Node) []int {
	var path []int
	current := node
	for current.Parent != nil {
		parent := current.Parent
		for i, ch := range parent.Children {
			if ch == current {
				path = append([]int{i}, path...)
				break
			}
		}
		current = parent
	}
	path = append([]int{0}, path...)
	return path
}
