package transfer

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/egfanboy/mediapire-manager/internal/node"
	mhApi "github.com/egfanboy/mediapire-media-host/pkg/api"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type downloaderQueue struct {
	queue      []node.NodeConfig
	transferId string
	content    map[node.NodeConfig][]byte
}

func (q *downloaderQueue) processNode(ctx context.Context, n node.NodeConfig) error {
	client := mhApi.NewClient(n)

	content, _, err := client.DownloadTransfer(ctx, q.transferId)
	if err != nil {
		return err
	}

	q.content[n] = content

	return nil
}

func (q *downloaderQueue) ProcessQueue(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	wg := sync.WaitGroup{}

	// get media from all nodes at the same time
	for _, nodeInQueue := range q.queue {
		wg.Add(1)

		go func(n node.NodeConfig) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return // Error somewhere, terminate
			default:
			}

			err := q.processNode(ctx, n)
			if err != nil {
				log.Err(err).Msgf("An error occured when downloading content from node %s", n.NodeHost)
				cancel()
			}
		}(nodeInQueue)

	}

	wg.Wait()

	if ctx.Err() != nil && ctx.Err() == context.Canceled {
		return errors.New("failed to download selected content from nodes")
	}

	return nil
}

func (q *downloaderQueue) GetContent(ctx context.Context) ([]byte, error) {
	numberOfNodes := len(q.queue)
	result := new(bytes.Buffer)
	zipWriter := zip.NewWriter(result)

	for _, nodeContent := range q.content {
		// this is the only node, just return the content
		if numberOfNodes == 1 {
			return nodeContent, nil
		}

		br := bytes.NewReader(nodeContent)

		zipReader, err := zip.NewReader(br, br.Size())
		if err != nil {
			return nil, err
		}

		for _, file := range zipReader.File {
			zipItem, err := file.Open()
			if err != nil {
				return nil, err
			}

			w, err := zipWriter.Create(file.Name)
			if err != nil {
				return nil, err
			}

			_, err = io.Copy(w, zipItem)
			if err != nil {
				return nil, err
			}
		}
	}

	err := zipWriter.Close()
	return result.Bytes(), err
}

type mediaDownloader struct{}

func (mediaDownloader) Download(ctx context.Context, transferId primitive.ObjectID, nodeIds []string) ([]byte, error) {
	log.Info().Msgf("Starting download of content for transfer %s", transferId.Hex())
	nodes := make([]node.NodeConfig, 0)

	nodeService, err := node.NewNodeService()
	if err != nil {
		return nil, err
	}

	allNodes, err := nodeService.GetAllNodes(ctx)
	if err != nil {
		return nil, err
	}

	for _, nodeId := range nodeIds {
		var node *node.NodeConfig

		for i := range allNodes {
			if allNodes[i].Id == nodeId {
				node = &allNodes[i]
				break
			}
		}

		if node != nil {
			nodes = append(nodes, *node)
		} else {
			return nil, fmt.Errorf("no node found with id %q", nodeId)
		}
	}

	queue := &downloaderQueue{queue: nodes, content: map[node.NodeConfig][]byte{}, transferId: transferId.Hex()}

	err = queue.ProcessQueue(ctx)
	if err != nil {
		log.Err(err).Msg("Failed to process download queue")

		return nil, err
	}

	content, err := queue.GetContent(ctx)
	if err != nil {
		log.Err(err).Msg("Failed to merge downloaded content")

		return nil, err
	}

	return content, nil
}
