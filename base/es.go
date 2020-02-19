package base

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func es1(es *elasticsearch.Client) {
	res, err := es.Index(
		"test",                                  // Index name
		strings.NewReader(`{"title" : "Test"}`), // Document body
		es.Index.WithDocumentID("1"),            // Document ID
		es.Index.WithRefresh("true"),            // Refresh
	)
	if err != nil {
		log.Fatal().Msgf("ERROR: %s", err)
	}
	defer res.Body.Close()

	log.Print(res)
}

func testElasticsearch() {
	es, _ := elasticsearch.NewDefaultClient()
	info, err := es.Info()
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	log.Print(info)
	r := map[string]interface{}{}

	if err := json.NewDecoder(info.Body).Decode(&r); err != nil {
		log.Error().Msg(err.Error())
	}

	var wg sync.WaitGroup

	for k, value := range r {

		log.Debug().Msg(k)

		wg.Add(1)

		go func(k string, value interface{}) {
			defer wg.Done()
			v1, ok := value.(string)
			if ok {
				log.Debug().Str("key", k).Msg(v1)
			}
			req := esapi.IndexRequest{
				Index: "test",
			}
			res, err := req.Do(context.Background(), es)
			if err != nil {
				//log.Fatalf("Error getting response: %s", err)
				log.Error().Msg(err.Error())
			}
			defer res.Body.Close()
			if res.IsError() {
				log.Print(114, res.Status())
				return
			}
			log.Print(res.Body)

		}(k, value)
	}

	wg.Wait()

	es1(es)

	b, err := json.Marshal(r)

	log.Debug().Msg(string(b))

}
