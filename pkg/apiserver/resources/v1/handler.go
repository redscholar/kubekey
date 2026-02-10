package v1

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/emicklei/go-restful/v3"
	capkkinfrav1beta1 "github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	"github.com/kubesphere/kubekey/v4/pkg/web/api"
	"github.com/kubesphere/kubekey/v4/pkg/web/query"
	"k8s.io/client-go/rest"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Handler handle web-installer resource apis
type Handler struct {
	workDir  string
	rootPath string
	client   ctrlclient.Client
	config   *rest.Config
}

// NewHandler create a new Handler
func NewHandler(client ctrlclient.Client, config *rest.Config, workDir, rootPath string) *Handler {
	return &Handler{
		workDir:  workDir,
		client:   client,
		config:   config,
		rootPath: rootPath,
	}
}

// GetSchemaSummary get resources summary data
// should return cluster summary and node summary
func (h *Handler) GetSchemaSummary(req *restful.Request, resp *restful.Response) {
	var kkClusterList capkkinfrav1beta1.KKClusterList
	if err := h.client.List(req.Request.Context(), &kkClusterList); err != nil {
		api.HandleBadRequest(resp, req, errors.Wrap(err, "failed to list kk clusters"))
		return
	}

	var kkMachineList capkkinfrav1beta1.KKMachineList
	if err := h.client.List(req.Request.Context(), &kkMachineList); err != nil {
		api.HandleBadRequest(resp, req, errors.Wrap(err, "failed to list kk machines"))
		return
	}

	var result = ResourcesSummaryResult{
		Clusters: ResourcesSummaryClusterResult{},
		Nodes:    ResourcesSummaryNodeResult{},
	}

	for _, item := range kkClusterList.Items {
		switch item.Status.Phase {
		case capkkinfrav1beta1.KKClusterPhaseNotInstall:
			result.Clusters.NotInstall++
		case capkkinfrav1beta1.KKClusterPhaseInitializing:
			result.Clusters.Initializing++
		case capkkinfrav1beta1.KKClusterPhaseRunning:
			result.Clusters.Running++
		case capkkinfrav1beta1.KKClusterPhaseUpgrading:
			result.Clusters.Upgrading++
		case capkkinfrav1beta1.KKClusterPhaseScaling:
			result.Clusters.Scaling++
		case v1beta1.StatusNodeError:
			result.Clusters.NodeError++
		case v1beta1.StatusUpgradeError:
			result.Clusters.UpgradeError++
		case v1beta1.StatusFailed:
			result.Clusters.Failed++
		}
	}

	for _, item := range kkMachineList.Items {
		switch item.Status.Status {
		case v1beta1.KKMachineStatusCreating:
			result.Nodes.Creating++
		case v1beta1.KKMachineStatusWarning:
			result.Nodes.Warning++
		case v1beta1.KKMachineStatusReady:
			result.Nodes.Ready++
		case v1beta1.KKMachineStatusRunning:
			result.Nodes.Running++
		case v1beta1.KKMachineStatusFault:
			result.Nodes.Fault++
		case v1beta1.KKMachineStatusUnschedulable:
			result.Nodes.Unschedulable++
		}
	}

	resp.WriteEntity(result)
}

// ListSchema lists all schema JSON files in the rootPath directory as a table.
// It supports filtering, sorting, and pagination via query parameters.
func (h *Handler) ListSchema(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	// Read all entries in the rootPath directory.
	entries, err := os.ReadDir(h.rootPath)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}
	schemaTable := make([]api.SchemaTable, 0)
	for _, entry := range entries {
		// Skip directories, non-JSON files, and special schema files.
		if entry.IsDir() ||
			!strings.HasSuffix(entry.Name(), ".json") ||
			entry.Name() == api.SchemaProductFile || entry.Name() == api.SchemaConfigFile {
			continue
		}
		// Read the JSON file.
		data, err := os.ReadFile(filepath.Join(h.rootPath, entry.Name()))
		if err != nil {
			api.HandleError(response, request, errors.Wrapf(err, "failed to read file for schema %q", entry.Name()))
			return
		}
		var schemaFile api.SchemaFile
		// Unmarshal the JSON data into a SchemaTable struct.
		if err := json.Unmarshal(data, &schemaFile); err != nil {
			api.HandleError(response, request, errors.Wrapf(err, "failed to unmarshal file for schema %q", entry.Name()))
			return
		}
		schema := api.SchemaFile2Table(schemaFile, filepath.Join(h.rootPath, api.SchemaConfigFile), entry.Name())
		schemaTable = append(schemaTable, schema)
	}
	// less is a comparison function for sorting SchemaTable items by a given field.
	less := func(left, right api.SchemaTable, sortBy string) bool {
		leftVal := query.GetFieldByJSONTag(reflect.ValueOf(left), sortBy)
		rightVal := query.GetFieldByJSONTag(reflect.ValueOf(right), sortBy)
		switch leftVal.Kind() {
		case reflect.String:
			return leftVal.String() > rightVal.String()
		case reflect.Int, reflect.Int64:
			return leftVal.Int() > rightVal.Int()
		default:
			// If the field is not a string or int, sort by Priority as a fallback.
			return left.Priority > right.Priority
		}
	}
	// filter is a function for filtering SchemaTable items by a given field and value.
	filter := func(o api.SchemaTable, f query.Filter) bool {
		val := query.GetFieldByJSONTag(reflect.ValueOf(o), f.Field)
		switch val.Kind() {
		case reflect.String:
			return strings.Contains(val.String(), f.Value)
		case reflect.Int:
			v, err := strconv.Atoi(f.Value)
			if err != nil {
				return false
			}
			return v == int(val.Int())
		default:
			return true
		}
	}

	// Use the DefaultList function to apply filtering, sorting, and pagination.
	// The results variable contains the filtered, sorted, and paginated schemaTable.
	results := query.DefaultList(schemaTable, queryParam, less, filter)
	_ = response.WriteEntity(results)
}
