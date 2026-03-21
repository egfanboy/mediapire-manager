package playback

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/egfanboy/mediapire-common/exceptions"
	commonRouter "github.com/egfanboy/mediapire-common/router"
	"github.com/egfanboy/mediapire-manager/internal/media"
	"github.com/egfanboy/mediapire-manager/internal/websocket"
	"github.com/egfanboy/mediapire-manager/pkg/types"
)

const (
	commandNext            = "next"
	commandPrevious        = "previous"
	commandSetRepeat       = "set_repeat"
	commandSetShuffle      = "set_shuffle"
	commandSetCurrentIndex = "set_current_index"
)

type PlaybackApi interface {
	GetSession(ctx context.Context) (types.PlaybackSessionState, error)
	SetQueue(ctx context.Context, request types.PlaybackSetQueueRequest) (types.PlaybackSessionState, error)
	StartFromMedia(ctx context.Context, request types.PlaybackStartRequest) (types.PlaybackSessionState, error)
	ApplyCommand(ctx context.Context, request types.PlaybackCommandRequest) (types.PlaybackSessionState, error)
}

type playbackService struct {
	repo         playbackRepository
	mediaService media.MediaApi
	nowFn        func() time.Time
}

func (s *playbackService) GetSession(ctx context.Context) (types.PlaybackSessionState, error) {
	session, err := s.ensureSession(ctx)
	if err != nil {
		return types.PlaybackSessionState{}, err
	}

	return s.buildResponseState(ctx, session)
}

func (s *playbackService) SetQueue(ctx context.Context, request types.PlaybackSetQueueRequest) (types.PlaybackSessionState, error) {
	session, err := s.ensureSession(ctx)
	if err != nil {
		return types.PlaybackSessionState{}, err
	}

	if err := s.validateQueueItems(ctx, request.Items); err != nil {
		return types.PlaybackSessionState{}, err
	}

	now := s.nowFn()
	if err := setQueue(session, request.Items, request.StartIndex, now.UnixNano()); err != nil {
		return types.PlaybackSessionState{}, exceptions.NewBadRequestException(err)
	}

	saved, err := s.saveMutation(ctx, session, now)
	if err != nil {
		return types.PlaybackSessionState{}, err
	}

	state, err := s.buildResponseState(ctx, saved)
	if err != nil {
		return types.PlaybackSessionState{}, err
	}
	_ = websocket.SendPlaybackSessionUpdated(state)

	return state, nil
}

func (s *playbackService) StartFromMedia(ctx context.Context, request types.PlaybackStartRequest) (types.PlaybackSessionState, error) {
	session, err := s.ensureSession(ctx)
	if err != nil {
		return types.PlaybackSessionState{}, err
	}

	filtering := types.ApiFilteringParams{}
	if request.SortBy != nil && strings.TrimSpace(*request.SortBy) != "" {
		filtering, err = types.NewApiFilteringParams(commonRouter.RouteParams{Params: map[string]string{"sortBy": strings.TrimSpace(*request.SortBy)}})
		if err != nil {
			return types.PlaybackSessionState{}, err
		}
	}

	mediaTypes := splitCommaSeparated(request.MediaType)
	mediaIds := splitCommaSeparated(request.MediaIds)

	items, err := s.getMediaForStart(ctx, mediaTypes, mediaIds, filtering)
	if err != nil {
		return types.PlaybackSessionState{}, err
	}

	if request.RepeatMode != nil {
		mode := strings.TrimSpace(*request.RepeatMode)
		switch mode {
		case string(repeatModeOff):
			session.RepeatMode = repeatModeOff
		case string(repeatModeOne):
			session.RepeatMode = repeatModeOne
		case string(repeatModeAll):
			session.RepeatMode = repeatModeAll
		default:
			return types.PlaybackSessionState{}, exceptions.NewBadRequestException(fmt.Errorf("invalid repeat mode %s", mode))
		}
	}

	if request.ShuffleEnabled != nil {
		session.ShuffleEnabled = *request.ShuffleEnabled
	}

	queue := make([]types.MediaItemMapping, 0, len(items))
	for _, item := range items {
		queue = append(queue, types.MediaItemMapping{NodeId: item.NodeId, MediaId: item.Id})
	}

	now := s.nowFn()
	if err := setQueue(session, queue, request.StartIndex, now.UnixNano()); err != nil {
		return types.PlaybackSessionState{}, exceptions.NewBadRequestException(err)
	}

	saved, err := s.saveMutation(ctx, session, now)
	if err != nil {
		return types.PlaybackSessionState{}, err
	}

	state, err := s.buildResponseState(ctx, saved)
	if err != nil {
		return types.PlaybackSessionState{}, err
	}
	_ = websocket.SendPlaybackSessionUpdated(state)

	return state, nil
}

func (s *playbackService) ApplyCommand(ctx context.Context, request types.PlaybackCommandRequest) (types.PlaybackSessionState, error) {
	session, err := s.ensureSession(ctx)
	if err != nil {
		return types.PlaybackSessionState{}, err
	}

	now := s.nowFn()
	err = s.applyCommandMutation(ctx, session, request, now)
	if err != nil {
		return types.PlaybackSessionState{}, err
	}

	saved, err := s.saveMutation(ctx, session, now)
	if err != nil {
		return types.PlaybackSessionState{}, err
	}

	state, err := s.buildResponseState(ctx, saved)
	if err != nil {
		return types.PlaybackSessionState{}, err
	}
	_ = websocket.SendPlaybackSessionUpdated(state)

	return state, nil
}

func (s *playbackService) navigate(
	ctx context.Context,
	session *playbackSession,
	delta int,
) error {
	if len(session.PlayOrder) == 0 || session.CurrentPlayOrderIndex < 0 || session.CurrentPlayOrderIndex >= len(session.PlayOrder) {
		return nil
	}

	previousPos := session.CurrentPlayOrderIndex
	pos := previousPos

	for checked := 0; checked < len(session.PlayOrder); checked++ {
		nextPos := pos + delta
		if nextPos < 0 || nextPos >= len(session.PlayOrder) {
			if session.RepeatMode == repeatModeAll {
				if delta > 0 {
					nextPos = 0
				} else {
					nextPos = len(session.PlayOrder) - 1
				}
			} else {
				break
			}
		}

		pos = nextPos

		exists, err := s.itemExistsAtOrderPos(ctx, session, pos)
		if err != nil {
			return err
		}
		if exists {
			session.CurrentPlayOrderIndex = pos
			return nil
		}
	}

	existsAtPrevious, err := s.itemExistsAtOrderPos(ctx, session, previousPos)
	if err != nil {
		return err
	}

	if existsAtPrevious {
		session.CurrentPlayOrderIndex = previousPos
	} else {
		session.CurrentPlayOrderIndex = -1
	}

	return nil
}

func (s *playbackService) itemExistsAtOrderPos(ctx context.Context, session *playbackSession, orderPos int) (bool, error) {
	if orderPos < 0 || orderPos >= len(session.PlayOrder) {
		return false, nil
	}

	queueIndex := session.PlayOrder[orderPos]
	if queueIndex < 0 || queueIndex >= len(session.Queue) {
		return false, nil
	}

	item := session.Queue[queueIndex]
	mediaItems, err := s.mediaService.GetMedia(ctx, []string{}, []string{item.NodeId}, []string{item.MediaId})
	if err != nil {
		// Treat fetch errors as not-found for skip behavior.
		return false, nil
	}

	return len(mediaItems) > 0, nil
}

func (s *playbackService) buildResponseState(ctx context.Context, session *playbackSession) (types.PlaybackSessionState, error) {
	state := session.toApiState()

	if state.CurrentItem == nil {
		return state, nil
	}

	currentMedia, err := s.mediaService.GetMedia(
		ctx,
		[]string{},
		[]string{state.CurrentItem.NodeId},
		[]string{state.CurrentItem.MediaId},
	)
	if err != nil {
		return types.PlaybackSessionState{}, err
	}

	if len(currentMedia) > 0 {
		state.CurrentMedia = &currentMedia[0]
	}

	return state, nil
}

func (s *playbackService) getMediaForStart(
	ctx context.Context,
	mediaTypes []string,
	mediaIds []string,
	filtering types.ApiFilteringParams,
) ([]types.MediaItem, error) {
	if filtering.SortByField == nil {
		return s.mediaService.GetMedia(ctx, mediaTypes, []string{}, mediaIds)
	}

	result, err := s.mediaService.GetMediaPaginated(ctx, mediaTypes, []string{}, filtering, nil)
	if err != nil {
		return nil, err
	}

	mediaResponse, ok := result.(types.MediaResponse)
	if !ok {
		return nil, errors.New("unexpected media response type")
	}

	if len(mediaIds) == 0 {
		return mediaResponse.Results, nil
	}

	allowedIds := make(map[string]struct{}, len(mediaIds))
	for _, id := range mediaIds {
		allowedIds[id] = struct{}{}
	}

	filtered := make([]types.MediaItem, 0, len(mediaResponse.Results))
	for _, item := range mediaResponse.Results {
		if _, ok := allowedIds[item.Id]; ok {
			filtered = append(filtered, item)
		}
	}

	return filtered, nil
}

func splitCommaSeparated(value *string) []string {
	if value == nil {
		return []string{}
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return []string{}
	}

	parts := strings.Split(trimmed, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		v := strings.TrimSpace(part)
		if v != "" {
			result = append(result, v)
		}
	}

	return result
}

func (s *playbackService) saveMutation(ctx context.Context, session *playbackSession, now time.Time) (*playbackSession, error) {
	session.UpdatedAt = now

	if err := s.repo.Save(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *playbackService) ensureSession(ctx context.Context) (*playbackSession, error) {
	session, err := s.repo.Get(ctx)
	if err == nil {
		// Extend idle lifetime on access so the session only expires when unused.
		if saveErr := s.repo.Save(ctx, session); saveErr != nil {
			return nil, saveErr
		}
		return session, nil
	}

	if !errors.Is(err, errSessionNotFound) {
		return nil, err
	}

	now := s.nowFn()
	newSession := newPlaybackSession(sessionDocumentId, now)
	if err := s.repo.Create(ctx, newSession); err != nil {
		return nil, err
	}

	return newSession, nil
}

func (s *playbackService) validateQueueItems(ctx context.Context, items []types.MediaItemMapping) error {
	for _, item := range items {
		found, err := s.mediaService.GetMedia(ctx, []string{}, []string{item.NodeId}, []string{item.MediaId})
		if err != nil {
			return err
		}

		if len(found) == 0 {
			return exceptions.NewBadRequestException(
				fmt.Errorf("media item %s on node %s does not exist", item.MediaId, item.NodeId),
			)
		}
	}

	return nil
}

func (s *playbackService) applyCommandMutation(
	ctx context.Context,
	session *playbackSession,
	request types.PlaybackCommandRequest,
	now time.Time,
) error {
	switch request.Command {
	case commandNext:
		return s.navigate(ctx, session, 1)
	case commandPrevious:
		return s.navigate(ctx, session, -1)
	case commandSetRepeat:
		if request.Payload.Mode == nil {
			return exceptions.NewBadRequestException(errors.New("payload.mode is required for set_repeat"))
		}

		switch *request.Payload.Mode {
		case string(repeatModeOff):
			session.RepeatMode = repeatModeOff
		case string(repeatModeOne):
			session.RepeatMode = repeatModeOne
		case string(repeatModeAll):
			session.RepeatMode = repeatModeAll
		default:
			return exceptions.NewBadRequestException(fmt.Errorf("invalid repeat mode %s", *request.Payload.Mode))
		}
	case commandSetShuffle:
		if request.Payload.Enabled == nil {
			return exceptions.NewBadRequestException(errors.New("payload.enabled is required for set_shuffle"))
		}

		currentQueueIndex := -1
		if session.CurrentPlayOrderIndex >= 0 && session.CurrentPlayOrderIndex < len(session.PlayOrder) {
			currentQueueIndex = session.PlayOrder[session.CurrentPlayOrderIndex]
		}

		session.ShuffleEnabled = *request.Payload.Enabled
		if len(session.Queue) > 0 {
			session.ShuffleSeed = now.UnixNano()
			session.PlayOrder = buildPlayOrder(len(session.Queue), session.ShuffleEnabled, session.ShuffleSeed)

			if currentQueueIndex >= 0 {
				if session.ShuffleEnabled {
					orderPos, err := getOrderPosByQueueIndex(session.PlayOrder, currentQueueIndex)
					if err != nil {
						return exceptions.NewBadRequestException(err)
					}
					session.PlayOrder[0], session.PlayOrder[orderPos] = session.PlayOrder[orderPos], session.PlayOrder[0]
					session.CurrentPlayOrderIndex = 0
					return nil
				}

				newPos, err := getOrderPosByQueueIndex(session.PlayOrder, currentQueueIndex)
				if err != nil {
					return exceptions.NewBadRequestException(err)
				}
				session.CurrentPlayOrderIndex = newPos
			}
		}
	case commandSetCurrentIndex:
		if request.Payload.Index == nil {
			return exceptions.NewBadRequestException(errors.New("payload.index is required for set_current_index"))
		}

		if err := setCurrentIndex(session, *request.Payload.Index); err != nil {
			return exceptions.NewBadRequestException(err)
		}
	default:
		return exceptions.NewBadRequestException(fmt.Errorf("unsupported command %s", request.Command))
	}

	return nil
}

func newPlaybackService(ctx context.Context) (PlaybackApi, error) {
	repo, err := newPlaybackRepository(ctx)
	if err != nil {
		return nil, err
	}

	mediaService, err := media.NewMediaService()
	if err != nil {
		return nil, err
	}

	return &playbackService{repo: repo, mediaService: mediaService, nowFn: time.Now}, nil
}
