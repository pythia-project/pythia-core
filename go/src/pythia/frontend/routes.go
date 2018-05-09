package frontend

import (
	"net/http"
	"github.com/gorilla/mux"
)

//Route struct to easily add new roots
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

//Routes is a list of Route
type Routes []Route

//NewRouter changed mux.Router func to work with the Rout struct
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	for _, route := range routes {
		router.Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	return router
}

/*
NEXT STEP

//MiddleWare check the IP of client with the list of IPs in conf.jdon
func MiddleWare(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		addr := GetClientIPs(r)
		IPConf := GetConf().IP

		for _, ipConf := range IPConf {
			for _, ipClient := range addr {
				if ipConf == ipClient {
					h.ServeHTTP(w, r)
					return
				}
			}
		}
		http.Error(w, "Unauthorized IP address", 401)
		return
	})
}

//GetClientIPs returns the IPs address of client
func GetClientIPs(r *http.Request) []string {
	//If X-FORWARDED-FOR structure is respected (first IP is the client's private IP address)
	//and separate with ", "
	//Header.Get will have all other IPs but not the last one used (last proxy or client if empty)
	var IPs []string
	if allIP := r.Header.Get("W-FORWARDED-FOR"); len(allIP) > 0 {
		IPs = strings.Split(allIP, ", ")
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	IPs = append(IPs, ip)
	return IPs
}
*/

var routes = Routes{
	Route{
		"Echo",
		"POST",
		"/api/echo",
		Echo,
	},
	Route{
		"Task",
		"POST",
		"/execute",
		Task,
	},
}
