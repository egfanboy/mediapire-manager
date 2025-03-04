package media_update

import (
	"fmt"
	"os"
	"path"

	"github.com/egfanboy/mediapire-media-host/pkg/types"
	mhTypes "github.com/egfanboy/mediapire-media-host/pkg/types"
	"github.com/rs/zerolog/log"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

const (
	keyVideoMetadata = "metadata:s:v"
	keyMetadata      = "metadata"
	keyMap           = "map"
	// c is for copy
	keyC = "c"
	// id3v2_version for mp3 metadata
	keyId3V2Version   = "id3v2_version"
	Id3V2VersionValue = "3"
)

type errorWriter struct {
	lastWrite []byte
}

// ffmpeg-go sends you all output regardless if it is an error or not.
// Just track the last write so we can use it if there is an error
func (w *errorWriter) Write(p []byte) (n int, err error) {
	w.lastWrite = p
	return len(p), nil
}

func cleanupUpdate(builder UpdateBuilder) {
	inputPath := getInputPath(builder.Item())

	err := os.Remove(inputPath)
	if err != nil && !os.IsNotExist(err) {
		log.Err(err).Msgf("failed to cleanup temporary input file for update %s", builder.Item().Id)
	}

	outputPath := getOutputPath(builder.Item())
	err = os.Remove(outputPath)
	if err != nil && !os.IsNotExist(err) {
		log.Err(err).Msgf("failed to cleanup temporary output file for update %s", builder.Item().Id)
	}
}

func getOutputPath(item mhTypes.MediaItemWithContent) string {
	return path.Join("/tmp", fmt.Sprintf("%s-temp-out.%s", item.Id, item.Extension))
}

func getInputPath(item mhTypes.MediaItemWithContent) string {
	return path.Join("/tmp", fmt.Sprintf("%s-temp-in.%s", item.Id, item.Extension))
}

func defaultGetInputsImpl(u *baseMediaUpdater) []*ffmpeg_go.Stream {
	result := []*ffmpeg_go.Stream{ffmpeg_go.Input(getInputPath(u.media))}
	if u.imagePath != nil {
		result = append(result, ffmpeg_go.Input(*u.imagePath))
	}
	return result
}

func mp3GetInputsImpl(u *baseMediaUpdater) []*ffmpeg_go.Stream {
	mp3FileStream := ffmpeg_go.Input(getInputPath(u.media))

	// If we want to change the image we need to take ONLY the audio portion of the mp3 file
	// if we don't ffmpeg will just append another album art stream in the metadata
	if u.imagePath != nil {
		return []*ffmpeg_go.Stream{mp3FileStream.Audio(), ffmpeg_go.Input(*u.imagePath)}
	}
	return []*ffmpeg_go.Stream{mp3FileStream}
}

type UpdateBuilder interface {
	GetInputs() []*ffmpeg_go.Stream
	BuildArgs() (ffmpeg_go.KwArgs, error)
	Item() types.MediaItemWithContent
}

type BaseUpdater interface {
	UpdateBuilder

	Title(title string) BaseUpdater
	Album(album string) BaseUpdater
	Artist(artist string) BaseUpdater
	Comment(comment string) BaseUpdater
	Genre(genre string) BaseUpdater
	TrackNumber(trackNumber int) BaseUpdater
	Art(imagePath string) BaseUpdater
}

type baseMediaUpdater struct {
	baseMetadata map[string]interface{}
	media        types.MediaItemWithContent

	imagePath *string
	metadata  []string
	// ffmpeg takes metadata for a video stream
	// ie: -metadata:s:v. Used in MP3 files to set the cover art by taking the input video stream as the source
	videoMetadata []string

	// custom functions based on media type
	getInputStreamsImpl func(self *baseMediaUpdater) []*ffmpeg_go.Stream
}

func (u *baseMediaUpdater) Title(title string) BaseUpdater {
	u.metadata = append(u.metadata, fmt.Sprintf(`title=%s`, title))

	return u
}

func (u *baseMediaUpdater) Artist(artist string) BaseUpdater {
	u.metadata = append(u.metadata, fmt.Sprintf(`artist=%s`, artist))

	return u
}

func (u *baseMediaUpdater) Album(album string) BaseUpdater {
	u.metadata = append(u.metadata, fmt.Sprintf(`album=%s`, album))

	return u
}

func (u *baseMediaUpdater) Comment(comment string) BaseUpdater {
	u.metadata = append(u.metadata, fmt.Sprintf(`comment=%s`, comment))

	return u
}

func (u *baseMediaUpdater) Genre(genre string) BaseUpdater {
	u.metadata = append(u.metadata, fmt.Sprintf(`genre=%s`, genre))

	return u
}

func (u *baseMediaUpdater) TrackNumber(trackNumber int) BaseUpdater {
	if u.media.Extension == "mp3" {
		log.Warn().Msgf("Item %s does not support updating the track number", u.media.Id)
	} else {
		u.metadata = append(u.metadata, fmt.Sprintf("track=%d", trackNumber))
	}

	return u
}

func (u *baseMediaUpdater) Art(imagePath string) BaseUpdater {
	if u.media.Extension == "mp3" {
		log.Warn().Msgf("Item %s does not support updating the album cover", u.media.Id)
	} else {
		u.imagePath = &imagePath
		u.videoMetadata = append(u.videoMetadata, `title=Album cover`)
		// // Note: for now only support the front cover metadata
		u.videoMetadata = append(u.videoMetadata, `comment=Cover (front)`)
	}

	return u
}

func (u *baseMediaUpdater) BuildArgs() (ffmpeg_go.KwArgs, error) {
	ffmpegArgs := ffmpeg_go.KwArgs{}

	for k, v := range u.baseMetadata {
		ffmpegArgs[k] = v
	}

	if len(u.videoMetadata) > 0 {
		ffmpegArgs[keyVideoMetadata] = u.videoMetadata
	}

	if len(u.metadata) > 0 {
		ffmpegArgs[keyMetadata] = u.metadata
	}

	return ffmpegArgs, nil
}

func (u *baseMediaUpdater) Item() mhTypes.MediaItemWithContent {
	return u.media
}

func (u *baseMediaUpdater) GetInputs() []*ffmpeg_go.Stream {
	return u.getInputStreamsImpl(u)
}

func UpdateMedia(builder UpdateBuilder) ([]byte, error) {
	inputPath := getInputPath(builder.Item())

	// ffmpeg needs the input to be actual files so write a temp file with the media content
	err := os.WriteFile(inputPath, builder.Item().Content, 0666)
	if err != nil {
		return nil, err
	}

	defer cleanupUpdate(builder)

	ffmpegArgs, err := builder.BuildArgs()
	if err != nil {
		return nil, err
	}

	outputPath := getOutputPath(builder.Item())
	w := &errorWriter{}
	err = ffmpeg_go.Output(builder.GetInputs(), outputPath, ffmpegArgs).
		OverWriteOutput().
		WithErrorOutput(w).
		Silent(true).
		Run()
	if err != nil {
		return nil, fmt.Errorf("err: %w. %s", err, string(w.lastWrite))
	}

	return os.ReadFile(outputPath)
}

func GetBuilder(media types.MediaItemWithContent) BaseUpdater {
	updater := &baseMediaUpdater{
		media: media,
		baseMetadata: map[string]interface{}{
			keyC: "copy",
		},
		getInputStreamsImpl: defaultGetInputsImpl,
	}

	if media.Extension == "mp3" {
		updater.getInputStreamsImpl = mp3GetInputsImpl
		// add base metadata related to mp3 only
		updater.baseMetadata[keyId3V2Version] = Id3V2VersionValue
	}

	return updater
}
