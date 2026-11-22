package tuingo

import "github.com/gdamore/tcell/v2"

// Gap is the columns between a card and its children: room for "─┬─▶".
const Gap = 4

// Node is a card in a tree: its id and the cards that hang off it.
type Node struct {
	ID       string
	Children []*Node
}

// DrawTree draws a tree of cards with its top-left at (x, y): the root, then each child one
// Gap to the right joined by an arrow, children stacked with a blank row between. make
// gives the card for an id. It returns the rows used.
func (p *CardPane) DrawTree(n *Node, x, y, w int, make func(id string) Card, arrow tcell.Style) int {
	return p.DrawTreeWithTitles(n, x, y, w, withCardLabel(make), arrow)
}

// DrawTreeWithTitles draws a tree whose cards receive separate titles.
func (p *CardPane) DrawTreeWithTitles(n *Node, x, y, w int, make func(id string) (Card, CardTitle), arrow tcell.Style) int {
	return p.drawTree(n, x, y, w, make, arrow, "")
}

func (p *CardPane) drawTree(n *Node, x, y, w int, make func(id string) (Card, CardTitle), arrow tcell.Style, parent string) int {
	card, title := make(n.ID)
	p.DrawCardWithTitle(n.ID, x, y, w, card, title, parent)
	if len(n.Children) == 0 {
		return CardHeight
	}
	junction, childX, used := x+w+1, x+w+Gap, 0
	var tops []int
	for _, kid := range n.Children {
		tops = append(tops, y+used)
		used += p.drawTree(kid, childX, y+used, w, make, arrow, n.ID) + 1
	}
	p.Put(x+w, tops[0], "─", arrow)
	for i, top := range tops {
		first, last := i == 0, i == len(tops)-1
		corner := "├"
		switch {
		case first && last:
			corner = "─"
		case first:
			corner = "┬"
		case last:
			corner = "└"
		}
		p.Put(junction, top, corner+"─▶", arrow)
	}
	for ly := tops[0] + 1; ly < tops[len(tops)-1]; ly++ {
		if p.At(junction, ly) == ' ' {
			p.Put(junction, ly, "│", arrow)
		}
	}
	if used-1 > CardHeight {
		return used - 1
	}
	return CardHeight
}

// DrawLane draws cards in a column from (x, y), a blank row between, and returns the rows
// used.
func (p *CardPane) DrawLane(ids []string, x, y, w int, make func(id string) Card) int {
	return p.DrawLaneWithTitles(ids, x, y, w, withCardLabel(make))
}

// DrawLaneWithTitles draws a column whose cards receive separate titles.
func (p *CardPane) DrawLaneWithTitles(ids []string, x, y, w int, make func(id string) (Card, CardTitle)) int {
	for i, id := range ids {
		card, title := make(id)
		p.DrawCardWithTitle(id, x, y+i*(CardHeight+1), w, card, title, "")
	}
	if len(ids) == 0 {
		return 0
	}
	return len(ids)*(CardHeight+1) - 1
}

func withCardLabel(make func(string) Card) func(string) (Card, CardTitle) {
	return func(id string) (Card, CardTitle) {
		card := make(id)
		return card, CardTitle{Label: card.Label}
	}
}
