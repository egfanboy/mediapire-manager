package changeset

import (
	"context"
	"fmt"

	"github.com/egfanboy/mediapire-manager/internal/media"
	"github.com/egfanboy/mediapire-manager/pkg/types"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChangesetApi interface {
	GetChangesets(ctx context.Context) ([]*Changeset, error)
	GetChangesetById(ctx context.Context, changesetId primitive.ObjectID) (*Changeset, error)
	CreateChangeset(ctx context.Context, request types.ChangesetCreateRequest) (*Changeset, error)
}

type service struct {
	repo         changesetRepository
	mediaService media.MediaApi
}

func (s *service) GetChangesets(ctx context.Context) (result []*Changeset, err error) {
	log.Info().Msg("Start: Get Changesets")

	result, err = s.repo.GetAll(ctx)
	if err != nil {
		return
	}

	log.Info().Msg("End: Get Changesets")
	return
}

func (s *service) GetChangesetById(ctx context.Context, changesetId primitive.ObjectID) (result *Changeset, err error) {
	log.Info().Msg("Start: Get Changeset by Id")

	result, err = s.repo.GetById(ctx, changesetId)
	if err != nil {
		return
	}

	log.Info().Msg("End: Get Changeset by Id")
	return
}

func (s *service) CreateChangeset(ctx context.Context, request types.ChangesetCreateRequest) (result *Changeset, err error) {
	log.Info().Msg("Start: Create Changeset")
	result, err = newChangesetFromRequest(request)
	if err != nil {
		log.Err(err).Msg("Failed to convert request to changeset")
		return nil, err
	}

	err = s.repo.Save(ctx, result)
	if err != nil {
		log.Err(err).Msg("Failed to save changeset record")
		return
	}

	// asynchronously start changeset
	go s.delegateChangeset(result)

	log.Info().Msg("End: Create Changeset")
	return
}

// runs asynchronously as a goroutine
func (s *service) delegateChangeset(cs *Changeset) error {
	ctx := context.Background()

	switch cs.Type {
	case TypeUpdate:
		{
			// Execute anonymous function so we can just catch any errors to update record
			err := func(ctx context.Context, cs *Changeset) error {
				changes, err := cs.GetChanges()
				if err != nil {

					return err
				}
				err = s.mediaService.InternalUpdateMedia(ctx, cs.Id.Hex(), changes)
				if err != nil {
					return err
				}

				return nil
			}(ctx, cs)
			if err != nil {
				cs.Status = StatusFailed
				cs.FailureReason = err.Error()
			}

			err = s.repo.Save(ctx, cs)
			if err != nil {
				log.Err(err).Msgf("Failed to update change set %s", cs.Id.Hex())

				return err
			}

			return nil
		}

	default:
		cs.Status = StatusFailed
		cs.FailureReason = fmt.Sprintf("changeset was for action %s which is not supported", cs.Type)
		err := s.repo.Save(ctx, cs)
		if err != nil {
			log.Err(err).Msgf("Failed to update change set %s", cs.Id.Hex())

			return err
		}

	}

	return nil
}

func newChangesetService(ctx context.Context) (ChangesetApi, error) {
	repo, err := newChangesetRepository(ctx)
	if err != nil {
		return nil, err
	}

	mediaService, err := media.NewMediaService()
	if err != nil {
		return nil, err
	}

	s := &service{
		repo:         repo,
		mediaService: mediaService,
	}

	return s, nil
}
