package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"io"
	"log"
	"os"
	"strings"
)

func main() {

	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	jsonValue, _, err := parse(string(input), 0)
	if err != nil {
		log.Fatal(err)
	}

	app := tview.NewApplication()

	inputField := tview.NewInputField()

	tree := tview.NewTreeView()

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(tree, 0, 1, true)
	flex.AddItem(inputField, 1, 1, false)

	app.SetRoot(flex, true).SetFocus(tree)

	root, err := CreateTreeNode(jsonValue)
	if err != nil {
		log.Fatal(err)
	}
	tree.SetRoot(root)
	AddChildren(root)

	tree.SetCurrentNode(root)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		children := node.GetChildren()
		if len(children) == 0 {
			AddChildren(node)
		} else {
			node.SetExpanded(!node.IsExpanded())
		}
	})

	tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft:
		case tcell.KeyRight:
			node := tree.GetCurrentNode()
			if node != nil {
				children := node.GetChildren()
				if len(children) == 0 {
					AddChildren(node)
				}
			}
		case tcell.KeyTab:
			app.SetFocus(inputField)
		}
		return event
	})

	inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(tree)
		}
		return event
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func CreateTreeNode(jsonValue JsonValue) (*tview.TreeNode, error) {
	var child *tview.TreeNode

	switch jsonValue.ValueType {
	case Array:
		child = tview.NewTreeNode("[...]")
	case Object:
		child = tview.NewTreeNode("{...}")
	default:
		child = tview.NewTreeNode(jsonValue.RawValue)
	}

	child.SetReference(jsonValue)
	return child, nil
}

func AddChildren(parentNode *tview.TreeNode) error {

	parentObj := parentNode.GetReference().(JsonValue)

	switch parentObj.ValueType {
	case Array:
		for _, a := range parentObj.ArrayMember {
			child, err := CreateTreeNode(a)
			if err != nil {
				return err
			}
			parentNode.AddChild(child)
		}
	case Object:
		maxLen := 0
		for _, pair := range parentObj.ObjectMember {
			if maxLen < len(pair.Key.RawValue) {
				maxLen = len(pair.Key.RawValue)
			}
		}

		for _, pair := range parentObj.ObjectMember {
			var child *tview.TreeNode

			switch pair.Value.ValueType {
			case Array:
				var s string
				if len(pair.Value.ArrayMember) == 0 {
					s = "[ ]"
				} else {
					s = "[...]"
				}
				child = tview.NewTreeNode(
					fmt.Sprintf(
						"%s%s : %s",
						pair.Key.RawValue,
						strings.Repeat(" ", maxLen-len(pair.Key.RawValue)),
						s,
					),
				)
			case Object:
				var s string
				if len(pair.Value.ObjectMember) == 0 {
					s = "{ }"
				} else {
					s = "{...}"
				}
				child = tview.NewTreeNode(
					fmt.Sprintf(
						"%s%s : %s",
						pair.Key.RawValue,
						strings.Repeat(" ", maxLen-len(pair.Key.RawValue)),
						s,
					),
				)
			default:
				child = tview.NewTreeNode(
					fmt.Sprintf(
						"%s%s : %s",
						pair.Key.RawValue,
						strings.Repeat(" ", maxLen-len(pair.Key.RawValue)),
						pair.Value.RawValue,
					),
				)
			}

			child.SetReference(pair.Value)
			parentNode.AddChild(child)
		}
	}
	return nil
}
