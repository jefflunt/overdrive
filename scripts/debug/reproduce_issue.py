
class TodoTree:
    def __init__(self):
        self.todos = []
        self.expanded = set()

    def render(self):
        print("Rendering...")
        for todo in self.todos:
            self.createTodoNode(todo)

    def createTodoNode(self, todo):
        is_expanded = todo['id'] in self.expanded
        print(f"Todo {todo['id']} (type: {type(todo['id'])}): is_expanded = {is_expanded}")
        if 'children' in todo:
            for child in todo['children']:
                self.createTodoNode(child)

    def toggleExpand(self, id_val):
        print(f"Toggling {id_val} (type: {type(id_val)})")
        if id_val in self.expanded:
            self.expanded.remove(id_val)
        else:
            self.expanded.add(id_val)
        self.render()

tree = TodoTree()
tree.todos = [{'id': '1', 'children': [{'id': '2'}]}]

print("Initial state:")
tree.render()

print("\nAction: toggleExpand('1')")
tree.toggleExpand('1')

print("\nWhat if todo.id was an integer?")
tree.todos = [{'id': 1, 'children': [{'id': 2}]}]
tree.expanded = set()
tree.render()
tree.toggleExpand('1')
