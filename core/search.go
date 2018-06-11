package core

import (
	"fmt"
	"net/rpc"

	// 	"github.com/qri-io/cafs"
	"github.com/qri-io/qri/repo"
	"github.com/qri-io/qri/repo/actions"
	"github.com/qri-io/registry/regclient"
)

// SearchRequests encapsulates business logic for the qri search
// command
type SearchRequests struct {
	cli  *rpc.Client
	repo repo.Repo
}

// NewSearchRequests creates a SearchRequests pointer from either a repo
// or an rpc.Client
func NewSearchRequests(r repo.Repo, cli *rpc.Client) *SearchRequests {
	if r != nil && cli != nil {
		panic(fmt.Errorf("both repo and client supplied to NewSearchRequests"))
	}
	return &SearchRequests{
		cli:  cli,
		repo: &actions.Registry{r},
	}
}

// CoreRequestsName implements the requests
func (sr SearchRequests) CoreRequestsName() string { return "search" }

// SearchParams defines paremeters for the search Method
type SearchParams struct {
	QueryString string
	Limit       int
	Offset      int
}

// SearchResult struct
type SearchResult struct {
	Type, ID string
	Value    interface{}
}

// Search queries for items on qri related to given parameters
func (sr *SearchRequests) Search(p *SearchParams, results *[]SearchResult) error {
	if sr.cli != nil {
		return sr.cli.Call("SearchRequests.Search", p, results)
	}
	reg := sr.repo.Registry()
	if p == nil {
		return fmt.Errorf("error: search params cannot be nil")
	}
	params := &regclient.SearchParams{p.QueryString, nil, p.Limit, p.Offset}

	regResults, err := reg.Search(params)
	if err != nil {
		return err
	}

	searchResults := make([]SearchResult, len(regResults))
	for i, result := range regResults {
		searchResults[i].Type = result.Type
		searchResults[i].ID = result.ID
		searchResults[i].Value = result.Value
	}
	*results = searchResults
	return nil
}

// // SearchRequests encapsulates business logic for the qri search
// // command
// type SearchRequests struct {
// 	store cafs.Filestore
// 	repo  repo.Repo
// 	// node  *p2p.QriNode
// 	cli *rpc.Client
// }

// // CoreRequestsName implements the requests
// func (d SearchRequests) CoreRequestsName() string { return "search" }

// // NewSearchRequests creates a SearchRequests pointer from either a repo
// // or an rpc.Client
// func NewSearchRequests(r repo.Repo, cli *rpc.Client) *SearchRequests {
// 	if r != nil && cli != nil {
// 		panic(fmt.Errorf("both repo and client supplied to NewSearchRequests"))
// 	}

// 	return &SearchRequests{
// 		repo: r,
// 		// node:  node,
// 		cli: cli,
// 	}
// }

// // Search queries for items on qri related to given parameters
// func (d *SearchRequests) Search(p *repo.SearchParams, res *[]repo.DatasetRef) error {
// 	if d.cli != nil {
// 		return d.cli.Call("SearchRequests.Search", p, res)
// 	}
// 	// if d.node != nil {
// 	// 	r, err := d.node.Search(p.Query, p.Limit, p.Offset)
// 	// 	if err != nil {
// 	// 		return err
// 	// 	}

// 	if searchable, ok := d.repo.(repo.Searchable); ok {
// 		results, err := searchable.Search(*p)
// 		if err != nil {
// 			log.Debug(err.Error())
// 			return fmt.Errorf("error searching: %s", err.Error())
// 		}
// 		*res = results
// 		return nil
// 	}

// 	return fmt.Errorf("this repo doesn't support search")
// }

// // ReindexSearchParams defines parmeters for
// // the Reindex method
// type ReindexSearchParams struct {
// 	// no args for reindex
// }

// // Reindex instructs a qri node to re-calculate it's search index
// func (d *SearchRequests) Reindex(p *ReindexSearchParams, done *bool) error {
// 	if d.cli != nil {
// 		return d.cli.Call("SearchRequests.Reindex", p, done)
// 	}

// 	if fsr, ok := d.repo.(*fsrepo.Repo); ok {
// 		err := fsr.UpdateSearchIndex(d.repo.Store())
// 		if err != nil {
// 			log.Debug(err.Error())
// 			return fmt.Errorf("error reindexing: %s", err.Error())
// 		}
// 		*done = true
// 		return nil
// 	}

// 	return fmt.Errorf("search reindexing is currently only supported on file-system repos")
// }
