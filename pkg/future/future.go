package future

import (
	"context"
	//"encoding/json"
	"net/http"
)

type Future[T any] interface {
	Wait() (*T, error)
}

type future[T any] struct {
	ctx context.Context
	ch  chan T
	err chan error
}

func New[T any](ctx context.Context, fn func() (T, error)) Future[T] {
	obj := &future[T]{
		ctx: ctx,
		ch:  make(chan T),
		err: make(chan error),
	}

	go func() {
		result, err := fn()
		if err != nil {
			obj.err <- err
			return
		}

		obj.ch <- result
	}()

	return obj
}

func (t future[T]) Wait() (*T, error) {
	select {
	case results := <-t.ch:
		return &results, nil
	case err := <-t.err:
		return nil, err
	case <-t.ctx.Done():
		return nil, t.ctx.Err()
	}
}

var httpClient = http.DefaultClient

func SetHTTPClient(client *http.Client) {
	httpClient = client
}

func NewHTTP[T any](ctx context.Context, req *http.Request, fn func(resp *http.Response, err error) (*T, error)) Future[T] {

	fut := &future[T]{
		ctx: ctx,
		ch:  make(chan T),
		err: make(chan error),
	}

	req = req.WithContext(ctx)
	go func() {
		result, err := fn(httpClient.Do(req))
		if err != nil {
			fut.err <- err
			return
		}

		fut.ch <- *result
	}()

	return fut
}

/*
func blah() {
	var results Results
	err := httpDo(ctx, req, func(resp *http.Response, err error) error {
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Parse the JSON search result.
		// https://developers.google.com/web-search/docs/#fonje
		var data struct {
			ResponseData struct {
				Results []struct {
					TitleNoFormatting string
					URL               string
				}
			}
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return err
		}
		for _, res := range data.ResponseData.Results {
			results = append(results, Result{Title: res.TitleNoFormatting, URL: res.URL})
		}
		return nil
	})
	// httpDo waits for the closure we provided to return, so it's safe to
	// read results here.
	return results, err
}

func httpDo(ctx context.Context, req *http.Request, f func(*http.Response, error) error) error {
	// Run the HTTP request in a goroutine and pass the response to f.
	c := make(chan error, 1)
	req = req.WithContext(ctx)
	go func() { c <- f(http.DefaultClient.Do(req)) }()
	select {
	case <-ctx.Done():
		<-c // Wait for f to return.
		return ctx.Err()
	case err := <-c:
		return err
	}
}
*/
