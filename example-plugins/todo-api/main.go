package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/calvinmclean/babyapi"
)

type TODO struct {
	babyapi.DefaultResource

	Title       string
	Description string
	Completed   *bool
	CreatedAt   time.Time
}

func (t *TODO) Patch(newTODO *TODO) *babyapi.ErrResponse {
	if newTODO.Title != "" {
		t.Title = newTODO.Title
	}
	if newTODO.Description != "" {
		t.Description = newTODO.Description
	}
	if newTODO.Completed != nil {
		t.Completed = newTODO.Completed
	}

	return nil
}

func (t *TODO) Bind(r *http.Request) error {
	err := t.DefaultResource.Bind(r)
	if err != nil {
		return err
	}

	switch r.Method {
	case http.MethodPost:
		t.CreatedAt = time.Now()
		fallthrough
	case http.MethodPut:
		if t.Title == "" {
			return errors.New("missing required title field")
		}
	}

	return nil
}

func Run(ctx context.Context, ip string) error {
	api := NewAPI().WithContext(ctx)
	return api.Serve(ip + ":8080")
}

func NewAPI() *babyapi.API[*TODO] {
	api := babyapi.NewAPI("TODOs", "/todos", func() *TODO { return &TODO{} })
	api.SetGetAllFilter(func(r *http.Request) babyapi.FilterFunc[*TODO] {
		return func(t *TODO) bool {
			getCompletedParam := r.URL.Query().Get("completed")
			// No filtering if param is not provided
			if getCompletedParam == "" {
				return true
			}

			if getCompletedParam == "true" {
				return t.Completed != nil && *t.Completed
			}

			return t.Completed == nil || !*t.Completed
		}
	})

	return api
}

func main() {
	api := NewAPI()
	api.RunCLI()
}
