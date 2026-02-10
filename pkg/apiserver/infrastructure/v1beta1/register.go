package v1beta1

import (
	"net/http"
	"strings"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	capkkinfrav1beta1 "github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v4/pkg/apiserver/runtime"
)

func AddToContainer(c *restful.Container, client ctrlclient.Client, config *rest.Config, workDir string) {
	h := NewHandler(client, config, workDir)
	webService := new(restful.WebService)
	webService.Path(strings.TrimRight(capkkinfrav1beta1.SchemeGroupVersion.String(), "/")).
		Produces(restful.MIME_JSON, string(types.MergePatchType), string(types.JSONPatchType))

	webService.Route(webService.POST("/kkcluster").
		Reads(v1beta1.KKCluster{}).
		To(h.CreateKKCluster).
		Returns(http.StatusOK, runtime.StatusOK, v1beta1.KKCluster{}).
		Doc("Create kk cluster").
		Metadata(restfulspec.KeyOpenAPITags, []string{runtime.CapkkTag}))

	webService.Route(webService.DELETE("namespaces/{namespace}/kkcluster/{kk-cluster-name}").
		To(h.DeleteKKCluster).
		Param(webService.PathParameter("namespace", "kk cluster namespace")).
		Param(webService.PathParameter("kk-cluster-name", "kk cluster name")).
		Returns(http.StatusOK, runtime.StatusOK, runtime.Result{}).
		Doc("delete kk cluster").
		Metadata(restfulspec.KeyOpenAPITags, []string{runtime.CapkkTag}))

	webService.Route(webService.POST("/kkmachine").
		Reads(v1beta1.KKMachine{}).
		To(h.CreateKKMachine).
		Returns(http.StatusOK, runtime.StatusOK, v1beta1.KKMachine{}).
		Doc("Create kk cluster").
		Metadata(restfulspec.KeyOpenAPITags, []string{runtime.CapkkTag}))

	webService.Route(webService.DELETE("namespaces/{namespace}/kkmachine/{kk-machine-name}").
		To(h.DeleteKKMachine).
		Param(webService.PathParameter("namespace", "kk machine namespace")).
		Param(webService.PathParameter("kk-machine-name", "kk machine name")).
		Param(webService.QueryParameter("kk-cluster-name", "kk cluster name")).
		Returns(http.StatusOK, runtime.StatusOK, runtime.Result{}).
		Doc("delete kk cluster").
		Metadata(restfulspec.KeyOpenAPITags, []string{runtime.CapkkTag}))

	webService.Route(webService.PATCH("namespaces/{namespace}/kkmachine/{kk-machine-name}").
		To(h.UpdateKKMachine).
		Param(webService.PathParameter("namespace", "kk machine namespace")).
		Param(webService.PathParameter("kk-machine-name", "kk machine name")).
		Param(webService.QueryParameter("kk-cluster-name", "kk cluster name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{runtime.CapkkTag}).
		Doc("patch a kk machine.").Operation("patchKKMachine").
		Consumes(string(types.JSONPatchType), string(types.MergePatchType), string(types.ApplyPatchType)).Produces(restful.MIME_JSON).
		Reads(v1beta1.KKMachine{}).
		Returns(http.StatusOK, runtime.StatusOK, v1beta1.KKMachine{}))

	webService.Route(webService.PATCH("/namespaces/{namespace}/kkcluster/{kk-cluster-name}").
		To(h.CreateConfig).
		Metadata(restfulspec.KeyOpenAPITags, []string{runtime.CapkkTag}).
		Doc("patch a kk machine template.").Operation("patchKKMachineTemplate").
		Consumes(string(types.JSONPatchType), string(types.MergePatchType), string(types.ApplyPatchType)).Produces(restful.MIME_JSON).
		Reads(v1beta1.KKCluster{}).
		Param(webService.PathParameter("namespace", "the namespace of the inventory")).
		Param(webService.PathParameter("kk-cluster-name", "the name of the inventory")).
		Returns(http.StatusOK, runtime.StatusOK, v1beta1.KKCluster{}))

	webService.Route(webService.GET("/namespaces/{namespace}/kkcluster/{kk-cluster-name}").
		To(h.GetKKCluster).
		Param(webService.PathParameter("namespace", "kk cluster namespace")).
		Param(webService.PathParameter("kk-cluster-name", "kk cluster  name")).
		Returns(http.StatusOK, runtime.StatusOK, v1beta1.KKCluster{}).
		Doc("describe kk cluster with config").
		Metadata(restfulspec.KeyOpenAPITags, []string{runtime.CapkkTag}))

	webService.Route(webService.GET("/namespaces/{namespace}/kkmachine/{kk-cluster-name}").
		To(h.GetKKMachineList).
		Param(webService.PathParameter("namespace", "kk cluster namespace")).
		Param(webService.PathParameter("kk-cluster-name", "kk cluster  name")).
		Returns(http.StatusOK, runtime.StatusOK, runtime.ListResult[v1beta1.KKMachine]{}).
		Doc("describe kk machine list").
		Metadata(restfulspec.KeyOpenAPITags, []string{runtime.CapkkTag}))

	webService.Route(webService.GET("/kkclusters").
		To(h.GetKKClusterList).
		Returns(http.StatusOK, runtime.StatusOK, runtime.ListResult[v1beta1.KKCluster]{}).
		Doc("describe kk machine list").
		Metadata(restfulspec.KeyOpenAPITags, []string{runtime.CapkkTag}))

	c.Add(webService)
}
