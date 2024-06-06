package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"net/http"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/query/pkg/query"
	"github.com/TrueBlocks/trueblocks-key/test/simulate_session/pkg/config"
	"github.com/TrueBlocks/trueblocks-key/test/simulate_session/pkg/scenario"
)

type Result struct {
	Ok       bool
	Error    ResultError
	Duration time.Duration
}

type ResultError struct {
	StatusCode int
	Text       string
}

func (re *ResultError) String() string {
	return fmt.Sprintf("%d: %s", re.StatusCode, re.Text)
}

func Run(cnf *config.Config, results chan<- Result) {
	defer close(results)

	ctx, cancel := context.WithTimeout(context.Background(), cnf.Duration)
	defer cancel()

	var wg sync.WaitGroup

	for _, s := range cnf.Scenarios {
		wg.Add(1)
		go func(s *scenario.Scenario) {
			defer wg.Done()
			RunSingle(s, cnf, ctx, results)
		}(&s)
	}

	wg.Wait()
}

func RunSingle(s *scenario.Scenario, cnf *config.Config, timeout context.Context, results chan<- Result) {
	// https://go.dev/wiki/RateLimiting
	// rateLimit := time.Second / time.Duration(cnf.Rate)
	// throttle := time.Tick(rateLimit)

	var pageNum int
	var nextPageId *query.PageId
	for {
		body, err := getBody(s, nextPageId)
		if err != nil {
			log.Fatalln("getBody:", err)
		}
		url := cnf.BaseUrl
		if s.DirectUser != "" {
			url += "/" + s.DirectUser
		}
		req, err := http.NewRequestWithContext(timeout, http.MethodPost, url, body)
		if err != nil {
			log.Fatalln("new request error:", err)
		}
		req.Header = s.Headers

		var res Result
		start := time.Now()
		resp, err := http.DefaultClient.Do(req)
		elapsed := time.Since(start)
		var statusCode int
		if resp != nil {
			statusCode = resp.StatusCode
		}
		if err != nil {
			res = Result{
				Error: ResultError{
					StatusCode: statusCode,
					Text:       err.Error(),
				},
				Duration: elapsed,
			}
		} else {
			defer resp.Body.Close()
			res = Result{
				Ok:       statusCode < 400,
				Duration: elapsed,
			}

			rpcResponse := new(query.RpcResponse[[]database.PublicAppearance])
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln("reading body:", err)
			}

			if statusCode >= 400 {
				res.Error = ResultError{
					StatusCode: statusCode,
					Text:       string(respBody),
				}
			} else {
				if err := json.Unmarshal(respBody, rpcResponse); err != nil {
					log.Fatalln("unmarshal body:", err)
				}

				if s.GoBackwards {
					nextPageId = rpcResponse.Meta.NextPageId
				} else {
					nextPageId = rpcResponse.Meta.PreviousPageId
				}
				// log.Println("got", len(rpcResponse.Data), "appearances, next page?", nextPageId)
			}
		}

		select {
		case <-timeout.Done():
			return
		case results <- res:
			if nextPageId == nil {
				log.Println("no next page, returning:", pageNum)
			} else {
				pageNum++
			}
		}
	}
}

func getBody(s *scenario.Scenario, nextPageId *query.PageId) (r *bytes.Reader, err error) {
	request := &query.RpcRequest{
		Method: query.MethodGetAppearances,
	}
	param := query.RpcGetAppearancesParam{
		Address: s.Address,
		PerPage: s.PerPage,
	}
	err = query.SetParams(
		request,
		[]query.RpcGetAppearancesParam{param},
	)
	if err != nil {
		return
	}
	pageSpecial := query.PageIdNoSpecial
	if s.GoBackwards && nextPageId == nil {
		pageSpecial = query.PageIdEarliest
	}
	if err = param.SetPageId(pageSpecial, nextPageId); err != nil {
		return
	}
	b, err := json.Marshal(request)
	if err != nil {
		return
	}
	r = bytes.NewReader(b)
	return
}
