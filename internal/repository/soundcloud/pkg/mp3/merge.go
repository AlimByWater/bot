package mp3

import (
	"bytes"
	"elysium/internal/repository/soundcloud/pkg/decrypt"
	"elysium/internal/repository/soundcloud/pkg/joiner"
	"elysium/internal/repository/soundcloud/pkg/pool"
	"elysium/internal/repository/soundcloud/pkg/zhttp"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/grafov/m3u8"
)

type Module struct {
	Zhttp        *zhttp.Zhttp
	Joiner       *joiner.Joiner
	keyCache     map[string][]byte
	keyCacheLock sync.Mutex
}

func NewModule(proxyUrl string, downloadPath, songName string) (*Module, error) {
	z, err := zhttp.New(30*time.Second, proxyUrl)
	if err != nil {
		return nil, fmt.Errorf("zhttp new: %w", err)
	}

	//outFile := downloadPath + strings.Replace(songName, "/", "\\", -1) + ".mp3"
	outFile := downloadPath + strconv.FormatInt(time.Now().Unix(), 10) + ".mp3"
	j, err := joiner.New(outFile)
	if err != nil {
		return nil, fmt.Errorf("joiner new: %w", err)
	}

	return &Module{
		Zhttp:        z,
		Joiner:       j,
		keyCache:     make(map[string][]byte),
		keyCacheLock: sync.Mutex{},
	}, nil

}

// TODO implement tests
func (m *Module) start(mpl *m3u8.MediaPlaylist) {
	// 30 go routines for now
	// TODO: find optimal go routine amount
	p := pool.New(100, m.download)

	go func() {
		var count = int(mpl.Count())
		for i := 0; i < count; i++ {
			p.Push([]interface{}{i, mpl.Segments[i], mpl.Key})
		}
		p.CloseQueue()
	}()

	go p.Run()
}

// TODO implement tests
func (m *Module) parseM3u8(m3u8Url string) (*m3u8.MediaPlaylist, error) {
	statusCode, data, err := m.Zhttp.Get(m3u8Url)
	if err != nil {
		return nil, err
	}

	if statusCode/100 != 2 || len(data) == 0 {
		return nil, fmt.Errorf("download m3u8 file failed, http code: " + strconv.Itoa(statusCode))
	}

	playlist, listType, err := m3u8.Decode(*bytes.NewBuffer(data), true)
	if err != nil {
		return nil, err
	}

	if listType == m3u8.MEDIA {
		obj, _ := url.Parse(m3u8Url)
		mpl := playlist.(*m3u8.MediaPlaylist)

		if mpl.Key != nil && mpl.Key.URI != "" {
			uri, err := m.formatURI(obj, mpl.Key.URI)
			if err != nil {
				return nil, err
			}
			mpl.Key.URI = uri
		}

		count := int(mpl.Count())
		for i := 0; i < count; i++ {
			segment := mpl.Segments[i]

			uri, err := m.formatURI(obj, segment.URI)
			if err != nil {
				return nil, err
			}
			segment.URI = uri

			if segment.Key != nil && segment.Key.URI != "" {
				uri, err := m.formatURI(obj, segment.Key.URI)
				if err != nil {
					return nil, err
				}
				segment.Key.URI = uri
			}

			mpl.Segments[i] = segment
		}

		return mpl, nil
	}

	return nil, errors.New("Unsupport m3u8 type")
}

// TODO implement tests
func (m *Module) getKey(url string) ([]byte, error) {
	m.keyCacheLock.Lock()
	defer m.keyCacheLock.Unlock()

	key := m.keyCache[url]
	if key != nil {
		return key, nil
	}

	statusCode, key, err := m.Zhttp.Get(url)
	if err != nil {
		return nil, err
	}

	if len(key) == 0 {
		return nil, errors.New("body is empty, http code: " + strconv.Itoa(statusCode))
	}

	m.keyCache[url] = key

	return key, nil
}

// TODO implement tests
func (m *Module) download(in interface{}) {
	params := in.([]interface{})
	id := params[0].(int)
	segment := params[1].(*m3u8.MediaSegment)
	globalKey := params[2].(*m3u8.Key)

	statusCode, data, err := m.Zhttp.Get(segment.URI)
	if err != nil {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Printf("%s Download failed: %s %s\n", red("[-]"), err, segment.URI)
	}

	if len(data) == 0 {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Printf("%s Download failed: body is empty, http code: %d\n", red("[-]"), statusCode)
	}

	var keyURL, ivStr string
	if segment.Key != nil && segment.Key.URI != "" {
		keyURL = segment.Key.URI
		ivStr = segment.Key.IV
	} else if globalKey != nil && globalKey.URI != "" {
		keyURL = globalKey.URI
		ivStr = globalKey.IV
	}

	if keyURL != "" {
		var key, iv []byte
		key, err = m.getKey(keyURL)
		if err != nil {
			fmt.Println("[-] Download key failed:", keyURL, err)
		}

		if ivStr != "" {
			iv, err = hex.DecodeString(strings.TrimPrefix(ivStr, "0x"))
			if err != nil {
				fmt.Println("[-] Decode iv failed:", err)
			}
		} else {
			iv = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, byte(id)}
		}

		data, err = decrypt.Decrypt(data, key, iv)
		if err != nil {
			red := color.New(color.FgRed).SprintFunc()
			fmt.Printf("%s Decrypt failed: %s\n", red("[-]"), err)
		}
	}

	m.Joiner.Join(id, data)
}

// TODO implement tests
func (m *Module) formatURI(base *url.URL, u string) (string, error) {
	if strings.HasPrefix(u, "http") {
		return u, nil
	}

	obj, err := base.Parse(u)
	if err != nil {
		return "", err
	}

	return obj.String(), nil
}

// Merge downloads and merges the mp3 files from the m3u8 file
func (m *Module) Merge(url string) (string, error) {
	mpl, err := m.parseM3u8(url)
	if err != nil {
		return "", fmt.Errorf("parse m3u8 file failed: %w", err)
	}

	if mpl.Count() > 0 {
		m.start(mpl)

		err = m.Joiner.Run(int(mpl.Count()))
		if err != nil {
			return "", fmt.Errorf("write to file failed: %w", err)
		}
	}

	return m.Joiner.Name(), nil
}
