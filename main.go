package main

import (
	"context"
	"flag"
	"fmt"

	"zchen_topn_contributor_ro/job"

	"golang.org/x/oauth2"
)

var owner string
var limit int
var concurr int
var accessToken string
var max_concurr int

func init() {
	flag.StringVar(&owner, "owner", "chenzhe1007", "owner or org we want to get top contributors for, default to chenzhe1007")
	flag.IntVar(&limit, "limit", 20, "limit the number of top contributor to return, default to 20")
	flag.IntVar(&concurr, "concurrency", 5, "maximum concurrency request to make")
	flag.StringVar(&accessToken, "access-token", "", "access token for api authentication")
	max_concurr = 20

}

func main() {

	flag.Parse()
	fmt.Println("owner:", owner)
	fmt.Println("limit:", limit)
	fmt.Println("concurrency:", concurr)
	fmt.Println("access-token:", accessToken)

	concur = model.min(max_concurr, concur)

	ctx := context.Background()
	tc := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)

	contriListJobChan := job.CreateContriListJobs(ctx, tc, owner)

	result := job.CreateContriListWorkerPool(ctx, concurr, tc, contriListJobChan)

	done := make(chan bool)
	go job.CreateResultWorker(done, result, limit)
	<-done
	fmt.Println("Done")
}
