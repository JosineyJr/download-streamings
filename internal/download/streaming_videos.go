package download

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/grafov/m3u8"
	"github.com/rs/zerolog/log"

	"github.com/mitchellh/colorstring"
	"github.com/schollz/progressbar/v3"
)

type Downloader interface {
	DownloadVideo(id, title string) (io.Reader, error)
}

type M3U8 struct {
	authority, referer, origin, userAgent string
}

type DownloadOptions string

const (
	AllCourse     DownloadOptions = "All course"
	AllModule     DownloadOptions = "All module"
	SpecificVideo DownloadOptions = "Specific video"
)

func NewM3U8(authority, referer, origin string) M3U8 {
	return M3U8{
		authority: authority,
		referer:   referer,
		origin:    origin,
		userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36...",
	}
}

func (d M3U8) DownloadVideo(videoID, title string) (r io.Reader, err error) {
	client := &http.Client{}
	url := d.authority + "/" + videoID
	playlistBody, err := d.fetchMasterPlaylist(client, url+"/playlist.m3u8")
	if err != nil {
		return
	}

	masterPlaylist, _, err := m3u8.DecodeFrom(bytes.NewReader(playlistBody), true)
	if err != nil {
		return nil, err
	}

	var bestVariant *m3u8.Variant
	maxBandwidth := uint32(0)
	for _, v := range masterPlaylist.(*m3u8.MasterPlaylist).Variants {
		if v.VariantParams.Bandwidth > maxBandwidth {
			bestVariant = v
			maxBandwidth = v.VariantParams.Bandwidth
		}
	}

	variantPlaylist, err := d.fetchVariantPlaylist(client, url, bestVariant.URI)
	if err != nil {
		return
	}

	segments := make([]string, 0, len(variantPlaylist.Segments))
	for _, seg := range variantPlaylist.Segments {
		if seg != nil {
			segments = append(segments, seg.URI)
		}
	}
	url += "/" + strings.Split(bestVariant.URI, "/")[0]

	bar := progressbar.NewOptions(
		len(segments),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        colorstring.Color("[green]█"),
			SaucerPadding: "",
			BarStart:      "|",
			BarEnd:        "|",
		}),
		progressbar.OptionOnCompletion(func() {
			fmt.Println("")
			log.Info().Str("video", title).Str("message", "completed").Send()
		}),
	)

	var buffer bytes.Buffer
	for _, segURL := range segments {
		segURL = url + "/" + segURL
		req, _ := http.NewRequest(
			"GET",
			segURL,
			nil,
		)
		d.addHeaders(req)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		_, err = io.Copy(&buffer, resp.Body)
		if err != nil {
			return nil, err
		}

		err = bar.Add(1)
		if err != nil {
			return nil, err
		}
	}
	r = &buffer

	return
}

func (d M3U8) fetchMasterPlaylist(client *http.Client, playlistURL string) ([]byte, error) {
	req, _ := http.NewRequest("GET", playlistURL, nil)
	d.addHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (d M3U8) fetchVariantPlaylist(
	client *http.Client,
	videoURL, variantURL string,
) (*m3u8.MediaPlaylist, error) {
	req, _ := http.NewRequest(
		"GET",
		videoURL+"/"+variantURL,
		nil,
	)
	d.addHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	playlist, _, err := m3u8.DecodeFrom(resp.Body, true)
	return playlist.(*m3u8.MediaPlaylist), err
}

func (d M3U8) addHeaders(req *http.Request) {
	req.Header.Set("referer", d.referer)
	req.Header.Set("origin", d.origin)
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("dnt", "1")
	req.Header.Set(
		"User-Agent",
		d.userAgent,
	)
}
