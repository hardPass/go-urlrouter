// Efficient URL routing using a Trie data structure.
//
// This Package implements a URL Router, but instead of using the usual
// "evaluate all the routes and return the first regexp that matches" strategy,
// it uses a Trie data structure to perform the routing. This is more efficient,
// and scales better for a large number of routes.
// It supports the :param and *splat placeholders in the route strings.
//
// Example:
//	router := urlrouter.Router{
//		Routes: []urlrouter.Route{
//			urlrouter.Route{
//				PathExp: "/resources/:id",
//				Dest:    "one_resource",
//			},
//			urlrouter.Route{
//				PathExp: "/resources",
//				Dest:    "all_resources",
//			},
//		},
//	}
//
//	err := router.Start()
//	if err != nil {
//		panic(err)
//	}
//
//	input := "http://example.org/resources/123"
//	route, params, err := router.FindRoute(input)
//	if err != nil {
//		panic(err)
//	}
//	fmt.Print(route.Dest)  // one_resource
//	fmt.Print(params["id"])  // 123
//
package urlrouter

import (
	"errors"
	"github.com/ant0ine/go-urlrouter/trie"
	"net/url"
)

// TODO
// support for http method routing ?
// support for #param placeholder ?

type Route struct {
	// A string defining the route, like "/resource/:id.json".
	// Placeholders supported are:
	// :param that matches any char to the first '/' or '.'
	// *splat that matches everything to the end of the string
	PathExp string
	// Can be anything useful to point to the code to run for this route.
	Dest interface{}
}

type Router struct {
	// list of Routes, the order matters, if multiple Routes match, the first defined will be used.
	Routes                   []Route
	disable_trie_compression bool
	index                    map[*Route]int
	trie                     *trie.Trie
}

// This validates the Routes and prepares the Trie data structure.
// It must be called once the Routes are defined and before trying to find Routes.
func (self *Router) Start() error {

	self.trie = trie.New()
	self.index = map[*Route]int{}
	unique := map[string]bool{}

	for i, _ := range self.Routes {
		// pointer to the Route
		route := &self.Routes[i]
		// unique
		if unique[route.PathExp] == true {
			return errors.New("duplicated PathExp")
		}
		unique[route.PathExp] = true
		// index
		self.index[route] = i
		// insert in the Trie
		err := self.trie.AddRoute(route.PathExp, route)
		if err != nil {
			return err
		}
	}

	if self.disable_trie_compression == false {
		self.trie.Compress()
	}

	// TODO validation of the PathExp ? start with a /
	// TODO url encoding

	return nil
}

// Return the first matching Route and the corresponding parameters for a given URL object.
func (self *Router) FindRouteFromURL(url_obj *url.URL) (*Route, map[string]string) {

	// lookup the routes in the Trie
	// TODO verify url encoding
	matches := self.trie.FindRoutes(url_obj.Path)

	// only return the first Route that matches
	min_index := -1
	matches_by_index := map[int]*trie.Match{}

	for _, match := range matches {
		route := match.Route.(*Route)
		route_index := self.index[route]
		matches_by_index[route_index] = match
		if min_index == -1 || route_index < min_index {
			min_index = route_index
		}
	}

	if min_index == -1 {
		// no route found
		return nil, nil
	}

	// and the corresponding params
	match := matches_by_index[min_index]

	return match.Route.(*Route), match.Params
}

// Parse the url string (complete or just the path) and return the first matching Route and the corresponding parameters.
func (self *Router) FindRoute(url_str string) (*Route, map[string]string, error) {

	// parse the url
	url_obj, err := url.Parse(url_str)
	if err != nil {
		return nil, nil, err
	}

	route, params := self.FindRouteFromURL(url_obj)
	return route, params, nil
}
