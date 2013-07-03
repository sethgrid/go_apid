package main

import "apid"
import "fmt"

func main() {
	// good functions to test
	// removeEventSubscription eventid, appid
	// getUserInfo userid/username
	//
	// TODO: XXX change the passed back value to something where I can foo['value']. simplejson? 

	foo, err := apid.Call("getUserInfo", map[string]interface{}{"userid": 180})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Loaded")
	fmt.Println(foo)

}
