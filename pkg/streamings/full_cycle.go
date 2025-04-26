package streamings

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type FullCycle struct {
	username, password string
	urls               map[string]string
	BearerToken        string `json:"token"`
	Curses             map[int]course
	Modules            []module
	Chapters           []chapter
}

type course struct {
	Category struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
	Classroom struct {
		ID int `json:"id"`
	} `json:"classroom"`
}

type module struct {
	ID   int    `json:"id"`
	Name string `json:"nome"`
	Desc string `json:"descricao"`
}

type chapter struct {
	ID       int       `json:"id"`
	Name     string    `json:"nome"`
	Contents []Content `json:"conteudos"`
}

type Content struct {
	Title   string `json:"titulo"`
	Type    uint8  `json:"tipo"`
	VideoID string `json:"video_url_bunny"`
	Order   int    `json:"ordem"`
}

func NewFullCycle(username, password string) (s FullCycle, err error) {
	s = FullCycle{
		username: username,
		password: password,
		urls: map[string]string{
			"login check":  "https://portal.fullcycle.com.br/api/login_check",
			"get curses":   "https://portal.fullcycle.com.br/api/cursos/my.json",
			"api origin":   "https://plataforma.fullcycle.com.br",
			"get modules":  "https://portal.fullcycle.com.br/api/cursos/turma/{Classroom}/categoria/{ID}/list.json",
			"get chapters": "https://portal.fullcycle.com.br/api/cursos/turma/{Classroom}/curso/{ID}/capitulos.json?expand_conteudos=1",
		},
	}

	err = s.getBearerToken()
	if err != nil {
		return
	}

	err = s.getCurses()

	return
}

func (s *FullCycle) getBearerToken() (err error) {
	// var requestBody bytes.Buffer
	// writer := multipart.NewWriter(&requestBody)
	// err = writer.WriteField("_username", s.username)
	// if err != nil {
	// 	return
	// }
	// err = writer.WriteField("_password", s.password)
	// if err != nil {
	// 	return
	// }
	// writer.Close()

	// req, err := http.NewRequest(
	// 	"POST",
	// 	s.urls["login check"],
	// 	&requestBody,
	// )
	// if err != nil {
	// 	fmt.Println("Error creating request:", err)
	// 	return
	// }

	// req.Header.Set("Content-Type", writer.FormDataContentType())

	// client := &http.Client{}
	// resp, err := client.Do(req)
	// if err != nil {
	// 	fmt.Println("Error sending request:", err)
	// 	return
	// }
	// defer resp.Body.Close()

	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return
	// }
	// err = json.Unmarshal(body, s)
	// if err != nil {
	// 	return
	// }

	// if s.BearerToken == "" {
	// 	return errors.ErrNotDefinedBearerToken
	// }

	s.BearerToken = "TOKEN GOES HERE"
	// s.BearerToken = "Bearer " + s.BearerToken
	return
}

func (s *FullCycle) getCurses() (err error) {
	req, err := http.NewRequest(
		"GET",
		s.urls["get curses"],
		nil,
	)
	if err != nil {
		return
	}

	s.addHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var curses []course
	err = json.Unmarshal(body, &curses)
	if err != nil {
		return
	}

	s.Curses = make(map[int]course, len(curses))
	for _, c := range curses {
		s.Curses[c.Category.ID] = c
	}

	return
}

func (s *FullCycle) GetModules(id int) (err error) {
	req, err := http.NewRequest(
		"GET",
		strings.Replace(
			strings.Replace(s.urls["get modules"], "{ID}", strconv.Itoa(id), 1),
			"{Classroom}",
			strconv.Itoa(s.Curses[id].Classroom.ID),
			1,
		),
		nil,
	)
	if err != nil {
		return
	}

	s.addHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &s.Modules)
	return
}

func (s *FullCycle) GetChapters(curseID, moduleID int) (err error) {
	req, err := http.NewRequest(
		"GET",
		strings.Replace(
			strings.Replace(s.urls["get chapters"], "{ID}", strconv.Itoa(moduleID), 1),
			"{Classroom}",
			strconv.Itoa(s.Curses[curseID].Classroom.ID),
			1,
		),
		nil,
	)
	if err != nil {
		return
	}

	s.addHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var cont struct {
		Content string `json:"content"`
	}

	err = json.Unmarshal(body, &cont)
	if err != nil {
		return
	}
	decodedData, err := base64.StdEncoding.DecodeString(cont.Content)
	if err != nil {
		return
	}

	decompressedData, err := s.decompressData(decodedData)
	if err != nil {
		return
	}

	err = json.Unmarshal(decompressedData, &s.Chapters)
	if err != nil {
		return
	}

	for i := range len(s.Chapters) {
		cValids := make([]Content, 0, len(s.Chapters[i].Contents))
		for j := range len(s.Chapters[i].Contents) {
			if s.Chapters[i].Contents[j].Type == 12 && s.Chapters[i].Contents[j].VideoID != "" {
				vID := strings.Split(s.Chapters[i].Contents[j].VideoID, "/")
				s.Chapters[i].Contents[j].VideoID = vID[len(vID)-1]
				cValids = append(cValids, s.Chapters[i].Contents[j])
			}
		}
		s.Chapters[i].Contents = cValids
	}

	return
}

func (s FullCycle) GetCourseSlice() (c []course) {
	c = make([]course, 0, len(s.Curses))
	for _, sc := range s.Curses {
		c = append(c, sc)
	}

	return
}

func (s FullCycle) addHeaders(req *http.Request) {
	req.Header.Set("Authorization", s.BearerToken)
	req.Header.Set("Origin", s.urls["api origin"])
	req.Header.Set("Referer", s.urls["api origin"])
	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36...",
	)
}

func (s FullCycle) decompressData(decodedData []byte) ([]byte, error) {
	var decompressedData bytes.Buffer
	gzipReader, err := gzip.NewReader(bytes.NewReader(decodedData))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(&decompressedData, gzipReader)
	if err != nil {
		return nil, err
	}

	err = gzipReader.Close()
	if err != nil {
		return nil, err
	}

	return decompressedData.Bytes(), err
}
