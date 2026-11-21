package render

import (
	"sort"
	"strings"

	"openmd/internal/scan"
)

type TreeNode struct {
	Name     string
	RelPath  string
	IsDir    bool
	DocID    string
	Title    string
	Expanded bool
	Active   bool
	Children []*TreeNode
}

func buildTree(docs []Document, activeID string) []*TreeNode {
	root := &TreeNode{IsDir: true}
	for _, doc := range docs {
		rel := doc.RelPath
		if rel == "" {
			continue
		}
		parts := strings.Split(rel, "/")
		current := root
		for i, part := range parts {
			if i == len(parts)-1 {
				current.Children = append(current.Children, &TreeNode{
					Name:    part,
					RelPath: rel,
					DocID:   doc.ID,
					Title:   doc.Title,
				})
				continue
			}
			var next *TreeNode
			for _, child := range current.Children {
				if child.IsDir && child.Name == part {
					next = child
					break
				}
			}
			if next == nil {
				next = &TreeNode{
					Name:    part,
					RelPath: strings.Join(parts[:i+1], "/"),
					IsDir:   true,
				}
				current.Children = append(current.Children, next)
			}
			current = next
		}
	}
	sortTree(root)
	markActivePath(root, activeID)
	return root.Children
}

func sortTree(node *TreeNode) {
	sort.SliceStable(node.Children, func(i, j int) bool {
		left, right := node.Children[i], node.Children[j]
		if left.IsDir != right.IsDir {
			return left.IsDir
		}
		return scan.NaturalLess(left.Name, right.Name)
	})
	for _, child := range node.Children {
		if child.IsDir {
			sortTree(child)
		}
	}
}

func markActivePath(node *TreeNode, activeID string) bool {
	active := false
	for _, child := range node.Children {
		if child.IsDir {
			if markActivePath(child, activeID) {
				child.Expanded = true
				active = true
			}
			continue
		}
		if child.DocID == activeID {
			child.Active = true
			active = true
		}
	}
	return active
}
