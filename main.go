package main

import (
	"fmt"
	"lastpass_provisioning/api"
	lastpassTime "lastpass_provisioning/lastpass_time"
	"os"
	"sync"
	"time"
)

// TODO Inactiveな人
func main() {
	// Client作成
	c, err := NewClient(nil)
	if err != nil {
		fmt.Errorf("Failed Building Client: %v", err.Error())
		os.Exit(1)
	}

	fmt.Println(" --------------------  Disabled Users -------------------- ")
	s := NewService(c)
	disabledUsers, err := s.GetDisabledUser()

	for _, user := range disabledUsers {
		fmt.Println(user.UserName)
	}

	fmt.Println(" --------------------  Inactive Users -------------------- ")
	inactiveUsers, err := s.GetInactiveUser()
	fmt.Println(len(inactiveUsers))

	fmt.Println(" --------------------  Admin Users -------------------- ")
	AdminUsers, err := s.GetAdminUserData()

	for _, admin := range AdminUsers {
		fmt.Println(admin.UserName)
	}

	// Get Shared Folder Data
	fmt.Println(" --------------------  Super Shared Folders -------------------- ")
	res, err := c.GetSharedFolderData()
	if err != nil {
		fmt.Println(err)
		return
	}
	var sharedFolders map[string]api.SharedFolder
	err = DecodeBody(res, &sharedFolders)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, sf := range sharedFolders {
		if sf.ShareFolderName == "Super-Admins" {
			for _, user := range sf.Users {
				fmt.Println(sf.ShareFolderName + " : " + user.UserName)
			}
		}
	}

	loc, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(loc)
	dayAgo := now.Add(-time.Duration(1) * time.Hour * 24)
	t := lastpassTime.JsonLastPassTime{now}
	f := lastpassTime.JsonLastPassTime{dayAgo}

	header := fmt.Sprintf(" -------------------- Events(%v ~ %v) --------------------", f.Format(), t.Format())
	fmt.Println(header)

	res, err = c.GetEventReport("", "", f, t)
	if err != nil {
		fmt.Println(err)
		return
	}

	var result api.Events
	err = DecodeBody(res, &result)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, event := range result.Events {
		if event.IsAuditEvent() {
			fmt.Println(event)
		}
	}

	GetLoginEvent := func(wg *sync.WaitGroup, q chan string) {
		defer wg.Done()
		for {
			userName, ok := <-q
			if !ok {
				return
			}
			res, err = c.GetEventReport(userName, "", f, t)
			fmt.Println(fmt.Sprintf(" --------------------------------------  %v Login History ------------------------------- ", userName))
			if err != nil {
				fmt.Println(err)
				return
			}
			err = DecodeBody(res, &result)
			if err != nil {
				fmt.Println(err)
				return
			}
			for _, event := range result.Events {
				fmt.Println(event)
			}
		}
	}

	var wg sync.WaitGroup
	q := make(chan string, 5)
	for i := 0; i < len(AdminUsers); i++ {
		wg.Add(1)
		go GetLoginEvent(&wg, q)
	}

	for _, admin := range AdminUsers {
		q <- admin.UserName
	}
	close(q)
	wg.Wait()

	// Move

	//var wgwg sync.WaitGroup
	//for _, h := range hoge	{
	//	wgwg.Add(1)
	//	go func(name string) {
	//		defer wgwg.Done()
	//
	//		_, err = c.ChangeGroupsMembership([]api.BelongingGroup{
	//			{
	//				name,
	//				[]string{"MFクラウドサービス開発本部"},
	//				[]string{},
	//			},
	//
	//		})
	//	}(h)
	//}
	//wgwg.Wait()

	//_, err = c.ChangeGroupsMembership([]api.BelongingGroup{
	//	{
	//		"ikeuchi.kenichi@moneyforward.co.jp",
	//		[]string{"MFクラウド事業推進本部"},
	//		[]string{},
	//	},
	//
	//})
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
}
