package main

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
)

// TODO Security Score
// Master pw reuses <-
// weak Master Passwords <- 取れそう?(MasterPasswordStrength, Mpstrength)
// weak security challenge <- 無理そう
func main() {
	app := cli.NewApp()
	app.Name = "lastpass"
	app.Version = fmt.Sprintf("%s (rev:%s)", version, gitcommit)
	app.Usage = "A CLI tool for Lastpass(Enterprise)"
	app.Author = "Money Forward Co., Ltd."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Load configuration from `FILE`",
		},
	}
	app.Commands = Commands
	app.Run(os.Args)

	// Client作成
	//c, err := NewClient(nil)
	//c, err := NewLastPassClientFromContext()
	//if err != nil {
	//	fmt.Errorf("Failed Building Client: %v", err.Error())
	//	os.Exit(1)
	//}
	//
	//fmt.Println(" --------------------  Disabled Users -------------------- ")
	//s := NewService(c)
	//disabledUsers, err := s.GetDisabledUser()
	//
	//for _, user := range disabledUsers {
	//	fmt.Println(user.UserName)
	//}
	//
	//fmt.Println(" --------------------  Inactive Users -------------------- ")
	//inactiveUsers, err := s.GetInactiveUser()
	//inactiveDep := make(map[string][]api.User)
	//for _, u := range inactiveUsers {
	//	for _, group := range u.Groups {
	//		inactiveDep[group] = append(inactiveDep[group], u)
	//	}
	//}
	//for dep, users := range inactiveDep {
	//	fmt.Println(fmt.Sprintf("%v : %v", dep, len(users)))
	//}
	//
	//fmt.Println(" --------------------  Admin Users -------------------- ")
	//AdminUsers, err := s.GetAdminUserData()
	//
	//for _, admin := range AdminUsers {
	//	fmt.Println(admin.UserName)
	//}
	//
	//// Get Shared Folder Data
	//fmt.Println(" --------------------  Super Shared Folders -------------------- ")
	//res, err := c.GetSharedFolderData()
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//var sharedFolders map[string]api.SharedFolder
	//err = DecodeBody(res, &sharedFolders)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//for _, sf := range sharedFolders {
	//	if sf.ShareFolderName == "Super-Admins" {
	//		for _, user := range sf.Users {
	//			fmt.Println(sf.ShareFolderName + " : " + user.UserName)
	//		}
	//	}
	//}
	//
	//loc, _ := time.LoadLocation("Asia/Tokyo")
	//now := time.Now().In(loc)
	//dayAgo := now.Add(-time.Duration(2) * time.Hour * 24)
	//t := lastpassTime.JsonLastPassTime{now}
	//f := lastpassTime.JsonLastPassTime{dayAgo}
	//
	//header := fmt.Sprintf(" -------------------- Events(%v ~ %v) --------------------", f.Format(), t.Format())
	//fmt.Println(header)
	//
	//res, err = c.GetEventReport("", "", f, t)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//var result api.Events
	//err = DecodeBody(res, &result)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//for _, event := range result.Events {
	//	if event.IsAuditEvent() {
	//		fmt.Println(event)
	//	}
	//}
	//
	//GetLoginEvent := func(wg *sync.WaitGroup, q chan string) {
	//	defer wg.Done()
	//	for {
	//		userName, ok := <-q
	//		if !ok {
	//			return
	//		}
	//		res, err = c.GetEventReport(userName, "", f, t)
	//		fmt.Println(fmt.Sprintf(" --------------------------------------  %v Login History ------------------------------- ", userName))
	//		if err != nil {
	//			fmt.Println(err)
	//			return
	//		}
	//		err = DecodeBody(res, &result)
	//		if err != nil {
	//			fmt.Println(err)
	//			return
	//		}
	//		for _, event := range result.Events {
	//			fmt.Println(event)
	//		}
	//	}
	//}
	//
	//var wg sync.WaitGroup
	//q := make(chan string, 5)
	//for i := 0; i < len(AdminUsers); i++ {
	//	wg.Add(1)
	//	go GetLoginEvent(&wg, q)
	//}
	//
	//for _, admin := range AdminUsers {
	//	q <- admin.UserName
	//}
	//close(q)
	//wg.Wait()
}
