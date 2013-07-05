package apid

// 3rd party
import "conf"
import "simplejson"

// standard
import "fmt"
import "flag"
import "os"
import "net/http"
import "io/ioutil"

const DEFAULT_CONF = "/etc/default/sendgrid.conf"
const DEFAULT_SERVER = "127.0.0.1"
const DEFAULT_PORT = "8082"

var apiServer string
var port *string

type ApidError struct {
	What string
}

type ApidFunction struct {
	functionName string
	returnKey    string
	params       map[string]interface{}
	cachable     int
	path         string
}

var functions *simplejson.Json

func (e *ApidError) Error() string {
	return fmt.Sprintf("%s", e.What)
}

// change return type to simplejson?
func Call(method string, params map[string]interface{}) (interface{}, error) {
	c := make(chan *simplejson.Json)
	go load(c)
	fmt.Println("called load")

	functions = <-c
	function, exists := functions.CheckGet(method)

	if exists == false {
		return nil, &ApidError{fmt.Sprintf("Method undefined: %s", method)}
	}

	var callInfo ApidFunction
	callInfo.functionName = function.Get("function").MustString()
	callInfo.returnKey = function.Get("return").MustString()
	callInfo.params = function.Get("params").MustMap()
	callInfo.cachable = function.Get("cachable").MustInt()
	callInfo.path = function.Get("path").MustString()

	// verify that the required params exist
	for k, _ := range callInfo.params {
		_, exists := params[k]

		if exists == false {
			return nil, &ApidError{fmt.Sprintf("Missing required parameter: %s", k)}
		}
	}

	// &{map[function:getUserInfo return:user params:map[] cachable:60 path:/api/user/get.json]}
	requestPath := formatRequest(callInfo, params)

	results := getSimpleJsonResponse(requestPath)

	return results, nil
}

func formatRequest(info ApidFunction, passedParams map[string]interface{}) string {
	params := "?"
	for k, v := range passedParams {
		params += k + "=" + fmt.Sprintf("%v", v) + "&"
	}
	params = params[:len(params)-1]

	return "http://" + apiServer + ":" + *port + info.path + params
}

func load(c chan *simplejson.Json) {
	// load the conf file
	conf_file := flag.String("conf", DEFAULT_CONF, "Conf file")
	port = flag.String("port", DEFAULT_PORT, "Default port")
	flag.Parse()

	file, err := conf.ReadConfigFile(*conf_file)
	if err != nil {
		fmt.Println("Error reading configuration file. Exiting.")
		os.Exit(1)
	}

	apiServer, _ = file.GetString("default", "api_server")
	fmt.Println(apiServer + ":" + *port)

	// grab apid functions
	functions_resource := "http://" + apiServer + ":" + *port + "/api/functions.json"
	myjson := getSimpleJsonResponse(functions_resource)

	ApidFunctions := myjson.Get("functions")

	c <- ApidFunctions
}

func getSimpleJsonResponse(path string) *simplejson.Json {
	response, err := http.Get(path)

	if err != nil {
		fmt.Println("Error requesting " + path)
		os.Exit(1)
	}
	defer response.Body.Close()

	// read in the response and convert json
	json_byte_string, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}

	myjson, err := simplejson.NewJson(json_byte_string)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	return myjson
}
