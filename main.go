package main

import (
	"fmt"
	"lastpass_provisioning/api"
	lastpassTime "lastpass_provisioning/lastpass_time"
	"os"
	"sync"
	"time"
)

func main() {
	// Client作成
	c, err := NewClient(nil)
	if err != nil {
		fmt.Errorf("Failed Building Client: %v", err.Error())
		os.Exit(1)
	}

	fmt.Println(" --------------------------------------  Admin Users ---------------------------------------- ")
	s := NewService(c)
	AdminUsers, err := s.GetAdminUserData()

	for _, admin := range AdminUsers {
		fmt.Println(admin.UserName)
	}

	// Get Shared Folder Data
	fmt.Println(" --------------------------------------  Super Shared Folders ------------------------------- ")
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

	fmt.Println(fmt.Sprintf(" -------------------------------- Events(%v ~ %v) -----------------------------------", f.Format(), t.Format()))
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

	// Delete
	//_, err = c.DeleteUser("takizawa.naoto@moneyforward.co.jp", Delete)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//_, err = c.DeleteUser("suga.kosuke@moneyforward.co.jp", Delete)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}

	// Move
	hoge := []string{
		"yamada.kazuya@moneyforward.co.jp",
		"taniguchi.toru@moneyforward.co.jp",
		"harada.toru@moneyforward.co.jp",
		"kaneko.yuichiro@moneyforward.co.jp",
		"fukahori.munetoshi@moneyforward.co.jp",
		"ogura.takashi@moneyforward.co.jp",
		"shinoda.takeshi@moneyforward.co.jp",
		"otsuka.hayato@moneyforward.co.jp",
		"takasugi.miyo@moneyforward.co.jp",
		"sato.ko@moneyforward.co.jp",
		"suda.michiko@moneyforward.co.jp",
		"izumiya.keisuke@moneyforward.co.jp",
		"akiyama.chie@moneyforward.co.jp",
		"oshima.mari@moneyforward.co.jp",
		"hamada.akira@moneyforward.co.jp",
		"takahashi.yuto@moneyforward.co.jp",
		"shibuya.ryo@moneyforward.co.jp",
		"ishii.hiroyuki@moneyforward.co.jp",
		"masuyama.daisuke@moneyforward.co.jp",
		"otsu.hiroaki@moneyforward.co.jp",
		"tsuji.shuhei@moneyforward.co.jp",
		"uemura.saeko@moneyforward.co.jp",
		"kuroda.naoki@moneyforward.co.jp",
		"nishinobu.takahiro@moneyforward.co.jp",
		"ito.daisuke@moneyforward.co.jp",
		"kashiwazaki.takahiro@moneyforward.co.jp",
		"aoyama.reiko@moneyforward.co.jp",
		"furuhama.yusuke@moneyforward.co.jp",
		"ito.yusuke@moneyforward.co.jp",
		"kato.takuya@moneyforward.co.jp",
		"nozaki.keigo@moneyforward.co.jp",
		"sogo.kenya@moneyforward.co.jp",
		"hashimoto.satoshi@moneyforward.co.jp",
		"iwamura.yuna@moneyforward.co.jp",
		"minegishi.kai@moneyforward.co.jp",
		"kuroki.yasuha@moneyforward.co.jp",
	}

	var wgwg sync.WaitGroup
	for _, h := range hoge	{
		wgwg.Add(1)
		go func(name string) {
			defer wgwg.Done()

			_, err = c.ChangeGroupsMembership([]api.BelongingGroup{
				{
					name,
					[]string{"MFクラウドサービス開発本部"},
					[]string{},
				},

			})
		}(h)
	}
	wgwg.Wait()

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

	// Add
	//_, err = c.BatchAddOrUpdateUsers(
	//	[]*api.User{
	//		{UserName:"takahashi.yuto@moneyforward.co.jp",Groups:[]string{"MFクラウドサービス開発本部"}},
	//		{UserName:"ishii.hiroyuki@moneyforward.co.jp",Groups:[]string{"MFクラウドサービス開発本部"}},
	//		{UserName:"suzuki.shota.340@moneyforward.co.jp",Groups:[]string{"PFMサービス開発本部"}},
	//		{UserName:"oba.akitaka@moneyforward.co.jp",Groups:[]string{"PFMサービス開発本部"}},
	//		{UserName:"ono.yumemi@moneyforward.co.jp",Groups:[]string{"アカウントアグリゲーション本部"}},
	//		{UserName:"takenaka.kazumasa@moneyforward.co.jp",Groups:[]string{"MFクラウド事業推進本部 - 事業戦略部"}},
	//		{UserName:"ukon.yuto@@moneyforward.co.jp", Groups:[]string{"MFクラウド事業推進本部 - ダイレクトセールス部"}},
	//		{UserName:"lee.choonghaeng@moneyforward.co.jp",Groups: []string{"MFクラウド事業推進本部 - MFクラウド事業戦略部"}},
	//		{UserName:"furuhama.yusuke@moneyforward.co.jp", Groups: []string{"社長室 - Chalin", "MFクラウドサービス開発本部"}},
	//	},
	//)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
}
