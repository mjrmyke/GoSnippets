// Myke Walker
// 11-8-2017
// Program to crawl a github repo by organization
//
// run the program by using the following
// go run main.go -org='orgname'
// organization aname defaults to customerio if no flag is present

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sort"

	"github.com/google/go-github/github"
)

func main() {
	//grab flag, or use default of customerio
	org := flag.String("org", "customerio", "organization to crawl on github")

	//parse flags
	flag.Parse()

	//init client and context
	ctx := context.Background()
	gclient := github.NewClient(nil)

	//call for all repos by the organization
	repos, _, err := gclient.Repositories.ListByOrg(ctx, *org, nil)
	if err != nil {
		log.Println("error requestion organizations", err)
	}

	//vars to track necessary requirements
	var maxwatchers int
	var mostwatchedrepo string
	var numopenissues int

	//range through repos
	for _, v := range repos {

		//determine most watched repo
		if v.GetWatchersCount() > maxwatchers {
			maxwatchers = v.GetWatchersCount()
			mostwatchedrepo = v.GetName()
		}

		//add the number of openissu
		numopenissues = numopenissues + v.GetOpenIssuesCount()

	}

	//sort the slice of repos in descending order by date created
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].GetCreatedAt().Unix() > repos[j].GetCreatedAt().Unix()
	})

	fmt.Println("Total number of open issues across the organization of ", *org, " is: ", numopenissues)
	fmt.Println("-------------------------------------------------------------------------")
	fmt.Println("Repos sorted by Date Created in descending order")
	for _, v := range repos {
		fmt.Println("Repo: ", v.GetName(), " Created On ", v.GetCreatedAt())
	}

	fmt.Println("-------------------------------------------------------------------------")
	fmt.Println("Repository with the maximum number of watchers is:", mostwatchedrepo)
	fmt.Println("The number of people watching the repo is:", maxwatchers)
	// fmt.Println("list of organizations:", repos)
}
