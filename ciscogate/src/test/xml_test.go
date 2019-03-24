package test

import (
	"fmt"
	"github.com/ugol/ciscogate/cmd"
	"io/ioutil"
	"testing"
)

func TestExtractToken(t *testing.T) {

	xmlBytes, err := ioutil.ReadFile("example.xml")
	if err != nil {
		panic(err)
	}

	token, err := cmd.ExtractToken(xmlBytes)
	if err != nil {
		panic(err)
	}
	expectedToken := "VHEFAAAAAAAAAAAAAAAAAMIejAwOdiZ6uxY4fuJCtYTqM+ZSjM8oPxorziue2jof75ECMSxd2n4CejTb/Az/FrWb+CMgFH51ee5G5aDaxEV3ox6aS7xcZDNrJ/iZCW9kLhvvy/YdAj/xY4q659HVgAObe15MnLgJiK5YYVPrIhT3Zhi4SbCpH3e4eFUm/+hMbKThXSAvD5DpWHTRB2QoTA=="
	if expectedToken != token {
		fmt.Printf("Received token was:\n%v", token)
		t.Errorf("Expected token was:\n%v", expectedToken)
	}
	//t.Errorf("Six should be a valid vote, but the vote returned: %v", valid)
	//t.Errorf("Seven shouldn't be a valid vote, but the vote returned: %v", invalid)

}
