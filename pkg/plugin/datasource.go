package plugin

import (
	"cloud.google.com/go/firestore"
	vkit "cloud.google.com/go/firestore/apiv1"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"time"
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

type FirestoreQueryCondition struct {
	Path      string
	Operator  string
	Value     string
	ValueType string
}

type FirestoreQueryOrderBy struct {
	Path      string
	Direction firestore.Direction
}

type FirestoreQuery struct {
	CollectionPath string
	Select         []string
	Where          []FirestoreQueryCondition
	OrderBy        []FirestoreQueryOrderBy
	Limit          json.Number
	IsCount        bool
}

type FirestoreSettings struct {
	ProjectId string
}

func (d *Datasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse

	// Unmarshal the JSON into our queryModel.
	var qm FirestoreQuery
	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, "json unmarshal: "+err.Error())
	}
	log.DefaultLogger.Debug("FirestoreQuery: ", qm)

	fsClient, err := newFirestoreClient(ctx, pCtx)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, "Invalid data source configuration: "+err.Error())
	}
	defer fsClient.Close()

	if len(qm.CollectionPath) > 0 {

		q := fsClient.Collection(qm.CollectionPath).Query

		if len(qm.Select) > 0 {
			q = q.Select(qm.Select...)
		}

		for _, condition := range qm.Where {
			q = q.Where(condition.Path, condition.Operator, condition.Value)
		}
		for _, orderBy := range qm.OrderBy {
			q = q.OrderBy(orderBy.Path, orderBy.Direction)
		}
		limit, err := qm.Limit.Int64()
		if err != nil {
			return backend.ErrDataResponse(backend.StatusBadRequest, "qm.Limit.Int64: "+err.Error())
		}
		if limit > 0 {
			q = q.Limit(int(limit))
		}

		log.DefaultLogger.Debug("Query ready!")
		documentItr := q.Documents(ctx)
		if err != nil {
			return backend.ErrDataResponse(backend.StatusBadRequest, "Query.Documents.GetAll : "+err.Error())
		}

		fieldValues := make(map[string]interface{})

		for {
			document, err := documentItr.Next()
			if err == iterator.Done {
				break
			}
			//values["Id"] = append(values["Id"], document.Ref.ID)
			data := document.Data()
			for key, val := range data {
				values, ok := fieldValues[key]
				if !ok {
					switch val.(type) {
					case bool:
						values = []bool{}
						break
					case float64:
						values = []float64{}
						break
					case time.Time:
						values = []time.Time{}
						break
					case map[string]interface{}, []map[string]interface{}, []interface{}:
						values = []json.RawMessage{}
					default:
						values = []string{}
					}
					fieldValues[key] = values
				}
				switch val.(type) {
				case bool:
					fieldValues[key] = append(values.([]bool), val.(bool))
					break
				case float64:
					fieldValues[key] = append(values.([]float64), val.(float64))
					break
				case time.Time:
					fieldValues[key] = append(values.([]time.Time), val.(time.Time))
					break
				case map[string]interface{}, []map[string]interface{}, []interface{}:
					jsonVal, err := json.Marshal(val)
					if err != nil {
						return backend.ErrDataResponse(backend.StatusBadRequest, "json.Marshal : "+key+err.Error())
					} else {
						fieldValues[key] = append(values.([]json.RawMessage), json.RawMessage(jsonVal))
					}
					break
				default:
					fieldValues[key] = append(values.([]string), fmt.Sprintf("%v", val))
				}
			}
		}

		// create data frame response.
		frame := data.NewFrame("response")
		for field, values := range fieldValues {
			frame.Fields = append(frame.Fields,
				data.NewField(field, nil, values),
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

	serviceAccount := pCtx.DataSourceInstanceSettings.DecryptedSecureJSONData["serviceAccount"]

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
	client, err := firestore.NewClient(ctx, settings.ProjectId, option.WithCredentials(creds))
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
		if err == nil {
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
