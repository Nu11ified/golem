package main

import (
	"golem-test-app/src/components"
	"strings"

	"github.com/Nu11ified/golem/css"
	"github.com/Nu11ified/golem/dom"
)

func App() *dom.Element {
	return dom.Div(
		dom.Class("container"),
		components.TodoListComponent(),
		components.CounterComponent(),
	)
}

func main() {
	InjectGlobalStyles()
	dom.Render(App(), "#app")
	select {}
}

// InjectGlobalStyles defines and injects all the application's CSS.
func InjectGlobalStyles() {
	var sb strings.Builder

	// --- FONT IMPORT ---
	// In a real app, you'd link this in your HTML's <head>.
	// For this demo, we inject it dynamically.
	sb.WriteString(`
		@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;700&display=swap');
	`)

	// --- CSS VARIABLES (THEME) ---
	sb.WriteString(`
		:root {
			--background: #1a1a1a;
			--surface: #2a2a2a;
			--primary: #6c5ce7;
			--primary-hover: #5a4cdb;
			--text-primary: #f0f0f0;
			--text-secondary: #a0a0a0;
			--border: #444;
			--success: #2ecc71;
			--danger: #e74c3c;
			--font-family: 'Inter', sans-serif;
		}
	`)

	// --- BASE & RESET STYLES ---
	sb.WriteString(`
		body {
			font-family: var(--font-family);
			background-color: var(--background);
			color: var(--text-primary);
			margin: 0;
			padding: 2rem;
			display: flex;
			justify-content: center;
			align-items: flex-start;
			min-height: 100vh;
		}
		* {
			box-sizing: border-box;
		}
	`)

	// --- MAIN APP CONTAINER ---
	sb.WriteString(`
		.container {
			width: 100%;
			max-width: 600px;
			display: flex;
			flex-direction: column;
			gap: 2rem;
		}
	`)

	// --- TODO APP STYLES ---
	sb.WriteString(`
		.todo-app, .counter-app {
			background-color: var(--surface);
			border-radius: 8px;
			padding: 1.5rem 2rem;
			box-shadow: 0 4px 12px rgba(0,0,0,0.2);
			border: 1px solid var(--border);
		}
		.todo-app h1, .counter-app h2 {
			margin-top: 0;
			border-bottom: 1px solid var(--border);
			padding-bottom: 1rem;
			margin-bottom: 1.5rem;
		}
		.input-container {
			display: flex;
			gap: 0.5rem;
			margin-bottom: 1.5rem;
		}
		.input-container input {
			flex-grow: 1;
			background: var(--background);
			border: 1px solid var(--border);
			color: var(--text-primary);
			padding: 0.75rem;
			border-radius: 6px;
			font-size: 1rem;
			transition: border-color 0.2s;
		}
		.input-container input:focus {
			outline: none;
			border-color: var(--primary);
		}
		ul {
			list-style: none;
			padding: 0;
			margin: 0;
			display: flex;
			flex-direction: column;
			gap: 0.75rem;
		}
		li {
			display: flex;
			align-items: center;
			gap: 0.75rem;
			padding: 0.75rem;
			background-color: var(--background);
			border-radius: 6px;
			transition: background-color 0.2s;
		}
		li:hover {
			background-color: #2f2f2f;
		}
		li.completed span {
			text-decoration: line-through;
			color: var(--text-secondary);
		}
		li span {
			flex-grow: 1;
		}
		li button {
			background: transparent;
			border: none;
			color: var(--text-secondary);
			cursor: pointer;
			font-size: 1rem;
			opacity: 0.5;
			transition: opacity 0.2s, color 0.2s;
		}
		li button:hover {
			opacity: 1;
			color: var(--danger);
		}
	`)

	// --- COUNTER & GENERAL BUTTON STYLES ---
	sb.WriteString(`
		.counter-app .dom-p {
			font-size: 1.2rem;
			background-color: var(--background);
			padding: 1rem;
			border-radius: 6px;
			text-align: center;
			margin-bottom: 1.5rem;
		}
		button {
			background-color: var(--primary);
			color: white;
			border: none;
			padding: 0.75rem 1.5rem;
			border-radius: 6px;
			font-size: 1rem;
			font-weight: 500;
			cursor: pointer;
			transition: background-color 0.2s;
		}
		button:hover {
			background-color: var(--primary-hover);
		}
		.counter-app > button {
			margin-right: 0.5rem;
		}
	`)

	css.InjectStyles(sb.String())
}
