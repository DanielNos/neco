package main

func Visualize(tree *Node) {
	moduleNode := tree.value.(*ModuleNode)

	println(moduleNode.name)
}
