package api

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Org struct {
	OUs []*OU `yaml:"organizations,flow"`
}

type OU struct {
	Name     string
	Members  []string `yaml:",flow"`
	Children []*OU
}

func FormUsers(ou, parentOU *OU) map[string]*User {
	users := make(map[string]*User)

	// Construct Members within Child OU
	if parentOU != nil {
		ou.Name = fmt.Sprintf("%v - %v", parentOU.Name, ou.Name)
	}
	for _, member := range ou.Members {
		if _, ok := users[member]; ok {
			users[member].Groups = append(users[member].Groups, ou.Name)
		} else {
			users[member] = &User{UserName: member, Groups: []string{ou.Name}}
		}
	}

	// Construct Members within Child OU
	for _, child_ou := range ou.Children {
		childUsers := FormUsers(child_ou, ou)
		for user, child := range childUsers {
			if v, ok := users[user]; ok {
				v.Groups = append(v.Groups, child.Groups...)
			} else {
				users[user] = child
			}
		}
	}
	return users
}

func ReadOrg(fileName string) Org {
	f, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	var org Org
	err = yaml.Unmarshal(f, &org)
	if err != nil {
		panic(err)
	}

	return org
}
