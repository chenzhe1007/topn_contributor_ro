package fetcher

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/tomnomnom/linkheader"
	"golang.org/x/oauth2"
)

//threshold below which we start to wait until the reset time from prev req
const RateLimitThreshold = 1000

type HttpFetcher interface {
	Next(nextUrl string) (bool, string, time.Duration, []byte, error)
}

type httpFetcher struct {
	httpClient *http.Client
}

// use the default 100 per page
func NewHttpFetcher(ctx context.Context, token oauth2.TokenSource) HttpFetcher {
	authClient := oauth2.NewClient(ctx, token)
	return httpFetcher{authClient}
}

func (rf httpFetcher) Next(nextUrl string) (bool, string, time.Duration, []byte, error) {

	resp, err := rf.httpClient.Get(nextUrl)
	parsedResp := make([]byte, 0)
	//TODO: handle error response
	if err != nil {
		return false, "", 0, parsedResp, err
	}

	parsedResp, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while parse resp body, respBody: %v, err: %v", resp.Body, err)
	}
	defer resp.Body.Close()

	hasMore, nextUrl := rf.hasNext(resp.Header.Get("Link"))

	remainStr := resp.Header.Get("X-Ratelimit-Remaining")
	resetStr := resp.Header.Get("X-Ratelimit-Reset")
	waitDuration, err := getWaitDuration(remainStr, resetStr)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while get wait duration, remainStr: %v, resetStr: %v, err: %v, Header: %v", remainStr, resetStr, err, resp.Header)
		return hasMore, nextUrl, waitDuration, parsedResp, err
	}

	return hasMore, nextUrl, waitDuration, parsedResp, err
}

func (rf httpFetcher) hasNext(links string) (bool, string) {
	for _, link := range linkheader.Parse(links) {
		if link.Rel == "next" {
			return true, link.URL
		}
	}

	return false, ""
}

func getWaitDuration(xRateRemain, xRateReset string) (time.Duration, error) {

	remainInt, err := strconv.Atoi(xRateRemain)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while converting x-ratelimit-remaining: %s, error: %v", xRateRemain, err)
		return 0, err
	}

	if remainInt < RateLimitThreshold {
		log.Println("x-Ratelimiting-Remain is less than threshold, enforcing a wait time, remain: ", xRateRemain)
		resetUnix, err := strconv.ParseInt(xRateReset, 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while converting x-ratelimit-reset: %s, error: %v", xRateReset, err)
			return 0, err
		}
		return time.Unix(resetUnix, 0).Sub(time.Now()), nil
	}
	return 0, err
}
