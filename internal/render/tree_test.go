package render

import "testing"

func TestBuildTreeGroupsDirectoriesBeforeFiles(t *testing.T) {
	docs := []Document{
		{ID: "doc-0", Title: "A", RelPath: "a.md"},
		{ID: "doc-1", Title: "Spec", RelPath: "a/spec.md"},
		{ID: "doc-2", Title: "Tasks", RelPath: "b/tasks.md"},
		{ID: "doc-3", Title: "Ten", RelPath: "b/10.md"},
	}

	tree := buildTree(docs, "doc-1")
	if len(tree) != 3 {
		t.Fatalf("root nodes = %d, want 3", len(tree))
	}
	if !tree[0].IsDir || tree[0].Name != "a" {
		t.Fatalf("first node = %#v", tree[0])
	}
	if !tree[1].IsDir || tree[1].Name != "b" {
		t.Fatalf("second node = %#v", tree[1])
	}
	if tree[2].IsDir || tree[2].Name != "a.md" {
		t.Fatalf("last node = %#v", tree[2])
	}
	if len(tree[1].Children) != 2 || tree[1].Children[0].Name != "10.md" || tree[1].Children[1].Name != "tasks.md" {
		t.Fatalf("b children = %#v", tree[1].Children)
	}
}

func TestBuildTreeMarksActivePathExpanded(t *testing.T) {
	docs := []Document{
		{ID: "doc-0", Title: "Root", RelPath: "README.md"},
		{ID: "doc-1", Title: "Leaf", RelPath: "guides/deep/leaf.md"},
	}
	tree := buildTree(docs, "doc-1")
	var guides *TreeNode
	for _, node := range tree {
		if node.Name == "guides" {
			guides = node
		}
	}
	if guides == nil || !guides.Expanded {
		t.Fatalf("guides node not expanded: %#v", guides)
	}
	deep := guides.Children[0]
	if !deep.Expanded {
		t.Fatalf("deep node not expanded: %#v", deep)
	}
	leaf := deep.Children[0]
	if !leaf.Active {
		t.Fatalf("leaf not active: %#v", leaf)
	}
	for _, node := range tree {
		if node.Name == "README.md" && node.Active {
			t.Fatalf("root README should not be active: %#v", node)
		}
	}
}

func TestBuildTreeKeepsDuplicateBasenamesDistinct(t *testing.T) {
	docs := []Document{
		{ID: "doc-0", Title: "A", RelPath: "duplicates/a.md"},
		{ID: "doc-1", Title: "B", RelPath: "duplicates/b/a.md"},
	}
	tree := buildTree(docs, "doc-1")
	duplicates := tree[0]
	if len(duplicates.Children) != 2 {
		t.Fatalf("children = %#v", duplicates.Children)
	}
	if duplicates.Children[0].Name != "b" || duplicates.Children[0].Children[0].DocID != "doc-1" {
		t.Fatalf("first child = %#v", duplicates.Children[0])
	}
	if duplicates.Children[1].Name != "a.md" || duplicates.Children[1].DocID != "doc-0" {
		t.Fatalf("second child = %#v", duplicates.Children[1])
	}
}
