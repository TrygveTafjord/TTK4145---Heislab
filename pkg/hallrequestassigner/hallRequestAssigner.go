package hallrequestassigner

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

func HallRequestAssigner(jsonBytes []byte) (output map[string][][2]bool){
	
	ret, err := exec.Command("./hall_request_assigner", "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return
	}
	
	output = make(map[string][][2]bool)

	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return
	}

	return output
}

//for testing purposes (paste in main): 