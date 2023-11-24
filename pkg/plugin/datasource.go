package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"cloud.google.com/go/firestore"
	vkit "cloud.google.com/go/firestore/apiv1"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/pgollangi/fireql"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces- only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// NewDatasource creates a new datasource instance.
func NewDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &Datasource{}, nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct{}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// when logging at a non-Debug level, make sure you don't include sensitive information in the message
	// (like the *backend.QueryDataRequest)
	log.DefaultLogger.Debug("QueryData called", "numQueries", len(req.Queries))

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type FirestoreQuery struct {
	Query string
}

type FirestoreSettings struct {
	ProjectId string
}

func (d *Datasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) (response backend.DataResponse) {
	defer func() {
		if err := recover(); err != nil {
			log.DefaultLogger.Error("panic occurred ", err)
			response = backend.ErrDataResponse(backend.StatusInternal, "internal server error")
		}
	}()
	response = d.queryInternal(ctx, pCtx, query)
	return response
}

func convertInt32To64(ar []int32) []float64 {
	newar := make([]float64, len(ar))
	var v int32
	var i int
	for i, v = range ar {
		newar[i] = float64(v)
	}
	return newar
}

func convertInt64To64(ar []int64) []float64 {
	newar := make([]float64, len(ar))
	var v int64
	var i int
	for i, v = range ar {
		newar[i] = float64(v)
	}
	return newar
}

func (d *Datasource) queryInternal(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse

	// Unmarshal the JSON into our queryModel.
	var qm FirestoreQuery
	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, "json unmarshal: "+err.Error())
	}
	log.DefaultLogger.Debug("FirestoreQuery: ", qm)

	var settings FirestoreSettings
	err = json.Unmarshal(pCtx.DataSourceInstanceSettings.JSONData, &settings)
	if err != nil {
		log.DefaultLogger.Error("Error parsing settings ", err)
		return backend.ErrDataResponse(backend.StatusBadRequest, "ProjectID: "+err.Error())
	}

	if len(settings.ProjectId) == 0 {
		return backend.ErrDataResponse(backend.StatusBadRequest, "ProjectID is required")
	}

	var options []fireql.Option
	if pCtx.DataSourceInstanceSettings.DecryptedSecureJSONData["serviceAccount"] != "" {
		options = append(options, fireql.OptionServiceAccount(pCtx.DataSourceInstanceSettings.DecryptedSecureJSONData["serviceAccount"]))
	}

	fQuery, err := fireql.New(settings.ProjectId, options...)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, "fireql.NewFireQL: "+err.Error())
	}

	log.DefaultLogger.Info("Created fireql.NewFireQLWithServiceAccountJSON")

	if len(qm.Query) > 0 {

		log.DefaultLogger.Info("Executing query", qm.Query)
		result, err := fQuery.Execute(qm.Query)
		if err != nil {
			return backend.ErrDataResponse(backend.StatusBadRequest, "fireql.Execute: "+err.Error())
		}

		fieldValues := make(map[string]interface{})

		for idx, column := range result.Columns {
			var values interface{}
			if len(result.Records) > 0 {
				for _, record := range result.Records {
					val := record[idx]
					switch val.(type) {
					case bool:
						if values == nil {
							values = []bool{}
						}
						values = append(values.([]bool), val.(bool))
						break
					case int:
						if values == nil {
							values = []int32{}
						}
						if reflect.TypeOf(values).Elem().Kind() == reflect.Float64 {
							values = append(values.([]float64), val.(float64))
						} else {
							values = append(values.([]int32), int32(val.(int)))
						}
						break
					case int32:
						if values == nil {
							values = []int32{}
						}
						if reflect.TypeOf(values).Elem().Kind() == reflect.Float64 {
							values = append(values.([]float64), val.(float64))
						} else {
							values = append(values.([]int32), val.(int32))
						}
						break
					case int64:
						if values == nil {
							values = []int64{}
						}
						if reflect.TypeOf(values).Elem().Kind() == reflect.Float64 {
							values = append(values.([]float64), val.(float64))
						} else {
							values = append(values.([]int64), val.(int64))
						}
						break
					case float64:
						if values == nil {
							values = []float64{}
						}
						if reflect.TypeOf(values).Elem().Kind() == reflect.Int {
							values = convertInt32To64(values.([]int32))
						}
						if reflect.TypeOf(values).Elem().Kind() == reflect.Int32 {
							values = convertInt32To64(values.([]int32))
						}
						if reflect.TypeOf(values).Elem().Kind() == reflect.Int64 {
							values = convertInt64To64(values.([]int64))
						}
						values = append(values.([]float64), val.(float64))
						break
					case time.Time:
						if values == nil {
							values = []time.Time{}
						}
						values = append(values.([]time.Time), val.(time.Time))
						break
					case map[string]interface{}, []map[string]interface{}, []interface{}:
						if values == nil {
							values = []json.RawMessage{}
						}
						jsonVal, err := json.Marshal(val)
						if err != nil {
							return backend.ErrDataResponse(backend.StatusBadRequest, "json.Marshal : "+column+err.Error())
						} else {
							values = append(values.([]json.RawMessage), json.RawMessage(jsonVal))
						}
						break
					default:
						if values == nil {
							values = []string{}
						}
						if val == nil {
							if reflect.TypeOf(values).Elem().Kind() == reflect.Int {
								values = append(values.([]int32), 0)
							}
							if reflect.TypeOf(values).Elem().Kind() == reflect.Int32 {
								values = append(values.([]int32), 0)
							}
							if reflect.TypeOf(values).Elem().Kind() == reflect.Int64 {
								values = append(values.([]int64), 0)
							}
							if reflect.TypeOf(values).Elem().Kind() == reflect.Float64 {
								values = append(values.([]float64), 0.0)
							}
						}
						values = append(values.([]string), fmt.Sprintf("%v", val))
					}
				}
			} else {
				values = []string{}
			}
			fieldValues[column] = values
		}

		// create data frame response.
		frame := data.NewFrame("response")
		for _, column := range result.Columns {
			frame.Fields = append(frame.Fields,
				data.NewField(column, nil, fieldValues[column]),
			)
		}
		// add the frames to the response.
		response.Frames = append(response.Frames, frame)
	}

	return response
}

func newFirestoreClient(ctx context.Context, pCtx backend.PluginContext) (*firestore.Client, error) {
	var settings FirestoreSettings
	err := json.Unmarshal(pCtx.DataSourceInstanceSettings.JSONData, &settings)
	if err != nil {
		log.DefaultLogger.Error("Error parsing settings ", err)
		return nil, fmt.Errorf("ProjectID: %v", err)
	}

	if len(settings.ProjectId) == 0 {
		return nil, errors.New("project Id is required")
	}

	var options []option.ClientOption
	serviceAccount := pCtx.DataSourceInstanceSettings.DecryptedSecureJSONData["serviceAccount"]

	if len(serviceAccount) > 0 {
		if !json.Valid([]byte(serviceAccount)) {
			return nil, errors.New("invalid service account, it is expected to be a JSON")
		}
		creds, err := google.CredentialsFromJSON(ctx, []byte(serviceAccount),
			vkit.DefaultAuthScopes()...,
		)
		if err != nil {
			log.DefaultLogger.Error("google.CredentialsFromJSON ", err)
			return nil, fmt.Errorf("ServiceAccount: %v", err)
		}
		options = append(options, option.WithCredentials(creds))
	}
	client, err := firestore.NewClient(ctx, settings.ProjectId, options...)
	if err != nil {
		log.DefaultLogger.Error("firestore.NewClient ", err)
		return nil, fmt.Errorf("firestore.NewClient: %v", err)
	}
	return client, nil
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	// when logging at a non-Debug level, make sure you don't include sensitive information in the message
	// (like the *backend.QueryDataRequest)
	log.DefaultLogger.Debug("CheckHealth called")

	var status = backend.HealthStatusOk
	var message = "Data source is working"

	client, healthErr := newFirestoreClient(ctx, req.PluginContext)

	if healthErr == nil {
		defer client.Close()
		collections := client.Collections(ctx)
		collection, err := collections.Next()
		if err == nil || errors.Is(err, iterator.Done) {
			log.DefaultLogger.Debug("First collections: ", collection.ID)
		} else {
			log.DefaultLogger.Error("client.Collections ", err)
			healthErr = fmt.Errorf("firestore.Collections: %v", err)
		}
	}

	if healthErr != nil {
		status = backend.HealthStatusError
		message = healthErr.Error()
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}
