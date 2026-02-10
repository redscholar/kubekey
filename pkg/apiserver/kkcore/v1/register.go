package v1

import (
	"net/http"
	"strings"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v4/pkg/apiserver/runtime"
)

func AddToContainer(c *restful.Container, client ctrlclient.Client, config *rest.Config, workDir string) {
	webService := new(restful.WebService)
	webService.Path(strings.TrimRight(kkcorev1.SchemeGroupVersion.String(), "/")).
		Produces(restful.MIME_JSON, string(types.MergePatchType), string(types.JSONPatchType))

	playbookHandler := NewPlaybookHandler(workDir, config, client)

	webService.Route(webService.POST("/playbooks").To(playbookHandler.Post).
		Metadata(restfulspec.KeyOpenAPITags, []string{runtime.KubeKeyTag}).
		Doc("create a playbook.").Operation("createPlaybook").
		Param(webService.QueryParameter("promise", "promise to execute playbook").Required(false).DefaultValue("true")).
		Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON).
		Reads(kkcorev1.Playbook{}).
		Returns(http.StatusOK, runtime.StatusOK, kkcorev1.Playbook{}))

	webService.Route(webService.GET("/namespaces/{namespace}/playbooks/{playbook}/log").To(playbookHandler.Log).
		Metadata(restfulspec.KeyOpenAPITags, []string{runtime.KubeKeyTag}).
		Doc("get a playbook execute log.").Operation("getPlaybookLog").
		Produces("text/plain").
		Param(webService.PathParameter("namespace", "the namespace of the playbook")).
		Param(webService.PathParameter("playbook", "the name of the playbook")).
		Returns(http.StatusOK, runtime.StatusOK, ""))

	c.Add(webService)
}
