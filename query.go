package salesforce

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
)

type queryResponse struct {
	TotalSize      int              `json:"totalSize"`
	Done           bool             `json:"done"`
	NextRecordsUrl string           `json:"nextRecordsUrl"`
	Records        []map[string]any `json:"records"`
}

func performQuery(auth *authentication, query string, sObject any) error {
	query = url.QueryEscape(query)
	queryResp := &queryResponse{
		Done:           false,
		NextRecordsUrl: "/query/?q=" + query,
	}

	for !queryResp.Done {
		resp, err := doRequest(auth, requestPayload{
			method:  http.MethodGet,
			uri:     queryResp.NextRecordsUrl,
			content: jsonType,
		})
		if err != nil {
			return err
		}

		respBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return readErr
		}

		tempQueryResp := &queryResponse{}
		queryResponseError := json.Unmarshal(respBody, &tempQueryResp)
		if queryResponseError != nil {
			return queryResponseError
		}

		queryResp.TotalSize = queryResp.TotalSize + tempQueryResp.TotalSize
		queryResp.Records = append(queryResp.Records, tempQueryResp.Records...)
		queryResp.Done = tempQueryResp.Done
		if !tempQueryResp.Done && tempQueryResp.NextRecordsUrl != "" {
			queryResp.NextRecordsUrl = strings.TrimPrefix(tempQueryResp.NextRecordsUrl, "/services/data/"+apiVersion)
		}
	}

	sObjectError := decode(queryResp.Records, sObject)
	if sObjectError != nil {
		return sObjectError
	}

	return nil
}

func ToTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		switch f.Kind() {
		case reflect.String:
			return time.Parse("2006-01-02T15:04:05.000-0700", data.(string))
		case reflect.Float64:
			return time.Unix(0, int64(data.(float64))*int64(time.Millisecond)), nil
		case reflect.Int64:
			return time.Unix(0, data.(int64)*int64(time.Millisecond)), nil
		default:
			return data, nil
		}
		// Convert it by parsing
	}
}

func decode(input interface{}, result interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			ToTimeHookFunc()),
		Result: result,
	})
	if err != nil {
		return err
	}

	if err := decoder.Decode(input); err != nil {
		return err
	}
	return err
}
