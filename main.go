package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/karlseguin/ccache/v2"
)

var cache = ccache.New(ccache.Configure().MaxSize(1000).ItemsToPrune(100))
var resultLimit = 10

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("Listening on port %s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

type Searcher struct {
	CompleteWorks string
	SuffixArray   *suffixarray.Index
}

type APIResponse struct {
	Result      []string `json:"results"`
	ResultCount int      `json:"resultCount"`
	Limit       int      `json:"resultLimit"`
	Page        int      `json:"page"`
	TotalPages  int      `json:"totalPages"`
}

func paginate(x []string, skip int, size int) []string {
	if skip > len(x) {
		skip = len(x)
	}

	end := skip + size
	if end > len(x) {
		end = len(x)
	}

	return x[skip:end]
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}
		var err error
		pageStr := r.URL.Query().Get("page")
		page := 1
		if pageStr != "" {
			page, err = strconv.Atoi(pageStr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Println("Erorr", err)
				fmt.Fprintf(w, "unable to parse page number")
				return
			}
		}

		querySearch := query[0]
		cacheData, err := cache.Fetch(querySearch, time.Minute*10, func() (interface{}, error) {
			return searcher.Search(querySearch)
		})
		results := cacheData.Value().([]string)
		resultLen := len(results)
		totalPages := resultLen / resultLimit
		res := APIResponse{
			Result:      paginate(results, page, resultLimit),
			ResultCount: resultLen,
			Limit:       resultLimit,
			Page:        page,
			TotalPages:  totalPages,
		}

		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err = enc.Encode(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

func (s *Searcher) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	s.CompleteWorks = string(dat)
	s.SuffixArray = suffixarray.New(dat)
	return nil
}

func (s *Searcher) Search(query string) ([]string, error) {
	reg, err := regexp.Compile(fmt.Sprintf("(?i).*%s.*", query))
	if err != nil {
		return []string{}, err
	}

	idxs := s.SuffixArray.FindAllIndex(reg, -1)
	results := []string{}
	for _, idx := range idxs {
		results = append(results, s.CompleteWorks[idx[0]-250:idx[1]+250])
	}
	return results, nil
}
