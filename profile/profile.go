package profile

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/yvasiyarov/playprof-client/parser"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type Profile struct {
	Metrics  *Metrics
	Resolver *Resolver
}

func NewProfile() *Profile {
	return &Profile{
		Metrics:  NewMetrics(),
		Resolver: NewResolver(),
	}
}

func (prof *Profile) ProfileByUrl(sourceUrl string, appId int64) error {

	if err := prof.loadMetricsByUrl(sourceUrl); err != nil {
		return err
	}

	sourceUrlParsed, err := url.Parse(sourceUrl)
	if err != nil {
		return err
	}

	symbolsUrl := sourceUrlParsed
	symbolsUrl.Path = "/debug/pprof/symbol"

	if symbolsData, err := prof.loadSymbolsByUrl(symbolsUrl.String()); err != nil {
		return err
	} else if err = prof.Resolver.LoadSymbols(symbolsData); err != nil {
		return err
	}

	return prof.sendProfile(appId)
}

func (prof *Profile) loadMetricsByUrl(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	switch string(data[:4]) {
	case "heap":
		err = prof.loadHeapMetrics(data)
	default:
		err = prof.loadCpuMetrics(data)
	}
	return err
}

func (prof *Profile) loadSymbolsByUrl(url string) ([]byte, error) {
	buf := new(bytes.Buffer)
	for i, addr := range prof.Metrics.Symbols() {
		if i > 0 {
			buf.WriteByte('+')
		}
		fmt.Fprintf(buf, "%#x", addr)
	}

	resp, err := http.Post(url, "text/plain", buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (prof *Profile) loadHeapMetrics(data []byte) error {
	p, err := parser.NewHeapProfParser(bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	for {
		rec, err := p.ReadRecord()
		if err == io.EOF {
			break
		}
		p.AdjustRecord(&rec,
			func(uint64) string { return "" })
		if err = prof.Metrics.Add(rec.Trace, rec.LiveObj, rec.LiveBytes, rec.AllocObj, rec.AllocBytes); err != nil {
			return err
		}
	}
	return nil
}

func (prof *Profile) loadCpuMetrics(data []byte) error {
	p, err := parser.NewCpuProfParser(bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	for {
		trace, count, err := p.ReadTrace()
		if trace == nil && err == io.EOF {
			break
		}
		if err = prof.Metrics.Add(trace, int64(count)); err != nil {
			return err
		}
	}
	return nil
}

func (prof *Profile) sendProfile(appId int64) error {
	data, err := prof.Serialise()
	if err != nil {
		return err
	}

	postValues := url.Values{
		"profile": []string{string(data)},
	}

	resp, err := http.PostForm(fmt.Sprintf("http://localhost:8080/api/%d/upload", appId), postValues)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Printf("Resp: %s\n", string(body))
	return nil
}

func (prof *Profile) Serialise() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(prof)
	return buf.Bytes(), err
}

func (prof *Profile) Unserialise(source []byte) error {
	buf := bytes.NewBuffer(source)
	enc := gob.NewDecoder(buf)

	return enc.Decode(prof)
}
