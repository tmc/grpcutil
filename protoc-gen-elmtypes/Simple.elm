-- this is a generated file
module Simple exposing (..)
import Json.Encode as JE
import Json.Decode exposing (..)

type SearchRequestCorpus = UNIVERSAL | WEB | IMAGES | LOCAL | NEWS | PRODUCTS | VIDEO

type alias SearchRequest = {
  query: Maybe String,
  page_number: Maybe Int,
  result_per_page: Maybe Int,
  corpus: Maybe SearchRequestCorpus
}

type alias SearchResponse = {
  results: Maybe List String,
  num_results: Maybe Int,
  original_request: Maybe SearchRequest
}



(maybe ("SearchRequestCorpus" := UNIVERSAL | WEB | IMAGES | LOCAL | NEWS | PRODUCTS | VIDEO))

searchRequest = object4 SearchRequest (maybe ("query" := string)) (maybe ("page_number" := int)) (maybe ("result_per_page" := int)) (maybe ("corpus" := SearchRequestCorpus))

searchResponse = object3 SearchResponse (maybe ("results" := (list string))) (maybe ("num_results" := int)) (maybe ("original_request" := SearchRequest))


