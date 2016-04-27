package gr

import (
	"errors"
	"fmt"

	"github.com/gopherjs/gopherjs/js"
)

var (
	react    = js.Global.Get("React")
	reactDOM = js.Global.Get("ReactDOM")
)

type Component interface {
	Node() *js.Object
}

// The root component.
type Root struct {
	// The React.createClass response.
	node *js.Object

	// The minimum interface needed to display something.
	r Renderer

	// Options
	exportName string
}

// TODO(bep) investigate modules.
func FromJS(path ...string) *Root {

	var component *js.Object

	for _, p := range path {
		if component != nil {
			component = component.Get(p)
		} else {
			component = js.Global.Get(p)
		}
	}

	if component == nil || component == js.Undefined {
		panic(fmt.Sprintf("JS component in path %v not found", path))
	}

	// TODO(bep): No concept of a Renderer implementation here. Do we need it?
	return &Root{node: component}
}

// Export is an option used to mark that the component should be exported to the
// JavaScript world as a global with the given name.
// TODO(bep) Again, explore the world of JavaScript modules.
func Export(name string) func(*Root) error {
	return func(r *Root) error {
		if name == "" {
			return errors.New("Must provide export name")
		}
		r.exportName = name
		return nil
	}
}

func New(r Renderer, options ...func(*Root) error) *Root {
	root := &Root{r: r}

	classProps := make(map[string]*js.Object)

	// Every component needs to render itself.
	classProps["render"] = makeRenderFunc(r.Render)

	// Optional lifecycle implementations below.
	if v, ok := r.(StateInitializer); ok {
		classProps["getInitialState"] = makeStateFunc(v.GetInitialState)
	}

	if v, ok := r.(ShouldComponentUpdate); ok {
		classProps["shouldComponentUpdate"] = makeComponentUpdateFunc(v.ShouldComponentUpdate)
	}

	if v, ok := r.(ComponentWillUpdate); ok {
		classProps["componentWillUpdate"] = makeComponentUpdateVoidFunc(v.ComponentWillUpdate)
	}

	if v, ok := r.(ComponentDidUpdate); ok {
		classProps["componentDidUpdate"] = makeComponentUpdateVoidFunc(v.ComponentDidUpdate)
	}

	if v, ok := r.(ComponentWillReceiveProps); ok {
		classProps["componentWillReceiveProps"] = makeComponentPropertyReceiverFunc(v.ComponentWillReceiveProps)
	}

	if v, ok := r.(ShouldComponentUpdate); ok {
		classProps["shouldComponentUpdate"] = makeComponentUpdateFunc(v.ShouldComponentUpdate)
	}

	if v, ok := r.(ComponentWillMount); ok {
		classProps["componentWillMount"] = makeVoidFunc(v.ComponentWillMount)
	}

	if v, ok := r.(ComponentDidMount); ok {
		classProps["componentDidMount"] = makeVoidFunc(v.ComponentDidMount)
	}

	if v, ok := r.(ComponentWillUnmount); ok {
		classProps["componentWillUnmount"] = makeVoidFunc(v.ComponentWillUnmount)
	}

	class := react.Call("createClass", classProps)
	root.node = react.Call("createFactory", class)

	for _, opt := range options {
		err := opt(root)
		if err != nil {
			panic(err)
		}
	}

	root.handleOptionsOnCreate()

	return root
}

func (r *Root) handleOptionsOnCreate() {
	if r.exportName != "" {
		js.Global.Set(r.exportName, r.node)
	}
}

// Implements the Component interface.
func (r *Root) Node() *js.Object {
	return r.node
}

// TODO(bep) ...
func (r *Root) CreateElement(props Props) *Element {
	elm := react.Call("createElement", r.node, props)
	return &Element{element: elm}
}

func (r *Root) Render(elementID string, props Props) {
	container := js.Global.Get("document").Call("getElementById", elementID)

	elm := react.Call("createElement", r.node, props)
	// TODO(bep) evaluate if the need the "this" returned on render.
	reactDOM.Call("render", elm, container)
}

type Props map[string]interface{}
type State map[string]interface{}

// HasChanged reports whether the value of the property with any of the given keys has changed.
func (p Props) HasChanged(nextProps Props, keys ...string) bool {
	return hasChanged(p, nextProps, keys...)
}

// HasChanged reports whether the value of the state with any of the given keys has changed.
func (s State) HasChanged(nextState State, keys ...string) bool {
	return hasChanged(s, nextState, keys...)
}

func hasChanged(m1, m2 map[string]interface{}, keys ...string) bool {
	for _, key := range keys {
		if m1[key] != m2[key] {
			return true
		}
	}
	return false
}

type This struct {
	this *js.Object
}

func (t *This) Props() Props {
	props := t.this.Get("props").Interface()
	if props == nil {
		return Props{}
	}
	return props.(map[string]interface{})
}

func (t *This) State() State {
	state := t.this.Get("state").Interface()
	if state == nil {
		return State{}
	}
	return state.(map[string]interface{})
}

func (t *This) StateInt(key string) int {
	if val, ok := t.State()[key]; ok {
		return int(val.(float64))
	}
	return 0
}

func (t *This) SetState(s State) {
	t.this.Call("setState", s)
}

// Lifecycle interfaces
//
// See http://facebook.github.io/react/docs/component-specs.html#lifecycle-methods

// TODO(bep) Consider some better names. Move to subpackage?

// Renderer is the core interface used to ... render a Component.
type Renderer interface {
	Render(this *This) Component
}

// StateInitializer sets up the initial state.
type StateInitializer interface {
	GetInitialState(this *This) State
}

// Invoked before rendering when new props or state are being received.
// This is not called for the initial render or when forceUpdate is used.
type ShouldComponentUpdate interface {
	ShouldComponentUpdate(this *This, nextProps Props, nextState State) bool
}

// Invoked immediately before rendering when new props or state are being received.
// This is not called for the initial render.
type ComponentWillUpdate interface {
	ComponentWillUpdate(this *This, nextProps Props, nextState State)
}

// Invoked when a component is receiving new props.
// This method is not called for the initial render.
type ComponentWillReceiveProps interface {
	ComponentWillReceiveProps(this *This, props Props)
}

// Invoked immediately after the component's updates are flushed to the DOM.
// This method is not called for the initial render.
type ComponentDidUpdate interface {
	ComponentDidUpdate(this *This, prevProps Props, prevState State)
}

// Invoked once, both on the client and server, immediately before the initial rendering occurs.
type ComponentWillMount interface {
	ComponentWillMount(this *This)
}

// Invoked immediately before a component is unmounted from the DOM.
type ComponentWillUnmount interface {
	ComponentWillUnmount(this *This)
}

// Invoked once, only on the client (not on the server),
// immediately after the initial rendering occurs.
type ComponentDidMount interface {
	ComponentDidMount(this *This)
}

func makeComponentUpdateFunc(f func(this *This, props Props, state State) bool) *js.Object {
	return js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		return f(extractComponentUpdateArgs(this, arguments))
	})
}

func makeComponentUpdateVoidFunc(f func(this *This, props Props, state State)) *js.Object {
	return js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		f(extractComponentUpdateArgs(this, arguments))
		return nil
	})
}

func makeComponentPropertyReceiverFunc(f func(this *This, props Props)) *js.Object {
	return js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		that, props, _ := extractComponentUpdateArgs(this, arguments)
		f(that, props)
		return nil
	})
}

func extractComponentUpdateArgs(this *js.Object, arguments []*js.Object) (*This, Props, State) {
	var (
		props Props
		state State
	)

	if arguments[0] != nil {
		props = arguments[0].Interface().(map[string]interface{})
	}
	if arguments[1] != nil {
		state = arguments[1].Interface().(map[string]interface{})
	}

	that := makeThis(this)

	return that, props, state
}

func makeVoidFunc(f func(this *This)) *js.Object {
	return js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		f(makeThis(this))
		return nil
	})
}

func makeStateFunc(f func(this *This) State) *js.Object {
	return js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		return f(makeThis(this))
	})
}

func makeRenderFunc(f func(this *This) Component) *js.Object {
	return js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		that := makeThis(this)
		comp := f(that)
		idFactory := &incrementor{}
		addMissingKeys(comp, idFactory)
		// TODO(bep) refactor
		if e, ok := comp.(*Element); ok {
			e.This = that
			addEventListeners(comp, that)
		}
		return comp.Node()
	})
}

func addEventListeners(c Component, that *This) {
	if e, ok := c.(*Element); ok {
		for _, l := range e.eventListeners {
			l.delegate = func(event *js.Object) {
				if l.preventDefault {
					event.Call("preventDefault")
				}
				l.listener(that, &Event{target: event.Get("target")})
			}

			e.properties[l.name] = l.delegate

		}
		for _, child := range e.children {
			addEventListeners(child, that)
		}
	}
}

func addMissingKeys(c Component, id *incrementor) {
	if e, ok := c.(*Element); ok {
		if e.properties == nil {
			e.properties = make(map[string]interface{})
		}
		if _, ok := e.properties["key"]; !ok {
			e.properties["key"] = id.next()
		}
		for _, c2 := range e.children {
			addMissingKeys(c2, id)
		}
	}
}

func makeThis(that *js.Object) *This {
	return &This{this: that}
}
