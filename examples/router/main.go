package main

import (
	"fmt"

	"strings"

	"github.com/bep/gr"
	"github.com/bep/gr/attr"
	"github.com/bep/gr/el"
	"github.com/bep/gr/evt"
	"github.com/bep/gr/examples"
	"github.com/bep/grouter"
)

var (
	component1   = gr.New(&clickCounter{title: "Counter 1", color: "#ff0066"})
	component2   = gr.New(&clickCounter{title: "Counter 2", color: "#339966"})
	component3   = gr.New(&clickCounter{title: "Counter 3", color: "#0099cc"})
	component3_2 = gr.New(&clickCounter{title: "Counter 3_2", color: "#ffcc66"})

	// WithRouter makes this.props.router happen.
	appComponent = gr.New(new(app), gr.Apply(grouter.WithRouter))

	router = grouter.New("/", appComponent,
		grouter.NewRoute("c1", grouter.Components{"main": component1}),
		grouter.NewRoute("c2", grouter.Components{"main": component2}),
		grouter.NewRoute("c3", grouter.Components{"main": component3, "sub": component3_2}),
	)
)

func main() {
	mainComponent := gr.New(new(testRouter))
	mainComponent.Render("react", gr.Props{})
}

type testRouter int

// Implements the Renderer interface.
func (r testRouter) Render(this *gr.This) gr.Component {
	return router
}

// Implements the ShouldComponentUpdate interface.
func (r testRouter) ShouldComponentUpdate(this *gr.This, nextProps gr.Props, nextState gr.State) bool {
	return true
}

type app int

// Implements the Renderer interface.
func (a app) Render(this *gr.This) gr.Component {
	// Receives the component as this.props.route.component
	return el.Div(
		el.Header1(gr.Text("Router")),
		el.UnorderedList(
			gr.CSS("nav", "nav-tabs"),
			createLinkListItem(this, "/c1", "Tab #1"),
			createLinkListItem(this, "/c2", "Tab #2"),
			createLinkListItem(this, "/c3", "Tab #3"),
		),
		this.Component("main"),
		this.Component("sub"),
	)
}

func createLinkListItem(this *gr.This, path, title string) gr.Modifier {
	return el.ListItem(
		grouter.MarkIfActive(this.Props(), path),
		attr.Role("presentation"),
		grouter.Link(path, title))
}

type clickCounter struct {
	title, color string
}

// Implements the StateInitializer interface.
func (c clickCounter) GetInitialState(this *gr.This) gr.State {
	println(c.title, "GetInitialState")
	return gr.State{
		"counter": 0,
	}
}

// Implements the Renderer interface.
func (c clickCounter) Render(this *gr.This) gr.Component {
	counter := this.State()["counter"]
	message := fmt.Sprintf("%s: %v", c.title, counter)

	elem := el.Div(
		el.Button(
			gr.CSS("btn", "btn-lg", "btn-warning"),
			gr.Style("color", c.color),
			gr.Style("fontWeight", "bold"),
			gr.Text(message),
			evt.Click(c.onClick)))

	return examples.Example(strings.Title(message), elem)
}

func (c clickCounter) onClick(this *gr.This, event *gr.Event) {
	this.SetState(gr.State{"counter": this.StateInt("counter") + 1})
}

// Implements the ShouldComponentUpdate interface.
func (e clickCounter) ShouldComponentUpdate(
	this *gr.This, nextProps gr.Props, nextState gr.State) bool {

	return this.State().HasChanged(nextState, "counter")
}