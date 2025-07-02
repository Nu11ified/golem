//go:build js && wasm

package components

import (
	"syscall/js"
	"time"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/state"
)

// Todo represents a single to-do item
type Todo struct {
	ID        int64  `json:"id"`
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}

const localStorageKey = "golem-todos"

// TodoListComponent creates a full-featured to-do list application
func TodoListComponent() *dom.Element {
	// --- STATE ---
	todos := state.NewObservable([]Todo{})
	input := state.NewObservable("")

	// --- PERSISTENCE ---
	persistence := state.NewPersistence()

	// Load initial state from localStorage
	loadInitialState := func() {
		var savedTodos []Todo
		err := persistence.LoadState(localStorageKey, &savedTodos)
		if err == nil {
			todos.Set(savedTodos)
		}
	}
	// Run on initial load
	go loadInitialState()

	// --- UI ELEMENTS ---
	todoListElement := dom.Ul()
	inputElement := dom.Input(
		dom.Placeholder("What needs to be done?"),
		dom.Autofocus(true),
		dom.OnInput(func(value string) {
			input.Set(value)
		}),
		dom.OnKeyDown(func(key string) {
			if key == "Enter" {
				addTodo(input, todos)
			}
		}),
	)

	// --- SUBSCRIPTIONS ---
	// Update UI and save to localStorage whenever todos change
	todos.Subscribe(func(newTodos, oldTodos []Todo) {
		renderTodos(todoListElement, newTodos, todos)
		go persistence.SaveState(localStorageKey, newTodos)
	})

	// --- RENDER ---
	return dom.Div(
		dom.Class("todo-app"),
		dom.H1("golem-todos"),
		dom.Div(
			dom.Class("input-container"),
			inputElement,
			dom.Button(
				"Add",
				dom.OnClick(func() {
					addTodo(input, todos)
				}),
			),
		),
		todoListElement,
	)
}

// addTodo adds a new item to the todo list
func addTodo(input *state.Observable[string], todos *state.Observable[[]Todo]) {
	text := input.Get()
	if text == "" {
		return
	}

	newTodo := Todo{
		ID:        time.Now().UnixNano(),
		Text:      text,
		Completed: false,
	}

	todos.Update(func(currentTodos []Todo) []Todo {
		return append(currentTodos, newTodo)
	})
	input.Set("") // Clear input.

	// We need to manually clear the input element's value in the DOM
	// because our state binding is one-way for inputs.
	js.Global().Get("document").Call("querySelector", `input[placeholder="What needs to be done?"]`).Set("value", "")
}

// renderTodos updates the DOM to display the list of todos
func renderTodos(ul *dom.Element, newTodos []Todo, todos *state.Observable[[]Todo]) {
	var children []*dom.Element
	for i := range newTodos {
		// Capture the todo for the closure
		todo := newTodos[i]

		children = append(children, dom.Li(
			dom.Class(getTodoClass(todo)),
			dom.Checkbox(
				dom.Checked(todo.Completed),
				dom.OnChange(func(checked bool) {
					toggleTodoCompletion(todo.ID, todos)
				}),
			),
			dom.Span(todo.Text),
			dom.Button(
				"❌",
				dom.OnClick(func() {
					removeTodo(todo.ID, todos)
				}),
			),
		))
	}
	ul.Children = children
	ul.Render()
}

// getTodoClass returns the CSS class for a todo item
func getTodoClass(todo Todo) string {
	if todo.Completed {
		return "completed"
	}
	return ""
}

// toggleTodoCompletion updates the completed status of a todo
func toggleTodoCompletion(id int64, todos *state.Observable[[]Todo]) {
	todos.Update(func(currentTodos []Todo) []Todo {
		for i, t := range currentTodos {
			if t.ID == id {
				currentTodos[i].Completed = !currentTodos[i].Completed
				break
			}
		}
		return currentTodos
	})
}

// removeTodo removes a todo item from the list
func removeTodo(id int64, todos *state.Observable[[]Todo]) {
	todos.Update(func(currentTodos []Todo) []Todo {
		var updatedTodos []Todo
		for _, t := range currentTodos {
			if t.ID != id {
				updatedTodos = append(updatedTodos, t)
			}
		}
		return updatedTodos
	})
}
