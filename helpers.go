package gr

import (
	"fmt"

	"strconv"

	"github.com/gopherjs/gopherjs/js"
)

//TODO make this work so the examples can be made without an index.html
/*func AddReactJS(version string) {
	for _, lib := range []string{"react-dom.js", "react.js"} {
		link := js.Global.Get("document").Call("createElement", "script")
		link.Set("src", fmt.Sprintf("//cdnjs.cloudflare.com/ajax/libs/react/%s/%s", version, lib))
		//js.Global.Get("document").Get("head").Call("appendChild", link)
		head := js.Global.Get("document").Get("head")
		head.Call("insertBefore", link, head.Get("firstChild"))
	}
}*/

func UnmountComponentAtNode(elementID string) bool {
	// TODO(bep) maybe incorporate this DOM element into the component
	container := js.Global.Get("document").Call("getElementById", elementID)
	return reactDOM.Call("unmountComponentAtNode", container).Bool()
}

type HostInfo struct {
	Path     string
	Port     int
	Host     string
	Href     string
	Protocol string
	Origin   string
}

// Location returns info about the current browser location.
func Location() HostInfo {
	l := js.Global.Get("window").Get("location").Interface().(map[string]interface{})
	loc := HostInfo{
		Path:     toString(l["pathname"]),
		Port:     toInt(l["port"]),
		Host:     toString(l["hostname"]),
		Href:     toString(l["href"]),
		Protocol: toString(l["protocol"]),
		Origin:   toString(l["origin"])}

	return loc
}

func toString(i interface{}) string {
	switch v := i.(type) {
	case string:
		return i.(string)
	case *js.Object:
		if v == js.Undefined {
			return ""
		}
		panic("Invalid string type")
	}

	return ""
}

func toInt(i interface{}) int {
	switch v := i.(type) {
	case int:
		return v
	case float32:
		return int(v)
	case float64:
		return int(v)
	case string:
		if v == "" {
			return 0
		}
		iv, err := strconv.ParseInt(v, 0, 0)
		if err == nil {
			return int(iv)
		}
		panic(err)
	default:
		panic(fmt.Sprintf("Unhandled number type: %T", v))
	}
}
