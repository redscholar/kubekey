package v1

import (
	"net/http"
	"strings"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	"github.com/kubesphere/kubekey/v4/pkg/apiserver/query"
	"github.com/kubesphere/kubekey/v4/pkg/apiserver/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ResourcesAPIPath defines the base path for resource-related endpoints.
// This path is used as the prefix for routes that serve static resources, schemas, and related files.
var SchemeGroupVersion = schema.GroupVersion{Group: "resources", Version: "v1"}

func AddToContainer(c *restful.Container, client ctrlclient.Client, cfg *rest.Config, workDir, schemaPath string) {
	webService := new(restful.WebService)
	webService.Path(strings.TrimRight(SchemeGroupVersion.String(), "/")).
		Produces(restful.MIME_JSON, string(types.MergePatchType), string(types.JSONPatchType))

	// only used for pre check host ,root path not needed
	h := NewHandler(client, cfg, workDir, schemaPath)

	webService.Route(webService.GET("/schema/summary").
		To(h.GetSchemaSummary).
		Returns(http.StatusOK, runtime.StatusOK, v1beta1.KKClusterList{}).
		Param(webService.QueryParameter("namespaces", "kk cluster name")).
		Param(webService.QueryParameter("kk-cluster-name", "kk cluster name")).
		Doc("describe kk machine list").
		Metadata(restfulspec.KeyOpenAPITags, []string{runtime.ResourceTag}))

	webService.Route(webService.POST("/ip").To(resourceHandler.PreCheckHost).
		Doc("pre check host ssh connect information").
		Metadata(restfulspec.KeyOpenAPITags, []string{runtime.ResourceTag}).
		Returns(http.StatusOK, runtime.StatusOK, runtime.ListResult[runtime.IPHostCheckResult]{}))

	webService.Route(webService.GET("/ip").To(resourceHandler.ListIP).
		Doc("list available ip from ip cidr").
		Metadata(restfulspec.KeyOpenAPITags, []string{runtime.ResourceTag}).
		Param(webService.QueryParameter("cidr", "the cidr for ip").Required(true)).
		Param(webService.QueryParameter("sshPort", "the ssh port for ip").Required(false)).
		Param(webService.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d")).
		Param(webService.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webService.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("false")).
		Param(webService.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=ip").Required(false).DefaultValue("ip")).
		Returns(http.StatusOK, runtime.StatusOK, runtime.ListResult[runtime.IPTable]{}))

	webService.Route(webService.GET("/schema/{subpath:*}").To(resourceHandler.SchemaInfo).
		Metadata(restfulspec.KeyOpenAPITags, []string{runtime.ResourceTag}))

	webService.Route(webService.GET("/schema").To(h.ListSchema).
		Doc("list all schema as table").
		Metadata(restfulspec.KeyOpenAPITags, []string{runtime.ResourceTag}).
		Param(webService.QueryParameter("cluster", "The namespace where the cluster resides").Required(false).DefaultValue("default")).
		Param(webService.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d")).
		Param(webService.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(webService.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("false")).
		Param(webService.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=priority")).
		Returns(http.StatusOK, runtime.StatusOK, runtime.ListResult[runtime.SchemaTable]{}))

	webService.Route(webService.GET("/schema/config").To(resourceHandler.ConfigInfo).
		Doc("get user-defined configuration information").
		Metadata(restfulspec.KeyOpenAPITags, []string{runtime.ResourceTag}))

	c.Add(webService)
}
