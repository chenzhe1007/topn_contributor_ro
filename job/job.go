package job

import (
	"container/heap"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
	"zchen_topn_contributor_ro/fetcher"
	"zchen_topn_contributor_ro/model"

	"golang.org/x/oauth2"
)

var REPOLIST_URL_PATTERN string
var CONTRILIST_URL_PATTERN string

func init() {
	REPOLIST_URL_PATTERN = "https://api.github.com/users/%v/repos"
	CONTRILIST_URL_PATTERN = "https://api.github.com/repos/%v/%v/contributors"
}

type RepoListJob struct {
	BaseUrl string
}

func CreateContriListJobs(ctx context.Context, token oauth2.TokenSource, owner string) chan ContriListJob {
	output := make(chan ContriListJob, 200)
	go createContriListJobHelper(ctx, token, owner, output)
	return output
}

func createContriListJobHelper(ctx context.Context, token oauth2.TokenSource, owner string, output chan ContriListJob) {

	repoListJob := NewRepoListJob(owner)
	repoListFetcher := fetcher.NewHttpFetcher(ctx, token)
	log.Println(fmt.Sprintf("Job to get repo list started, url: %v", repoListJob.BaseUrl))

	baseUrl := repoListJob.BaseUrl
	total := 0

	hasMore, nextUrl, waitDuration, resp, err := repoListFetcher.Next(baseUrl)
	for {
		hasMore, nextUrl, waitDuration, resp, err = repoListFetcher.Next(baseUrl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while requesting, err: %v, url: %v", err, baseUrl)
			break
		}
		baseUrl = nextUrl
		time.Sleep(waitDuration)

		jobs, err := repoListJob.ProcessResp(resp, owner)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while requesting, err: %v, url: %v", err, baseUrl)
			continue
		}
		total += len(jobs)
		for _, curJob := range jobs {
			output <- curJob
		}

		if !hasMore {
			break
		}
	}
	log.Println(fmt.Println("%v contributor list job created", total))
	close(output)
}

func (rj RepoListJob) ProcessResp(resp []byte, owner string) ([]ContriListJob, error) {

	var repos []model.RepoName
	err := json.Unmarshal(resp, &repos)
	if err != nil {
		return nil, err
	}
	newJob := make([]ContriListJob, 0)
	for _, repo := range repos {
		contriListUrl := formContriListUrl(owner, repo.Name)
		newJob = append(newJob, NewContriListJob(contriListUrl))
	}
	return newJob, nil
}

func NewRepoListJob(owner string) RepoListJob {
	baseUrl := formRepoListUrl(owner)
	return RepoListJob{baseUrl}
}

func formRepoListUrl(owner string) string {
	return fmt.Sprintf(REPOLIST_URL_PATTERN, owner)
}

type ContriListJob struct {
	BaseUrl string
}

func formContriListUrl(owner string, repoName string) string {
	return fmt.Sprintf(CONTRILIST_URL_PATTERN, owner, repoName)
}

func (cj ContriListJob) ProcessResp(resp []byte) ([]model.Contributor, error) {

	var contributors []model.Contributor
	err := json.Unmarshal(resp, &contributors)
	if err != nil {
		return nil, err
	}
	return contributors, nil
}

func NewContriListJob(baseUrl string) ContriListJob {
	return ContriListJob{baseUrl}
}

func PublishToResult(ctx context.Context, wg *sync.WaitGroup, token oauth2.TokenSource, jobs chan ContriListJob, results chan []model.Contributor) {

	for curJob := range jobs {
		log.Println("Job to get conributor list for repo: ", curJob)
		contriListFetcher := fetcher.NewHttpFetcher(ctx, token)
		baseUrl := curJob.BaseUrl
		for {
			hasMore, nextUrl, waitDuration, resp, err := contriListFetcher.Next(baseUrl)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while requesting, err: %v, url: %v", err, baseUrl)
				break
			}
			baseUrl = nextUrl
			time.Sleep(waitDuration)
			contributors, err := curJob.ProcessResp(resp)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while unmarshal resp, err: %v, resp: %v", err, resp)
				continue
			}
			results <- contributors
			if !hasMore {
				break
			}
		}
	}
	wg.Done()
}

func CreateContriListWorkerPool(ctx context.Context, numOfWorker int, token oauth2.TokenSource, jobs chan ContriListJob) chan []model.Contributor {

	results := make(chan []model.Contributor, numOfWorker*1000)
	var wg sync.WaitGroup
	for i := 0; i < numOfWorker; i++ {
		wg.Add(1)
		go PublishToResult(ctx, &wg, token, jobs, results)
	}
	wg.Wait()
	close(results)
	return results
}

func CreateResultWorker(done chan bool, results chan []model.Contributor, top int) {
	fmt.Println("started result worker")
	countMap := make(map[string]int)
	for contributors := range results {
		for _, contributor := range contributors {
			countMap[contributor.Name] += contributor.Contribution
		}
	}
	fmt.Println("map.size: ", len(countMap))

	pq := &model.ContributorHeap{}
	for key, val := range countMap {
		pq.Push(model.Contributor{key, val})
		if pq.Len() > top {
			heap.Pop(pq)
		}
	}
	topn := make([]model.Contributor, pq.Len(), pq.Len())
	for i := pq.Len() - 1; pq.Len() > 0; i-- {
		topn[i] = heap.Pop(pq).(model.Contributor)
	}

	for _, contributor := range topn {
		fmt.Println(fmt.Sprintf("%v : %v", contributor.Name, contributor.Contribution))
	}
	done <- true
}
