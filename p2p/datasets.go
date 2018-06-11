package p2p

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/qri-io/dataset/dsfs"
	"github.com/qri-io/qri/repo"

	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
)

// MtDatasets is a dataset list message
const MtDatasets = MsgType("list_datasets")

// listMax is the highest number of entries a list request should return
const listMax = 30

// DatasetsListParams encapsulates options for requesting datasets
type DatasetsListParams struct {
	Limit  int
	Offset int
	RPC    bool
}

// RequestDatasetsList gets a list of a peer's datasets
func (n *QriNode) RequestDatasetsList(pid peer.ID, p DatasetsListParams) ([]repo.DatasetRef, error) {
	log.Debugf("%s RequestDatasetList: %s", n.ID, pid)

	if pid == n.ID {
		// requesting self isn't a network operation
		return n.Repo.References(p.Limit, p.Offset)
	}

	if !n.Online {
		return nil, fmt.Errorf("not connected to p2p network")
	}

	req, err := NewJSONBodyMessage(n.ID, MtDatasets, p)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}

	req = req.WithHeaders("phase", "request", "RPC", fmt.Sprintf("%v", p.RPC))

	replies := make(chan Message)
	err = n.SendMessage(req, replies, pid)
	if err != nil {
		log.Debug(err.Error())
		return nil, fmt.Errorf("send dataset info message error: %s", err.Error())
	}

	res := <-replies
	ref := []repo.DatasetRef{}
	err = json.Unmarshal(res.Body, &ref)
	return ref, err
}

func (n *QriNode) handleDatasetsList(ws *WrappedStream, msg Message) (hangup bool) {
	hangup = true
	switch msg.Header("phase") {
	case "request":
		dlp := DatasetsListParams{}
		if err := json.Unmarshal(msg.Body, &dlp); err != nil {
			log.Debugf("%s %s", n.ID, err.Error())
			return
		}

		if dlp.Limit == 0 || dlp.Limit > listMax {
			dlp.Limit = listMax
		}

		refs, err := n.Repo.References(dlp.Limit, dlp.Offset)
		if err != nil {
			log.Debug(err.Error())
			return
		}

		for i, ref := range refs {
			if i >= dlp.Limit {
				break
			}
			ds, err := dsfs.LoadDataset(n.Repo.Store(), datastore.NewKey(ref.Path))
			if err != nil {
				log.Info("error loading dataset at path:", ref.Path)
				return
			}
			refs[i].Dataset = ds.Encode()
			// The gob encoder that encodes from go to the terminal does not understand how to
			// handle `map[string]interface{}`s that contain a field of `map[string]interface{}`
			// in this case, the offender is the Structure.Schema. If we are using RPC, let's clear
			// the schema since, for now, we don't need it to list a peer's datasets
			if msg.Header("RPC") == "true" && refs[i].Dataset.Structure != nil && refs[i].Dataset.Structure.Schema != nil {
				refs[i].Dataset.Structure.Schema = map[string]interface{}{}
			}
		}

		reply, err := msg.UpdateJSON(refs)
		reply = reply.WithHeaders("phase", "response")
		if err := ws.sendMessage(reply); err != nil {
			log.Debug(err.Error())
			return
		}
	}

	return
}
