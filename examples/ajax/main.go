package main

import (
	"encoding/json"
	"net/http"

	"github.com/bep/gr"
	"github.com/bep/gr/attr"
	"github.com/bep/gr/el"
	"github.com/bep/gr/examples"
)

// Ajax loading example similar to:
// https://facebook.github.io/react/tips/initial-ajax.html
//
// NOTE: These initial examples are technical in nature, and do not attempt
// to establish a best practice of how to create React apps.

func main() {
	component := gr.New(new(userGists))
	component.Render("react", gr.Props{})

	// block
	w := make(chan struct{})
	<-w
}

type gist struct {
	URL         string `json:"url"`
	ID          string `json:"id"`
	HtmlUrl     string `json:"html_url"`
	CreatedAt   string `json:"created_at"`
	Description string `json:"description"`
}

type userGists int

// Implements the Renderer interface.
func (g userGists) Render(this *gr.This) gr.Component {

	elem := el.Div()

	if s, ok := this.State()["gists"]; ok {
		// The nice Gist type is lost once we entered the JavaScript world.
		//
		// What we get now is:
		//
		// []interface{} with the individual Gists as map[string]interface{}
		//
		// Let that serve as a note to self that this may not be the optimal way.
		// I imagine most of the UI will happen in JavaScript and the business logic
		// and the orchestration, including injecting data reuired by the components.
		gists := s.([]interface{})

		table := el.Table(
			gr.CSS("table", "table-striped"),
			gr.Style("width", "50%"),
			el.TableHead(el.TableRow(
				el.TableHeader(gr.Text("Description")),
				el.TableHeader(gr.Text("URL")),
			)))

		body := el.TableBody()

		for _, g := range gists {
			tr := tableRow(g)
			tr.Modify(body)
		}

		// TODO(bep) "body modifies table" doesn't sound right/good. Rename ...
		body.Modify(table)
		table.Modify(elem)
	}

	return examples.Example("Ajax (some random Gists)", elem)

}

func tableRow(i interface{}) *gr.Element {
	gist := i.(map[string]interface{})

	return el.TableRow(
		el.TableData(gr.Text(gist["Description"])),
		el.TableData(
			el.Anchor(attr.HRef(gist["HtmlUrl"]),
				attr.Target("_blank"), gr.Text("View"))),
	)
}

// Implements the ComponentDidMount interface
func (g userGists) ComponentDidMount(this *gr.This) {
	println("ComponentDidMount")

	var gists []gist

	// Note that ComponentDidMount is assumed to block, so no need to spin up a
	// goroutine for this.
	resp, err := http.Get("https://api.github.com/users/bradfitz/gists")

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&gists)

	if err != nil {
		panic(err)
	}

	this.SetState(gr.State{"gists": gists})
}

// Implements the ComponentWillUnmount interface
func (g userGists) ComponentWillUnmount(this *gr.This) {
	println("ComponentWillUnmount")
	// TODO(bep): HTTP Cancelation
}

// Implements the ShouldComponentUpdate interface.
func (g userGists) ShouldComponentUpdate(this *gr.This, nextProps gr.Props, nextState gr.State) bool {
	return this.State().HasChanged(nextState, "gists")
}
